package usecase

import (
	"errors"
	"testing"
)

func TestNewIdentity(t *testing.T) {
	t.Parallel()

	identity, err := NewIdentity(
		"  https://example.supabase.co/auth/v1/  ",
		" user-id ",
		" user@example.com ",
	)
	if err != nil {
		t.Fatalf("NewIdentity() error = %v", err)
	}
	if identity.Issuer() != "https://example.supabase.co/auth/v1/" {
		t.Fatalf("Issuer() = %q", identity.Issuer())
	}
	if identity.Subject() != "user-id" {
		t.Fatalf("Subject() = %q", identity.Subject())
	}
	if identity.Email() != "user@example.com" {
		t.Fatalf("Email() = %q", identity.Email())
	}

	for _, test := range []struct {
		name    string
		issuer  string
		subject string
		email   string
	}{
		{name: "missing issuer", subject: "user-id", email: "user@example.com"},
		{name: "missing subject", issuer: "https://example.supabase.co/auth/v1", email: "user@example.com"},
		{name: "missing email", issuer: "https://example.supabase.co/auth/v1", subject: "user-id"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewIdentity(test.issuer, test.subject, test.email)
			if !errors.Is(err, ErrInvalidIdentity) {
				t.Fatalf("NewIdentity() error = %v, want ErrInvalidIdentity", err)
			}
		})
	}
}
