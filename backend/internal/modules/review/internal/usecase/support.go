package usecase

import (
	"cmp"
	"context"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"time"
	// 実行環境にタイムゾーンデータが無くても、ユーザーのタイムゾーンで日付を数えられるようにする。
	_ "time/tzdata"

	"github.com/gesop0n/spaco/backend/internal/modules/catalog"
	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/domain"
	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/domain/scheduler"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

// ReviewItemViewは、画面に返す復習対象で、問題の基本情報と最後の記録を含む。
type ReviewItemView struct {
	Item    domain.ReviewItem
	Problem catalog.Problem
	// RegisteredOnは、ユーザーのタイムゾーンでの登録日である。
	RegisteredOn scheduler.Day
	// LastLogは、実施日時が最も新しい記録である。未記録ならnil。
	LastLog *domain.ReviewLog
}

func newReviewItemViews(
	items []domain.ReviewItem,
	problems map[catalog.ProblemID]catalog.Problem,
	latestLogs map[domain.ReviewItemID]domain.ReviewLog,
	location *time.Location,
) ([]ReviewItemView, error) {
	views := make([]ReviewItemView, 0, len(items))
	for _, item := range items {
		problem, ok := problems[item.ProblemID()]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrProblemInfoMissing, item.ProblemID())
		}
		view := ReviewItemView{
			Item:         item,
			Problem:      problem,
			RegisteredOn: scheduler.DayOf(item.RegisteredAt(), location),
		}
		if log, ok := latestLogs[item.ID()]; ok {
			view.LastLog = &log
		}
		views = append(views, view)
	}
	return views, nil
}

// findProblemsForItemsは、復習対象が指す問題の基本情報をcatalog moduleから受け取る。
func findProblemsForItems(
	ctx context.Context,
	finder IProblemFinder,
	items []domain.ReviewItem,
) (map[catalog.ProblemID]catalog.Problem, error) {
	if len(items) == 0 {
		return map[catalog.ProblemID]catalog.Problem{}, nil
	}
	ids := make([]catalog.ProblemID, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ProblemID())
	}
	problems, err := finder.FindProblems(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("find problems: %w", err)
	}
	return problemsByID(problems), nil
}

func problemsByID(problems []catalog.Problem) map[catalog.ProblemID]catalog.Problem {
	byID := make(map[catalog.ProblemID]catalog.Problem, len(problems))
	for _, problem := range problems {
		byID[problem.ID()] = problem
	}
	return byID
}

func findLatestLogs(
	ctx context.Context,
	finder ILatestReviewLogFinder,
	userID identifier.UserID,
	items []domain.ReviewItem,
) (map[domain.ReviewItemID]domain.ReviewLog, error) {
	if len(items) == 0 {
		return map[domain.ReviewItemID]domain.ReviewLog{}, nil
	}
	ids := make([]domain.ReviewItemID, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID())
	}
	logs, err := finder.FindLatestReviewLogs(ctx, userID, ids)
	if err != nil {
		return nil, fmt.Errorf("find latest review logs: %w", err)
	}
	return latestLogsByItem(logs), nil
}

// latestLogsByItemは、実施日時が古い順の記録から、復習対象ごとに最後の記録を選ぶ。
func latestLogsByItem(logs []domain.ReviewLog) map[domain.ReviewItemID]domain.ReviewLog {
	latest := make(map[domain.ReviewItemID]domain.ReviewLog, len(logs))
	for _, log := range logs {
		latest[log.ReviewItemID()] = log
	}
	return latest
}

func loadLocation(
	ctx context.Context,
	reader ITimeZoneReader,
	userID identifier.UserID,
) (*time.Location, string, error) {
	name, err := reader.TimeZone(ctx, userID)
	if err != nil {
		return nil, "", fmt.Errorf("read time zone: %w", err)
	}
	location, err := time.LoadLocation(name)
	if err != nil {
		return nil, "", fmt.Errorf("load time zone %q: %w", name, err)
	}
	return location, name, nil
}

// parseProblemIDsは、問題IDを検証し、指定された順序を保ったまま重複を除く。
func parseProblemIDs(values []string, maxCount int) ([]catalog.ProblemID, error) {
	if len(values) == 0 {
		return nil, ErrNoProblems
	}
	if len(values) > maxCount {
		return nil, fmt.Errorf("%w: up to %d problems", ErrTooManyProblems, maxCount)
	}
	ids := make([]catalog.ProblemID, 0, len(values))
	seen := make(map[catalog.ProblemID]struct{}, len(values))
	for _, value := range values {
		id, err := catalog.ParseProblemID(strings.TrimSpace(value))
		if err != nil {
			return nil, err
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids, nil
}

func orderByProblemIDs(items []domain.ReviewItem, problemIDs []catalog.ProblemID) []domain.ReviewItem {
	positions := make(map[catalog.ProblemID]int, len(problemIDs))
	for index, id := range problemIDs {
		positions[id] = index
	}
	ordered := slices.Clone(items)
	slices.SortStableFunc(ordered, func(a, b domain.ReviewItem) int {
		return cmp.Compare(positions[a.ProblemID()], positions[b.ProblemID()])
	})
	return ordered
}

type dependency struct {
	name  string
	value any
}

func requireDependencies(useCase string, dependencies ...dependency) error {
	for _, dependency := range dependencies {
		if isNil(dependency.value) {
			return fmt.Errorf("create %s use case: %s is required", useCase, dependency.name)
		}
	}
	return nil
}

func isNil(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
