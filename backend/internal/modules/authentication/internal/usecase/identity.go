package usecase

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrInvalidIdentityは、外部認証identityのissuerまたはsubjectが不正なことを表す。
	ErrInvalidIdentity = errors.New("invalid external identity")
	// ErrInvalidTokenは、access tokenを信頼できないことを表す。
	ErrInvalidToken = errors.New("invalid access token")
)

// Identityは、外部認証基盤におけるユーザーを識別する値である。
// subjectはissuer内でのみ一意な可能性があるため、必ず両方を保持する。
type Identity struct {
	issuer  string
	subject string
	email   string
}

// NewIdentityは、検証済みclaimのissuer、subject、emailからIdentityを生成する。
func NewIdentity(issuer, subject, email string) (Identity, error) {
	issuer = strings.TrimSpace(issuer)
	subject = strings.TrimSpace(subject)
	email = strings.TrimSpace(email)
	if issuer == "" || subject == "" || email == "" {
		return Identity{}, fmt.Errorf("%w: issuer, subject, and email are required", ErrInvalidIdentity)
	}

	return Identity{issuer: issuer, subject: subject, email: email}, nil
}

func (i Identity) Issuer() string  { return i.issuer }
func (i Identity) Subject() string { return i.subject }
func (i Identity) Email() string   { return i.email }
