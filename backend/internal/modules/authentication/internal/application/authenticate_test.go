package application

import (
	"context"
	"errors"
	"testing"

	"github.com/gesop0n/spaco/backend/internal/modules/authentication"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

func TestNewIdentity(t *testing.T) {
	t.Parallel()

	identity, err := NewIdentity("  https://example.supabase.co/auth/v1/  ", " user-id ")
	if err != nil {
		t.Fatalf("NewIdentity() error = %v", err)
	}
	if identity.Issuer() != "https://example.supabase.co/auth/v1/" {
		t.Fatalf("Issuer() = %q", identity.Issuer())
	}
	if identity.Subject() != "user-id" {
		t.Fatalf("Subject() = %q", identity.Subject())
	}

	for _, test := range []struct {
		name    string
		issuer  string
		subject string
	}{
		{name: "missing issuer", subject: "user-id"},
		{name: "missing subject", issuer: "https://example.supabase.co/auth/v1"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewIdentity(test.issuer, test.subject)
			if !errors.Is(err, ErrInvalidIdentity) {
				t.Fatalf("NewIdentity() error = %v, want ErrInvalidIdentity", err)
			}
		})
	}
}

func TestServiceAuthenticate(t *testing.T) {
	t.Parallel()

	identity := mustIdentity(t, "https://example.supabase.co/auth/v1", "external-user-id")
	wantUserID := identifier.NewUserID()
	var verifiedToken string
	var resolvedIssuer string
	var resolvedSubject string

	service, err := NewService(
		TokenVerifierFunc(func(_ context.Context, token string) (Identity, error) {
			verifiedToken = token
			return identity, nil
		}),
		authentication.UserResolverFunc(func(
			_ context.Context,
			issuer string,
			subject string,
		) (identifier.UserID, error) {
			resolvedIssuer = issuer
			resolvedSubject = subject
			return wantUserID, nil
		}),
	)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	gotUserID, err := service.Authenticate(context.Background(), "access-token")
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if gotUserID != wantUserID {
		t.Fatalf("Authenticate() = %v, want %v", gotUserID, wantUserID)
	}
	if verifiedToken != "access-token" {
		t.Fatalf("verified token = %q, want access-token", verifiedToken)
	}
	if resolvedIssuer != identity.Issuer() || resolvedSubject != identity.Subject() {
		t.Fatalf(
			"resolved identity = (%q, %q), want (%q, %q)",
			resolvedIssuer,
			resolvedSubject,
			identity.Issuer(),
			identity.Subject(),
		)
	}
}

func TestServiceAuthenticateErrors(t *testing.T) {
	t.Parallel()

	identity := mustIdentity(t, "https://example.supabase.co/auth/v1", "external-user-id")
	verificationErr := errors.New("verification failed")
	resolverErr := errors.New("resolver failed")

	tests := []struct {
		name      string
		verifier  ITokenVerifier
		resolver  authentication.IUserResolver
		wantError error
	}{
		{
			name:      "token verification failure",
			verifier:  verifierReturning(Identity{}, verificationErr),
			resolver:  resolverReturning(identifier.NewUserID(), nil),
			wantError: verificationErr,
		},
		{
			name:      "invalid verifier result",
			verifier:  verifierReturning(Identity{}, nil),
			resolver:  resolverReturning(identifier.NewUserID(), nil),
			wantError: ErrInvalidVerifierResult,
		},
		{
			name:      "resolver failure",
			verifier:  verifierReturning(identity, nil),
			resolver:  resolverReturning(identifier.UserID{}, resolverErr),
			wantError: resolverErr,
		},
		{
			name:      "zero user ID",
			verifier:  verifierReturning(identity, nil),
			resolver:  resolverReturning(identifier.UserID{}, nil),
			wantError: ErrInvalidResolvedUser,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			service, err := NewService(test.verifier, test.resolver)
			if err != nil {
				t.Fatalf("NewService() error = %v", err)
			}
			_, err = service.Authenticate(context.Background(), "access-token")
			if !errors.Is(err, test.wantError) {
				t.Fatalf("Authenticate() error = %v, want %v", err, test.wantError)
			}
		})
	}
}

func TestServiceAuthenticatePreservesCanceledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	service, err := NewService(
		verifierReturning(Identity{}, errors.New("must not be called")),
		resolverReturning(identifier.UserID{}, errors.New("must not be called")),
	)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	if _, err := service.Authenticate(ctx, "access-token"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Authenticate() error = %v, want context.Canceled", err)
	}
}

func TestNewServiceRequiresDependencies(t *testing.T) {
	t.Parallel()

	identity := mustIdentity(t, "https://example.supabase.co/auth/v1", "external-user-id")
	if _, err := NewService(nil, resolverReturning(identifier.NewUserID(), nil)); err == nil {
		t.Fatal("NewService(nil, resolver) error = nil")
	}
	if _, err := NewService(verifierReturning(identity, nil), nil); err == nil {
		t.Fatal("NewService(verifier, nil) error = nil")
	}
}

func verifierReturning(identity Identity, err error) ITokenVerifier {
	return TokenVerifierFunc(func(context.Context, string) (Identity, error) {
		return identity, err
	})
}

func resolverReturning(userID identifier.UserID, err error) authentication.IUserResolver {
	return authentication.UserResolverFunc(func(
		context.Context,
		string,
		string,
	) (identifier.UserID, error) {
		return userID, err
	})
}

func mustIdentity(t *testing.T, issuer, subject string) Identity {
	t.Helper()
	identity, err := NewIdentity(issuer, subject)
	if err != nil {
		t.Fatalf("NewIdentity() error = %v", err)
	}
	return identity
}
