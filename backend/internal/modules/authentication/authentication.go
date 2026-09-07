// Package authentication は、外部のaccess tokenからアプリ内ユーザーを認証する
// 機能と、ほかのmoduleへ公開する契約を提供する。
package authentication

import (
	"context"

	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

// IUserResolver は、検証済みの外部認証identityをアプリ内UserIDへ変換するinterfaceである。
// account moduleがこのinterfaceを実装し、authentication moduleへ注入する。
type IUserResolver interface {
	ResolveUser(
		ctx context.Context,
		issuer string,
		subject string,
	) (identifier.UserID, error)
}

// UserResolverFunc は、関数をIUserResolverとして利用できるようにするadapterである。
type UserResolverFunc func(context.Context, string, string) (identifier.UserID, error)

func (f UserResolverFunc) ResolveUser(
	ctx context.Context,
	issuer string,
	subject string,
) (identifier.UserID, error) {
	return f(ctx, issuer, subject)
}
