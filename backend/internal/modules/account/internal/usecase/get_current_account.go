package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/gesop0n/spaco/backend/internal/modules/account/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

// GetCurrentAccountは、認証済みユーザーのAccountを取得するユースケースである。
type GetCurrentAccount struct {
	repository IGetCurrentAccountRepository
}

func NewGetCurrentAccount(repository IGetCurrentAccountRepository) (*GetCurrentAccount, error) {
	if repository == nil {
		return nil, errors.New("create get current account use case: repository is required")
	}
	return &GetCurrentAccount{repository: repository}, nil
}

func (u *GetCurrentAccount) Execute(
	ctx context.Context,
	userID identifier.UserID,
) (domain.Account, error) {
	if userID.IsZero() {
		return domain.Account{}, fmt.Errorf("get current account: user id is required")
	}

	account, err := u.repository.FindByID(ctx, userID)
	if err != nil {
		return domain.Account{}, fmt.Errorf("get current account: %w", err)
	}
	return account, nil
}
