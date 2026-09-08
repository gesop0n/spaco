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

	_, err = updateProfile.Execute(context.Background(), identifier.NewUserID(), "太郎", "tourist", "invalid/zone")
	if !errors.Is(err, domain.ErrInvalidTimeZone) {
		t.Fatalf("Execute() error = %v, want ErrInvalidTimeZone", err)
	}
	if repositoryCalled {
		t.Fatal("repository was called for invalid input")
	}
}

func TestUpdateProfileExecuteWithoutAtCoderID(t *testing.T) {
	t.Parallel()
	userID := identifier.NewUserID()
	repositoryCalled := false
	updateProfile, err := NewUpdateProfile(repositoryStub{
		update: func(_ context.Context, gotID identifier.UserID, profile domain.Profile) (domain.Account, error) {
			repositoryCalled = true
			if gotID != userID || profile.Username() != "太郎" || profile.TimeZone() != "UTC" {
				t.Fatalf("unexpected profile: %v, %q, %q", gotID, profile.Username(), profile.TimeZone())
			}
			if _, exists := profile.AtCoderID(); exists {
				t.Fatal("blank AtCoder ID must be unset")
			}
			username := profile.Username()
			return domain.RehydrateAccount(gotID, &username, nil, profile.TimeZone())
		},
	})
	if err != nil {
		t.Fatalf("NewUpdateProfile() error = %v", err)
	}
	account, err := updateProfile.Execute(context.Background(), userID, " 太郎 ", "  ", " UTC ")
	if err != nil || !repositoryCalled || !account.SetupCompleted() {
		t.Fatalf("Execute() = %v, %v; repositoryCalled = %v", account, err, repositoryCalled)
	}

	repositoryCalled = false
	_, err = updateProfile.Execute(context.Background(), userID, " ", "", "UTC")
	if !errors.Is(err, domain.ErrInvalidUsername) || repositoryCalled {
		t.Fatalf("blank username: error = %v, repositoryCalled = %v", err, repositoryCalled)
	}
}
