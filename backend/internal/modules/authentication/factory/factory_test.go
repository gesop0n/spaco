package factory

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gesop0n/spaco/backend/internal/modules/authentication"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

func TestNew(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"keys":[]}`))
	}))
	t.Cleanup(server.Close)

	module, err := New(
		context.Background(),
		Config{
			Issuer:     server.URL + "/auth/v1",
			JWKSURL:    server.URL + "/jwks",
			HTTPClient: server.Client(),
		},
		authentication.UserResolverFunc(func(
			context.Context,
			string,
			string,
			string,
		) (identifier.UserID, error) {
			return identifier.NewUserID(), nil
		}),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if module.Interceptor() == nil {
		t.Fatal("Interceptor() = nil")
	}
	if err := module.Close(context.Background()); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func TestNewRequiresResolver(t *testing.T) {
	t.Parallel()
	if _, err := New(context.Background(), Config{}, nil); err == nil {
		t.Fatal("New(config, nil) error = nil")
	}
}
