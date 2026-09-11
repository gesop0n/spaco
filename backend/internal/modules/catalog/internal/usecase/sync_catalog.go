package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/gesop0n/spaco/backend/internal/modules/catalog"
)

// SyncCatalogは、AtCoder Problemsのコンテスト・問題情報をDBへ取り込むユースケースである。
// 追加と更新だけを行い、AtCoder Problemsから消えた行や、URLから登録した問題は削除しない。
type SyncCatalog struct {
	client     IAtCoderProblemsClient
	repository ISyncCatalogRepository
}

func NewSyncCatalog(client IAtCoderProblemsClient, repository ISyncCatalogRepository) (*SyncCatalog, error) {
	if client == nil {
		return nil, errors.New("create sync catalog use case: client is required")
	}
	if repository == nil {
		return nil, errors.New("create sync catalog use case: repository is required")
	}
	return &SyncCatalog{client: client, repository: repository}, nil
}

func (u *SyncCatalog) Execute(ctx context.Context) (catalog.SyncResult, error) {
	snapshot, err := u.client.FetchSnapshot(ctx)
	if err != nil {
		return catalog.SyncResult{}, fmt.Errorf("sync catalog: fetch: %w", err)
	}
	normalized := snapshot.Normalized()
	// 取得に失敗して空になったデータを、同期できたものとして扱わない。
	if len(normalized.Contests) == 0 || len(normalized.Problems) == 0 {
		return catalog.SyncResult{}, fmt.Errorf("sync catalog: %w", ErrEmptySnapshot)
	}
	if err := u.repository.SaveSnapshot(ctx, normalized); err != nil {
		return catalog.SyncResult{}, fmt.Errorf("sync catalog: save: %w", err)
	}
	return catalog.SyncResult{
		Contests:        len(normalized.Contests),
		Problems:        len(normalized.Problems),
		ContestProblems: len(normalized.ContestProblems),
		Skipped:         normalized.Skipped,
	}, nil
}
