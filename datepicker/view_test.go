package datepicker

import (
	"strings"
	"testing"
	"time"
)

func TestPlainLinesNovember2026(t *testing.T) {
	cal := newCalendar(day(2026, time.November, 15))
	lines := cal.plainLines()

	want := []string{
		weekdayHeader,
		"  1   2   3   4   5   6   7",
		"  8   9  10  11  12  13  14",
		" 15  16  17  18  19  20  21",
		" 22  23  24  25  26  27  28",
		" 29  30" + strings.Repeat(" ", calendarWidth-len(" 29  30")),
	}
	if len(lines) != len(want)+1 {
		t.Fatalf("got %d lines, want %d", len(lines), len(want)+1)
	}

	header := lines[0]
	if len(header) != calendarWidth {
		t.Errorf("header width = %d, want %d: %q", len(header), calendarWidth, header)
	}
	if !strings.HasPrefix(header, "<") || !strings.HasSuffix(header, ">") {
		t.Errorf("header must be enclosed by arrows: %q", header)
	}
	if got := strings.Index(header, "2026 Nov"); got != 9 {
		t.Errorf("title starts at %d, want 9: %q", got, header)
	}

	for i, wantLine := range want {
		if got := lines[i+1]; got != wantLine {
			t.Errorf("line %d = %q, want %q", i+1, got, wantLine)
		}
	}
}

func TestLinesStyles(t *testing.T) {
	cal := newCalendar(day(2026, time.November, 15))

	sameDay := cal.lines(day(2026, time.November, 15))
	if !containsLine(sameDay, "\x1b[1;4;7m 15\x1b[0m") {
		t.Errorf("cursor on today should be reverse, underlined and bold: %q", sameDay)
	}

	cursorOnly := cal.lines(day(2026, time.November, 20))
	if !containsLine(cursorOnly, "\x1b[7m 15\x1b[0m") {
		t.Errorf("cursor should be reverse video: %q", cursorOnly)
	}
	if !containsLine(cursorOnly, "\x1b[4m 20\x1b[0m") {
		t.Errorf("today should be underlined: %q", cursorOnly)
	}
}

func TestLinesHeightMatchesMonthRows(t *testing.T) {
	cal := newCalendar(day(2026, time.February, 1))
	if got, want := len(cal.lines(time.Now())), cal.height(); got != want {
		t.Fatalf("lines = %d, height = %d", got, want)
	}
}

func containsLine(lines []string, substr string) bool {
	for _, line := range lines {
		if strings.Contains(line, substr) {
			return true
		}
	}
	return false
}
