package datepicker

import (
	"fmt"
	"strings"
	"time"
)

const weekdayHeader = "Sun Mon Tue Wed Thu Fri Sat"

type actionKind int

const (
	actionNone actionKind = iota
	actionPrevMonth
	actionNextMonth
	actionSelect
)

type action struct {
	kind actionKind
	date time.Time
}

// plainLines renders the frame without escape sequences. It is used by tests
// and by documentation.
func (c *calendar) plainLines() []string {
	lines := make([]string, 0, c.height())
	lines = append(lines, c.headerLine(), weekdayHeader)
	for _, week := range c.weeks() {
		var b strings.Builder
		for i, day := range week {
			if i > 0 {
				b.WriteByte(' ')
			}
			if day == 0 {
				b.WriteString("   ")
				continue
			}
			fmt.Fprintf(&b, "%3d", day)
		}
		lines = append(lines, b.String())
	}
	return lines
}

// lines renders the single-date frame with SGR styles. Cursor and today never
// change the visible width because the escape sequences are zero-width.
func (c *calendar) lines(now time.Time) []string {
	today := dateOnly(now)
	return c.renderLines(func(date time.Time) string {
		switch {
		case sameDate(date, c.cursor) && sameDate(date, today):
			return "\x1b[1;4;7m"
		case sameDate(date, c.cursor):
			return "\x1b[7m"
		case sameDate(date, today):
			return "\x1b[4m"
		}
		return ""
	})
}

// renderLines renders the frame using style for each day cell. A nil style
// renders plain text.
func (c *calendar) renderLines(style func(time.Time) string) []string {
	loc := c.month.Location()

	lines := make([]string, 0, c.height())
	lines = append(lines, c.headerLine(), weekdayHeader)
	for _, week := range c.weeks() {
		var b strings.Builder
		for i, day := range week {
			if i > 0 {
				b.WriteByte(' ')
			}
			if day == 0 {
				b.WriteString("   ")
				continue
			}

			date := time.Date(c.month.Year(), c.month.Month(), day, 0, 0, 0, 0, loc)
			cell := fmt.Sprintf("%3d", day)
			code := ""
			if style != nil {
				code = style(date)
			}
			if code == "" {
				b.WriteString(cell)
				continue
			}
			b.WriteString(code)
			b.WriteString(cell)
			b.WriteString("\x1b[0m")
		}
		lines = append(lines, b.String())
	}
	return lines
}

func (c *calendar) headerLine() string {
	title := c.month.Format("2006 Jan")
	buf := []byte(strings.Repeat(" ", calendarWidth))
	buf[0] = '<'
	buf[calendarWidth-1] = '>'
	copy(buf[(calendarWidth-len(title))/2:], title)
	return string(buf)
}

// hit maps a 1-based terminal position to a calendar action. The frame is
// always drawn at the top-left corner of the alternate screen.
func (c *calendar) hit(row, col int) action {
	if row < 1 || col < 1 || col > calendarWidth {
		return action{}
	}

	switch row {
	case 1:
		switch {
		case col <= 3:
			return action{kind: actionPrevMonth}
		case col >= calendarWidth-2:
			return action{kind: actionNextMonth}
		}
		return action{}
	case 2:
		return action{}
	}

	cell := (col - 1) / cellWidth
	if cell > 6 {
		return action{}
	}
	date, ok := c.dateAt(row-3, cell)
	if !ok {
		return action{}
	}
	return action{kind: actionSelect, date: date}
}
