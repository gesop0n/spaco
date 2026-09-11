package scheduler

import "math"

// parametersは、FSRS-6のモデルパラメータw0〜w20である。
type parameters [21]float64

// defaultParametersは、Ankiが使うfsrs-rs 6.6.2の既定パラメータである。
// 復習履歴からパラメータを最適化する機能を持つまでは、全ユーザーにこの値を使う。
var defaultParameters = parameters{
	0.212, 1.2931, 2.3065, 8.2956, 6.4133, 0.8334, 3.0194, 0.001, 1.8722, 0.1666, 0.796,
	1.4835, 0.0614, 0.2629, 1.6483, 0.6014, 1.8729, 0.5425, 0.0912, 0.0658, 0.1542,
}

// fsrs-rsが記憶状態に適用する範囲。
const (
	minimumStability  = 0.001
	maximumStability  = 36500.0
	minimumDifficulty = 1.0
	maximumDifficulty = 10.0
)

// MemoryStateは、FSRSで推定した記憶の状態である。
type MemoryState struct {
	// Stabilityは、思い出せる確率が90%まで下がるまでの日数である。
	Stability float64
	// Difficultyは、覚えにくさを1〜10で表す。
	Difficulty float64
}

// itemStateは、fsrs-rsのItemStateと同じく、ある評価で回答した後の記憶状態と間隔を表す。
type itemState struct {
	memory   MemoryState
	interval float64
}

// nextStatesは、fsrs-rsのFSRS::next_statesと同じく、Again〜Easyそれぞれで回答した結果を返す。
// previousがnilなら、初めての回答として初期値を使う。
func (w parameters) nextStates(
	previous *MemoryState,
	elapsedDays int,
	desiredRetention float64,
) [4]itemState {
	var states [4]itemState
	for rating := Again; rating <= Easy; rating++ {
		memory := w.nextMemoryState(previous, float64(elapsedDays), rating)
		states[rating-1] = itemState{
			memory:   memory,
			interval: w.interval(memory.Stability, desiredRetention),
		}
	}
	return states
}

func (w parameters) nextMemoryState(
	previous *MemoryState,
	elapsedDays float64,
	rating Rating,
) MemoryState {
	if previous == nil {
		return MemoryState{
			Stability:  clamp(w[rating-1], minimumStability, maximumStability),
			Difficulty: clamp(w.initialDifficulty(rating), minimumDifficulty, maximumDifficulty),
		}
	}

	stability := clamp(previous.Stability, minimumStability, maximumStability)
	difficulty := clamp(previous.Difficulty, minimumDifficulty, maximumDifficulty)
	retrievability := w.retrievability(elapsedDays, stability)
	var nextStability float64
	switch {
	case elapsedDays == 0:
		// 同じ日のうちに再度回答した場合は、短期記憶の式を使う。
		nextStability = w.shortTermStability(stability, rating)
	case rating == Again:
		nextStability = w.stabilityAfterFailure(stability, difficulty, retrievability)
	default:
		nextStability = w.stabilityAfterSuccess(stability, difficulty, retrievability, rating)
	}
	return MemoryState{
		Stability: clamp(nextStability, minimumStability, maximumStability),
		Difficulty: clamp(
			w.meanReversion(w.nextDifficulty(difficulty, rating)),
			minimumDifficulty,
			maximumDifficulty,
		),
	}
}

func (w parameters) decay() float64 { return -w[20] }

// factorは、stabilityと同じ日数が経ったときに、思い出せる確率を90%にする係数である。
func (w parameters) factor() float64 { return math.Pow(0.9, 1/w.decay()) - 1 }

// retrievabilityは、べき乗の忘却曲線から、経過日数後に思い出せる確率を返す。
func (w parameters) retrievability(elapsedDays, stability float64) float64 {
	return math.Pow(elapsedDays/stability*w.factor()+1, w.decay())
}

// intervalは、思い出せる確率がdesiredRetentionまで下がる日数を返す。
func (w parameters) interval(stability, desiredRetention float64) float64 {
	return stability / w.factor() * (math.Pow(desiredRetention, 1/w.decay()) - 1)
}

func (w parameters) initialDifficulty(rating Rating) float64 {
	return w[4] - math.Exp(w[5]*float64(rating-1)) + 1
}

func (w parameters) nextDifficulty(difficulty float64, rating Rating) float64 {
	delta := -w[6] * float64(rating-Good)
	// 上限の10に近いほど、難しさの増え方を小さくする。
	return difficulty + (10-difficulty)*delta/9
}

// meanReversionは、難しさを「Easyで初めて回答したときの値」へ少しずつ戻す。
func (w parameters) meanReversion(difficulty float64) float64 {
	return w[7]*(w.initialDifficulty(Easy)-difficulty) + difficulty
}

func (w parameters) stabilityAfterSuccess(
	stability float64,
	difficulty float64,
	retrievability float64,
	rating Rating,
) float64 {
	hardPenalty, easyBonus := 1.0, 1.0
	switch rating {
	case Hard:
		hardPenalty = w[15]
	case Easy:
		easyBonus = w[16]
	}
	return stability * (math.Exp(w[8])*
		(11-difficulty)*
		math.Pow(stability, -w[9])*
		(math.Exp((1-retrievability)*w[10])-1)*
		hardPenalty*
		easyBonus + 1)
}

func (w parameters) stabilityAfterFailure(stability, difficulty, retrievability float64) float64 {
	next := w[11] *
		math.Pow(difficulty, -w[12]) *
		(math.Pow(stability+1, w[13]) - 1) *
		math.Exp((1-retrievability)*w[14])
	// 忘れた後のstabilityには、忘れる前の値にもとづく上限を設ける。
	return math.Min(next, stability/math.Exp(w[17]*w[18]))
}

func (w parameters) shortTermStability(stability float64, rating Rating) float64 {
	increase := math.Exp(w[17]*(float64(rating-Good)+w[18])) * math.Pow(stability, -w[19])
	if rating >= Hard {
		increase = math.Max(increase, 1)
	}
	return stability * increase
}

func clamp(value, lower, upper float64) float64 {
	return math.Min(math.Max(value, lower), upper)
}
