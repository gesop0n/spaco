package rpc

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"connectrpc.com/connect"

	"github.com/gesop0n/spaco/backend/internal/modules/authentication"
	"github.com/gesop0n/spaco/backend/internal/modules/authentication/internal/usecase"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

func TestBearerToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		authorization string
		want          string
		wantErr       error
	}{
		{name: "standard", authorization: "Bearer header.payload.signature", want: "header.payload.signature"},
		{name: "case insensitive scheme", authorization: "bearer token", want: "token"},
		{name: "optional whitespace", authorization: "  Bearer\t token  ", want: "token"},
		{name: "missing", wantErr: ErrMissingAuthorization},
		{name: "wrong scheme", authorization: "Basic token", wantErr: ErrMalformedBearerToken},
		{name: "token missing", authorization: "Bearer", wantErr: ErrMalformedBearerToken},
		{name: "too many values", authorization: "Bearer one two", wantErr: ErrMalformedBearerToken},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := bearerToken(test.authorization)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("bearerToken() error = %v, want %v", err, test.wantErr)
			}
			if got != test.want {
				t.Fatalf("bearerToken() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestAuthInterceptorUnary(t *testing.T) {
	t.Parallel()

	wantUserID := identifier.NewUserID()
	var authenticatedToken string
	interceptor, err := NewAuthInterceptor(AuthenticateUseCaseFunc(func(
		_ context.Context,
		token string,
	) (identifier.UserID, error) {
		authenticatedToken = token
		return wantUserID, nil
	}))
	if err != nil {
		t.Fatalf("NewAuthInterceptor() error = %v", err)
	}

	called := false
	wrapped := interceptor.WrapUnary(func(ctx context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		called = true
		gotUserID, ok := authentication.UserIDFromContext(ctx)
		if !ok || gotUserID != wantUserID {
			t.Fatalf("UserIDFromContext() = %v, %v; want %v, true", gotUserID, ok, wantUserID)
		}
		return connect.NewResponse(&struct{}{}), nil
	})
	request := connect.NewRequest(&struct{}{})
	request.Header().Set("Authorization", "Bearer access-token")

	if _, err := wrapped(context.Background(), request); err != nil {
		t.Fatalf("wrapped() error = %v", err)
	}
	if !called {
		t.Fatal("wrapped handler was not called")
	}
	if authenticatedToken != "access-token" {
		t.Fatalf("authenticated token = %q, want access-token", authenticatedToken)
	}
}

func TestAuthInterceptorErrors(t *testing.T) {
	t.Parallel()

	internalErr := errors.New("database unavailable")
	tests := []struct {
		name        string
		header      string
		result      identifier.UserID
		authErr     error
		wantCode    connect.Code
		wantMessage string
		wantCause   error
	}{
		{
			name:        "missing authorization",
			result:      identifier.NewUserID(),
			wantCode:    connect.CodeUnauthenticated,
			wantMessage: "authentication required",
			wantCause:   ErrMissingAuthorization,
		},
		{
			name:        "invalid token",
			header:      "Bearer invalid",
			authErr:     usecase.ErrInvalidToken,
			wantCode:    connect.CodeUnauthenticated,
			wantMessage: "authentication required",
			wantCause:   usecase.ErrInvalidToken,
		},
		{
			name:        "use case failure",
			header:      "Bearer valid",
			authErr:     internalErr,
			wantCode:    connect.CodeInternal,
			wantMessage: "failed to authenticate user",
			wantCause:   internalErr,
		},
		{
			name:        "zero user ID",
			header:      "Bearer valid",
			wantCode:    connect.CodeInternal,
			wantMessage: "failed to authenticate user",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			interceptor, err := NewAuthInterceptor(authenticatorReturning(test.result, test.authErr))
			if err != nil {
				t.Fatalf("NewAuthInterceptor() error = %v", err)
			}
			wrapped := interceptor.WrapUnary(func(context.Context, connect.AnyRequest) (connect.AnyResponse, error) {
				t.Fatal("failed authentication reached handler")
				return nil, nil
			})
			request := connect.NewRequest(&struct{}{})
			if test.header != "" {
				request.Header().Set("Authorization", test.header)
			}

			_, gotErr := wrapped(context.Background(), request)
			if connect.CodeOf(gotErr) != test.wantCode {
				t.Fatalf("wrapped() code = %v, want %v (error: %v)", connect.CodeOf(gotErr), test.wantCode, gotErr)
			}
			var connectErr *connect.Error
			if !errors.As(gotErr, &connectErr) || connectErr.Message() != test.wantMessage {
				t.Fatalf("wrapped() error = %v, want message %q", gotErr, test.wantMessage)
			}
			if test.wantCause != nil && !errors.Is(gotErr, test.wantCause) {
				t.Fatalf("wrapped() error = %v, want wrapped cause %v", gotErr, test.wantCause)
			}
			if test.wantCode == connect.CodeUnauthenticated && connectErr.Meta().Get("WWW-Authenticate") != "Bearer" {
				t.Fatalf("wrapped() WWW-Authenticate = %q, want Bearer", connectErr.Meta().Get("WWW-Authenticate"))
			}
		})
	}
}

