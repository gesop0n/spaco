// Package application は、authentication moduleのuse caseを実装する。
package application

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

// Serviceは、access tokenの検証とアプリ内UserIDの解決を順番に実行する。
type Service struct {
	verifier ITokenVerifier
	resolver authentication.IUserResolver
}

// NewServiceは、authentication use caseを生成する。
func NewService(
	verifier ITokenVerifier,
	resolver authentication.IUserResolver,
) (*Service, error) {
	if verifier == nil {
		return nil, errors.New("create authentication service: token verifier is required")
	}
	if resolver == nil {
		return nil, errors.New("create authentication service: user resolver is required")
	}

	return &Service{verifier: verifier, resolver: resolver}, nil
}

// Authenticateは、access tokenから認証済みのアプリ内UserIDを返す。
func (s *Service) Authenticate(
	ctx context.Context,
	rawToken string,
) (identifier.UserID, error) {
	if err := ctx.Err(); err != nil {
		return identifier.UserID{}, err
	}

	// tokenの形式や署名方式などの技術的な検証はadapterへ委譲する。
	identity, err := s.verifier.Verify(ctx, rawToken)
	if err != nil {
		return identifier.UserID{}, fmt.Errorf("verify access token: %w", err)
	}
	if identity.Issuer() == "" || identity.Subject() == "" {
		return identifier.UserID{}, ErrInvalidVerifierResult
	}

	// 外部identityとアプリ内ユーザーの対応付けはaccount moduleへ委譲する。
	userID, err := s.resolver.ResolveUser(ctx, identity.Issuer(), identity.Subject())
	if err != nil {
		return identifier.UserID{}, fmt.Errorf("resolve authenticated user: %w", err)
	}
	if userID.IsZero() {
		return identifier.UserID{}, ErrInvalidResolvedUser
	}

	return userID, nil
}
