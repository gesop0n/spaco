package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/gesop0n/spaco/backend/internal/modules/catalog"
)

// EnsureProblemは、問題URLから読み取った問題を返すユースケースである。
// AtCoder Problemsにまだ無い問題も最小限の情報で登録し、後の同期で問題名などを補う。
type EnsureProblem struct {
	repository IEnsureProblemRepository
}

func NewEnsureProblem(repository IEnsureProblemRepository) (*EnsureProblem, error) {
	if repository == nil {
		return nil, errors.New("create ensure problem use case: repository is required")
	}
	return &EnsureProblem{repository: repository}, nil
}

func (u *EnsureProblem) Execute(ctx context.Context, ref catalog.ProblemRef) (catalog.Problem, error) {
	if ref.ProblemID().IsZero() || ref.ContestID().IsZero() {
		return catalog.Problem{}, fmt.Errorf("ensure problem: %w", catalog.ErrInvalidProblemURL)
	}
	problem, err := u.repository.EnsureProblem(ctx, ref)
	if err != nil {
		return catalog.Problem{}, fmt.Errorf("ensure problem: %w", err)
	}
	return problem, nil
}
