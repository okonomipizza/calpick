package datepicker

import "time"

const (
	// calendarWidth is the visible width of the calendar in terminal columns.
	calendarWidth = 27
	// calendarMinHeight is header + weekday + six week rows.
	calendarMinHeight = 8
	// cellWidth is the pitch of one day cell: three characters plus a gap.
	cellWidth = 4
)

// calendar is the pure state of the month view. All times are local midnight.
type calendar struct {
	month  time.Time // first day of the displayed month
	cursor time.Time // highlighted date
}

func newCalendar(t time.Time) calendar {
	cursor := dateOnly(t)
	return calendar{month: firstOfMonth(cursor), cursor: cursor}
}

func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func firstOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func daysInMonth(year int, month time.Month, loc *time.Location) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, loc).Day()
}

// weeks lays the month out in Sunday-first rows. Zero means an empty cell.
func (c *calendar) weeks() [][]int {
	loc := c.month.Location()
	leading := int(c.month.Weekday()) // time.Sunday == 0
	days := daysInMonth(c.month.Year(), c.month.Month(), loc)
	rows := (leading + days + 6) / 7

	weeks := make([][]int, rows)
	for i := range weeks {
		weeks[i] = make([]int, 7)
	}
	for day := 1; day <= days; day++ {
		index := leading + day - 1
		weeks[index/7][index%7] = day
	}
	return weeks
}

// shiftMonth changes the displayed month and keeps the cursor day when it
// exists in the new month.
func (c *calendar) shiftMonth(delta int) {
	month := firstOfMonth(c.month.AddDate(0, delta, 0))
	day := c.cursor.Day()
	if last := daysInMonth(month.Year(), month.Month(), month.Location()); day > last {
		day = last
	}
	c.month = month
	c.cursor = time.Date(month.Year(), month.Month(), day, 0, 0, 0, 0, month.Location())
}

// moveDays moves the cursor and follows it across month boundaries.
func (c *calendar) moveDays(delta int) {
	c.cursor = c.cursor.AddDate(0, 0, delta)
	c.month = firstOfMonth(c.cursor)
}

func (c *calendar) dateAt(row, col int) (time.Time, bool) {
	weeks := c.weeks()
	if row < 0 || row >= len(weeks) || col < 0 || col >= 7 {
		return time.Time{}, false
	}
	day := weeks[row][col]
	if day == 0 {
		return time.Time{}, false
	}
	return time.Date(c.month.Year(), c.month.Month(), day, 0, 0, 0, 0, c.month.Location()), true
}

// height is the number of terminal rows the current month needs.
func (c *calendar) height() int {
	return 2 + len(c.weeks())
}

func sameDate(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}
