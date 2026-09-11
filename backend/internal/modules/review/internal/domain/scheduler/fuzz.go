package scheduler

import "math"

// fuzzRangesは、Ankiが間隔の長さに応じて加えるfuzzの割合である。
var fuzzRanges = [...]struct{ start, end, factor float64 }{
	{start: 2.5, end: 7, factor: 0.15},
	{start: 7, end: 20, factor: 0.1},
	{start: 20, end: math.MaxFloat64, factor: 0.05},
}

// Fuzzは、Ankiのfuzz_factorに相当する。
// 同じ日に予定が集中しないよう、許容範囲の中から間隔を選ぶための[0, 1)の値を持つ。
type Fuzz struct {
	factor  float64
	enabled bool
}

// NoFuzzは、間隔を四捨五入するだけのFuzzを返す。
func NoFuzz() Fuzz { return Fuzz{} }

// FuzzFromSeedは、seedから決まる値でfuzzをかける。
// Ankiと同じく復習対象と復習回数からseedを作れば、プレビューと記録で同じ間隔になる。
func FuzzFromSeed(seed uint64) Fuzz {
	// splitmix64で偏りを減らしてから、上位53bitを[0, 1)の浮動小数点数に変換する。
	z := seed + 0x9e3779b97f4a7c15
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	z ^= z >> 31
	return Fuzz{factor: float64(z>>11) / (1 << 53), enabled: true}
}

// applyは、Ankiのwith_review_fuzzと同じく、minimum以上maximum以下で間隔を決める。
func (f Fuzz) apply(interval float64, minimum, maximum int) int {
	if !f.enabled {
		return clampInt(int(math.Round(interval)), minimum, maximum)
	}
	lower, upper := constrainedFuzzBounds(interval, minimum, maximum)
	return int(math.Floor(float64(lower) + f.factor*float64(1+upper-lower)))
}

// constrainedFuzzBoundsは、fuzzの範囲をminimumとmaximumの内側に収める。
// 範囲が1日に潰れる場合は、maximumが許せば上限を1日広げる。
func constrainedFuzzBounds(interval float64, minimum, maximum int) (int, int) {
	minimum = min(minimum, maximum)
	interval = clamp(interval, float64(minimum), float64(maximum))
	delta := fuzzDelta(interval)
	lower := clampInt(int(math.Round(interval-delta)), minimum, maximum)
	upper := clampInt(int(math.Round(interval+delta)), minimum, maximum)
	if upper == lower && upper > 2 && upper < maximum {
		upper = lower + 1
	}
	return lower, upper
}

// fuzzDeltaは、間隔を前後にずらす日数を返す。2.5日未満の間隔はずらさない。
func fuzzDelta(interval float64) float64 {
	if interval < 2.5 {
		return 0
	}
	delta := 1.0
	for _, fuzzRange := range fuzzRanges {
		delta += fuzzRange.factor * math.Max(math.Min(interval, fuzzRange.end)-fuzzRange.start, 0)
	}
	return delta
}

// minimumReviewFuzzIntervalは、fuzzによって前回の間隔より短くならないための下限を返す。
func minimumReviewFuzzInterval(interval float64, previousInterval, maximumInterval int) int {
	_, upper := constrainedFuzzBounds(interval, 1, maximumInterval)
	switch {
	case int(math.Round(interval)) > previousInterval:
		return previousInterval + 1
	case previousInterval <= upper:
		// 間隔があまり伸びなかった場合も、fuzzで前回より短くはしない。
		return previousInterval
	default:
		// パラメータの変更などで間隔が大きく縮んだ場合は、下限を設けない。
		return 0
	}
}

func clampInt(value, lower, upper int) int {
	return min(max(value, lower), upper)
}
