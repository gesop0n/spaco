package scheduler

import (
	"errors"
	"testing"
)

func TestNextSchedulesFirstReview(t *testing.T) {
	t.Parallel()

	today := mustParseDay(t, "2026-09-11")
	for _, test := range []struct {
		name         string
		rating       Rating
		wantInterval int
	}{
		// FSRSの初期stabilityは0.212日だが、日単位の復習なので1日にする。
		{name: "again", rating: Again, wantInterval: 1},
		{name: "hard", rating: Hard, wantInterval: 1},
		{name: "good", rating: Good, wantInterval: 2},
		{name: "easy", rating: Easy, wantInterval: 8},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			next, err := NewScheduler().Next(NewCard(today), test.rating, today, NoFuzz())
			if err != nil {
				t.Fatalf("Next() error = %v", err)
			}
			if next.State != StateReview || next.IntervalDays != test.wantInterval ||
				next.Due != today.AddDays(test.wantInterval) || next.LastReviewedOn != today ||
				next.Reps != 1 || next.Lapses != 0 {
				t.Fatalf("Next() = %+v, want interval %d", next, test.wantInterval)
			}
			want := defaultParameters.nextStates(nil, 0, 0.9)[test.rating-1].memory
			if next.Memory == nil || *next.Memory != want {
				t.Fatalf("Memory = %+v, want %+v", next.Memory, want)
			}
		})
	}
}

func TestNextAgainDuringReviewCountsLapseWithoutFuzz(t *testing.T) {
	t.Parallel()

	today := mustParseDay(t, "2026-09-11")
	card := reviewCard(today.AddDays(-10), 10, MemoryState{Stability: 10, Difficulty: 5})
	withoutFuzz, err := NewScheduler().Next(card, Again, today, NoFuzz())
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}
	if withoutFuzz.Lapses != 1 || withoutFuzz.Reps != card.Reps+1 {
		t.Fatalf("Next() = %+v, want one more lapse and rep", withoutFuzz)
	}
	for seed := range uint64(50) {
		withFuzz, err := NewScheduler().Next(card, Again, today, FuzzFromSeed(seed))
		if err != nil {
			t.Fatalf("Next() error = %v", err)
		}
		if withFuzz.IntervalDays != withoutFuzz.IntervalDays {
			t.Fatalf("seed %d interval = %d, want %d without fuzz", seed, withFuzz.IntervalDays, withoutFuzz.IntervalDays)
		}
	}
}

func TestNextKeepsHardGoodEasyOrdered(t *testing.T) {
	t.Parallel()

	today := mustParseDay(t, "2026-09-11")
	for _, elapsedDays := range []int{0, 3, 10, 30} {
		card := reviewCard(today.AddDays(-elapsedDays), 10, MemoryState{Stability: 10, Difficulty: 5})
		for seed := range uint64(200) {
			intervals := make(map[Rating]int, 3)
			for _, rating := range []Rating{Hard, Good, Easy} {
				next, err := NewScheduler().Next(card, rating, today, FuzzFromSeed(seed))
				if err != nil {
					t.Fatalf("Next() error = %v", err)
				}
				intervals[rating] = next.IntervalDays
			}
			if intervals[Hard] >= intervals[Good] || intervals[Good] >= intervals[Easy] {
				t.Fatalf("elapsed %d, seed %d: intervals = %v, want hard < good < easy", elapsedDays, seed, intervals)
			}
		}
	}
}

func TestNextDoesNotLetSameDayReviewShortenPreviousInterval(t *testing.T) {
	t.Parallel()

	today := mustParseDay(t, "2026-09-11")
	card := reviewCard(today, 10, MemoryState{Stability: 10, Difficulty: 5})
	for rating, want := range map[Rating]int{Hard: 10, Good: 11, Easy: 16} {
		next, err := NewScheduler().Next(card, rating, today, NoFuzz())
		if err != nil {
			t.Fatalf("Next() error = %v", err)
		}
		if next.IntervalDays != want {
			t.Errorf("rating %d interval = %d, want %d", rating, next.IntervalDays, want)
		}
	}
}

func TestNextExtendsIntervalWhileSolvedIndependently(t *testing.T) {
	t.Parallel()

	day := mustParseDay(t, "2026-09-11")
	card := NewCard(day)
	previousInterval := 0
	for range 6 {
		next, err := NewScheduler().Next(card, Good, card.Due, NoFuzz())
		if err != nil {
			t.Fatalf("Next() error = %v", err)
		}
		if next.IntervalDays <= previousInterval {
			t.Fatalf("interval = %d after %d, want longer interval", next.IntervalDays, previousInterval)
		}
		previousInterval = next.IntervalDays
		card = next
	}
}

func TestNextRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	today := mustParseDay(t, "2026-09-11")
	if _, err := NewScheduler().Next(NewCard(today), Rating(5), today, NoFuzz()); !errors.Is(err, ErrInvalidRating) {
		t.Fatalf("invalid rating error = %v, want ErrInvalidRating", err)
	}
	broken := NewCard(today)
	broken.Reps = 1
	if _, err := NewScheduler().Next(broken, Good, today, NoFuzz()); !errors.Is(err, ErrInvalidCard) {
		t.Fatalf("invalid card error = %v, want ErrInvalidCard", err)
	}
}

func TestElapsedDays(t *testing.T) {
	t.Parallel()

	today := mustParseDay(t, "2026-09-11")
	if got := ElapsedDays(NewCard(today.AddDays(-3)), today); got != 0 {
		t.Fatalf("new card ElapsedDays() = %d, want 0", got)
	}
	card := reviewCard(today.AddDays(-4), 2, MemoryState{Stability: 2, Difficulty: 5})
	if got := ElapsedDays(card, today); got != 4 {
		t.Fatalf("ElapsedDays() = %d, want 4", got)
	}
	if got := ElapsedDays(card, today.AddDays(-5)); got != 0 {
		t.Fatalf("ElapsedDays() before last review = %d, want 0", got)
	}
}

func reviewCard(lastReviewedOn Day, intervalDays int, memory MemoryState) Card {
	return Card{
		State:          StateReview,
		Due:            lastReviewedOn.AddDays(intervalDays),
		IntervalDays:   intervalDays,
		Memory:         &memory,
		LastReviewedOn: lastReviewedOn,
		Reps:           3,
	}
}

func mustParseDay(t *testing.T, value string) Day {
	t.Helper()
	day, err := ParseDay(value)
	if err != nil {
		t.Fatalf("ParseDay() error = %v", err)
	}
	return day
}
