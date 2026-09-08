// Package supabase は、Supabase Authをauthentication moduleへ接続するadapterを提供する。
package supabase

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/lestrrat-go/httprc/v3"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"

	"github.com/gesop0n/spaco/backend/internal/modules/authentication/internal/usecase"
)

const (
	DefaultAudience        = "authenticated"
	defaultClockSkew       = 30 * time.Second
	minJWKSRefreshInterval = time.Minute
	maxJWKSRefreshInterval = 10 * time.Minute
)

// TokenVerifierConfig は、Supabase AuthのJWT検証設定を保持する。
// JWKSURLを省略した場合はIssuerから自動生成する。
type TokenVerifierConfig struct {
	Issuer     string
	Audience   string
	JWKSURL    string
	ClockSkew  time.Duration
	HTTPClient *http.Client
}

// TokenVerifier は、projectのローテーション可能なJWKSを利用して、
// 非対称鍵で署名されたSupabase JWTを検証する。
// legacyな共通鍵方式のHS256は意図的に対象外としている。
type TokenVerifier struct {
	issuer    string
	audience  string
	clockSkew time.Duration
	keySet    jwk.Set
	cache     *jwk.Cache
}

var _ usecase.ITokenVerifier = (*TokenVerifier)(nil)

// NewTokenVerifier は、最初のJWKSを取得し、バックグラウンド更新を開始する。
// ctxにはauthentication moduleと同じ寿命のcontextを渡し、shutdown時にCloseを呼び出す。
func NewTokenVerifier(
	ctx context.Context,
	config TokenVerifierConfig,
) (*TokenVerifier, error) {
	issuer, jwksURL, audience, clockSkew, err := normalizeSupabaseVerifierConfig(config)
	if err != nil {
		return nil, err
	}

	// JWKS cacheはauthentication module全体で1つだけ生成し、requestごとの取得を避ける。
	cache, err := jwk.NewCache(ctx, httprc.NewClient())
	if err != nil {
		return nil, fmt.Errorf("create Supabase JWKS cache: %w", err)
	}

	registerOptions := []jwk.RegisterOption{
		jwk.WithMinInterval(minJWKSRefreshInterval),
		jwk.WithMaxInterval(maxJWKSRefreshInterval),
	}
	if config.HTTPClient != nil {
		registerOptions = append(registerOptions, jwk.WithHTTPClient(config.HTTPClient))
	}

	// Registerは初回JWKSの取得完了まで待つため、起動時に設定・通信不良を検出できる。
	if err := cache.Register(ctx, jwksURL, registerOptions...); err != nil {
		shutdownErr := cache.Shutdown(context.Background())
		return nil, errors.Join(
			fmt.Errorf("register Supabase JWKS endpoint: %w", err),
			shutdownErr,
		)
	}

	keySet, err := cache.CachedSet(jwksURL)
	if err != nil {
		shutdownErr := cache.Shutdown(context.Background())
		return nil, errors.Join(
			fmt.Errorf("open Supabase JWKS cache: %w", err),
			shutdownErr,
		)
	}

	return &TokenVerifier{
		issuer:    issuer,
		audience:  audience,
		clockSkew: clockSkew,
		keySet:    keySet,
		cache:     cache,
	}, nil
}

func (v *TokenVerifier) Verify(
	ctx context.Context,
	rawToken string,
) (usecase.Identity, error) {
	if err := ctx.Err(); err != nil {
		return usecase.Identity{}, err
	}
	if strings.TrimSpace(rawToken) == "" {
		return usecase.Identity{}, usecase.ErrInvalidToken
	}

	// 署名だけでなく、発行者・利用対象・期限・必須claimも同時に検証する。
	token, err := jwt.Parse(
		[]byte(rawToken),
		jwt.WithKeySet(v.keySet),
		jwt.WithIssuer(v.issuer),
		jwt.WithAudience(v.audience),
		jwt.WithAcceptableSkew(v.clockSkew),
		jwt.WithContext(ctx),
		jwt.WithRequiredClaim(jwt.IssuerKey),
		jwt.WithRequiredClaim(jwt.SubjectKey),
		jwt.WithRequiredClaim(jwt.AudienceKey),
		jwt.WithRequiredClaim(jwt.ExpirationKey),
	)
	if err != nil {
		return usecase.Identity{}, fmt.Errorf("%w: %v", usecase.ErrInvalidToken, err)
	}

	issuer, issuerOK := token.Issuer()
	subject, subjectOK := token.Subject()
	if !issuerOK || !subjectOK {
		return usecase.Identity{}, usecase.ErrInvalidToken
	}

	identity, err := usecase.NewIdentity(issuer, subject)
	if err != nil {
		return usecase.Identity{}, fmt.Errorf("%w: %v", usecase.ErrInvalidToken, err)
	}

	return identity, nil
}

// Close は、JWKSのバックグラウンド更新を停止し、workerの終了を待つ。
func (v *TokenVerifier) Close(ctx context.Context) error {
	if v == nil || v.cache == nil {
		return nil
	}
	if err := v.cache.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown Supabase JWKS cache: %w", err)
	}
	return nil
}

func normalizeSupabaseVerifierConfig(
	config TokenVerifierConfig,
) (issuer, jwksURL, audience string, clockSkew time.Duration, err error) {
	issuer = strings.TrimRight(strings.TrimSpace(config.Issuer), "/")
	if issuer == "" {
		return "", "", "", 0, errors.New("configure Supabase token verifier: issuer is required")
	}
	if err := validateHTTPURL("issuer", issuer); err != nil {
		return "", "", "", 0, err
	}

	jwksURL = strings.TrimSpace(config.JWKSURL)
	if jwksURL == "" {
		jwksURL = issuer + "/.well-known/jwks.json"
	}
	if err := validateHTTPURL("JWKS URL", jwksURL); err != nil {
		return "", "", "", 0, err
	}

	audience = strings.TrimSpace(config.Audience)
	if audience == "" {
		audience = DefaultAudience
	}

	clockSkew = config.ClockSkew
	if clockSkew < 0 {
		return "", "", "", 0, errors.New("configure Supabase token verifier: clock skew cannot be negative")
	}
	if clockSkew == 0 {
		clockSkew = defaultClockSkew
	}

	return issuer, jwksURL, audience, clockSkew, nil
}

func validateHTTPURL(name, value string) error {
	parsedURL, err := url.ParseRequestURI(value)
	if err != nil ||
		parsedURL.Host == "" ||
		parsedURL.User != nil ||
		parsedURL.RawQuery != "" ||
		parsedURL.Fragment != "" ||
		(parsedURL.Scheme != "https" && parsedURL.Scheme != "http") {
		return fmt.Errorf("configure Supabase token verifier: %s must be an absolute HTTP(S) URL", name)
	}
	return nil
}
