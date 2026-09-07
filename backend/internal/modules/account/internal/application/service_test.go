package application

import (
	"context"
	"errors"
	"testing"

	"github.com/gesop0n/spaco/backend/internal/modules/account/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

type repositoryStub struct {
	resolve func(context.Context, domain.AuthIdentity) (identifier.UserID, error)
	find    func(context.Context, identifier.UserID) (domain.Account, error)
	update  func(context.Context, identifier.UserID, domain.Profile) (domain.Account, error)
}

func (s repositoryStub) ResolveOrCreateByAuthIdentity(
	ctx context.Context,
	identity domain.AuthIdentity,
) (identifier.UserID, error) {
	return s.resolve(ctx, identity)
}

func (s repositoryStub) FindByID(
	ctx context.Context,
	userID identifier.UserID,
) (domain.Account, error) {
	return s.find(ctx, userID)
}

func (s repositoryStub) UpdateProfile(
	ctx context.Context,
	userID identifier.UserID,
	profile domain.Profile,
) (domain.Account, error) {
	return s.update(ctx, userID, profile)
}

func TestServiceResolveUserValidatesIdentity(t *testing.T) {
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
	service, err := NewService(repository)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	got, err := service.ResolveUser(
		context.Background(),
		" https://example.test/auth/v1 ",
		" subject-1 ",
	)
	if err != nil {
		t.Fatalf("ResolveUser() error = %v", err)
	}
	if got != wantUserID {
		t.Fatalf("ResolveUser() = %v, want %v", got, wantUserID)
	}
}

func TestServiceUpdateProfileRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	repositoryCalled := false
	repository := repositoryStub{
		update: func(context.Context, identifier.UserID, domain.Profile) (domain.Account, error) {
			repositoryCalled = true
			return domain.Account{}, nil
		},
	}
	service, err := NewService(repository)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	_, err = service.UpdateProfile(context.Background(), identifier.NewUserID(), "tourist", "invalid/zone")
	if !errors.Is(err, domain.ErrInvalidTimeZone) {
		t.Fatalf("UpdateProfile() error = %v, want ErrInvalidTimeZone", err)
	}
	if repositoryCalled {
		t.Fatal("repository was called for invalid input")
	}
}
