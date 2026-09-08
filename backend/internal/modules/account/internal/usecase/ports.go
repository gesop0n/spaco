// Package usecase は、account moduleのユースケースと必要なportを定義する。
package usecase

import (
	"context"

	"github.com/gesop0n/spaco/backend/internal/modules/account/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

// PostgreSQL固有の型は、いずれのrepository interfaceにも持ち込まない。

type IResolveUserRepository interface {
	ResolveOrCreateByAuthIdentity(
		context.Context,
		domain.AuthIdentity,
	) (identifier.UserID, error)
}

type IGetCurrentAccountRepository interface {
	FindByID(context.Context, identifier.UserID) (domain.Account, error)
}

type IUpdateProfileRepository interface {
	UpdateProfile(
		context.Context,
		identifier.UserID,
		domain.Profile,
	) (domain.Account, error)
}
