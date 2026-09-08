// Package usecase は、authentication moduleのユースケースを実装する。
package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/gesop0n/spaco/backend/internal/modules/authentication"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

var (
	ErrInvalidVerifierResult = errors.New("token verifier returned an invalid identity")
	ErrInvalidResolvedUser   = errors.New("user resolver returned a zero user ID")
)

// Authenticateは、access tokenからアプリ内UserIDを確定するユースケースである。
type Authenticate struct {
	verifier ITokenVerifier
	resolver authentication.IUserResolver
}

func NewAuthenticate(
	verifier ITokenVerifier,
	resolver authentication.IUserResolver,
) (*Authenticate, error) {
	if verifier == nil {
		return nil, errors.New("create authenticate use case: token verifier is required")
	}
	if resolver == nil {
		return nil, errors.New("create authenticate use case: user resolver is required")
	}

	return &Authenticate{verifier: verifier, resolver: resolver}, nil
}

// Executeは、access tokenから認証済みのアプリ内UserIDを返す。
func (u *Authenticate) Execute(
	ctx context.Context,
	rawToken string,
) (identifier.UserID, error) {
	if err := ctx.Err(); err != nil {
		return identifier.UserID{}, err
	}

	// tokenの形式や署名方式などの技術的な検証はadapterへ委譲する。
	identity, err := u.verifier.Verify(ctx, rawToken)
	if err != nil {
		return identifier.UserID{}, fmt.Errorf("verify access token: %w", err)
	}
	if identity.Issuer() == "" || identity.Subject() == "" {
		return identifier.UserID{}, ErrInvalidVerifierResult
	}

	// 外部identityとアプリ内ユーザーの対応付けはaccount moduleへ委譲する。
	userID, err := u.resolver.ResolveUser(ctx, identity.Issuer(), identity.Subject())
	if err != nil {
		return identifier.UserID{}, fmt.Errorf("resolve authenticated user: %w", err)
	}
	if userID.IsZero() {
		return identifier.UserID{}, ErrInvalidResolvedUser
	}

	return userID, nil
}
