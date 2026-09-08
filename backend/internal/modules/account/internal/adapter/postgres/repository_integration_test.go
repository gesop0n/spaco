package postgres

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gesop0n/spaco/backend/internal/modules/account/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

// TEST_DATABASE_URLが設定された場合だけ、migration適用済みPostgreSQLとの接続を確認する。
func TestRepositoryIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("pgxpool.New() error = %v", err)
	}
	defer pool.Close()

	repository, err := NewRepository(pool)
	if err != nil {
		t.Fatalf("NewRepository() error = %v", err)
	}
	issuer := fmt.Sprintf("https://integration-%s.test/auth/v1", uuid.NewString())
	identity, err := domain.NewAuthIdentity(issuer, "subject-1")
	if err != nil {
		t.Fatalf("NewAuthIdentity() error = %v", err)
	}
	// poolを閉じる前に、このtestが作成したAccountだけを削除する。
	defer func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		_, _ = pool.Exec(
			cleanupContext,
			`DELETE FROM accounts
			 WHERE id IN (SELECT user_id FROM auth_identities WHERE issuer = $1)`,
			issuer,
		)
	}()

	// 同じidentityを並行解決しても、全requestが同じUserIDを取得することを確認する。
	const requestCount = 8
	userIDs := make(chan string, requestCount)
	errorsChannel := make(chan error, requestCount)
	var waitGroup sync.WaitGroup
	for range requestCount {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			userID, resolveErr := repository.ResolveOrCreateByAuthIdentity(ctx, identity)
			if resolveErr != nil {
				errorsChannel <- resolveErr
				return
			}
			userIDs <- userID.String()
		}()
	}
	waitGroup.Wait()
	close(userIDs)
	close(errorsChannel)
	for resolveErr := range errorsChannel {
		t.Errorf("ResolveOrCreateByAuthIdentity() error = %v", resolveErr)
	}
	if t.Failed() {
		return
	}

	var resolvedUserID string
	for userID := range userIDs {
		if resolvedUserID == "" {
			resolvedUserID = userID
		}
		if userID != resolvedUserID {
			t.Fatalf("resolved user IDs differ: %q and %q", resolvedUserID, userID)
		}
	}

	var accountCount int
	if err := pool.QueryRow(
		ctx,
		`SELECT count(*)
		 FROM accounts AS account
		 JOIN auth_identities AS identity ON identity.user_id = account.id
		 WHERE identity.issuer = $1 AND identity.subject = $2`,
		identity.Issuer(),
		identity.Subject(),
	).Scan(&accountCount); err != nil {
		t.Fatalf("count provisioned accounts error = %v", err)
	}
	if accountCount != 1 {
		t.Fatalf("provisioned account count = %d, want 1", accountCount)
	}

	userID, err := identifier.ParseUserID(resolvedUserID)
	if err != nil {
		t.Fatalf("ParseUserID() error = %v", err)
	}
	account, err := repository.FindByID(ctx, userID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if account.SetupCompleted() {
		t.Fatal("new account SetupCompleted() = true, want false")
	}

	// AtCoder IDなしで初期設定し、追加・解除とユーザー名変更を永続化できる。
	for _, update := range []struct {
		username  string
		atCoderID string
	}{
		{username: "復習 太郎"},
		{username: "復習 太郎", atCoderID: "tourist"},
		{username: "新しい名前"},
	} {
		profile, err := domain.NewProfile(update.username, update.atCoderID, "Asia/Tokyo")
		if err != nil {
			t.Fatalf("NewProfile() error = %v", err)
		}
		updated, err := repository.UpdateProfile(ctx, userID, profile)
		if err != nil {
			t.Fatalf("UpdateProfile() error = %v", err)
		}
		stored, err := repository.FindByID(ctx, userID)
		if err != nil {
			t.Fatalf("FindByID() error = %v", err)
		}
		for _, result := range []domain.Account{updated, stored} {
			if name, ok := result.Username(); !ok || name != update.username {
				t.Fatalf("Username() = %q, %v; want %q", name, ok, update.username)
			}
			if atCoderID, ok := result.AtCoderID(); atCoderID != update.atCoderID || ok != (update.atCoderID != "") {
				t.Fatalf("AtCoderID() = %q, %v; want %q", atCoderID, ok, update.atCoderID)
			}
			if !result.SetupCompleted() {
				t.Fatal("updated account SetupCompleted() = false, want true")
			}
		}
	}
	var atCoderIDIsNull bool
	if err := pool.QueryRow(ctx, `SELECT atcoder_id IS NULL FROM accounts WHERE id = $1`, userID.UUID()).Scan(&atCoderIDIsNull); err != nil || !atCoderIDIsNull {
		t.Fatalf("cleared AtCoder ID must be SQL NULL: isNull = %v, error = %v", atCoderIDIsNull, err)
	}
}
