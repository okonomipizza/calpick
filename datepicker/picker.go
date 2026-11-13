package datepicker

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrCanceled is returned when the user cancels the picker.
var ErrCanceled = errors.New("datepicker: canceled")

type config struct {
	initial time.Time
	until   time.Time
}

// Option configures Pick or PickRange. Option values are produced by this
// package (currently WithInitial and WithInitialRange); external packages
// cannot define new options.
type Option func(*config)

// WithInitial sets the date and month shown when the picker starts.
// The time of day is ignored; the zero time means today.
func WithInitial(t time.Time) Option {
	return func(c *config) {
		c.initial = t
	}
}

// WithInitialRange sets the From and Until dates shown by PickRange. A zero
// until means the same day as from; reversed dates are sorted.
func WithInitialRange(from, until time.Time) Option {
	return func(c *config) {
		c.initial = from
		c.until = until
	}
}

// Pick shows an interactive calendar on the controlling terminal and blocks
// until the user selects a date, cancels (ErrCanceled), or an error occurs.
//
// It draws on the controlling terminal, so the caller remains free to write
// the result to stdout. The terminal state is restored before Pick returns.
// The returned time is midnight in the initial date's location (local time by
// default). Pick is not safe for concurrent use.
func Pick(opts ...Option) (time.Time, error) {
	cfg := config{initial: time.Now()}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.initial.IsZero() {
		cfg.initial = time.Now()
	}

	term, err := openTerminal()
	if err != nil {
		return time.Time{}, err
	}
	defer term.Stop()

	interruptRequested.Store(false)
	resizeRequested.Store(false)
	stopSignals := watchSignals()
	defer stopSignals()

	if err := term.Start(); err != nil {
		return time.Time{}, err
	}

	return run(term, newCalendar(cfg.initial), time.Now)
}

func run(term terminal, cal calendar, now func() time.Time) (time.Time, error) {
	dirty := true
	for {
		if interruptRequested.Swap(false) {
			return time.Time{}, ErrCanceled
		}
		if resizeRequested.Swap(false) {
			dirty = true
		}

		width, height, sizeErr := term.Size()
		// Keep one spare column: writing the last calendar cell in the
		// terminal's final column behaves differently across terminals and
		// can wrap or scroll the frame. The safe height also covers months
		// that need a sixth week.
		usable := sizeErr == nil && width > calendarWidth && height >= calendarMinHeight

		if dirty {
			if sizeErr != nil {
				return time.Time{}, sizeErr
			}
			if usable {
				if err := term.Draw(frame(cal.lines(now()))); err != nil {
					return time.Time{}, err
				}
			} else {
				message := fmt.Sprintf("terminal too small: need at least %dx%d", calendarWidth+1, calendarMinHeight)
				if err := term.Draw("\x1b[H\x1b[2J" + message); err != nil {
					return time.Time{}, err
				}
			}
			dirty = false
		}

		ev, ok, err := term.PollEvent(100 * time.Millisecond)
		if err != nil {
			return time.Time{}, err
		}
		if !ok {
			continue
		}
		dirty = true

		if !usable {
			if ev.kind == evKey && ev.key == keyCancel {
				return time.Time{}, ErrCanceled
			}
			continue
		}

		switch ev.kind {
		case evResize:
			// The next iteration redraws with the new size.
		case evKey:
			switch ev.key {
			case keyCancel:
				return time.Time{}, ErrCanceled
			case keyEnter:
				return cal.cursor, nil
			case keyLeft:
				cal.moveDays(-1)
			case keyRight:
				cal.moveDays(1)
			case keyUp:
				cal.moveDays(-7)
			case keyDown:
				cal.moveDays(7)
			case keyPageUp:
				cal.shiftMonth(-1)
			case keyPageDown:
				cal.shiftMonth(1)
			}
		case evMouse:
			switch act := cal.hit(ev.row, ev.col); act.kind {
			case actionPrevMonth:
				cal.shiftMonth(-1)
			case actionNextMonth:
				cal.shiftMonth(1)
			case actionSelect:
				return act.date, nil
			}
		}
	}
}

// frame positions every row absolutely. Relying on newline movement after a
// full-width row differs between terminals (notably conhost at the last
// column), which would shift clicks onto the wrong row.
func frame(lines []string) string {
	var b strings.Builder
	for i, line := range lines {
		fmt.Fprintf(&b, "\x1b[%d;1H", i+1)
		b.WriteString(line)
		b.WriteString("\x1b[K")
	}
	b.WriteString("\x1b[J")
	return b.String()
}
