// Package schedulerは、FSRSを有効にしたAnkiと同じ規則で、次の復習日を決める。
// DB、現在時刻の取得、タイムゾーンには依存させず、純粋な計算だけを置く。
package scheduler

import (
	"errors"
	"fmt"
	"math"
)

var (
	ErrInvalidCard   = errors.New("invalid scheduling card")
	ErrInvalidRating = errors.New("invalid rating")
)

// Stateは、Ankiのカードの種類に相当する。
type State string

const (
	// StateNewは、登録後に一度も再挑戦していない状態である。
	StateNew State = "new"
	// StateReviewは、再挑戦を記録し、FSRSの記憶状態を持つ状態である。
	StateReview State = "review"
)

// Ratingは、Ankiの回答ボタンに相当するFSRSへの入力である。
type Rating int

const (
	Again Rating = 1
	Hard  Rating = 2
	Good  Rating = 3
	Easy  Rating = 4
)

// Cardは、スケジューリングに必要な状態で、Ankiのcardに相当する。
type Card struct {
	State State
	// Dueは、次に復習する日である。
	Due Day
	// IntervalDaysは、直前の記録で決めた間隔で、Ankiのscheduled_daysに相当する。
	IntervalDays int
	// Memoryは、StateReviewのときだけ設定する。
	Memory *MemoryState
	// LastReviewedOnは、最後に再挑戦した日で、StateReviewのときだけ使う。
	LastReviewedOn Day
	Reps           int
	Lapses         int
}

// NewCardは、dueに初めて出題する新規カードを返す。
func NewCard(due Day) Card {
	return Card{State: StateNew, Due: due}
}

// Validateは、保存済みの値から復元したカードが矛盾していないかを確認する。
func (c Card) Validate() error {
	switch c.State {
	case StateNew:
		if c.Memory != nil || c.Reps != 0 {
			return fmt.Errorf("%w: new card must not have review history", ErrInvalidCard)
		}
	case StateReview:
		if c.Memory == nil || c.Reps <= 0 {
			return fmt.Errorf("%w: review card requires memory state", ErrInvalidCard)
		}
		if !(c.Memory.Stability > 0) || math.IsInf(c.Memory.Stability, 0) ||
			!(c.Memory.Difficulty >= minimumDifficulty && c.Memory.Difficulty <= maximumDifficulty) {
			return fmt.Errorf("%w: memory state is out of range", ErrInvalidCard)
		}
	default:
		return fmt.Errorf("%w: unknown state %q", ErrInvalidCard, c.State)
	}
	if c.IntervalDays < 0 || c.Lapses < 0 {
		return fmt.Errorf("%w: interval and lapses must not be negative", ErrInvalidCard)
	}
	return nil
}

// ElapsedDaysは、FSRSに渡す、前回の再挑戦からの経過日数を返す。
func ElapsedDays(card Card, today Day) int {
	if card.State == StateNew {
		return 0
	}
	return max(today.DaysSince(card.LastReviewedOn), 0)
}

// Schedulerは、FSRSを有効にしたAnkiのスケジューラを、日単位の復習に合わせて移植したものである。
// AtCoderの問題を同じ日に何度も解き直す運用は想定しないため、Ankiの学習ステップと、
// 1日未満の間隔で再出題するshort-term schedulingは使わず、間隔は最短でも1日とする。
type Scheduler struct {
	parameters       parameters
	desiredRetention float64
	maximumInterval  int
}

// NewSchedulerは、Ankiの既定値で作る。
// FSRS-6の既定パラメータ、目標の記憶率90%、最大間隔36500日を使う。
func NewScheduler() Scheduler {
	return Scheduler{
		parameters:       defaultParameters,
		desiredRetention: 0.9,
		maximumInterval:  36500,
	}
}

// Nextは、todayにratingで回答した後のカードを返す。
// fuzzには、同じカードと復習回数に対して常に同じ値を渡す。
func (s Scheduler) Next(card Card, rating Rating, today Day, fuzz Fuzz) (Card, error) {
	if rating < Again || rating > Easy {
		return Card{}, fmt.Errorf("%w: %d", ErrInvalidRating, rating)
	}
	if err := card.Validate(); err != nil {
		return Card{}, err
	}

	states := s.parameters.nextStates(card.Memory, ElapsedDays(card, today), s.desiredRetention)
	next := card
	var interval int
	switch {
	case card.State == StateNew:
		interval = s.firstInterval(states, rating, fuzz)
	case rating == Again:
		next.Lapses++
		interval = s.lapseInterval(states[Again-1].interval)
	default:
		interval = s.passingReviewIntervals(card.IntervalDays, states, fuzz)[rating-Hard]
	}

	memory := states[rating-1].memory
	next.State = StateReview
	next.Memory = &memory
	next.IntervalDays = interval
	next.Due = today.AddDays(interval)
	next.LastReviewedOn = today
	next.Reps++
	return next, nil
}

// firstIntervalは、学習ステップを持たない新規カードに対するAnkiの間隔である。
func (s Scheduler) firstInterval(states [4]itemState, rating Rating, fuzz Fuzz) int {
	if rating != Easy {
		return fuzz.apply(math.Max(math.Round(states[rating-1].interval), 1), 1, s.maximumInterval)
	}
	// Easyは、Goodで回答した場合より必ず長くする。
	good := fuzz.apply(states[Good-1].interval, 1, s.maximumInterval)
	easy := math.Max(math.Round(states[Easy-1].interval), 1)
	return fuzz.apply(easy, good+1, s.maximumInterval)
}

// lapseIntervalは、復習中にAgainで回答したときの間隔である。
// FSRSのAnkiは、忘れたカードが再学習を終えるときにfuzzをかけるため、ここではかけない。
func (s Scheduler) lapseInterval(interval float64) int {
	return min(int(math.Max(math.Round(interval), 1)), s.maximumInterval)
}

// passingReviewIntervalsは、Hard・Good・Easyの間隔を返す。
// Ankiと同じく、Hard < Good < Easyとなるよう、前の評価の間隔を下限に使う。
func (s Scheduler) passingReviewIntervals(
	previousInterval int,
	states [4]itemState,
	fuzz Fuzz,
) [3]int {
	var intervals [3]int
	lowerBound := 1
	for index, rating := range []Rating{Hard, Good, Easy} {
		interval := states[rating-1].interval
		minimum := max(minimumReviewFuzzInterval(interval, previousInterval, s.maximumInterval), lowerBound)
		intervals[index] = fuzz.apply(interval, clampInt(minimum, 1, s.maximumInterval), s.maximumInterval)
		lowerBound = intervals[index] + 1
	}
	return intervals
}
