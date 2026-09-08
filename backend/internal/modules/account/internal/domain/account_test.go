package domain

import (
	"errors"
	"strings"
	"testing"

	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

func TestNewAccountCreatesIncompleteAccount(t *testing.T) {
	t.Parallel()

	account, err := NewAccount(identifier.NewUserID(), "user@example.com")
	if err != nil {
		t.Fatalf("NewAccount() error = %v", err)
	}
	if account.SetupCompleted() {
		t.Fatal("SetupCompleted() = true, want false")
	}
	if account.TimeZone() != "Asia/Tokyo" {
		t.Fatalf("TimeZone() = %q, want Asia/Tokyo", account.TimeZone())
	}
	if account.Email() != "user@example.com" {
		t.Fatalf("Email() = %q, want user@example.com", account.Email())
	}
	if _, ok := account.AtCoderID(); ok {
		t.Fatal("AtCoderID() exists, want unset")
	}
	if _, ok := account.Username(); ok {
		t.Fatal("Username() exists, want unset")
	}
}

func TestNewAccountRequiresEmail(t *testing.T) {
	t.Parallel()

	_, err := NewAccount(identifier.NewUserID(), " \t ")
	if !errors.Is(err, ErrInvalidEmail) {
		t.Fatalf("NewAccount() error = %v, want ErrInvalidEmail", err)
	}
}

func TestNewProfileValidatesInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		username  string
		atCoderID string
		timeZone  string
		wantError error
	}{
		{name: "valid", username: "復習 太郎", atCoderID: "tourist", timeZone: "Asia/Tokyo"},
		{name: "AtCoder ID omitted", username: "太郎", timeZone: "Asia/Tokyo"},
		{name: "blank AtCoder ID means unset", username: "太郎", atCoderID: " \t ", timeZone: "Asia/Tokyo"},
		{name: "normalizes input", username: " 太郎 ", atCoderID: " tourist ", timeZone: " UTC "},
		{name: "missing username", timeZone: "Asia/Tokyo", wantError: ErrInvalidUsername},
		{name: "space in atcoder id", username: "太郎", atCoderID: "invalid id", timeZone: "Asia/Tokyo", wantError: ErrInvalidAtCoderID},
		{name: "too long atcoder id", username: "太郎", atCoderID: strings.Repeat("a", 65), timeZone: "Asia/Tokyo", wantError: ErrInvalidAtCoderID},
		{name: "unknown time zone", username: "太郎", timeZone: "Mars/Olympus", wantError: ErrInvalidTimeZone},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			profile, err := NewProfile(test.username, test.atCoderID, test.timeZone)
			if test.wantError != nil {
				if !errors.Is(err, test.wantError) {
					t.Fatalf("NewProfile() error = %v, want %v", err, test.wantError)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewProfile() error = %v", err)
			}
			atCoderID, exists := profile.AtCoderID()
			wantAtCoderID := strings.TrimSpace(test.atCoderID)
			if atCoderID != wantAtCoderID || exists != (wantAtCoderID != "") {
				t.Fatalf("AtCoderID() = (%q, %v), want (%q, %v)", atCoderID, exists, wantAtCoderID, wantAtCoderID != "")
			}
			if profile.Username() != strings.TrimSpace(test.username) || profile.TimeZone() != strings.TrimSpace(test.timeZone) {
				t.Fatalf("NewProfile() = (%q, %q)", profile.Username(), profile.TimeZone())
			}
		})
	}
}

func TestRehydrateAccountRestoresSetupState(t *testing.T) {
	t.Parallel()

	username := "太郎"
	atCoderID := "tourist"
	for _, test := range []struct {
		name      string
		username  *string
		atCoderID *string
		wantSetup bool
	}{
		{name: "new account"},
		{name: "existing account needs username", atCoderID: &atCoderID},
		{name: "username without AtCoder ID is complete", username: &username, wantSetup: true},
		{name: "both set", username: &username, atCoderID: &atCoderID, wantSetup: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			account, err := RehydrateAccount(
				identifier.NewUserID(),
				"user@example.com",
				test.username,
				test.atCoderID,
				"UTC",
			)
			if err != nil {
				t.Fatalf("RehydrateAccount() error = %v", err)
			}
			if account.SetupCompleted() != test.wantSetup {
				t.Fatalf("SetupCompleted() = %v, want %v", account.SetupCompleted(), test.wantSetup)
			}
			gotName, hasName := account.Username()
			if hasName != (test.username != nil) || (hasName && gotName != *test.username) {
				t.Fatalf("Username() = (%q, %v)", gotName, hasName)
			}
		})
	}
}
