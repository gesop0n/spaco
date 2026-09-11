package scheduler

import (
	"fmt"
	"time"
)

const secondsPerDay = 24 * 60 * 60

// Dayは、ユーザーのタイムゾーンで数えた暦日を、1970-01-01からの日数で表す。
// Ankiが予定日を日数で持つのと同じく、間隔の加算や経過日数を整数で計算する。
type Day int32

// DayOfは、locationでtが属する暦日を返す。
func DayOf(t time.Time, location *time.Location) Day {
	year, month, day := t.In(location).Date()
	return Day(time.Date(year, month, day, 0, 0, 0, 0, time.UTC).Unix() / secondsPerDay)
}

// ParseDayは、YYYY-MM-DD形式の暦日を読み取る。
func ParseDay(value string) (Day, error) {
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return 0, fmt.Errorf("parse day %q: %w", value, err)
	}
	return DayOf(parsed, time.UTC), nil
}

func (d Day) AddDays(days int) Day { return d + Day(days) }

// DaysSinceは、otherからdまでの日数を返す。
func (d Day) DaysSince(other Day) int { return int(d - other) }

// Timeは、暦日をUTCの0時として返す。DBのdate型との変換に使う。
func (d Day) Time() time.Time { return time.Unix(int64(d)*secondsPerDay, 0).UTC() }

// StartInは、locationでのこの日の0時を返す。
func (d Day) StartIn(location *time.Location) time.Time {
	year, month, day := d.Time().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, location)
}

func (d Day) String() string { return d.Time().Format(time.DateOnly) }
