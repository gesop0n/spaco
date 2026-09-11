// Package postgresは、review usecaseが要求するrepositoryをPostgreSQLで実装する。
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gesop0n/spaco/backend/internal/modules/catalog"
	reviewsqlc "github.com/gesop0n/spaco/backend/internal/modules/review/internal/adapter/postgres/sqlc"
	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/domain/scheduler"
	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/usecase"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

const (
	uniqueViolationCode     = "23505"
	requestUniqueConstraint = "review_logs_request_unique"
)

type Repository struct {
	pool    *pgxpool.Pool
	queries *reviewsqlc.Queries
}

var (
	_ usecase.IRegisterProblemsRepository         = (*Repository)(nil)
	_ usecase.IListRegisteredProblemIDsRepository = (*Repository)(nil)
	_ usecase.IGetTodayReviewsRepository          = (*Repository)(nil)
	_ usecase.IListReviewItemsRepository          = (*Repository)(nil)
	_ usecase.IGetReviewItemRepository            = (*Repository)(nil)
	_ usecase.IPreviewReviewScheduleRepository    = (*Repository)(nil)
	_ usecase.IRecordReviewRepository             = (*Repository)(nil)
	_ usecase.IChangeReviewItemPauseRepository    = (*Repository)(nil)
)

func NewRepository(pool *pgxpool.Pool) (*Repository, error) {
	if pool == nil {
		return nil, errors.New("create review repository: pool is required")
	}
	return &Repository{pool: pool, queries: reviewsqlc.New(pool)}, nil
}

// CreateReviewItemsは、同じユーザーの同じ問題が未登録の復習対象だけを保存する。
func (r *Repository) CreateReviewItems(ctx context.Context, items []domain.ReviewItem) ([]domain.ReviewItem, error) {
	if len(items) == 0 {
		return nil, nil
	}
	params := reviewsqlc.CreateReviewItemsParams{
		Ids:               make([]string, len(items)),
		UserIds:           make([]string, len(items)),
		ProblemIds:        make([]string, len(items)),
		RegistrationNotes: make([]string, len(items)),
		RegisteredAts:     make([]time.Time, len(items)),
		DueOns:            make([]pgtype.Date, len(items)),
	}
	for index, item := range items {
		params.Ids[index] = item.ID().String()
		params.UserIds[index] = item.UserID().String()
		params.ProblemIds[index] = item.ProblemID().String()
		params.RegistrationNotes[index] = item.RegistrationNote()
		params.RegisteredAts[index] = item.RegisteredAt()
		params.DueOns[index] = dateOf(item.DueOn())
	}
	rows, err := r.queries.CreateReviewItems(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create review items: %w", err)
	}
	return rehydrateReviewItems(rows)
}

func (r *Repository) FindReviewItemsByProblemIDs(
	ctx context.Context,
	userID identifier.UserID,
	problemIDs []catalog.ProblemID,
) ([]domain.ReviewItem, error) {
	rows, err := r.queries.FindReviewItemsByProblemIDs(ctx, reviewsqlc.FindReviewItemsByProblemIDsParams{
		UserID:     userID.UUID(),
		ProblemIds: problemIDStrings(problemIDs),
	})
	if err != nil {
		return nil, fmt.Errorf("find review items by problem ids: %w", err)
	}
	return rehydrateReviewItems(rows)
}

