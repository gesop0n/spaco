package scheduler

import (
	"fmt"
	"math"
	"testing"
)

// 期待値は、Ankiが使うfsrs-rs 6.6.2のテストとドキュメントから引用した。

func TestNextStatesForFirstReview(t *testing.T) {
	t.Parallel()

	states := defaultParameters.nextStates(nil, 0, 0.9)
	want := [4]MemoryState{
		{Stability: 0.212, Difficulty: 6.4133},
		{Stability: 1.2931, Difficulty: 5.1121707},
		{Stability: 2.3065, Difficulty: 2.118104},
		{Stability: 8.2956, Difficulty: 1.0},
	}
	for index, state := range states {
		rating := Rating(index + 1)
		assertClose(t, fmt.Sprintf("rating %d stability", rating), state.memory.Stability, want[index].Stability)
		assertClose(t, fmt.Sprintf("rating %d difficulty", rating), state.memory.Difficulty, want[index].Difficulty)
		// 目標の記憶率が90%のとき、間隔はstabilityと等しい。
		assertClose(t, fmt.Sprintf("rating %d interval", rating), state.interval, want[index].Stability)
	}
}

func TestMemoryStateAfterReviewSequence(t *testing.T) {
	t.Parallel()

	withoutShortTerm := defaultParameters
	withoutShortTerm[17], withoutShortTerm[18], withoutShortTerm[19] = 0, 0, 0
	for _, test := range []struct {
		name           string
		parameters     parameters
		wantStability  float64
		wantDifficulty float64
	}{
		{name: "default", parameters: defaultParameters, wantStability: 53.62691, wantDifficulty: 6.3574867},
		{name: "without short-term stability", parameters: withoutShortTerm, wantStability: 53.335106, wantDifficulty: 6.3574867},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ratings := []Rating{Again, Good, Good, Good, Good, Good}
			elapsedDays := []int{0, 0, 1, 3, 8, 21}
			var memory *MemoryState
			for index, rating := range ratings {
				next := test.parameters.nextStates(memory, elapsedDays[index], 0.9)[rating-1].memory
				memory = &next
			}
			assertClose(t, "stability", memory.Stability, test.wantStability)
			assertClose(t, "difficulty", memory.Difficulty, test.wantDifficulty)
		})
	}
}

func TestIntervalForDesiredRetention(t *testing.T) {
	t.Parallel()

	want := []float64{3116766, 34793, 2508, 387, 90, 27, 9, 3, 1, 1}
	for index, expected := range want {
		retention := float64(index+1) / 10
		got := math.Max(math.Round(defaultParameters.interval(1, retention)), 1)
		assertClose(t, fmt.Sprintf("interval at retention %.1f", retention), got, expected)
	}
}

func assertClose(t *testing.T, name string, got, want float64) {
	t.Helper()
	// fsrs-rsはf32で計算するため、相対誤差1e-4までを同じ値とみなす。
	if math.Abs(got-want) > 1e-4*math.Max(1, math.Abs(want)) {
		t.Fatalf("%s = %v, want %v", name, got, want)
	}
}
