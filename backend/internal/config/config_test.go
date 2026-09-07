package config

import (
	"testing"
	"time"
)

func TestAuthenticationConfigDerivesJWTURLs(t *testing.T) {
	t.Parallel()

	config := AuthenticationConfig{SupabaseURL: "https://project.supabase.co/"}
	if got := config.IssuerURL(); got != "https://project.supabase.co/auth/v1" {
		t.Fatalf("IssuerURL() = %q", got)
	}
	if got := config.JSONWebKeySetURL(); got != "https://project.supabase.co/auth/v1/.well-known/jwks.json" {
		t.Fatalf("JSONWebKeySetURL() = %q", got)
	}
}

func TestValidateRejectsCORSOriginWithPath(t *testing.T) {
	t.Parallel()

	err := validate(Config{
		Server: ServerConfig{
			Address:         ":8080",
			AllowedOrigins:  []string{"https://example.com/path"},
			ShutdownTimeout: time.Second,
		},
		Authentication: AuthenticationConfig{
			SupabaseURL: "https://project.supabase.co",
			ClockSkew:   time.Second,
		},
	})
	if err == nil {
		t.Fatal("validate() error = nil, want error")
	}
}
