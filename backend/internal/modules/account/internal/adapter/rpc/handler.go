// Package rpc は、AccountのユースケースをConnectRPCへ接続するadapterを提供する。
package rpc

import (
	"context"
	"errors"

	"connectrpc.com/connect"

	accountv1 "github.com/gesop0n/spaco/backend/generated/spaco/account/v1"
	"github.com/gesop0n/spaco/backend/internal/modules/account/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/modules/account/internal/usecase"
	"github.com/gesop0n/spaco/backend/internal/modules/authentication"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

// IGetCurrentAccountUseCaseは、RPC adapterが必要とするAccount参照を表すinterfaceである。
type IGetCurrentAccountUseCase interface {
	Execute(context.Context, identifier.UserID) (domain.Account, error)
}

// IUpdateProfileUseCaseは、RPC adapterが必要とするプロフィール更新を表すinterfaceである。
type IUpdateProfileUseCase interface {
	Execute(context.Context, identifier.UserID, string, string, string) (domain.Account, error)
}

type Handler struct {
	getCurrentAccount IGetCurrentAccountUseCase
	updateProfile     IUpdateProfileUseCase
}

func NewHandler(
	getCurrentAccount IGetCurrentAccountUseCase,
	updateProfile IUpdateProfileUseCase,
) (*Handler, error) {
	if getCurrentAccount == nil {
		return nil, errors.New("create account handler: get current account use case is required")
	}
	if updateProfile == nil {
		return nil, errors.New("create account handler: update profile use case is required")
	}
	return &Handler{
		getCurrentAccount: getCurrentAccount,
		updateProfile:     updateProfile,
	}, nil
}

func (h *Handler) GetCurrentAccount(
	ctx context.Context,
	_ *connect.Request[accountv1.GetCurrentAccountRequest],
) (*connect.Response[accountv1.GetCurrentAccountResponse], error) {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return nil, err
	}

	account, err := h.getCurrentAccount.Execute(ctx, userID)
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&accountv1.GetCurrentAccountResponse{
		Account: accountMessage(account),
	}), nil
}

func (h *Handler) UpdateProfile(
	ctx context.Context,
	request *connect.Request[accountv1.UpdateProfileRequest],
) (*connect.Response[accountv1.UpdateProfileResponse], error) {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return nil, err
	}
	if request == nil || request.Msg == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("request is required"))
	}

	account, err := h.updateProfile.Execute(
		ctx,
		userID,
		request.Msg.GetUsername(),
		request.Msg.GetAtcoderId(),
		request.Msg.GetTimeZone(),
	)
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&accountv1.UpdateProfileResponse{
		Account: accountMessage(account),
	}), nil
}

func authenticatedUserID(ctx context.Context) (identifier.UserID, error) {
	userID, ok := authentication.UserIDFromContext(ctx)
	if !ok {
		return identifier.UserID{}, connect.NewError(
			connect.CodeUnauthenticated,
			errors.New("authentication required"),
		)
	}
	return userID, nil
}

func accountMessage(account domain.Account) *accountv1.Account {
	message := &accountv1.Account{
		Id:             account.ID().String(),
		Email:          account.Email(),
		TimeZone:       account.TimeZone(),
		SetupCompleted: account.SetupCompleted(),
	}
	if username, ok := account.Username(); ok {
		message.Username = &username
	}
	if atCoderID, ok := account.AtCoderID(); ok {
		message.AtcoderId = &atCoderID
	}
	return message
}

func connectError(err error) *connect.Error {
	switch {
	case errors.Is(err, context.Canceled):
		return connect.NewError(connect.CodeCanceled, context.Canceled)
	case errors.Is(err, context.DeadlineExceeded):
		return connect.NewError(connect.CodeDeadlineExceeded, context.DeadlineExceeded)
	case errors.Is(err, domain.ErrInvalidUsername), errors.Is(err, domain.ErrInvalidAtCoderID), errors.Is(err, domain.ErrInvalidTimeZone):
		return connect.NewError(connect.CodeInvalidArgument, errors.New("invalid profile"))
	case errors.Is(err, usecase.ErrAccountNotFound):
		return connect.NewError(connect.CodeNotFound, errors.New("account not found"))
	default:
		// DBや内部構造の詳細はclientへ公開しない。
		return connect.NewError(connect.CodeInternal, errors.New("account operation failed"))
	}
}
