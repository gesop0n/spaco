// Package factory は、authentication module内部のuse caseとadapterを組み立てる。
package factory

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"

	"github.com/gesop0n/spaco/backend/internal/modules/authentication"
	"github.com/gesop0n/spaco/backend/internal/modules/authentication/internal/adapter/rpc"
	"github.com/gesop0n/spaco/backend/internal/modules/authentication/internal/adapter/supabase"
	"github.com/gesop0n/spaco/backend/internal/modules/authentication/internal/application"
)

// Configは、authentication moduleの外部接続設定を保持する。
type Config struct {
	Issuer     string
	Audience   string
	JWKSURL    string
	ClockSkew  time.Duration
	HTTPClient *http.Client
}

// Moduleは、組み立て済みのauthentication moduleである。
type Module struct {
	interceptor connect.Interceptor
	verifier    *supabase.TokenVerifier
}

// Newは、Supabase verifier、authentication service、ConnectRPC interceptorを
// 組み立てる。ctxにはapplicationと同じ寿命のcontextを渡す。
func New(
	ctx context.Context,
	config Config,
	resolver authentication.IUserResolver,
) (*Module, error) {
	if resolver == nil {
		return nil, errors.New("create authentication module: user resolver is required")
	}

	verifier, err := supabase.NewTokenVerifier(ctx, supabase.TokenVerifierConfig{
		Issuer:     config.Issuer,
		Audience:   config.Audience,
		JWKSURL:    config.JWKSURL,
		ClockSkew:  config.ClockSkew,
		HTTPClient: config.HTTPClient,
	})
	if err != nil {
		return nil, fmt.Errorf("create authentication module: %w", err)
	}

	service, err := application.NewService(verifier, resolver)
	if err != nil {
		closeErr := verifier.Close(context.Background())
		return nil, errors.Join(fmt.Errorf("create authentication service: %w", err), closeErr)
	}
	interceptor, err := rpc.NewAuthInterceptor(service)
	if err != nil {
		closeErr := verifier.Close(context.Background())
		return nil, errors.Join(fmt.Errorf("create authentication interceptor: %w", err), closeErr)
	}

	return &Module{interceptor: interceptor, verifier: verifier}, nil
}

// Interceptorは、認証が必要なConnectRPC handlerへ設定するinterceptorを返す。
func (m *Module) Interceptor() connect.Interceptor {
	return m.interceptor
}

// Closeは、authentication moduleが使用するバックグラウンドworkerを停止する。
func (m *Module) Close(ctx context.Context) error {
	if m == nil || m.verifier == nil {
		return nil
	}
	return m.verifier.Close(ctx)
}
