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

func TestResolveServerAddress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		address string
		port    string
		want    string
		wantErr bool
	}{
		{name: "Cloud RunのPORTを使う", port: "9090", want: ":9090"},
		{name: "PORTがなければ8080を使う", want: ":8080"},
		{name: "SERVER_ADDRESSを優先する", address: "127.0.0.1:3000", port: "9090", want: "127.0.0.1:3000"},
		{name: "数値でないPORTを拒否する", port: "http", wantErr: true},
		{name: "範囲外のPORTを拒否する", port: "65536", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := resolveServerAddress(test.address, test.port)
			if test.wantErr {
				if err == nil {
					t.Fatal("resolveServerAddress() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveServerAddress() error = %v", err)
			}
			if got != test.want {
				t.Fatalf("resolveServerAddress() = %q, want %q", got, test.want)
			}
		})
	}
}
