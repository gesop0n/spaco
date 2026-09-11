package domain

import (
	"fmt"

	"github.com/gesop0n/spaco/backend/internal/modules/review/internal/domain/scheduler"
)

// Resultは、再挑戦の結果である。
type Result string

const (
	// ResultIndependentは、解説・ヒントを使わずにACできた結果である。
	ResultIndependent Result = "independent"
	// ResultAssistedは、解説・ヒントを見てACした結果である。
	ResultAssisted Result = "assisted"
	// ResultUnsolvedは、ACできなかった結果である。
	ResultUnsolved Result = "unsolved"
)

// Difficultyは、自力でACできたときの手応えである。
type Difficulty string

const (
	DifficultyHard Difficulty = "hard"
	DifficultyGood Difficulty = "good"
	DifficultyEasy Difficulty = "easy"
)

// Outcomeは、再挑戦の結果と手応えの組で、FSRSへ渡す評価を決める。
type Outcome struct {
	result Result
	// difficultyは、ResultIndependentのときだけ設定する。
	difficulty Difficulty
}

// NewOutcomeは、結果と手応えの組を検証する。自力ACで手応えが未指定なら、普通に解けたとして扱う。
func NewOutcome(result Result, difficulty Difficulty) (Outcome, error) {
	switch result {
	case ResultIndependent:
		switch difficulty {
		case "":
			difficulty = DifficultyGood
		case DifficultyHard, DifficultyGood, DifficultyEasy:
		default:
			return Outcome{}, fmt.Errorf("%w: unknown difficulty %q", ErrInvalidOutcome, difficulty)
		}
	case ResultAssisted, ResultUnsolved:
		if difficulty != "" {
			return Outcome{}, fmt.Errorf("%w: difficulty is only for independent AC", ErrInvalidOutcome)
		}
	default:
		return Outcome{}, fmt.Errorf("%w: unknown result %q", ErrInvalidOutcome, result)
	}
	return Outcome{result: result, difficulty: difficulty}, nil
}

// OutcomeFromRatingは、保存済みの結果とFSRSの評価から、手応えを含むOutcomeを復元する。
func OutcomeFromRating(result Result, rating scheduler.Rating) (Outcome, error) {
	if result != ResultIndependent {
		if rating != scheduler.Again {
			return Outcome{}, fmt.Errorf("%w: %s must be rated again", ErrInvalidOutcome, result)
		}
		return NewOutcome(result, "")
	}
	switch rating {
	case scheduler.Hard:
		return NewOutcome(result, DifficultyHard)
	case scheduler.Good:
		return NewOutcome(result, DifficultyGood)
	case scheduler.Easy:
		return NewOutcome(result, DifficultyEasy)
	default:
		return Outcome{}, fmt.Errorf("%w: independent AC must be rated hard, good, or easy", ErrInvalidOutcome)
	}
}

// Outcomesは、次回予定日のプレビューで示す5通りの結果である。
func Outcomes() []Outcome {
	return []Outcome{
		{result: ResultUnsolved},
		{result: ResultAssisted},
		{result: ResultIndependent, difficulty: DifficultyHard},
		{result: ResultIndependent, difficulty: DifficultyGood},
		{result: ResultIndependent, difficulty: DifficultyEasy},
	}
}

func (o Outcome) Result() Result { return o.result }

// Difficultyは、自力ACのときだけ手応えを返す。
func (o Outcome) Difficulty() (Difficulty, bool) { return o.difficulty, o.difficulty != "" }

// Ratingは、FSRSへ渡す評価を返す。
// 解説・ヒントを見たACは、自力で思い出せた成功として扱わず、ACできなかった場合と同じAgainにする。
func (o Outcome) Rating() scheduler.Rating {
	switch o.difficulty {
	case DifficultyHard:
		return scheduler.Hard
	case DifficultyGood:
		return scheduler.Good
	case DifficultyEasy:
		return scheduler.Easy
	default:
		return scheduler.Again
	}
}
