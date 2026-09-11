// Package account は、Account機能と他moduleへ公開する契約を提供する。
package account

import (
	"context"
	"errors"
	"net/http"

	"connectrpc.com/connect"

	"github.com/gesop0n/spaco/backend/generated/spaco/account/v1/accountv1connect"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

// ResolveUserFuncは、外部identityをアプリ内UserIDへ変換する関数である。
type ResolveUserFunc func(context.Context, string, string, string) (identifier.UserID, error)

// TimeZoneFuncは、ユーザーが設定したIANAタイムゾーン名を返す関数である。
type TimeZoneFunc func(context.Context, identifier.UserID) (string, error)

// Moduleは、組み立て済みaccount moduleの公開窓口である。
// use caseやrepositoryの具象型はmodule外へ公開しない。
type Module struct {
	handler     accountv1connect.AccountServiceHandler
	resolveUser ResolveUserFunc
	timeZone    TimeZoneFunc
}

func NewModule(
	handler accountv1connect.AccountServiceHandler,
	resolveUser ResolveUserFunc,
	timeZone TimeZoneFunc,
) (*Module, error) {
	if handler == nil {
		return nil, errors.New("create account module: handler is required")
	}
	if resolveUser == nil {
		return nil, errors.New("create account module: user resolver is required")
	}
	if timeZone == nil {
		return nil, errors.New("create account module: time zone reader is required")
	}
	return &Module{handler: handler, resolveUser: resolveUser, timeZone: timeZone}, nil
}

// TimeZoneは、ユーザーが設定したIANAタイムゾーン名を返す。
// 他moduleは日付の計算に使うだけで、タイムゾーンの設定・変更はaccountだけが行う。
func (m *Module) TimeZone(ctx context.Context, userID identifier.UserID) (string, error) {
	return m.timeZone(ctx, userID)
}

// ResolveUserによりModule自身がauthentication.IUserResolverを満たす。
func (m *Module) ResolveUser(
	ctx context.Context,
	issuer string,
	subject string,
	email string,
) (identifier.UserID, error) {
	return m.resolveUser(ctx, issuer, subject, email)
}

// ConnectHandlerは、AccountServiceをmountするpathとHTTP handlerを返す。
func (m *Module) ConnectHandler(
	options ...connect.HandlerOption,
) (string, http.Handler) {
	return accountv1connect.NewAccountServiceHandler(m.handler, options...)
}
