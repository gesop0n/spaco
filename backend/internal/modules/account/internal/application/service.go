package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/gesop0n/spaco/backend/internal/modules/account/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

var ErrAccountNotFound = errors.New("account not found")

// Serviceは、外部identityの解決とAccountの参照・更新を実行する。
type Service struct {
	repository IAccountRepository
}

func NewService(repository IAccountRepository) (*Service, error) {
	if repository == nil {
		return nil, errors.New("create account service: repository is required")
	}
	return &Service{repository: repository}, nil
}

// ResolveUserは、外部identityに対応するUserIDを返す。
// 初回だけAccountと対応関係を同じtransaction内で自動作成する。
func (s *Service) ResolveUser(
	ctx context.Context,
	issuer string,
	subject string,
) (identifier.UserID, error) {
	identity, err := domain.NewAuthIdentity(issuer, subject)
	if err != nil {
		return identifier.UserID{}, err
	}

	userID, err := s.repository.ResolveOrCreateByAuthIdentity(ctx, identity)
	if err != nil {
		return identifier.UserID{}, fmt.Errorf("resolve account by auth identity: %w", err)
	}
	return userID, nil
}

func (s *Service) GetCurrentAccount(
	ctx context.Context,
	userID identifier.UserID,
) (domain.Account, error) {
	if userID.IsZero() {
		return domain.Account{}, fmt.Errorf("get current account: user id is required")
	}

	account, err := s.repository.FindByID(ctx, userID)
	if err != nil {
		return domain.Account{}, fmt.Errorf("get current account: %w", err)
	}
	return account, nil
}

func (s *Service) UpdateProfile(
	ctx context.Context,
	userID identifier.UserID,
	atCoderID string,
	timeZone string,
) (domain.Account, error) {
	if userID.IsZero() {
		return domain.Account{}, fmt.Errorf("update profile: user id is required")
	}
	profile, err := domain.NewProfile(atCoderID, timeZone)
	if err != nil {
		return domain.Account{}, fmt.Errorf("update profile: %w", err)
	}

	account, err := s.repository.UpdateProfile(ctx, userID, profile)
	if err != nil {
		return domain.Account{}, fmt.Errorf("update profile: %w", err)
	}
	return account, nil
}
