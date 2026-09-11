package scheduler

import (
	"fmt"
	"testing"
)

// 期待値は、Ankiのrslib/src/scheduler/states/fuzz.rsのテストから引用した。

func TestFuzzApplyWithoutFuzzRoundsAndClamps(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		interval float64
		minimum  int
		maximum  int
		want     int
	}{
		{interval: 1.5, minimum: 1, maximum: 100, want: 2},
		{interval: 0.1, minimum: 1, maximum: 100, want: 1},
		{interval: 101, minimum: 1, maximum: 100, want: 100},
	} {
		if got := NoFuzz().apply(test.interval, test.minimum, test.maximum); got != test.want {
			t.Errorf("apply(%v, %d, %d) = %d, want %d", test.interval, test.minimum, test.maximum, got, test.want)
		}
	}
}

func TestFuzzApplyMatchesAnki(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		interval float64
		minimum  int
		maximum  int
		lower    int
		middle   int
		upper    int
	}{
		// 2.5日未満はずらさない。
		{interval: 1.0, minimum: 1, maximum: 1000, lower: 1, middle: 1, upper: 1},
		{interval: 2.49, minimum: 1, maximum: 1000, lower: 2, middle: 2, upper: 2},
		// 2.5日以上は1日に、帯ごとの日数×割合を加えた幅でずらす。
		{interval: 2.5, minimum: 1, maximum: 1000, lower: 2, middle: 3, upper: 4},
		{interval: 7.0, minimum: 1, maximum: 1000, lower: 5, middle: 7, upper: 9},
		{interval: 17.0, minimum: 1, maximum: 1000, lower: 14, middle: 17, upper: 20},
		{interval: 37.0, minimum: 1, maximum: 1000, lower: 33, middle: 37, upper: 41},
		// 下限に合わせた結果、範囲が1日に潰れる場合は広げる。
		{interval: 2.0, minimum: 2, maximum: 1000, lower: 2, middle: 2, upper: 2},
		{interval: 2.0, minimum: 3, maximum: 1000, lower: 3, middle: 4, upper: 4},
		{interval: 2.0, minimum: 3, maximum: 3, lower: 3, middle: 3, upper: 3},
		// 帯の境界付近。
		{interval: 6.9, minimum: 3, maximum: 1000, lower: 5, middle: 7, upper: 9},
		{interval: 7.0, minimum: 3, maximum: 1000, lower: 5, middle: 7, upper: 9},
		{interval: 7.1, minimum: 3, maximum: 1000, lower: 5, middle: 7, upper: 9},
		{interval: 19.9, minimum: 3, maximum: 1000, lower: 17, middle: 20, upper: 23},
		{interval: 20.0, minimum: 3, maximum: 1000, lower: 17, middle: 20, upper: 23},
		{interval: 20.1, minimum: 3, maximum: 1000, lower: 17, middle: 20, upper: 23},
		// 上限・下限を守り、範囲内では均等に選ぶ。
		{interval: 100.0, minimum: 101, maximum: 1000, lower: 101, middle: 105, upper: 108},
		{interval: 100.0, minimum: 1, maximum: 99, lower: 92, middle: 96, upper: 99},
		{interval: 100.0, minimum: 97, maximum: 103, lower: 97, middle: 100, upper: 103},
	} {
		for factor, want := range map[float64]int{0: test.lower, 0.5: test.middle, 0.99: test.upper} {
			t.Run(fmt.Sprintf("%v in [%d,%d] with %v", test.interval, test.minimum, test.maximum, factor), func(t *testing.T) {
				t.Parallel()
				fuzz := Fuzz{factor: factor, enabled: true}
				if got := fuzz.apply(test.interval, test.minimum, test.maximum); got != want {
					t.Fatalf("apply() = %d, want %d", got, want)
				}
			})
		}
	}
}

func TestConstrainedFuzzBoundsDoesNotPanicWhenMinimumExceedsMaximum(t *testing.T) {
	t.Parallel()

	lower, upper := constrainedFuzzBounds(1.0, 3, 2)
	if lower != 2 || upper != 2 {
		t.Fatalf("constrainedFuzzBounds() = %d, %d; want 2, 2", lower, upper)
	}
}

func TestMinimumReviewFuzzIntervalPreservesPreviousOnlyWithinRange(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		interval float64
		previous int
		want     int
	}{
		{interval: 2.7269483, previous: 4, want: 4},
		{interval: 2.7269483, previous: 5, want: 0},
		{interval: 4.591988, previous: 4, want: 5},
	} {
		if got := minimumReviewFuzzInterval(test.interval, test.previous, 36500); got != test.want {
			t.Errorf("minimumReviewFuzzInterval(%v, %d) = %d, want %d", test.interval, test.previous, got, test.want)
		}
	}
}

func TestFuzzFromSeedIsDeterministicAndInRange(t *testing.T) {
	t.Parallel()

	for seed := range uint64(1000) {
		fuzz := FuzzFromSeed(seed)
		if fuzz != FuzzFromSeed(seed) {
			t.Fatalf("FuzzFromSeed(%d) is not deterministic", seed)
		}
		if !fuzz.enabled || fuzz.factor < 0 || fuzz.factor >= 1 {
			t.Fatalf("FuzzFromSeed(%d) = %+v, want factor in [0, 1)", seed, fuzz)
		}
	}
}
