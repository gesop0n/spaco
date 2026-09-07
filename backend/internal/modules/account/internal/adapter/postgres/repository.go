// Package postgres は、account applicationが要求するrepositoryをPostgreSQLで実装する。
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	accountsqlc "github.com/gesop0n/spaco/backend/internal/modules/account/internal/adapter/postgres/sqlc"
	"github.com/gesop0n/spaco/backend/internal/modules/account/internal/application"
	"github.com/gesop0n/spaco/backend/internal/modules/account/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

type Repository struct {
	pool    *pgxpool.Pool
	queries *accountsqlc.Queries
}

var _ application.IAccountRepository = (*Repository)(nil)

func NewRepository(pool *pgxpool.Pool) (*Repository, error) {
	if pool == nil {
		return nil, errors.New("create account repository: pool is required")
	}
	return &Repository{
		pool:    pool,
		queries: accountsqlc.New(pool),
	}, nil
}

// ResolveOrCreateByAuthIdentityは、identityに対応するUserIDを返す。
// 初回アクセスが並行しても、DBの一意制約とtransactionによってAccountを一つだけ作成する。
func (r *Repository) ResolveOrCreateByAuthIdentity(
	ctx context.Context,
	identity domain.AuthIdentity,
) (identifier.UserID, error) {
	// 通常経路ではtransactionを開始せず、認証requestごとのDB負荷を抑える。
	existingUserID, err := findUserIDByIdentity(ctx, r.queries, identity)
	if err == nil {
		return existingUserID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return identifier.UserID{}, fmt.Errorf("find auth identity: %w", err)
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return identifier.UserID{}, fmt.Errorf("begin resolve auth identity transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	// fast pathの直後に別requestが作成している可能性があるため、transaction内で再確認する。
	transactionQueries := r.queries.WithTx(tx)
	existingUserID, err = findUserIDByIdentity(ctx, transactionQueries, identity)
	if err == nil {
		if commitErr := tx.Commit(ctx); commitErr != nil {
			return identifier.UserID{}, fmt.Errorf("commit existing auth identity: %w", commitErr)
		}
		return existingUserID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return identifier.UserID{}, fmt.Errorf("find auth identity: %w", err)
	}

	newAccount, err := domain.NewAccount(identifier.NewUserID())
	if err != nil {
		return identifier.UserID{}, fmt.Errorf("create account: %w", err)
	}
	if err := transactionQueries.CreateAccount(ctx, accountsqlc.CreateAccountParams{
		ID:       newAccount.ID().UUID(),
		TimeZone: newAccount.TimeZone(),
	}); err != nil {
		return identifier.UserID{}, fmt.Errorf("insert account: %w", err)
	}

	rowsAffected, err := transactionQueries.CreateAuthIdentity(
		ctx,
		accountsqlc.CreateAuthIdentityParams{
			Issuer:  identity.Issuer(),
			Subject: identity.Subject(),
			UserID:  newAccount.ID().UUID(),
		},
	)
	if err != nil {
		return identifier.UserID{}, fmt.Errorf("insert auth identity: %w", err)
	}

	resolvedUserID := newAccount.ID()
	if rowsAffected == 0 {
		// 別requestが先に同じidentityを作成した場合、今回だけ作ったAccountを残さない。
		if err := transactionQueries.DeleteAccount(ctx, newAccount.ID().UUID()); err != nil {
			return identifier.UserID{}, fmt.Errorf("delete duplicate account: %w", err)
		}

		resolvedUserID, err = findUserIDByIdentity(ctx, transactionQueries, identity)
		if err != nil {
			return identifier.UserID{}, fmt.Errorf("find concurrently created auth identity: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return identifier.UserID{}, fmt.Errorf("commit resolved auth identity: %w", err)
	}
	return resolvedUserID, nil
}

func (r *Repository) FindByID(
	ctx context.Context,
	userID identifier.UserID,
) (domain.Account, error) {
	row, err := r.queries.FindAccountByID(ctx, userID.UUID())
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Account{}, application.ErrAccountNotFound
	}
	if err != nil {
		return domain.Account{}, fmt.Errorf("find account by id: %w", err)
	}
	return rehydrateAccount(row.ID, row.AtcoderID, row.TimeZone)
}

func (r *Repository) UpdateProfile(
	ctx context.Context,
	userID identifier.UserID,
	profile domain.Profile,
) (domain.Account, error) {
	atCoderID := profile.AtCoderID()
	row, err := r.queries.UpdateAccountProfile(
		ctx,
		accountsqlc.UpdateAccountProfileParams{
			ID:        userID.UUID(),
			AtcoderID: &atCoderID,
			TimeZone:  profile.TimeZone(),
		},
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Account{}, application.ErrAccountNotFound
	}
	if err != nil {
		return domain.Account{}, fmt.Errorf("update account profile: %w", err)
	}
	return rehydrateAccount(row.ID, row.AtcoderID, row.TimeZone)
}

func rehydrateAccount(
	userIDValue uuid.UUID,
	atCoderID *string,
	timeZone string,
) (domain.Account, error) {
	userID, err := identifier.ParseUserID(userIDValue.String())
	if err != nil {
		return domain.Account{}, fmt.Errorf("parse stored user id: %w", err)
	}
	account, err := domain.RehydrateAccount(userID, atCoderID, timeZone)
	if err != nil {
		return domain.Account{}, fmt.Errorf("rehydrate stored account: %w", err)
	}
	return account, nil
}

func findUserIDByIdentity(
	ctx context.Context,
	queries accountsqlc.Querier,
	identity domain.AuthIdentity,
) (identifier.UserID, error) {
	userIDValue, err := queries.FindUserIDByAuthIdentity(
		ctx,
		accountsqlc.FindUserIDByAuthIdentityParams{
			Issuer:  identity.Issuer(),
			Subject: identity.Subject(),
		},
	)
	if err != nil {
		return identifier.UserID{}, err
	}

	userID, err := identifier.ParseUserID(userIDValue.String())
	if err != nil {
		return identifier.UserID{}, fmt.Errorf("parse stored user id: %w", err)
	}
	return userID, nil
}
