package app

import (
	"context"
	"log/slog"
	"time"

	"connectrpc.com/connect"
)

// rpcLoggingInterceptorは、ConnectRPCのprocedure単位で結果と処理時間を記録する。
// tokenやrequest/response bodyは、認証情報・個人情報を含み得るため記録しない。
type rpcLoggingInterceptor struct {
	logger *slog.Logger
}

var _ connect.Interceptor = (*rpcLoggingInterceptor)(nil)

func newRPCLoggingInterceptor(logger *slog.Logger) *rpcLoggingInterceptor {
	if logger == nil {
		logger = slog.Default()
	}
	return &rpcLoggingInterceptor{logger: logger}
}

func (i *rpcLoggingInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		startedAt := time.Now()
		response, err := next(ctx, request)
		i.log(ctx, request.Spec(), request.Peer(), time.Since(startedAt), err)
		return response, err
	}
}

// server側では使用しないclient streamは変更せず、そのまま次へ渡す。
func (i *rpcLoggingInterceptor) WrapStreamingClient(
	next connect.StreamingClientFunc,
) connect.StreamingClientFunc {
	return next
}

func (i *rpcLoggingInterceptor) WrapStreamingHandler(
	next connect.StreamingHandlerFunc,
) connect.StreamingHandlerFunc {
	return func(ctx context.Context, connection connect.StreamingHandlerConn) error {
		startedAt := time.Now()
		err := next(ctx, connection)
		i.log(ctx, connection.Spec(), connection.Peer(), time.Since(startedAt), err)
		return err
	}
}

func (i *rpcLoggingInterceptor) log(
	ctx context.Context,
	spec connect.Spec,
	peer connect.Peer,
	duration time.Duration,
	err error,
) {
	code := "ok"
	level := slog.LevelInfo
	if err != nil {
		connectCode := connect.CodeOf(err)
		code = connectCode.String()
		level = rpcLogLevel(connectCode)
	}

	attributes := []slog.Attr{
		slog.String("procedure", spec.Procedure),
		slog.String("stream_type", spec.StreamType.String()),
		slog.String("protocol", peer.Protocol),
		slog.String("code", code),
		slog.Duration("duration", duration),
	}
	if err != nil {
		attributes = append(attributes, slog.String("error", err.Error()))
	}

	i.logger.LogAttrs(ctx, level, "rpc completed", attributes...)
}

func rpcLogLevel(code connect.Code) slog.Level {
	switch code {
	case connect.CodeCanceled:
		return slog.LevelInfo
	case connect.CodeUnknown,
		connect.CodeDeadlineExceeded,
		connect.CodeResourceExhausted,
		connect.CodeInternal,
		connect.CodeUnavailable,
		connect.CodeDataLoss:
		return slog.LevelError
	default:
		// 認証失敗や入力不備など、client側で対処可能な失敗はwarnとして区別する。
		return slog.LevelWarn
	}
}
