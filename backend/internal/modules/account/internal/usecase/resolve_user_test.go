package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/gesop0n/spaco/backend/internal/modules/account/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

func TestResolveUserExecuteValidatesIdentity(t *testing.T) {
	t.Parallel()

	wantUserID := identifier.NewUserID()
	repository := repositoryStub{
		resolve: func(_ context.Context, identity domain.AuthIdentity) (identifier.UserID, error) {
			if identity.Issuer() != "https://example.test/auth/v1" ||
				identity.Subject() != "subject-1" ||
				identity.Email() != "user@example.com" {
				t.Fatalf("identity = (%q, %q, %q)", identity.Issuer(), identity.Subject(), identity.Email())
			}
			return wantUserID, nil
		},
	}
	resolveUser, err := NewResolveUser(repository)
	if err != nil {
		t.Fatalf("NewResolveUser() error = %v", err)
	}

	got, err := resolveUser.Execute(
		context.Background(),
		" https://example.test/auth/v1 ",
		" subject-1 ",
		" user@example.com ",
	)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got != wantUserID {
		t.Fatalf("Execute() = %v, want %v", got, wantUserID)
	}
}

func TestResolveUserExecuteRejectsMissingEmail(t *testing.T) {
	t.Parallel()

	repositoryCalled := false
	resolveUser, err := NewResolveUser(repositoryStub{
		resolve: func(context.Context, domain.AuthIdentity) (identifier.UserID, error) {
			repositoryCalled = true
			return identifier.NewUserID(), nil
		},
	})
	if err != nil {
		t.Fatalf("NewResolveUser() error = %v", err)
	}

	_, err = resolveUser.Execute(
		context.Background(),
		"https://example.test/auth/v1",
		"subject-1",
		"",
	)
	if !errors.Is(err, domain.ErrInvalidAuthIdentity) {
		t.Fatalf("Execute() error = %v, want ErrInvalidAuthIdentity", err)
	}
	if repositoryCalled {
		t.Fatal("repository was called for an identity without email")
	}
}
