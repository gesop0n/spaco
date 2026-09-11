package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/gesop0n/spaco/backend/internal/modules/catalog"
	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

const maxProblemsPerLookup = 200

// ListRegisteredProblemIDsは、指定した問題のうち登録済みの問題IDを返すユースケースである。
// 問題一覧で「登録済み」を表示するために使う。
type ListRegisteredProblemIDs struct {
	repository IListRegisteredProblemIDsRepository
}

func NewListRegisteredProblemIDs(repository IListRegisteredProblemIDsRepository) (*ListRegisteredProblemIDs, error) {
	if err := requireDependencies(
		"list registered problem ids",
		dependency{name: "repository", value: repository},
	); err != nil {
		return nil, err
	}
	return &ListRegisteredProblemIDs{repository: repository}, nil
}

func (u *ListRegisteredProblemIDs) Execute(
	ctx context.Context,
	userID identifier.UserID,
	rawProblemIDs []string,
) ([]catalog.ProblemID, error) {
	if userID.IsZero() {
		return nil, errors.New("list registered problem ids: user id is required")
	}
	if len(rawProblemIDs) == 0 {
		return []catalog.ProblemID{}, nil
	}
	problemIDs, err := parseProblemIDs(rawProblemIDs, maxProblemsPerLookup)
	if err != nil {
		return nil, fmt.Errorf("list registered problem ids: %w", err)
	}
	registered, err := u.repository.ListRegisteredProblemIDs(ctx, userID, problemIDs)
	if err != nil {
		return nil, fmt.Errorf("list registered problem ids: %w", err)
	}
	return registered, nil
}
