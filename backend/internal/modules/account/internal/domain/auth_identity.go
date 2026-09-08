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
	email   string
}

func NewAuthIdentity(issuer, subject, email string) (AuthIdentity, error) {
	issuer = strings.TrimSpace(issuer)
	subject = strings.TrimSpace(subject)
	email = strings.TrimSpace(email)
	if issuer == "" || subject == "" || email == "" {
		return AuthIdentity{}, fmt.Errorf(
			"%w: issuer, subject, and email are required",
			ErrInvalidAuthIdentity,
		)
	}

	return AuthIdentity{issuer: issuer, subject: subject, email: email}, nil
}

func (i AuthIdentity) Issuer() string  { return i.issuer }
func (i AuthIdentity) Subject() string { return i.subject }
func (i AuthIdentity) Email() string   { return i.email }
