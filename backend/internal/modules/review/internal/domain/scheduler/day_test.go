package scheduler

import (
	"testing"
	"time"
)

func TestDayOfUsesLocation(t *testing.T) {
	t.Parallel()

	instant := time.Date(2026, 9, 8, 15, 30, 0, 0, time.UTC)
	tokyo, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatalf("LoadLocation() error = %v", err)
	}
	losAngeles, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatalf("LoadLocation() error = %v", err)
	}
	if got := DayOf(instant, tokyo).String(); got != "2026-09-09" {
		t.Fatalf("Tokyo day = %s, want 2026-09-09", got)
	}
	if got := DayOf(instant, losAngeles).String(); got != "2026-09-08" {
		t.Fatalf("Los Angeles day = %s, want 2026-09-08", got)
	}
	if got := DayOf(instant, tokyo).StartIn(tokyo); !got.Equal(time.Date(2026, 9, 9, 0, 0, 0, 0, tokyo)) {
		t.Fatalf("StartIn() = %v, want midnight in Tokyo", got)
	}
}

func TestDayArithmeticAcrossMonthAndLeapDay(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		day  string
		add  int
		want string
	}{
		{day: "2026-12-31", add: 1, want: "2027-01-01"},
		{day: "2024-02-28", add: 1, want: "2024-02-29"},
		{day: "2026-02-28", add: 1, want: "2026-03-01"},
		{day: "2026-03-01", add: -1, want: "2026-02-28"},
	} {
		day := mustParseDay(t, test.day)
		if got := day.AddDays(test.add).String(); got != test.want {
			t.Errorf("%s + %d = %s, want %s", test.day, test.add, got, test.want)
		}
		if got := day.AddDays(test.add).DaysSince(day); got != test.add {
			t.Errorf("DaysSince() = %d, want %d", got, test.add)
		}
	}
	if _, err := ParseDay("2026-02-29"); err == nil {
		t.Fatal("ParseDay() accepted a non-existent date")
	}
}
