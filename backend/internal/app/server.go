// Package app は、各moduleとHTTP serverを組み立ててアプリケーションを起動する。
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gesop0n/spaco/backend/internal/config"
	accountfactory "github.com/gesop0n/spaco/backend/internal/modules/account/factory"
	authenticationfactory "github.com/gesop0n/spaco/backend/internal/modules/authentication/factory"
)

type Server struct {
	httpServer      *http.Server
	authentication  *authenticationfactory.Module
	database        *pgxpool.Pool
	shutdownTimeout time.Duration
}

// NewServerは、DB、業務module、認証interceptor、HTTP routingを接続する。
func NewServer(ctx context.Context, cfg config.Config) (*Server, error) {
	databaseConfig, err := databasePoolConfig(cfg.Database.URL)
	if err != nil {
		return nil, err
	}
	database, err := pgxpool.NewWithConfig(ctx, databaseConfig)
	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}
	if err := database.Ping(ctx); err != nil {
		database.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	accountModule, err := accountfactory.New(database)
	if err != nil {
		database.Close()
		return nil, err
	}
	authenticationModule, err := authenticationfactory.New(
		ctx,
		authenticationfactory.Config{
			Issuer:    cfg.Authentication.IssuerURL(),
			Audience:  cfg.Authentication.Audience,
			JWKSURL:   cfg.Authentication.JSONWebKeySetURL(),
			ClockSkew: cfg.Authentication.ClockSkew,
		},
		accountModule,
	)
	if err != nil {
		database.Close()
		return nil, err
	}

	accountPath, accountHandler := accountModule.ConnectHandler(
		connect.WithInterceptors(
			// loggingを外側に置き、認証Interceptorが返す失敗も記録する。
			newRPCLoggingInterceptor(slog.Default()),
			authenticationModule.Interceptor(),
		),
	)
	router := newRouter(accountPath, accountHandler, cfg.Server.AllowedOrigins)

	return &Server{
		httpServer: &http.Server{
			Addr:              cfg.Server.Address,
			Handler:           router,
			ReadHeaderTimeout: 5 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
		authentication:  authenticationModule,
		database:        database,
		shutdownTimeout: cfg.Server.ShutdownTimeout,
	}, nil
}

func databasePoolConfig(databaseURL string) (*pgxpool.Config, error) {
	databaseConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}

	// Supabaseのtransaction poolerはconnectionを共有するため、connection固有の
	// named prepared statementを利用しないexec modeでqueryを送信する。
	databaseConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec

	return databaseConfig, nil
}

// RunはHTTP serverを開始し、終了signal時に処理中requestを待って停止する。
func (s *Server) Run(ctx context.Context) (runErr error) {
	defer func() {
		closeContext, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
		defer cancel()
		runErr = errors.Join(runErr, s.authentication.Close(closeContext))
		s.database.Close()
	}()

	listener, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", s.httpServer.Addr, err)
	}

	serveError := make(chan error, 1)
	go func() {
		serveError <- s.httpServer.Serve(listener)
	}()

	slog.Info("server started", "address", listener.Addr().String())
	select {
	case err := <-serveError:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			runErr = fmt.Errorf("serve HTTP: %w", err)
		}
	case <-ctx.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
		defer cancel()
		if err := s.httpServer.Shutdown(shutdownContext); err != nil {
			runErr = fmt.Errorf("shutdown HTTP server: %w", err)
			// graceful shutdownが時間切れになった場合は、待機を終えるため接続を閉じる。
			_ = s.httpServer.Close()
		}
		if err := <-serveError; err != nil && !errors.Is(err, http.ErrServerClosed) {
			runErr = errors.Join(runErr, fmt.Errorf("serve HTTP: %w", err))
		}
	}

	return runErr
}
