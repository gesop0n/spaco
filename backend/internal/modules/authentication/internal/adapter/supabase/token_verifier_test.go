package supabase

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"

	"github.com/gesop0n/spaco/backend/internal/modules/authentication/internal/usecase"
)

func TestTokenVerifier(t *testing.T) {
	privateKey, jwksJSON := testSigningKeyAndJWKS(t, "test-key")
	var requests atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requests.Add(1)
		if request.URL.Path != "/auth/v1/.well-known/jwks.json" {
			http.NotFound(writer, request)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write(jwksJSON)
	}))
	t.Cleanup(server.Close)

	issuer := server.URL + "/auth/v1"
	verifier, err := NewTokenVerifier(context.Background(), TokenVerifierConfig{
		Issuer:     issuer + "/",
		HTTPClient: server.Client(),
	})
	if err != nil {
		t.Fatalf("NewTokenVerifier() error = %v", err)
	}
	t.Cleanup(func() {
		if err := verifier.Close(context.Background()); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	validToken := signTestToken(t, privateKey, tokenClaims{
		issuer:     issuer,
		subject:    "external-user-id",
		audience:   DefaultAudience,
		expiration: time.Now().Add(time.Hour),
	})
	identity, err := verifier.Verify(context.Background(), validToken)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if identity.Issuer() != issuer {
		t.Fatalf("Identity.Issuer() = %q, want %q", identity.Issuer(), issuer)
	}
	if identity.Subject() != "external-user-id" {
		t.Fatalf("Identity.Subject() = %q, want %q", identity.Subject(), "external-user-id")
	}
	if requests.Load() == 0 {
		t.Fatal("JWKS endpoint was not requested during verifier construction")
	}

	invalidTokens := []struct {
		name   string
		claims tokenClaims
	}{
		{
			name: "wrong issuer",
			claims: tokenClaims{
				issuer: "https://another-project.supabase.co/auth/v1", subject: "external-user-id",
				audience: DefaultAudience, expiration: time.Now().Add(time.Hour),
			},
		},
		{
			name: "wrong audience",
			claims: tokenClaims{
				issuer: issuer, subject: "external-user-id", audience: "anon",
				expiration: time.Now().Add(time.Hour),
			},
		},
		{
			name: "expired",
			claims: tokenClaims{
				issuer: issuer, subject: "external-user-id", audience: DefaultAudience,
				expiration: time.Now().Add(-time.Minute),
			},
		},
		{
			name: "missing subject",
			claims: tokenClaims{
				issuer: issuer, audience: DefaultAudience,
				expiration: time.Now().Add(time.Hour),
			},
		},
		{
			name: "missing expiration",
			claims: tokenClaims{
				issuer: issuer, subject: "external-user-id", audience: DefaultAudience,
			},
		},
	}

	for _, test := range invalidTokens {
		t.Run(test.name, func(t *testing.T) {
			rawToken := signTestToken(t, privateKey, test.claims)
			_, err := verifier.Verify(context.Background(), rawToken)
			if !errors.Is(err, usecase.ErrInvalidToken) {
				t.Fatalf("Verify() error = %v, want ErrInvalidToken", err)
			}
		})
	}

	for _, rawToken := range []string{"", "not-a-jwt"} {
		_, err := verifier.Verify(context.Background(), rawToken)
		if !errors.Is(err, usecase.ErrInvalidToken) {
			t.Fatalf("Verify(%q) error = %v, want ErrInvalidToken", rawToken, err)
		}
	}

	canceledContext, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := verifier.Verify(canceledContext, validToken); !errors.Is(err, context.Canceled) {
		t.Fatalf("Verify(canceled context) error = %v, want context.Canceled", err)
	}
}

func TestTokenVerifierRejectsWrongSignature(t *testing.T) {
	trustedKey, jwksJSON := testSigningKeyAndJWKS(t, "trusted-key")
	untrustedKey, _ := testSigningKeyAndJWKS(t, "trusted-key")
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write(jwksJSON)
	}))
	t.Cleanup(server.Close)

	issuer := server.URL + "/auth/v1"
	verifier, err := NewTokenVerifier(context.Background(), TokenVerifierConfig{
		Issuer:     issuer,
		JWKSURL:    server.URL + "/jwks",
		HTTPClient: server.Client(),
	})
	if err != nil {
		t.Fatalf("NewTokenVerifier() error = %v", err)
	}
	t.Cleanup(func() { _ = verifier.Close(context.Background()) })

	claims := tokenClaims{
		issuer: issuer, subject: "external-user-id", audience: DefaultAudience,
		expiration: time.Now().Add(time.Hour),
	}
	if _, err := verifier.Verify(context.Background(), signTestToken(t, untrustedKey, claims)); !errors.Is(err, usecase.ErrInvalidToken) {
		t.Fatalf("Verify(untrusted token) error = %v, want ErrInvalidToken", err)
	}
	if _, err := verifier.Verify(context.Background(), signTestToken(t, trustedKey, claims)); err != nil {
		t.Fatalf("Verify(trusted token) error = %v", err)
	}
}

func TestNewTokenVerifierValidatesConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config TokenVerifierConfig
	}{
		{name: "issuer missing"},
		{name: "issuer relative", config: TokenVerifierConfig{Issuer: "/auth/v1"}},
		{name: "issuer unsupported scheme", config: TokenVerifierConfig{Issuer: "ftp://example.com/auth/v1"}},
		{name: "issuer with credentials", config: TokenVerifierConfig{Issuer: "https://user@example.com/auth/v1"}},
		{name: "issuer with query", config: TokenVerifierConfig{Issuer: "https://example.com/auth/v1?key=value"}},
		{
			name: "JWKS URL relative",
			config: TokenVerifierConfig{
				Issuer: "https://example.supabase.co/auth/v1", JWKSURL: "/jwks",
			},
		},
		{
			name: "negative clock skew",
			config: TokenVerifierConfig{
				Issuer: "https://example.supabase.co/auth/v1", ClockSkew: -time.Second,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewTokenVerifier(context.Background(), test.config)
			if err == nil || !strings.HasPrefix(err.Error(), "configure Supabase token verifier:") {
				t.Fatalf("NewTokenVerifier() error = %v, want config error", err)
			}
		})
	}
}

type tokenClaims struct {
	issuer     string
	subject    string
	audience   string
	expiration time.Time
}

func testSigningKeyAndJWKS(t *testing.T, keyID string) (jwk.Key, []byte) {
	t.Helper()

	rawPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	privateKey, err := jwk.Import(rawPrivateKey)
	if err != nil {
		t.Fatalf("jwk.Import() error = %v", err)
	}
	if err := privateKey.Set(jwk.KeyIDKey, keyID); err != nil {
		t.Fatalf("privateKey.Set(kid) error = %v", err)
	}
	if err := privateKey.Set(jwk.AlgorithmKey, jwa.ES256()); err != nil {
		t.Fatalf("privateKey.Set(alg) error = %v", err)
	}

	publicKey, err := jwk.PublicKeyOf(privateKey)
	if err != nil {
		t.Fatalf("jwk.PublicKeyOf() error = %v", err)
	}
	if err := publicKey.Set(jwk.KeyIDKey, keyID); err != nil {
		t.Fatalf("publicKey.Set(kid) error = %v", err)
	}
	if err := publicKey.Set(jwk.AlgorithmKey, jwa.ES256()); err != nil {
		t.Fatalf("publicKey.Set(alg) error = %v", err)
	}
	keySet := jwk.NewSet()
	if err := keySet.AddKey(publicKey); err != nil {
		t.Fatalf("keySet.AddKey() error = %v", err)
	}
	jwksJSON, err := json.Marshal(keySet)
	if err != nil {
		t.Fatalf("json.Marshal(JWKS) error = %v", err)
	}

	return privateKey, jwksJSON
}

func signTestToken(t *testing.T, privateKey jwk.Key, claims tokenClaims) string {
	t.Helper()

	builder := jwt.NewBuilder().Issuer(claims.issuer).Audience([]string{claims.audience})
	if claims.subject != "" {
		builder.Subject(claims.subject)
	}
	if !claims.expiration.IsZero() {
		builder.Expiration(claims.expiration)
	}
	token, err := builder.Build()
	if err != nil {
		t.Fatalf("token builder error = %v", err)
	}
	signed, err := jwt.Sign(token, jwt.WithKey(jwa.ES256(), privateKey))
	if err != nil {
		t.Fatalf("jwt.Sign() error = %v", err)
	}
	return string(signed)
}
