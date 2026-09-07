package authentication

import (
	"context"

	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

type userIDContextKey struct{}

// WithUserID は、認証済みUserIDをcontextへ格納する。
// authentication moduleのtransport adapterから利用することを想定している。
func WithUserID(ctx context.Context, userID identifier.UserID) context.Context {
	return context.WithValue(ctx, userIDContextKey{}, userID)
}

// UserIDFromContext は、認証処理がcontextに格納したアプリ内UserIDを返す。
func UserIDFromContext(ctx context.Context) (identifier.UserID, bool) {
	userID, ok := ctx.Value(userIDContextKey{}).(identifier.UserID)
	return userID, ok && !userID.IsZero()
}
