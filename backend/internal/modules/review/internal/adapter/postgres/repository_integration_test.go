package postgres

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gesop0n/spaco/backend/internal/modules/catalog"
	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/domain/scheduler"
	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/usecase"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

// TEST_DATABASE_URLが設定された場合だけ、migration適用済みPostgreSQLで確認する。
func TestRepositoryIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}
	// 本番と同じく、named prepared statementを使わないexec modeで確認する。
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("NewWithConfig() error = %v", err)
	}
	defer pool.Close()
	repository, err := NewRepository(pool)
	if err != nil {
		t.Fatalf("NewRepository() error = %v", err)
	}

	userID := identifier.NewUserID()
	otherUserID := identifier.NewUserID()
	// このtestが作成したユーザーの復習対象だけを削除する。
	defer func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		for _, id := range []identifier.UserID{userID, otherUserID} {
			_, _ = pool.Exec(cleanupContext, `DELETE FROM review_items WHERE user_id = $1`, id.UUID())
		}
	}()

	tokyo, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatalf("LoadLocation() error = %v", err)
	}
	registeredAt := time.Date(2026, 9, 1, 3, 0, 0, 123456789, time.UTC)
	problemA, problemB := mustProblemID(t, "abc350_a"), mustProblemID(t, "abc350_b")
	itemA := mustRegister(t, userID, problemA, registeredAt, tokyo)
	itemB := mustRegister(t, userID, problemB, registeredAt.Add(time.Minute), tokyo)

	created, err := repository.CreateReviewItems(ctx, []domain.ReviewItem{itemA, itemB})
	if err != nil {
		t.Fatalf("CreateReviewItems() error = %v", err)
	}
	if len(created) != 2 {
		t.Fatalf("created = %d, want 2", len(created))
	}
	for _, item := range created {
		if !item.RegisteredAt().Equal(itemA.RegisteredAt()) && !item.RegisteredAt().Equal(itemB.RegisteredAt()) {
			t.Fatalf("stored registered at = %v", item.RegisteredAt())
		}
		if item.DueOn() != itemA.DueOn() {
			t.Fatalf("stored due = %s, want %s", item.DueOn(), itemA.DueOn())
		}
	}

	// 同じユーザーの同じ問題は登録しないが、別のユーザーは登録できる。
	duplicated, err := repository.CreateReviewItems(ctx, []domain.ReviewItem{mustRegister(t, userID, problemA, registeredAt, tokyo)})
	if err != nil || len(duplicated) != 0 {
		t.Fatalf("duplicated CreateReviewItems() = %d, %v", len(duplicated), err)
	}
	others, err := repository.CreateReviewItems(ctx, []domain.ReviewItem{mustRegister(t, otherUserID, problemA, registeredAt, tokyo)})
	if err != nil || len(others) != 1 {
		t.Fatalf("other user's CreateReviewItems() = %d, %v", len(others), err)
	}

	registered, err := repository.ListRegisteredProblemIDs(ctx, userID, []catalog.ProblemID{problemA, mustProblemID(t, "abc350_c")})
	if err != nil || len(registered) != 1 || registered[0] != problemA {
		t.Fatalf("ListRegisteredProblemIDs() = %v, %v", registered, err)
	}

	// 二重送信を並行して送っても、記録は1件だけで、次回予定日は一度しか進まない。
	requestID := mustRequestID(t)
	performedAt := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	record := recordFunc(requestID, performedAt, tokyo)
	const requestCount = 8
	results := make(chan usecase.RecordedReview, requestCount)
	errorsChannel := make(chan error, requestCount)
	var waitGroup sync.WaitGroup
	for range requestCount {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			recorded, err := repository.RecordReview(ctx, userID, itemA.ID(), requestID, record)
			if err != nil {
				errorsChannel <- err
				return
			}
			results <- recorded
		}()
	}
	waitGroup.Wait()
	close(results)
	close(errorsChannel)
	for err := range errorsChannel {
		t.Fatalf("RecordReview() error = %v", err)
	}
	var logID domain.ReviewLogID
	replayed := 0
	for result := range results {
		if logID.IsZero() {
			logID = result.Log.ID()
		}
		if result.Log.ID() != logID || result.Item.Reps() != 1 {
			t.Fatalf("recorded result = %+v", result)
		}
		if result.Replayed {
			replayed++
		}
	}
	if replayed != requestCount-1 {
		t.Fatalf("replayed = %d, want %d", replayed, requestCount-1)
	}

	if _, err := repository.RecordReview(ctx, userID, itemB.ID(), requestID, recordFunc(requestID, performedAt, tokyo)); !errors.Is(err, usecase.ErrRequestIDConflict) {
		t.Fatalf("reused request id error = %v, want ErrRequestIDConflict", err)
	}
	if _, err := repository.RecordReview(ctx, otherUserID, itemA.ID(), mustRequestID(t), recordFunc(requestID, performedAt, tokyo)); !errors.Is(err, usecase.ErrReviewItemNotFound) {
		t.Fatalf("other user's RecordReview() error = %v, want ErrReviewItemNotFound", err)
	}

	recordedA, err := repository.FindReviewItemByID(ctx, userID, itemA.ID())
	if err != nil {
		t.Fatalf("FindReviewItemByID() error = %v", err)
	}
	if recordedA.State() != scheduler.StateReview || recordedA.Reps() != 1 {
		t.Fatalf("recorded item = %+v", recordedA)
	}
	if lastReviewedAt, ok := recordedA.LastReviewedAt(); !ok || !lastReviewedAt.Equal(performedAt) {
		t.Fatalf("LastReviewedAt() = %v, %v", lastReviewedAt, ok)
	}
	if _, ok := recordedA.Memory(); !ok {
		t.Fatal("recorded item must have memory state")
	}

	latest, err := repository.FindLatestReviewLogs(ctx, userID, []domain.ReviewItemID{itemA.ID(), itemB.ID()})
	if err != nil || len(latest) != 1 || latest[0].ID() != logID || latest[0].NextDueOn() != recordedA.DueOn() {
		t.Fatalf("FindLatestReviewLogs() = %+v, %v", latest, err)
	}
	logs, err := repository.ListReviewLogs(ctx, userID)
	if err != nil || len(logs) != 1 {
		t.Fatalf("ListReviewLogs() = %d, %v", len(logs), err)
	}
	for _, test := range []struct {
		from  time.Time
		until time.Time
		want  int
	}{
		{from: performedAt, until: performedAt.Add(time.Hour), want: 1},
		{from: performedAt.Add(-time.Hour), until: performedAt, want: 0},
	} {
		count, err := repository.CountReviewLogsPerformedBetween(ctx, userID, test.from, test.until)
		if err != nil || count != test.want {
			t.Fatalf("CountReviewLogsPerformedBetween(%v, %v) = %d, %v; want %d", test.from, test.until, count, err, test.want)
		}
	}

	// 予定日が古い順に並び、一時停止中は今日の復習から外れる。
	due, err := repository.ListDueReviewItems(ctx, userID, recordedA.DueOn())
	if err != nil || len(due) != 2 || due[0].ID() != itemB.ID() || due[1].ID() != itemA.ID() {
		t.Fatalf("ListDueReviewItems() = %+v, %v", due, err)
	}
	upcoming, err := repository.ListUpcomingReviewItems(ctx, userID, itemB.DueOn(), 3)
	if err != nil || len(upcoming) != 1 || upcoming[0].ID() != itemA.ID() {
		t.Fatalf("ListUpcomingReviewItems() = %+v, %v", upcoming, err)
	}
	paused, err := repository.UpdateReviewItemPausedAt(ctx, itemB.Pause(performedAt))
	if err != nil || !paused.Paused() {
		t.Fatalf("UpdateReviewItemPausedAt(pause) = %+v, %v", paused, err)
	}
	due, err = repository.ListDueReviewItems(ctx, userID, recordedA.DueOn())
	if err != nil || len(due) != 1 || due[0].ID() != itemA.ID() {
		t.Fatalf("ListDueReviewItems() after pause = %+v, %v", due, err)
	}
	resumed, err := repository.UpdateReviewItemPausedAt(ctx, paused.Resume())
	if err != nil || resumed.Paused() {
		t.Fatalf("UpdateReviewItemPausedAt(resume) = %+v, %v", resumed, err)
	}

	found, err := repository.FindReviewItemsByProblemIDs(ctx, userID, []catalog.ProblemID{problemB})
	if err != nil || len(found) != 1 || found[0].ID() != itemB.ID() {
		t.Fatalf("FindReviewItemsByProblemIDs() = %+v, %v", found, err)
	}
	items, err := repository.ListReviewItems(ctx, userID)
	if err != nil || len(items) != 2 || items[0].ID() != itemB.ID() {
		t.Fatalf("ListReviewItems() = %+v, %v; want newest registration first", items, err)
	}
}

