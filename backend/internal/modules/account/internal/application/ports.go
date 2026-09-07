// Package application は、account moduleのユースケースと必要なportを定義する。
package application

import (
	"context"

	"github.com/gesop0n/spaco/backend/internal/modules/account/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

// IAccountRepositoryは、Accountの永続化に必要な操作だけを表すinterfaceである。
// PostgreSQL固有の型をapplication層へ持ち込まない。
type IAccountRepository interface {
	ResolveOrCreateByAuthIdentity(
		context.Context,
		domain.AuthIdentity,
	) (identifier.UserID, error)
	FindByID(context.Context, identifier.UserID) (domain.Account, error)
	UpdateProfile(
		context.Context,
		identifier.UserID,
		domain.Profile,
	) (domain.Account, error)
}
