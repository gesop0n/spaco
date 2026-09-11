package domain

import "github.com/gesop0n/spaco/backend/internal/modules/catalog"

// Snapshotは、AtCoder Problemsから一度に取得したコンテスト・問題・対応表である。
type Snapshot struct {
	Contests        []Contest
	Problems        []catalog.Problem
	ContestProblems []ContestProblem
	// Skippedは、取得したデータのうち、形式が不正で取り込めなかった行数である。
	Skipped int
}

// Normalizedは、保存できる形に整えたSnapshotを返す。
// IDが重複する行は最初の行だけを残し、対応表からは同じSnapshotに無いコンテスト・問題を指す行を除く。
func (s Snapshot) Normalized() Snapshot {
	normalized := Snapshot{Skipped: s.Skipped}

	contestIDs := make(map[catalog.ContestID]struct{}, len(s.Contests))
	for _, contest := range s.Contests {
		if _, exists := contestIDs[contest.ID()]; exists {
			normalized.Skipped++
			continue
		}
		contestIDs[contest.ID()] = struct{}{}
		normalized.Contests = append(normalized.Contests, contest)
	}

	problemIDs := make(map[catalog.ProblemID]struct{}, len(s.Problems))
	for _, problem := range s.Problems {
		if _, exists := problemIDs[problem.ID()]; exists {
			normalized.Skipped++
			continue
		}
		problemIDs[problem.ID()] = struct{}{}
		normalized.Problems = append(normalized.Problems, problem)
	}

	type pairKey struct {
		contestID catalog.ContestID
		problemID catalog.ProblemID
	}
	pairs := make(map[pairKey]struct{}, len(s.ContestProblems))
	for _, pair := range s.ContestProblems {
		key := pairKey{contestID: pair.ContestID(), problemID: pair.ProblemID()}
		_, duplicated := pairs[key]
		_, contestExists := contestIDs[pair.ContestID()]
		_, problemExists := problemIDs[pair.ProblemID()]
		if duplicated || !contestExists || !problemExists {
			normalized.Skipped++
			continue
		}
		pairs[key] = struct{}{}
		normalized.ContestProblems = append(normalized.ContestProblems, pair)
	}
	return normalized
}