func (r *Repository) FindReviewItemByID(
	ctx context.Context,
	userID identifier.UserID,
	reviewItemID domain.ReviewItemID,
) (domain.ReviewItem, error) {
	row, err := r.queries.FindReviewItemByID(ctx, reviewsqlc.FindReviewItemByIDParams{
		UserID: userID.UUID(),
		ID:     reviewItemID.UUID(),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ReviewItem{}, usecase.ErrReviewItemNotFound
	}
	if err != nil {
		return domain.ReviewItem{}, fmt.Errorf("find review item by id: %w", err)
	}
	return rehydrateReviewItem(row)
}

func (r *Repository) ListRegisteredProblemIDs(
	ctx context.Context,
	userID identifier.UserID,
	problemIDs []catalog.ProblemID,
) ([]catalog.ProblemID, error) {
	values, err := r.queries.ListRegisteredProblemIDs(ctx, reviewsqlc.ListRegisteredProblemIDsParams{
		UserID:     userID.UUID(),
		ProblemIds: problemIDStrings(problemIDs),
	})
	if err != nil {
		return nil, fmt.Errorf("list registered problem ids: %w", err)
	}
	registered := make([]catalog.ProblemID, 0, len(values))
	for _, value := range values {
		problemID, err := catalog.ParseProblemID(value)
		if err != nil {
			return nil, fmt.Errorf("parse stored problem id: %w", err)
		}
		registered = append(registered, problemID)
	}
	return registered, nil
}

func (r *Repository) ListDueReviewItems(
	ctx context.Context,
	userID identifier.UserID,
	today scheduler.Day,
) ([]domain.ReviewItem, error) {
	rows, err := r.queries.ListDueReviewItems(ctx, reviewsqlc.ListDueReviewItemsParams{
		UserID: userID.UUID(),
		Today:  dateOf(today),
	})
	if err != nil {
		return nil, fmt.Errorf("list due review items: %w", err)
	}
	return rehydrateReviewItems(rows)
}

func (r *Repository) ListUpcomingReviewItems(
	ctx context.Context,
	userID identifier.UserID,
	today scheduler.Day,
	limit int,
) ([]domain.ReviewItem, error) {
	rows, err := r.queries.ListUpcomingReviewItems(ctx, reviewsqlc.ListUpcomingReviewItemsParams{
		UserID:   userID.UUID(),
		Today:    dateOf(today),
		MaxItems: int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list upcoming review items: %w", err)
	}
	return rehydrateReviewItems(rows)
}

func (r *Repository) CountReviewLogsPerformedBetween(
	ctx context.Context,
	userID identifier.UserID,
	from time.Time,
	until time.Time,
) (int, error) {
	count, err := r.queries.CountReviewLogsPerformedBetween(ctx, reviewsqlc.CountReviewLogsPerformedBetweenParams{
		UserID:         userID.UUID(),
		PerformedFrom:  from,
		PerformedUntil: until,
	})
	if err != nil {
		return 0, fmt.Errorf("count review logs: %w", err)
	}
	return int(count), nil
}

func (r *Repository) ListReviewItems(ctx context.Context, userID identifier.UserID) ([]domain.ReviewItem, error) {
	rows, err := r.queries.ListReviewItems(ctx, userID.UUID())
	if err != nil {
		return nil, fmt.Errorf("list review items: %w", err)
	}
	return rehydrateReviewItems(rows)
}

func (r *Repository) ListReviewLogs(ctx context.Context, userID identifier.UserID) ([]domain.ReviewLog, error) {
	rows, err := r.queries.ListReviewLogs(ctx, userID.UUID())
	if err != nil {
		return nil, fmt.Errorf("list review logs: %w", err)
	}
	return rehydrateReviewLogs(rows)
}

func (r *Repository) FindLatestReviewLogs(
	ctx context.Context,
	userID identifier.UserID,
	reviewItemIDs []domain.ReviewItemID,
) ([]domain.ReviewLog, error) {
	if len(reviewItemIDs) == 0 {
		return nil, nil
	}
	ids := make([]string, len(reviewItemIDs))
	for index, id := range reviewItemIDs {
		ids[index] = id.String()
	}
	rows, err := r.queries.FindLatestReviewLogs(ctx, reviewsqlc.FindLatestReviewLogsParams{
		UserID:        userID.UUID(),
		ReviewItemIds: ids,
	})
	if err != nil {
		return nil, fmt.Errorf("find latest review logs: %w", err)
	}
	return rehydrateReviewLogs(rows)
}

func (r *Repository) UpdateReviewItemPausedAt(ctx context.Context, item domain.ReviewItem) (domain.ReviewItem, error) {
	params := reviewsqlc.UpdateReviewItemPausedAtParams{
		UserID: item.UserID().UUID(),
		ID:     item.ID().UUID(),
	}
	if pausedAt, ok := item.PausedAt(); ok {
		params.PausedAt = &pausedAt
	}
	row, err := r.queries.UpdateReviewItemPausedAt(ctx, params)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ReviewItem{}, usecase.ErrReviewItemNotFound
	}
	if err != nil {
		return domain.ReviewItem{}, fmt.Errorf("update review item paused at: %w", err)
	}
	return rehydrateReviewItem(row)
}

// RecordReviewは、復習対象の行をロックしてから記録する。
// 同じ復習対象への記録を直列にし、保存ボタンの二重送信で次回予定日を二度進めないようにする。
func (r *Repository) RecordReview(
	ctx context.Context,
	userID identifier.UserID,
	reviewItemID domain.ReviewItemID,
	requestID domain.RequestID,
	record usecase.RecordReviewFunc,
) (usecase.RecordedReview, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return usecase.RecordedReview{}, fmt.Errorf("begin record review transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	queries := r.queries.WithTx(tx)

	row, err := queries.LockReviewItemByID(ctx, reviewsqlc.LockReviewItemByIDParams{
		UserID: userID.UUID(),
		ID:     reviewItemID.UUID(),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return usecase.RecordedReview{}, usecase.ErrReviewItemNotFound
	}
	if err != nil {
		return usecase.RecordedReview{}, fmt.Errorf("lock review item: %w", err)
	}
	item, err := rehydrateReviewItem(row)
	if err != nil {
		return usecase.RecordedReview{}, err
	}

	// ロックを取った後に確認するため、先に記録した同じrequest_idの結果を必ず読める。
	existingRow, err := queries.FindReviewLogByRequestID(ctx, reviewsqlc.FindReviewLogByRequestIDParams{
		UserID:    userID.UUID(),
		RequestID: requestID.UUID(),
	})
	switch {
	case err == nil:
		existing, err := rehydrateReviewLog(existingRow)
		if err != nil {
			return usecase.RecordedReview{}, err
		}
		if existing.ReviewItemID() != reviewItemID {
			return usecase.RecordedReview{}, usecase.ErrRequestIDConflict
		}
		return usecase.RecordedReview{Item: item, Log: existing, Replayed: true}, nil
	case !errors.Is(err, pgx.ErrNoRows):
		return usecase.RecordedReview{}, fmt.Errorf("find review log by request id: %w", err)
	}

	next, log, err := record(item)
	if err != nil {
		return usecase.RecordedReview{}, err
	}
	updatedRow, err := queries.UpdateReviewItemSchedule(ctx, scheduleParams(next))
	if err != nil {
		return usecase.RecordedReview{}, fmt.Errorf("update review item schedule: %w", err)
	}
	createdRow, err := queries.CreateReviewLog(ctx, reviewLogParams(log))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode && pgErr.ConstraintName == requestUniqueConstraint {
			// 同じrequest_idを、別の復習対象への記録に並行して使った場合。
			return usecase.RecordedReview{}, usecase.ErrRequestIDConflict
		}
		return usecase.RecordedReview{}, fmt.Errorf("create review log: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return usecase.RecordedReview{}, fmt.Errorf("commit record review: %w", err)
	}

	updated, err := rehydrateReviewItem(updatedRow)
	if err != nil {
		return usecase.RecordedReview{}, err
	}
	created, err := rehydrateReviewLog(createdRow)
	if err != nil {
		return usecase.RecordedReview{}, err
	}
	return usecase.RecordedReview{Item: updated, Log: created}, nil
}

func scheduleParams(item domain.ReviewItem) reviewsqlc.UpdateReviewItemScheduleParams {
	params := reviewsqlc.UpdateReviewItemScheduleParams{
		State:        string(item.State()),
		DueOn:        dateOf(item.DueOn()),
		IntervalDays: int32(item.IntervalDays()),
		Reps:         int32(item.Reps()),
		Lapses:       int32(item.Lapses()),
		UserID:       item.UserID().UUID(),
		ID:           item.ID().UUID(),
	}
	if memory, ok := item.Memory(); ok {
		params.Stability = &memory.Stability
		params.Difficulty = &memory.Difficulty
	}
	if lastReviewedAt, ok := item.LastReviewedAt(); ok {
		params.LastReviewedAt = &lastReviewedAt
	}
	return params
}

func reviewLogParams(log domain.ReviewLog) reviewsqlc.CreateReviewLogParams {
	memory := log.Memory()
	return reviewsqlc.CreateReviewLogParams{
		ID:               log.ID().UUID(),
		ReviewItemID:     log.ReviewItemID().UUID(),
		UserID:           log.UserID().UUID(),
		RequestID:        log.RequestID().UUID(),
		Result:           string(log.Outcome().Result()),
		Rating:           int16(log.Outcome().Rating()),
		PerformedAt:      log.PerformedAt(),
		Note:             log.Note(),
		StateBefore:      string(log.StateBefore()),
		ElapsedDays:      int32(log.ElapsedDays()),
		LastIntervalDays: int32(log.LastIntervalDays()),
		IntervalDays:     int32(log.IntervalDays()),
		Stability:        memory.Stability,
		Difficulty:       memory.Difficulty,
		NextDueOn:        dateOf(log.NextDueOn()),
	}
}

func rehydrateReviewItems(rows []reviewsqlc.ReviewItem) ([]domain.ReviewItem, error) {
	items := make([]domain.ReviewItem, 0, len(rows))
	for _, row := range rows {
		item, err := rehydrateReviewItem(row)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func rehydrateReviewItem(row reviewsqlc.ReviewItem) (domain.ReviewItem, error) {
	id, err := domain.ParseReviewItemID(row.ID.String())
	if err != nil {
		return domain.ReviewItem{}, fmt.Errorf("parse stored review item id: %w", err)
	}
	userID, err := identifier.ParseUserID(row.UserID.String())
	if err != nil {
		return domain.ReviewItem{}, fmt.Errorf("parse stored user id: %w", err)
	}
	problemID, err := catalog.ParseProblemID(row.ProblemID)
	if err != nil {
		return domain.ReviewItem{}, fmt.Errorf("parse stored problem id: %w", err)
	}
	dueOn, err := dayOf(row.DueOn)
	if err != nil {
		return domain.ReviewItem{}, fmt.Errorf("parse stored due date: %w", err)
	}
	var memory *scheduler.MemoryState
	if row.Stability != nil || row.Difficulty != nil {
		if row.Stability == nil || row.Difficulty == nil {
			return domain.ReviewItem{}, errors.New("stored memory state must have both stability and difficulty")
		}
		memory = &scheduler.MemoryState{Stability: *row.Stability, Difficulty: *row.Difficulty}
	}
	item, err := domain.RehydrateReviewItem(domain.StoredReviewItem{
		ID:               id,
		UserID:           userID,
		ProblemID:        problemID,
		RegistrationNote: row.RegistrationNote,
		RegisteredAt:     row.RegisteredAt.UTC(),
		PausedAt:         utcPointer(row.PausedAt),
		State:            scheduler.State(row.State),
		DueOn:            dueOn,
		IntervalDays:     int(row.IntervalDays),
		Memory:           memory,
		Reps:             int(row.Reps),
		Lapses:           int(row.Lapses),
		LastReviewedAt:   utcPointer(row.LastReviewedAt),
	})
	if err != nil {
		return domain.ReviewItem{}, fmt.Errorf("rehydrate stored review item: %w", err)
	}
	return item, nil
}

func rehydrateReviewLogs(rows []reviewsqlc.ReviewLog) ([]domain.ReviewLog, error) {
	logs := make([]domain.ReviewLog, 0, len(rows))
	for _, row := range rows {
		log, err := rehydrateReviewLog(row)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}

func rehydrateReviewLog(row reviewsqlc.ReviewLog) (domain.ReviewLog, error) {
	id, err := domain.ParseReviewLogID(row.ID.String())
	if err != nil {
		return domain.ReviewLog{}, fmt.Errorf("parse stored review log id: %w", err)
	}
	reviewItemID, err := domain.ParseReviewItemID(row.ReviewItemID.String())
	if err != nil {
		return domain.ReviewLog{}, fmt.Errorf("parse stored review item id: %w", err)
	}
	userID, err := identifier.ParseUserID(row.UserID.String())
	if err != nil {
		return domain.ReviewLog{}, fmt.Errorf("parse stored user id: %w", err)
	}
	requestID, err := domain.ParseRequestID(row.RequestID.String())
	if err != nil {
		return domain.ReviewLog{}, fmt.Errorf("parse stored request id: %w", err)
	}
	nextDueOn, err := dayOf(row.NextDueOn)
	if err != nil {
		return domain.ReviewLog{}, fmt.Errorf("parse stored next due date: %w", err)
	}
	log, err := domain.RehydrateReviewLog(domain.StoredReviewLog{
		ID:               id,
		ReviewItemID:     reviewItemID,
		UserID:           userID,
		RequestID:        requestID,
		Result:           domain.Result(row.Result),
		Rating:           scheduler.Rating(row.Rating),
		PerformedAt:      row.PerformedAt.UTC(),
		Note:             row.Note,
		StateBefore:      scheduler.State(row.StateBefore),
		ElapsedDays:      int(row.ElapsedDays),
		LastIntervalDays: int(row.LastIntervalDays),
		IntervalDays:     int(row.IntervalDays),
		Memory:           scheduler.MemoryState{Stability: row.Stability, Difficulty: row.Difficulty},
		NextDueOn:        nextDueOn,
	})
	if err != nil {
		return domain.ReviewLog{}, fmt.Errorf("rehydrate stored review log: %w", err)
	}
	return log, nil
}

func problemIDStrings(problemIDs []catalog.ProblemID) []string {
	values := make([]string, len(problemIDs))
	for index, problemID := range problemIDs {
		values[index] = problemID.String()
	}
	return values
}

// dateOfは、暦日をPostgreSQLのdate型に変換する。
func dateOf(day scheduler.Day) pgtype.Date {
	return pgtype.Date{Time: day.Time(), Valid: true}
}

func dayOf(date pgtype.Date) (scheduler.Day, error) {
	if !date.Valid || date.InfinityModifier != pgtype.Finite {
		return 0, errors.New("date must be a finite value")
	}
	return scheduler.DayOf(date.Time, time.UTC), nil
}

func utcPointer(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	value := t.UTC()
	return &value
}
