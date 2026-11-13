package datepicker

import (
	"strings"
	"testing"
	"time"
)

func TestNewRangeStateDefaults(t *testing.T) {
	r := newRangeState(day(2026, time.November, 15), time.Time{})

	if got, want := r.from, day(2026, time.November, 15); !got.Equal(want) {
		t.Fatalf("from = %v, want %v", got, want)
	}
	if got, want := r.until, day(2026, time.November, 15); !got.Equal(want) {
		t.Fatalf("until = %v, want %v", got, want)
	}
	if !r.haveUntil {
		t.Fatal("initial range should have Until selected")
	}
	if got, want := r.cal.cursor, day(2026, time.November, 15); !got.Equal(want) {
		t.Fatalf("cursor = %v, want %v", got, want)
	}
}

func TestNewRangeStateSortsInitialRange(t *testing.T) {
	r := newRangeState(day(2026, time.November, 20), day(2026, time.November, 15))
	if got, want := r.from, day(2026, time.November, 15); !got.Equal(want) {
		t.Fatalf("from = %v, want %v", got, want)
	}
	if got, want := r.until, day(2026, time.November, 20); !got.Equal(want) {
		t.Fatalf("until = %v, want %v", got, want)
	}
}

func TestNewRangeStateUsesOneLocation(t *testing.T) {
	plus14 := time.FixedZone("+14", 14*60*60)
	minus11 := time.FixedZone("-11", -11*60*60)
	from := time.Date(2026, time.November, 16, 0, 0, 0, 0, plus14)
	until := time.Date(2026, time.November, 17, 0, 0, 0, 0, minus11)

	r := newRangeState(from, until)
	if r.from.Location() != plus14 || r.until.Location() != plus14 {
		t.Fatalf("locations = %v, %v; want both +14:00", r.from.Location(), r.until.Location())
	}
	if want := time.Date(2026, time.November, 17, 0, 0, 0, 0, plus14); !r.until.Equal(want) {
		t.Fatalf("until = %v, want %v", r.until, want)
	}
}

func TestRangeSelectSequence(t *testing.T) {
	r := newRangeState(day(2026, time.November, 15), time.Time{})
	r.clickDate(day(2026, time.November, 10))

	if got, want := r.from, day(2026, time.November, 10); !got.Equal(want) {
		t.Fatalf("from = %v, want %v", got, want)
	}
	if r.haveUntil {
		t.Fatal("Until should be cleared by the first click")
	}

	r.clickDate(day(2026, time.November, 20))
	if !r.haveUntil {
		t.Fatal("Until should be selected by the second click")
	}
	if got, want := r.until, day(2026, time.November, 20); !got.Equal(want) {
		t.Fatalf("until = %v, want %v", got, want)
	}

	r.clickDate(day(2026, time.December, 1))
	if r.haveUntil {
		t.Fatal("a third click should start a new range")
	}
	if got, want := r.from, day(2026, time.December, 1); !got.Equal(want) {
		t.Fatalf("from = %v, want %v", got, want)
	}
}

func TestRangeSelectSortsEndpoints(t *testing.T) {
	r := newRangeState(day(2026, time.November, 15), time.Time{})
	r.clickDate(day(2026, time.November, 20))
	r.clickDate(day(2026, time.November, 10))

	if got, want := r.from, day(2026, time.November, 10); !got.Equal(want) {
		t.Fatalf("from = %v, want %v", got, want)
	}
	if got, want := r.until, day(2026, time.November, 20); !got.Equal(want) {
		t.Fatalf("until = %v, want %v", got, want)
	}
}

func TestRangeNavigateKeepsEndpoints(t *testing.T) {
	r := newRangeState(day(2026, time.November, 15), day(2026, time.November, 20))

	r.navigate(keyRight)
	r.navigate(keyPageDown)

	if got, want := r.from, day(2026, time.November, 15); !got.Equal(want) {
		t.Fatalf("from = %v, want %v", got, want)
	}
	if got, want := r.until, day(2026, time.November, 20); !got.Equal(want) {
		t.Fatalf("until = %v, want %v", got, want)
	}
	if got, want := r.cal.cursor, day(2026, time.December, 16); !got.Equal(want) {
		t.Fatalf("cursor = %v, want %v", got, want)
	}
}

