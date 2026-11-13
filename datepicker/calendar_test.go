package datepicker

import (
	"reflect"
	"testing"
	"time"
)

func day(year int, month time.Month, d int) time.Time {
	return time.Date(year, month, d, 0, 0, 0, 0, time.UTC)
}

func TestWeeksStartOnSunday(t *testing.T) {
	cal := newCalendar(day(2026, time.November, 15))

	want := [][]int{
		{1, 2, 3, 4, 5, 6, 7},
		{8, 9, 10, 11, 12, 13, 14},
		{15, 16, 17, 18, 19, 20, 21},
		{22, 23, 24, 25, 26, 27, 28},
		{29, 30, 0, 0, 0, 0, 0},
	}
	if got := cal.weeks(); !reflect.DeepEqual(got, want) {
		t.Fatalf("weeks() = %v, want %v", got, want)
	}
}

func TestFebruaryRows(t *testing.T) {
	february2026 := newCalendar(day(2026, time.February, 10))
	if got, want := len(february2026.weeks()), 4; got != want {
		t.Fatalf("February 2026 has %d rows, want %d", got, want)
	}
	if got, want := february2026.height(), 6; got != want {
		t.Fatalf("February 2026 height = %d, want %d", got, want)
	}

	february2028 := newCalendar(day(2028, time.February, 1))
	weeks := february2028.weeks()
	if got, want := weeks[4][2], 29; got != want {
		t.Fatalf("February 2028 day 29 position = %d, want %d", got, want)
	}
}

func TestShiftMonthClampsCursorDay(t *testing.T) {
	cal := newCalendar(day(2026, time.January, 31))
	cal.shiftMonth(1)
	if got, want := cal.cursor, day(2026, time.February, 28); !got.Equal(want) {
		t.Fatalf("cursor = %v, want %v", got, want)
	}
	cal.shiftMonth(1)
	if got, want := cal.cursor, day(2026, time.March, 28); !got.Equal(want) {
		t.Fatalf("cursor = %v, want %v", got, want)
	}
}

func TestShiftMonthAcrossYear(t *testing.T) {
	cal := newCalendar(day(2026, time.December, 15))
	cal.shiftMonth(1)
	if got, want := cal.month, day(2027, time.January, 1); !got.Equal(want) {
		t.Fatalf("month = %v, want %v", got, want)
	}
	cal.shiftMonth(-1)
	if got, want := cal.month, day(2026, time.December, 1); !got.Equal(want) {
		t.Fatalf("month = %v, want %v", got, want)
	}
}

func TestMoveDaysFollowsMonth(t *testing.T) {
	cal := newCalendar(day(2026, time.November, 30))
	cal.moveDays(1)
	if got, want := cal.cursor, day(2026, time.December, 1); !got.Equal(want) {
		t.Fatalf("cursor = %v, want %v", got, want)
	}
	if got, want := cal.month, day(2026, time.December, 1); !got.Equal(want) {
		t.Fatalf("month = %v, want %v", got, want)
	}
}

func TestDateAt(t *testing.T) {
	cal := newCalendar(day(2026, time.November, 15))

	tests := []struct {
		row, col int
		want     time.Time
		ok       bool
	}{
		{0, 0, day(2026, time.November, 1), true},
		{2, 0, day(2026, time.November, 15), true},
		{4, 1, day(2026, time.November, 30), true},
		{4, 2, time.Time{}, false},
		{5, 0, time.Time{}, false},
		{0, 7, time.Time{}, false},
	}
	for _, tt := range tests {
		got, ok := cal.dateAt(tt.row, tt.col)
		if ok != tt.ok || !got.Equal(tt.want) {
			t.Errorf("dateAt(%d,%d) = %v,%v, want %v,%v", tt.row, tt.col, got, ok, tt.want, tt.ok)
		}
	}
}

func TestHit(t *testing.T) {
	cal := newCalendar(day(2026, time.November, 15))

	tests := []struct {
		row, col int
		want     actionKind
		date     time.Time
	}{
		{1, 1, actionPrevMonth, time.Time{}},
		{1, 3, actionPrevMonth, time.Time{}},
		{1, 25, actionNextMonth, time.Time{}},
		{1, 27, actionNextMonth, time.Time{}},
		{1, 10, actionNone, time.Time{}},
		{2, 5, actionNone, time.Time{}},
		{3, 1, actionSelect, day(2026, time.November, 1)},
		{3, 5, actionSelect, day(2026, time.November, 2)},
		{7, 1, actionSelect, day(2026, time.November, 29)},
		{7, 25, actionNone, time.Time{}},
		{9, 1, actionNone, time.Time{}},
	}
	for _, tt := range tests {
		got := cal.hit(tt.row, tt.col)
		if got.kind != tt.want || !got.date.Equal(tt.date) {
			t.Errorf("hit(%d,%d) = kind %v date %v, want kind %v date %v",
				tt.row, tt.col, got.kind, got.date, tt.want, tt.date)
		}
	}
}

func TestHitRejectsOutOfRange(t *testing.T) {
	cal := newCalendar(day(2026, time.November, 15))
	for _, pos := range [][2]int{{0, 1}, {1, 0}, {1, 28}, {3, 28}} {
		if got := cal.hit(pos[0], pos[1]); got.kind != actionNone {
			t.Errorf("hit(%d,%d) = %v, want actionNone", pos[0], pos[1], got.kind)
		}
	}
}
