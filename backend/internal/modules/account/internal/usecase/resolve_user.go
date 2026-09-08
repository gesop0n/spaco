package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/gesop0n/spaco/backend/internal/modules/account/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

// ResolveUserは、外部identityに対応するUserIDを解決するユースケースである。
type ResolveUser struct {
	repository IResolveUserRepository
}

func NewResolveUser(repository IResolveUserRepository) (*ResolveUser, error) {
	if repository == nil {
		return nil, errors.New("create resolve user use case: repository is required")
	}
	return &ResolveUser{repository: repository}, nil
}

// Executeは、外部identityに対応するUserIDを返す。
// 初回だけAccountと対応関係を同じtransaction内で自動作成する。
func (u *ResolveUser) Execute(
	ctx context.Context,
	issuer string,
	subject string,
) (identifier.UserID, error) {
	identity, err := domain.NewAuthIdentity(issuer, subject)
	if err != nil {
		return identifier.UserID{}, err
	}

	userID, err := u.repository.ResolveOrCreateByAuthIdentity(ctx, identity)
	if err != nil {
		return identifier.UserID{}, fmt.Errorf("resolve account by auth identity: %w", err)
	}
	return userID, nil
}
