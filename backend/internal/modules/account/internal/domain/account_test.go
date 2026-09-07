package domain

import (
	"errors"
	"testing"

	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

func TestNewAccountCreatesIncompleteAccount(t *testing.T) {
	t.Parallel()

	account, err := NewAccount(identifier.NewUserID())
	if err != nil {
		t.Fatalf("NewAccount() error = %v", err)
	}
	if account.SetupCompleted() {
		t.Fatal("SetupCompleted() = true, want false")
	}
	if account.TimeZone() != "Asia/Tokyo" {
		t.Fatalf("TimeZone() = %q, want Asia/Tokyo", account.TimeZone())
	}
	if _, ok := account.AtCoderID(); ok {
		t.Fatal("AtCoderID() exists, want unset")
	}
}

func TestNewProfileValidatesInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		atCoderID string
		timeZone  string
		wantError error
	}{
		{name: "valid", atCoderID: "tourist", timeZone: "Asia/Tokyo"},
		{name: "blank atcoder id", atCoderID: " ", timeZone: "Asia/Tokyo", wantError: ErrInvalidAtCoderID},
		{name: "space in atcoder id", atCoderID: "invalid id", timeZone: "Asia/Tokyo", wantError: ErrInvalidAtCoderID},
		{name: "unknown time zone", atCoderID: "tourist", timeZone: "Mars/Olympus", wantError: ErrInvalidTimeZone},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			profile, err := NewProfile(test.atCoderID, test.timeZone)
			if test.wantError != nil {
				if !errors.Is(err, test.wantError) {
					t.Fatalf("NewProfile() error = %v, want %v", err, test.wantError)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewProfile() error = %v", err)
			}
			if profile.AtCoderID() != test.atCoderID || profile.TimeZone() != test.timeZone {
				t.Fatalf("NewProfile() = (%q, %q)", profile.AtCoderID(), profile.TimeZone())
			}
		})
	}
}

func TestRehydrateAccountRestoresSetupState(t *testing.T) {
	t.Parallel()

	atCoderID := "tourist"
	account, err := RehydrateAccount(identifier.NewUserID(), &atCoderID, "UTC")
	if err != nil {
		t.Fatalf("RehydrateAccount() error = %v", err)
	}
	if !account.SetupCompleted() {
		t.Fatal("SetupCompleted() = false, want true")
	}
}
