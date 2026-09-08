package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/gesop0n/spaco/backend/internal/modules/authentication"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

func TestAuthenticateExecute(t *testing.T) {
	t.Parallel()

	identity := mustIdentity(t, "https://example.supabase.co/auth/v1", "external-user-id")
	wantUserID := identifier.NewUserID()
	var verifiedToken string
	var resolvedIssuer string
	var resolvedSubject string
	var resolvedEmail string

	authenticate, err := NewAuthenticate(
		TokenVerifierFunc(func(_ context.Context, token string) (Identity, error) {
			verifiedToken = token
			return identity, nil
		}),
		authentication.UserResolverFunc(func(
			_ context.Context,
			issuer string,
			subject string,
			email string,
		) (identifier.UserID, error) {
			resolvedIssuer = issuer
			resolvedSubject = subject
			resolvedEmail = email
			return wantUserID, nil
		}),
	)
	if err != nil {
		t.Fatalf("NewAuthenticate() error = %v", err)
	}

	gotUserID, err := authenticate.Execute(context.Background(), "access-token")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if gotUserID != wantUserID {
		t.Fatalf("Execute() = %v, want %v", gotUserID, wantUserID)
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
	if resolvedEmail != identity.Email() {
		t.Fatalf("resolved email = %q, want %q", resolvedEmail, identity.Email())
	}
}

func TestAuthenticateExecuteErrors(t *testing.T) {
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

			authenticate, err := NewAuthenticate(test.verifier, test.resolver)
			if err != nil {
				t.Fatalf("NewAuthenticate() error = %v", err)
			}
			_, err = authenticate.Execute(context.Background(), "access-token")
			if !errors.Is(err, test.wantError) {
				t.Fatalf("Execute() error = %v, want %v", err, test.wantError)
			}
		})
	}
}

func TestAuthenticateExecutePreservesCanceledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	authenticate, err := NewAuthenticate(
		verifierReturning(Identity{}, errors.New("must not be called")),
		resolverReturning(identifier.UserID{}, errors.New("must not be called")),
	)
	if err != nil {
		t.Fatalf("NewAuthenticate() error = %v", err)
	}

	if _, err := authenticate.Execute(ctx, "access-token"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Execute() error = %v, want context.Canceled", err)
	}
}

func TestNewAuthenticateRequiresDependencies(t *testing.T) {
	t.Parallel()

	identity := mustIdentity(t, "https://example.supabase.co/auth/v1", "external-user-id")
	if _, err := NewAuthenticate(nil, resolverReturning(identifier.NewUserID(), nil)); err == nil {
		t.Fatal("NewAuthenticate(nil, resolver) error = nil")
	}
	if _, err := NewAuthenticate(verifierReturning(identity, nil), nil); err == nil {
		t.Fatal("NewAuthenticate(verifier, nil) error = nil")
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
		string,
	) (identifier.UserID, error) {
		return userID, err
	})
}

func mustIdentity(t *testing.T, issuer, subject string) Identity {
	t.Helper()
	identity, err := NewIdentity(issuer, subject, "user@example.com")
	if err != nil {
		t.Fatalf("NewIdentity() error = %v", err)
	}
	return identity
}
