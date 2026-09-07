package application

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
}

// NewIdentityは、検証済みclaimのissuerとsubjectからIdentityを生成する。
func NewIdentity(issuer, subject string) (Identity, error) {
	issuer = strings.TrimSpace(issuer)
	subject = strings.TrimSpace(subject)
	if issuer == "" || subject == "" {
		return Identity{}, fmt.Errorf("%w: issuer and subject are required", ErrInvalidIdentity)
	}

	return Identity{issuer: issuer, subject: subject}, nil
}

func (i Identity) Issuer() string  { return i.issuer }
func (i Identity) Subject() string { return i.subject }