func TestRangeConfirmUsesCursorWhenUntilMissing(t *testing.T) {
	r := newRangeState(day(2026, time.November, 10), time.Time{})
	r.clickDate(day(2026, time.November, 12))
	r.cal.cursor = day(2026, time.November, 15)
	r.cal.month = firstOfMonth(r.cal.cursor)

	from, until := r.confirm()
	if want := day(2026, time.November, 12); !from.Equal(want) {
		t.Fatalf("from = %v, want %v", from, want)
	}
	if want := day(2026, time.November, 15); !until.Equal(want) {
		t.Fatalf("until = %v, want %v", until, want)
	}
}

func TestRangeStatusLine(t *testing.T) {
	r := newRangeState(day(2026, time.November, 15), day(2026, time.November, 20))
	if got, want := r.statusLine(), "2026-11-15 -> 2026-11-20"; got != want {
		t.Fatalf("status = %q, want %q", got, want)
	}

	r.clickDate(day(2026, time.November, 18))
	if got, want := r.statusLine(), "2026-11-18 -> --"; got != want {
		t.Fatalf("status = %q, want %q", got, want)
	}
}

func TestRangeStyleFor(t *testing.T) {
	r := newRangeState(day(2026, time.November, 15), day(2026, time.November, 20))
	r.cal.cursor = day(2026, time.November, 16)
	today := day(2026, time.November, 16)

	tests := []struct {
		name string
		date time.Time
		want string
	}{
		{"endpoint", day(2026, time.November, 15), "\x1b[44;97m"},
		{"cursor in range and today", day(2026, time.November, 16), "\x1b[1;4;100;97m"},
		{"in range not cursor", day(2026, time.November, 17), "\x1b[100;97m"},
		{"endpoint not cursor", day(2026, time.November, 20), "\x1b[44;97m"},
		{"outside", day(2026, time.November, 26), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := r.styleFor(tt.date, today); got != tt.want {
				t.Fatalf("styleFor = %q, want %q", got, tt.want)
			}
		})
	}

	outside := newRangeState(day(2026, time.November, 15), day(2026, time.November, 20))
	outside.cal.cursor = day(2026, time.November, 25)
	if got, want := outside.styleFor(day(2026, time.November, 25), today), "\x1b[7m"; got != want {
		t.Fatalf("cursor outside style = %q, want %q", got, want)
	}

	same := newRangeState(day(2026, time.November, 15), time.Time{})
	same.cal.cursor = day(2026, time.November, 15)
	if got, want := same.styleFor(day(2026, time.November, 15), day(2026, time.November, 15)), "\x1b[1;4;44;97m"; got != want {
		t.Fatalf("same-day style = %q, want %q", got, want)
	}
}

func TestRangeLinesIncludeStatus(t *testing.T) {
	r := newRangeState(day(2026, time.November, 15), day(2026, time.November, 20))
	lines := r.lines(day(2026, time.November, 16))

	if got, want := len(lines), r.requiredHeight(); got != want {
		t.Fatalf("lines = %d, want %d", got, want)
	}
	if !strings.Contains(lines[0], "2026 Nov") {
		t.Fatalf("header = %q", lines[0])
	}
	if got, want := lines[len(lines)-1], "2026-11-15 -> 2026-11-20"; got != want {
		t.Fatalf("status line = %q, want %q", got, want)
	}
}

func TestRangeTooSmallMessage(t *testing.T) {
	r := newRangeState(day(2026, time.November, 15), day(2026, time.November, 20))
	message := r.tooSmallMessage()
	for _, want := range []string{"28x8", "PageUp"} {
		if !strings.Contains(message, want) {
			t.Errorf("message = %q, want it to contain %q", message, want)
		}
	}
}
