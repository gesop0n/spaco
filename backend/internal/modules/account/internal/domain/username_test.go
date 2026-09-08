package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestNewUsername(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name    string
		input   string
		want    string
		invalid bool
	}{
		{name: "Japanese display name", input: " 復習 太郎 ", want: "復習 太郎"},
		{name: "40 Unicode characters", input: strings.Repeat("あ", 40), want: strings.Repeat("あ", 40)},
		{name: "40 emoji", input: strings.Repeat("🌱", 40), want: strings.Repeat("🌱", 40)},
		{name: "empty", invalid: true},
		{name: "whitespace only", input: " \t\u3000 ", invalid: true},
		{name: "too long", input: strings.Repeat("あ", 41), invalid: true},
		{name: "control character", input: "a\x00b", invalid: true},
		{name: "line break", input: "a\nb", invalid: true},
		{name: "invalid UTF-8", input: "\xff", invalid: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			username, err := NewUsername(test.input)
			if test.invalid {
				if !errors.Is(err, ErrInvalidUsername) {
					t.Fatalf("NewUsername() error = %v, want ErrInvalidUsername", err)
				}
				return
			}
			if err != nil || username.String() != test.want {
				t.Fatalf("NewUsername() = %q, %v; want %q, nil", username.String(), err, test.want)
			}
		})
	}
}
