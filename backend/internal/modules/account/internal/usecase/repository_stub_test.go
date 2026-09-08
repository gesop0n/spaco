package usecase

import (
	"context"

	"github.com/gesop0n/spaco/backend/internal/modules/account/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

type repositoryStub struct {
	resolve func(context.Context, domain.AuthIdentity) (identifier.UserID, error)
	find    func(context.Context, identifier.UserID) (domain.Account, error)
	update  func(context.Context, identifier.UserID, domain.Profile) (domain.Account, error)
}

func (s repositoryStub) ResolveOrCreateByAuthIdentity(
	ctx context.Context,
	identity domain.AuthIdentity,
) (identifier.UserID, error) {
	return s.resolve(ctx, identity)
}

func (s repositoryStub) FindByID(
	ctx context.Context,
	userID identifier.UserID,
) (domain.Account, error) {
	return s.find(ctx, userID)
}

func (s repositoryStub) UpdateProfile(
	ctx context.Context,
	userID identifier.UserID,
	profile domain.Profile,
) (domain.Account, error) {
	return s.update(ctx, userID, profile)
}
