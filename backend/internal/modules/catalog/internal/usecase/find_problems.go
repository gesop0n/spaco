package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/gesop0n/spaco/backend/internal/modules/catalog"
)

// FindProblemsは、他moduleから渡された問題IDの基本情報を返すユースケースである。
type FindProblems struct {
	repository IFindProblemsRepository
}

func NewFindProblems(repository IFindProblemsRepository) (*FindProblems, error) {
	if repository == nil {
		return nil, errors.New("create find problems use case: repository is required")
	}
	return &FindProblems{repository: repository}, nil
}

// Executeは、登録済みの問題だけを返す。重複したIDは1件にまとめる。
func (u *FindProblems) Execute(ctx context.Context, ids []catalog.ProblemID) ([]catalog.Problem, error) {
	unique := make([]catalog.ProblemID, 0, len(ids))
	seen := make(map[catalog.ProblemID]struct{}, len(ids))
	for _, id := range ids {
		if id.IsZero() {
			return nil, fmt.Errorf("find problems: %w", catalog.ErrInvalidProblemID)
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	if len(unique) == 0 {
		return nil, nil
	}

	problems, err := u.repository.FindProblemsByIDs(ctx, unique)
	if err != nil {
		return nil, fmt.Errorf("find problems: %w", err)
	}
	return problems, nil
}
