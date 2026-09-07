// Package rpc は、authentication use caseをConnectRPCへ接続するadapterを提供する。
package rpc

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"connectrpc.com/connect"

	"github.com/gesop0n/spaco/backend/internal/modules/authentication"
	"github.com/gesop0n/spaco/backend/internal/modules/authentication/internal/application"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

var (
	ErrMissingAuthorization = errors.New("authorization header is required")
	ErrMalformedBearerToken = errors.New("authorization header must contain one bearer token")
)

// IAuthenticatorは、access tokenからアプリ内UserIDを認証するinterfaceである。
// application.Serviceがこのinterfaceを実装する。
type IAuthenticator interface {
	Authenticate(context.Context, string) (identifier.UserID, error)
}

// AuthenticatorFuncは、関数をIAuthenticatorとして利用できるようにするadapterである。
type AuthenticatorFunc func(context.Context, string) (identifier.UserID, error)

func (f AuthenticatorFunc) Authenticate(
	ctx context.Context,
	token string,
) (identifier.UserID, error) {
	return f(ctx, token)
}

type authInterceptor struct {
	authenticator IAuthenticator
}

var _ connect.Interceptor = (*authInterceptor)(nil)

// NewAuthInterceptorは、unary RPCとstreaming RPCのhandlerを認証で保護する。
// アプリ内ユーザーの認証が必要なprocedureにだけ適用する。
func NewAuthInterceptor(authenticator IAuthenticator) (connect.Interceptor, error) {
	if authenticator == nil {
		return nil, errors.New("create auth interceptor: authenticator is required")
	}

	return &authInterceptor{authenticator: authenticator}, nil
}

func (i *authInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		authenticatedContext, err := i.authenticate(ctx, req.Header())
		if err != nil {
			return nil, err
		}

		return next(authenticatedContext, req)
	}
}

// このinterceptorはserver handler用なので、外向きのstreaming clientは変更しない。
func (*authInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (i *authInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		authenticatedContext, err := i.authenticate(ctx, conn.RequestHeader())
		if err != nil {
			return err
		}

		return next(authenticatedContext, conn)
	}
}

func (i *authInterceptor) authenticate(
	ctx context.Context,
	header http.Header,
) (context.Context, error) {
	// transport層ではBearer tokenの取り出しとConnectのエラー変換だけを行う。
	token, err := bearerToken(header.Get("Authorization"))
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	userID, err := i.authenticator.Authenticate(ctx, token)
	if err != nil {
		if contextError := connectContextError(err); contextError != nil {
			return nil, contextError
		}
		if errors.Is(err, application.ErrInvalidToken) {
			return nil, unauthenticatedError(err)
		}
		return nil, connect.NewError(
			connect.CodeInternal,
			&opaqueError{message: "failed to authenticate user", cause: err},
		)
	}
	if userID.IsZero() {
		return nil, connect.NewError(
			connect.CodeInternal,
			&opaqueError{
				message: "failed to authenticate user",
				cause:   errors.New("authenticator returned a zero user ID"),
			},
		)
	}

	// 業務handlerには外部identityではなく、アプリ内UserIDだけを渡す。
	return authentication.WithUserID(ctx, userID), nil
}

func bearerToken(authorization string) (string, error) {
	fields := strings.Fields(authorization)
	if len(fields) == 0 {
		return "", ErrMissingAuthorization
	}
	if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") || fields[1] == "" {
		return "", ErrMalformedBearerToken
	}

	return fields[1], nil
}

func unauthenticatedError(cause error) *connect.Error {
	// 詳細な検証エラーはclientへ公開せず、内部ではUnwrap経由で追跡可能にする。
	err := connect.NewError(
		connect.CodeUnauthenticated,
		&opaqueError{message: "authentication required", cause: cause},
	)
	err.Meta().Set("WWW-Authenticate", "Bearer")
	return err
}

func connectContextError(err error) *connect.Error {
	switch {
	case errors.Is(err, context.Canceled):
		return connect.NewError(connect.CodeCanceled, context.Canceled)
	case errors.Is(err, context.DeadlineExceeded):
		return connect.NewError(connect.CodeDeadlineExceeded, context.DeadlineExceeded)
	default:
		return nil
	}
}

// opaqueErrorは、内部エラーをerrors.Is/errors.Asやログから追跡可能にしつつ、
// 実装詳細がRPC clientへ送信されないように公開メッセージを固定する。
type opaqueError struct {
	message string
	cause   error
}

func (e *opaqueError) Error() string { return e.message }
func (e *opaqueError) Unwrap() error { return e.cause }
