package domain

import (
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidAuthIdentity = errors.New("invalid auth identity")

// AuthIdentityは、外部認証基盤内のユーザーを一意に識別する値である。
// subjectはissuer内でのみ一意なため、必ずissuerと組み合わせて扱う。
type AuthIdentity struct {
	issuer  string
	subject string
}

func NewAuthIdentity(issuer, subject string) (AuthIdentity, error) {
	issuer = strings.TrimSpace(issuer)
	subject = strings.TrimSpace(subject)
	if issuer == "" || subject == "" {
		return AuthIdentity{}, fmt.Errorf(
			"%w: issuer and subject are required",
			ErrInvalidAuthIdentity,
		)
	}

	return AuthIdentity{issuer: issuer, subject: subject}, nil
}

func (i AuthIdentity) Issuer() string  { return i.issuer }
func (i AuthIdentity) Subject() string { return i.subject }
