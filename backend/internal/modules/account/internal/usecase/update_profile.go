package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/gesop0n/spaco/backend/internal/modules/account/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

// UpdateProfileは、認証済みユーザーのプロフィールを更新するユースケースである。
type UpdateProfile struct {
	repository IUpdateProfileRepository
}

func NewUpdateProfile(repository IUpdateProfileRepository) (*UpdateProfile, error) {
	if repository == nil {
		return nil, errors.New("create update profile use case: repository is required")
	}
	return &UpdateProfile{repository: repository}, nil
}

func (u *UpdateProfile) Execute(
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

	account, err := u.repository.UpdateProfile(ctx, userID, profile)
	if err != nil {
		return domain.Account{}, fmt.Errorf("update profile: %w", err)
	}
	return account, nil
}
