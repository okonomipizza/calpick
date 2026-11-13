package datepicker

import (
	"fmt"
	"strings"
	"time"
)

// rangeState is a single-calendar range picker. Clicking selects From first,
// then Until; a further click starts a new range. The calendar cursor moves
// independently from the selected endpoints.
type rangeState struct {
	cal       calendar
	from      time.Time
	until     time.Time
	haveUntil bool
}

func newRangeState(from, until time.Time) rangeState {
	from = dateOnly(from)
	if until.IsZero() {
		until = from
	} else {
		// Keep the calendar date the caller wrote, but use one location for
		// all comparisons and for the returned values.
		until = time.Date(until.Year(), until.Month(), until.Day(), 0, 0, 0, 0, from.Location())
	}
	if until.Before(from) {
		from, until = until, from
	}
	return rangeState{
		cal:       newCalendar(from),
		from:      from,
		until:     until,
		haveUntil: true,
	}
}

// navigate moves the cursor or the displayed month. It never changes the
// selected endpoints.
func (r *rangeState) navigate(key keyKind) {
	switch key {
	case keyLeft:
		r.cal.moveDays(-1)
	case keyRight:
		r.cal.moveDays(1)
	case keyUp:
		r.cal.moveDays(-7)
	case keyDown:
		r.cal.moveDays(7)
	case keyPageUp:
		r.cal.shiftMonth(-1)
	case keyPageDown:
		r.cal.shiftMonth(1)
	}
}

// selectCursor applies a click or Space to the cursor date.
func (r *rangeState) selectCursor() {
	date := r.cal.cursor
	if r.haveUntil {
		r.from = date
		r.until = time.Time{}
		r.haveUntil = false
		return
	}

	r.until = date
	r.haveUntil = true
	if r.until.Before(r.from) {
		r.from, r.until = r.until, r.from
	}
}

func (r *rangeState) clickDate(date time.Time) {
	r.cal.cursor = dateOnly(date)
	r.cal.month = firstOfMonth(r.cal.cursor)
	r.selectCursor()
}

// confirm returns an ordered range. If Until has not been selected yet, the
// cursor position becomes Until.
func (r *rangeState) confirm() (time.Time, time.Time) {
	if !r.haveUntil {
		r.until = r.cal.cursor
		if r.until.Before(r.from) {
			r.from, r.until = r.until, r.from
		}
		r.haveUntil = true
	}
	return r.from, r.until
}

func (r *rangeState) requiredHeight() int {
	// One row for the range status line below the calendar.
	return r.cal.height() + 1
}

func (r *rangeState) tooSmallMessage() string {
	return fmt.Sprintf(
		"terminal too small: need at least %dx%d; PageUp/PageDown still change months",
		calendarWidth+1, r.requiredHeight(),
	)
}

func (r *rangeState) statusLine() string {
	if !r.haveUntil {
		return r.from.Format("2006-01-02") + " -> --"
	}
	return r.from.Format("2006-01-02") + " -> " + r.until.Format("2006-01-02")
}

// styleFor returns the SGR prefix for one day cell. Endpoints and in-range
// days use a background; the cursor is reverse video unless it sits on a
// selected day, where bold keeps the background visible.
func (r *rangeState) styleFor(date, today time.Time) string {
	endpoint := sameDate(date, r.from) || (r.haveUntil && sameDate(date, r.until))
	inRange := r.haveUntil && date.After(r.from) && date.Before(r.until)
	cursor := sameDate(date, r.cal.cursor)

	var codes []string
	switch {
	case cursor && !endpoint && !inRange:
		codes = append(codes, "7")
	case cursor:
		codes = append(codes, "1")
	}
	if sameDate(date, today) {
		codes = append(codes, "4")
	}
	switch {
	case endpoint:
		codes = append(codes, "44", "97")
	case inRange:
		codes = append(codes, "100", "97")
	}
	if len(codes) == 0 {
		return ""
	}
	return "\x1b[" + strings.Join(codes, ";") + "m"
}

func (r *rangeState) lines(today time.Time) []string {
	calLines := r.cal.renderLines(func(date time.Time) string {
		return r.styleFor(date, today)
	})
	return append(calLines, r.statusLine())
}
