package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/gesop0n/spaco/backend/internal/modules/account/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

func TestUpdateProfileExecuteRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	repositoryCalled := false
	repository := repositoryStub{
		update: func(context.Context, identifier.UserID, domain.Profile) (domain.Account, error) {
			repositoryCalled = true
			return domain.Account{}, nil
		},
	}
	updateProfile, err := NewUpdateProfile(repository)
	if err != nil {
		t.Fatalf("NewUpdateProfile() error = %v", err)
	}

	_, err = updateProfile.Execute(context.Background(), identifier.NewUserID(), "tourist", "invalid/zone")
	if !errors.Is(err, domain.ErrInvalidTimeZone) {
		t.Fatalf("Execute() error = %v, want ErrInvalidTimeZone", err)
	}
	if repositoryCalled {
		t.Fatal("repository was called for invalid input")
	}
}