func TestAuthInterceptorPreservesContextErrors(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name     string
		cause    error
		wantCode connect.Code
	}{
		{name: "canceled", cause: context.Canceled, wantCode: connect.CodeCanceled},
		{name: "deadline", cause: context.DeadlineExceeded, wantCode: connect.CodeDeadlineExceeded},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			interceptor, err := NewAuthInterceptor(authenticatorReturning(identifier.UserID{}, test.cause))
			if err != nil {
				t.Fatalf("NewAuthInterceptor() error = %v", err)
			}
			request := connect.NewRequest(&struct{}{})
			request.Header().Set("Authorization", "Bearer access-token")
			_, gotErr := interceptor.WrapUnary(func(context.Context, connect.AnyRequest) (connect.AnyResponse, error) {
				t.Fatal("failed authentication reached handler")
				return nil, nil
			})(context.Background(), request)
			if connect.CodeOf(gotErr) != test.wantCode {
				t.Fatalf("wrapped() code = %v, want %v", connect.CodeOf(gotErr), test.wantCode)
			}
		})
	}
}

func TestAuthInterceptorStreamingHandler(t *testing.T) {
	t.Parallel()

	wantUserID := identifier.NewUserID()
	interceptor, err := NewAuthInterceptor(authenticatorReturning(wantUserID, nil))
	if err != nil {
		t.Fatalf("NewAuthInterceptor() error = %v", err)
	}
	called := false
	wrapped := interceptor.WrapStreamingHandler(func(ctx context.Context, _ connect.StreamingHandlerConn) error {
		called = true
		gotUserID, ok := authentication.UserIDFromContext(ctx)
		if !ok || gotUserID != wantUserID {
			t.Fatalf("UserIDFromContext() = %v, %v; want %v, true", gotUserID, ok, wantUserID)
		}
		return nil
	})
	conn := &testStreamingHandlerConn{requestHeader: make(http.Header)}
	conn.requestHeader.Set("Authorization", "Bearer access-token")

	if err := wrapped(context.Background(), conn); err != nil {
		t.Fatalf("wrapped() error = %v", err)
	}
	if !called {
		t.Fatal("wrapped streaming handler was not called")
	}
}

func TestNewAuthInterceptorRequiresAuthenticator(t *testing.T) {
	t.Parallel()
	if _, err := NewAuthInterceptor(nil); err == nil {
		t.Fatal("NewAuthInterceptor(nil) error = nil")
	}
}

func authenticatorReturning(userID identifier.UserID, err error) IAuthenticateUseCase {
	return AuthenticateUseCaseFunc(func(context.Context, string) (identifier.UserID, error) {
		return userID, err
	})
}

type testStreamingHandlerConn struct {
	requestHeader  http.Header
	responseHeader http.Header
	trailer        http.Header
}

func (*testStreamingHandlerConn) Spec() connect.Spec { return connect.Spec{} }
func (*testStreamingHandlerConn) Peer() connect.Peer { return connect.Peer{} }
func (*testStreamingHandlerConn) Receive(any) error  { return nil }
func (c *testStreamingHandlerConn) RequestHeader() http.Header {
	return c.requestHeader
}
func (*testStreamingHandlerConn) Send(any) error { return nil }
func (c *testStreamingHandlerConn) ResponseHeader() http.Header {
	if c.responseHeader == nil {
		c.responseHeader = make(http.Header)
	}
	return c.responseHeader
}
func (c *testStreamingHandlerConn) ResponseTrailer() http.Header {
	if c.trailer == nil {
		c.trailer = make(http.Header)
	}
	return c.trailer
}
