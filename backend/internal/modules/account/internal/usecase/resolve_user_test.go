package usecase

import (
	"context"
	"testing"

	"github.com/gesop0n/spaco/backend/internal/modules/account/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

func TestResolveUserExecuteValidatesIdentity(t *testing.T) {
	t.Parallel()

	wantUserID := identifier.NewUserID()
	repository := repositoryStub{
		resolve: func(_ context.Context, identity domain.AuthIdentity) (identifier.UserID, error) {
			if identity.Issuer() != "https://example.test/auth/v1" || identity.Subject() != "subject-1" {
				t.Fatalf("identity = (%q, %q)", identity.Issuer(), identity.Subject())
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
	)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got != wantUserID {
		t.Fatalf("Execute() = %v, want %v", got, wantUserID)
	}
}
