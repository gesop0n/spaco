package app

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
)

func TestRPCLoggingInterceptorLogsSuccessfulCall(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	interceptor := newRPCLoggingInterceptor(slog.New(slog.NewJSONHandler(&output, nil)))
	interceptor.log(
		context.Background(),
		connect.Spec{
			Procedure:  "/spaco.account.v1.AccountService/GetCurrentAccount",
			StreamType: connect.StreamTypeUnary,
		},
		connect.Peer{Protocol: connect.ProtocolConnect},
		25*time.Millisecond,
		nil,
	)

	logLine := output.String()
	for _, expected := range []string{
		`"level":"INFO"`,
		`"msg":"rpc completed"`,
		`"procedure":"/spaco.account.v1.AccountService/GetCurrentAccount"`,
		`"stream_type":"unary"`,
		`"protocol":"connect"`,
		`"code":"ok"`,
	} {
		if !strings.Contains(logLine, expected) {
			t.Errorf("log = %s, want to contain %s", logLine, expected)
		}
	}
}

func TestRPCLoggingInterceptorLogsInternalError(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	interceptor := newRPCLoggingInterceptor(slog.New(slog.NewJSONHandler(&output, nil)))
	err := connect.NewError(connect.CodeInternal, errors.New("account operation failed"))
	interceptor.log(
		context.Background(),
		connect.Spec{
			Procedure:  "/spaco.account.v1.AccountService/UpdateProfile",
			StreamType: connect.StreamTypeUnary,
		},
		connect.Peer{Protocol: connect.ProtocolConnect},
		time.Millisecond,
		err,
	)

	logLine := output.String()
	for _, expected := range []string{
		`"level":"ERROR"`,
		`"code":"internal"`,
		`"error":"internal: account operation failed"`,
	} {
		if !strings.Contains(logLine, expected) {
			t.Errorf("log = %s, want to contain %s", logLine, expected)
		}
	}
}

func TestRPCLogLevelTreatsAuthenticationFailureAsWarning(t *testing.T) {
	t.Parallel()

	if got := rpcLogLevel(connect.CodeUnauthenticated); got != slog.LevelWarn {
		t.Fatalf("rpcLogLevel(unauthenticated) = %v, want %v", got, slog.LevelWarn)
	}
}