func recordFunc(requestID domain.RequestID, performedAt time.Time, location *time.Location) usecase.RecordReviewFunc {
	return func(item domain.ReviewItem) (domain.ReviewItem, domain.ReviewLog, error) {
		outcome, err := domain.NewOutcome(domain.ResultIndependent, domain.DifficultyGood)
		if err != nil {
			return domain.ReviewItem{}, domain.ReviewLog{}, err
		}
		return item.Record(domain.NewReviewLogID(), domain.ReviewInput{
			RequestID:   requestID,
			Outcome:     outcome,
			PerformedAt: performedAt,
			Note:        "自力で解けた",
		}, performedAt, location, scheduler.NewScheduler())
	}
}

func mustRegister(
	t *testing.T,
	userID identifier.UserID,
	problemID catalog.ProblemID,
	registeredAt time.Time,
	location *time.Location,
) domain.ReviewItem {
	t.Helper()
	item, err := domain.RegisterReviewItem(domain.NewReviewItemID(), userID, problemID, "解説を読んだ", registeredAt, location)
	if err != nil {
		t.Fatalf("RegisterReviewItem() error = %v", err)
	}
	return item
}

func mustProblemID(t *testing.T, value string) catalog.ProblemID {
	t.Helper()
	id, err := catalog.ParseProblemID(value)
	if err != nil {
		t.Fatalf("ParseProblemID() error = %v", err)
	}
	return id
}

func mustRequestID(t *testing.T) domain.RequestID {
	t.Helper()
	id, err := domain.ParseRequestID(uuid.NewString())
	if err != nil {
		t.Fatalf("ParseRequestID() error = %v", err)
	}
	return id
}
