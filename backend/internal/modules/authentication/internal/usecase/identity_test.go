package usecase

import (
	"errors"
	"testing"
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
