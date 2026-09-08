package rpc

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"

	accountv1 "github.com/gesop0n/spaco/backend/generated/spaco/account/v1"
	"github.com/gesop0n/spaco/backend/internal/modules/account/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/modules/authentication"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

type getCurrentAccountUseCaseStub struct {
	execute func(context.Context, identifier.UserID) (domain.Account, error)
}

func (s getCurrentAccountUseCaseStub) Execute(
	ctx context.Context,
	userID identifier.UserID,
) (domain.Account, error) {
	return s.execute(ctx, userID)
}

type updateProfileUseCaseStub struct {
	execute func(context.Context, identifier.UserID, string, string) (domain.Account, error)
}

func (s updateProfileUseCaseStub) Execute(
	ctx context.Context,
	userID identifier.UserID,
	atCoderID string,
	timeZone string,
) (domain.Account, error) {
	return s.execute(ctx, userID, atCoderID, timeZone)
}

func TestHandlerGetCurrentAccountReturnsAuthenticatedUser(t *testing.T) {
	t.Parallel()

	userID := identifier.NewUserID()
	atCoderID := "tourist"
	account, err := domain.RehydrateAccount(userID, &atCoderID, "Asia/Tokyo")
	if err != nil {
		t.Fatalf("RehydrateAccount() error = %v", err)
	}
	handler, err := NewHandler(
		getCurrentAccountUseCaseStub{
			execute: func(_ context.Context, gotUserID identifier.UserID) (domain.Account, error) {
				if gotUserID != userID {
					t.Fatalf("userID = %v, want %v", gotUserID, userID)
				}
				return account, nil
			},
		},
		updateProfileUseCaseStub{},
	)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	ctx := authentication.WithUserID(context.Background(), userID)
	response, err := handler.GetCurrentAccount(
		ctx,
		connect.NewRequest(&accountv1.GetCurrentAccountRequest{}),
	)
	if err != nil {
		t.Fatalf("GetCurrentAccount() error = %v", err)
	}
	if response.Msg.Account.GetId() != userID.String() || !response.Msg.Account.GetSetupCompleted() {
		t.Fatalf("account = %+v", response.Msg.Account)
	}
	if response.Msg.Account.GetAtcoderId() != atCoderID {
		t.Fatalf("atcoder_id = %q, want %q", response.Msg.Account.GetAtcoderId(), atCoderID)
	}
}

func TestHandlerGetCurrentAccountRejectsMissingUser(t *testing.T) {
	t.Parallel()

	handler, err := NewHandler(getCurrentAccountUseCaseStub{}, updateProfileUseCaseStub{})
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	_, err = handler.GetCurrentAccount(
		context.Background(),
		connect.NewRequest(&accountv1.GetCurrentAccountRequest{}),
	)
	assertConnectCode(t, err, connect.CodeUnauthenticated)
}

func TestHandlerUpdateProfileMapsDomainError(t *testing.T) {
	t.Parallel()

	handler, err := NewHandler(
		getCurrentAccountUseCaseStub{},
		updateProfileUseCaseStub{
			execute: func(context.Context, identifier.UserID, string, string) (domain.Account, error) {
				return domain.Account{}, domain.ErrInvalidTimeZone
			},
		},
	)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	ctx := authentication.WithUserID(context.Background(), identifier.NewUserID())
	_, err = handler.UpdateProfile(
		ctx,
		connect.NewRequest(&accountv1.UpdateProfileRequest{
			AtcoderId: "tourist",
			TimeZone:  "invalid/zone",
		}),
	)
	assertConnectCode(t, err, connect.CodeInvalidArgument)
}

func assertConnectCode(t *testing.T, err error, want connect.Code) {
	t.Helper()
	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("error = %v, want *connect.Error", err)
	}
	if connectErr.Code() != want {
		t.Fatalf("code = %v, want %v", connectErr.Code(), want)
	}
}
