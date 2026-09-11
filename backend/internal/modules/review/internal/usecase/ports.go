// Package usecaseは、review moduleのユースケースと必要なportを定義する。
package usecase

import (
	"context"
	"time"

	"github.com/gesop0n/spaco/backend/internal/modules/catalog"
	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/domain/scheduler"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

// PostgreSQL固有の型は、いずれのinterfaceにも持ち込まない。

// Clockは、現在時刻を返す関数である。テストで時刻を固定できるよう注入する。
type Clock func() time.Time

// IProblemFinderは、catalog moduleから問題の基本情報を受け取るinterfaceである。
type IProblemFinder interface {
	FindProblems(context.Context, []catalog.ProblemID) ([]catalog.Problem, error)
}

// IProblemEnsurerは、問題URLから読み取った問題をcatalog moduleに登録させるinterfaceである。
type IProblemEnsurer interface {
	EnsureProblem(context.Context, catalog.ProblemRef) (catalog.Problem, error)
}

// ITimeZoneReaderは、account moduleからユーザーのタイムゾーンを受け取るinterfaceである。
type ITimeZoneReader interface {
	TimeZone(context.Context, identifier.UserID) (string, error)
}

type ILatestReviewLogFinder interface {
	FindLatestReviewLogs(context.Context, identifier.UserID, []domain.ReviewItemID) ([]domain.ReviewLog, error)
}

type IReviewItemsByProblemFinder interface {
	FindReviewItemsByProblemIDs(context.Context, identifier.UserID, []catalog.ProblemID) ([]domain.ReviewItem, error)
}

type IRegisterProblemsRepository interface {
	// CreateReviewItemsは、未登録の復習対象だけを保存し、保存した復習対象を返す。
	CreateReviewItems(context.Context, []domain.ReviewItem) ([]domain.ReviewItem, error)
	IReviewItemsByProblemFinder
	ILatestReviewLogFinder
}

type IListRegisteredProblemIDsRepository interface {
	ListRegisteredProblemIDs(context.Context, identifier.UserID, []catalog.ProblemID) ([]catalog.ProblemID, error)
}

type IGetTodayReviewsRepository interface {
	ListDueReviewItems(context.Context, identifier.UserID, scheduler.Day) ([]domain.ReviewItem, error)
	ListUpcomingReviewItems(context.Context, identifier.UserID, scheduler.Day, int) ([]domain.ReviewItem, error)
	CountReviewLogsPerformedBetween(context.Context, identifier.UserID, time.Time, time.Time) (int, error)
	ILatestReviewLogFinder
}

type IListReviewItemsRepository interface {
	ListReviewItems(context.Context, identifier.UserID) ([]domain.ReviewItem, error)
	ListReviewLogs(context.Context, identifier.UserID) ([]domain.ReviewLog, error)
}

type IGetReviewItemRepository interface {
	IReviewItemsByProblemFinder
	ILatestReviewLogFinder
}

type IPreviewReviewScheduleRepository interface {
	FindReviewItemByID(context.Context, identifier.UserID, domain.ReviewItemID) (domain.ReviewItem, error)
}

// RecordReviewFuncは、ロックした復習対象から、保存する状態と記録を決める関数である。
type RecordReviewFunc func(domain.ReviewItem) (domain.ReviewItem, domain.ReviewLog, error)

// RecordedReviewは、今回保存した記録、または同じrequest_idで保存済みだった記録である。
type RecordedReview struct {
	Item     domain.ReviewItem
	Log      domain.ReviewLog
	Replayed bool
}

type IRecordReviewRepository interface {
	// RecordReviewは、復習対象をロックしてから、同じrequest_idの記録があればそれを返し、
	// なければrecordで決めた状態と記録を1つのtransactionで保存する。
	RecordReview(
		context.Context,
		identifier.UserID,
		domain.ReviewItemID,
		domain.RequestID,
		RecordReviewFunc,
	) (RecordedReview, error)
	ILatestReviewLogFinder
}

type IChangeReviewItemPauseRepository interface {
	FindReviewItemByID(context.Context, identifier.UserID, domain.ReviewItemID) (domain.ReviewItem, error)
	UpdateReviewItemPausedAt(context.Context, domain.ReviewItem) (domain.ReviewItem, error)
	ILatestReviewLogFinder
}
