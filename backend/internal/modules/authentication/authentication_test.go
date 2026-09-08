package authentication

import (
	"context"
	"testing"

	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

func TestUserIDContext(t *testing.T) {
	t.Parallel()

	if _, ok := UserIDFromContext(context.Background()); ok {
		t.Fatal("UserIDFromContext() found an unset user ID")
	}

	want := identifier.NewUserID()
	ctx := WithUserID(context.Background(), want)
	got, ok := UserIDFromContext(ctx)
	if !ok || got != want {
		t.Fatalf("UserIDFromContext() = %v, %v; want %v, true", got, ok, want)
	}
}

func TestUserResolverFunc(t *testing.T) {
	t.Parallel()

	want := identifier.NewUserID()
	resolver := UserResolverFunc(func(_ context.Context, issuer, subject, email string) (identifier.UserID, error) {
		if issuer != "https://example.supabase.co/auth/v1" {
			t.Fatalf("issuer = %q", issuer)
		}
		if subject != "external-user-id" {
			t.Fatalf("subject = %q", subject)
		}
		if email != "user@example.com" {
			t.Fatalf("email = %q", email)
		}
		return want, nil
	})

	got, err := resolver.ResolveUser(
		context.Background(),
		"https://example.supabase.co/auth/v1",
		"external-user-id",
		"user@example.com",
	)
	if err != nil {
		t.Fatalf("ResolveUser() error = %v", err)
	}
	if got != want {
		t.Fatalf("ResolveUser() = %#v, want %#v", got, want)
	}
}
