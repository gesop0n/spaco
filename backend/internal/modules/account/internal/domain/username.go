package domain

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

var ErrInvalidUsername = errors.New("invalid username")

// Usernameは、アプリ内で表示する名前である。他ユーザーとの重複を許可する。
type Username struct{ value string }

func NewUsername(value string) (Username, error) {
	value = strings.TrimSpace(value)
	if value == "" || !utf8.ValidString(value) || utf8.RuneCountInString(value) > 40 ||
		strings.IndexFunc(value, unicode.IsControl) >= 0 {
		return Username{}, fmt.Errorf("%w: must be 1-40 characters without control characters", ErrInvalidUsername)
	}
	return Username{value: value}, nil
}

func (name Username) String() string { return name.value }
