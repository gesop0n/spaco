package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gesop0n/spaco/backend/internal/modules/catalog"
	"github.com/gesop0n/spaco/backend/internal/modules/catalog/internal/domain"
)

// ContestProblemsは、コンテストと、そのコンテストでの問題番号を持つ問題の一覧である。
type ContestProblems struct {
	Contest  domain.Contest
	Problems []catalog.Problem
}

// ListContestProblemsは、コンテストに含まれる問題を返すユースケースである。
type ListContestProblems struct {
	repository IListContestProblemsRepository
}

func NewListContestProblems(repository IListContestProblemsRepository) (*ListContestProblems, error) {
	if repository == nil {
		return nil, errors.New("create list contest problems use case: repository is required")
	}
	return &ListContestProblems{repository: repository}, nil
}

func (u *ListContestProblems) Execute(ctx context.Context, rawContestID string) (ContestProblems, error) {
	contestID, err := catalog.ParseContestID(strings.TrimSpace(rawContestID))
	if err != nil {
		return ContestProblems{}, fmt.Errorf("list contest problems: %w", err)
	}
	contest, err := u.repository.FindContestByID(ctx, contestID)
	if err != nil {
		return ContestProblems{}, fmt.Errorf("list contest problems: %w", err)
	}
	problems, err := u.repository.ListContestProblems(ctx, contestID)
	if err != nil {
		return ContestProblems{}, fmt.Errorf("list contest problems: %w", err)
	}
	return ContestProblems{Contest: contest, Problems: problems}, nil
}
