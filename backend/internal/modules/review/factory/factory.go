// Package factoryは、review module内部のuse caseとadapterを組み立てる。
package factory

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gesop0n/spaco/backend/internal/modules/catalog"
	"github.com/gesop0n/spaco/backend/internal/modules/review"
	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/adapter/postgres"
	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/adapter/rpc"
	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/domain/scheduler"
	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/usecase"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

// ProblemCatalogは、review moduleがcatalog moduleに求める操作である。
type ProblemCatalog interface {
	FindProblems(context.Context, []catalog.ProblemID) ([]catalog.Problem, error)
	EnsureProblem(context.Context, catalog.ProblemRef) (catalog.Problem, error)
}

// TimeZoneReaderは、review moduleがaccount moduleに求める操作である。
type TimeZoneReader interface {
	TimeZone(context.Context, identifier.UserID) (string, error)
}

// Newは、PostgreSQL repository、FSRSのスケジューラ、use case、ConnectRPC handlerを接続する。
func New(pool *pgxpool.Pool, problems ProblemCatalog, timeZones TimeZoneReader) (*review.Module, error) {
	if problems == nil || timeZones == nil {
		return nil, errors.New("create review module: problem catalog and time zone reader are required")
	}
	repository, err := postgres.NewRepository(pool)
	if err != nil {
		return nil, fmt.Errorf("create review module: %w", err)
	}
	clock := usecase.Clock(time.Now)
	reviewScheduler := scheduler.NewScheduler()

	registerProblems, registerProblemsErr := usecase.NewRegisterProblems(repository, problems, timeZones, clock)
	registerProblemByURL, registerProblemByURLErr := usecase.NewRegisterProblemByURL(repository, problems, timeZones, clock)
	listRegisteredProblemIDs, listRegisteredProblemIDsErr := usecase.NewListRegisteredProblemIDs(repository)
	getTodayReviews, getTodayReviewsErr := usecase.NewGetTodayReviews(repository, problems, timeZones, clock)
	listReviewItems, listReviewItemsErr := usecase.NewListReviewItems(repository, problems, timeZones, clock)
	getReviewItem, getReviewItemErr := usecase.NewGetReviewItem(repository, problems, timeZones)
	previewReviewSchedule, previewReviewScheduleErr := usecase.NewPreviewReviewSchedule(repository, timeZones, clock, reviewScheduler)
	recordReview, recordReviewErr := usecase.NewRecordReview(repository, problems, timeZones, clock, reviewScheduler)
	pauseReviewItem, pauseReviewItemErr := usecase.NewPauseReviewItem(repository, problems, timeZones, clock)
	resumeReviewItem, resumeReviewItemErr := usecase.NewResumeReviewItem(repository, problems, timeZones, clock)
	if err := errors.Join(
		registerProblemsErr,
		registerProblemByURLErr,
		listRegisteredProblemIDsErr,
		getTodayReviewsErr,
		listReviewItemsErr,
		getReviewItemErr,
		previewReviewScheduleErr,
		recordReviewErr,
		pauseReviewItemErr,
		resumeReviewItemErr,
	); err != nil {
		return nil, fmt.Errorf("create review module: %w", err)
	}

	handler, err := rpc.NewHandler(rpc.UseCases{
		RegisterProblems:         registerProblems,
		RegisterProblemByURL:     registerProblemByURL,
		ListRegisteredProblemIDs: listRegisteredProblemIDs,
		GetTodayReviews:          getTodayReviews,
		ListReviewItems:          listReviewItems,
		GetReviewItem:            getReviewItem,
		PreviewReviewSchedule:    previewReviewSchedule,
		RecordReview:             recordReview,
		PauseReviewItem:          pauseReviewItem,
		ResumeReviewItem:         resumeReviewItem,
	})
	if err != nil {
		return nil, fmt.Errorf("create review module: %w", err)
	}
	module, err := review.NewModule(handler)
	if err != nil {
		return nil, fmt.Errorf("create review module: %w", err)
	}
	return module, nil
}
