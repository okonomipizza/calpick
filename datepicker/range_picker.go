package datepicker

import (
	"time"
)

// PickRange shows one interactive calendar on the controlling terminal and
// blocks until the user presses Enter, cancels (ErrCanceled), or an error
// occurs.
//
// Click a date to set From, then click another date to set Until. Clicking
// again starts a new range. The range is shown below the calendar and Enter
// confirms it; if Until has not been selected, the cursor date becomes Until.
// The returned times are midnight in the initial date's location.
func PickRange(opts ...Option) (from, until time.Time, err error) {
	cfg := config{initial: time.Now()}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.initial.IsZero() {
		cfg.initial = time.Now()
	}
	state := newRangeState(cfg.initial, cfg.until)

	term, err := openTerminal()
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	defer term.Stop()

	interruptRequested.Store(false)
	resizeRequested.Store(false)
	stopSignals := watchSignals()
	defer stopSignals()

	if err := term.Start(); err != nil {
		return time.Time{}, time.Time{}, err
	}

	return runRange(term, state, time.Now)
}

func runRange(term terminal, r rangeState, now func() time.Time) (time.Time, time.Time, error) {
	dirty := true
	for {
		if interruptRequested.Swap(false) {
			return time.Time{}, time.Time{}, ErrCanceled
		}
		if resizeRequested.Swap(false) {
			dirty = true
		}

		width, height, sizeErr := term.Size()
		fits := sizeErr == nil && width > calendarWidth && height >= r.requiredHeight()

		if dirty {
			if sizeErr != nil {
				return time.Time{}, time.Time{}, sizeErr
			}
			if fits {
				if err := term.Draw(frame(r.lines(dateOnly(now())))); err != nil {
					return time.Time{}, time.Time{}, err
				}
			} else {
				if err := term.Draw("\x1b[H\x1b[2J" + r.tooSmallMessage()); err != nil {
					return time.Time{}, time.Time{}, err
				}
			}
			dirty = false
		}

		ev, ok, err := term.PollEvent(100 * time.Millisecond)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		if !ok {
			continue
		}
		dirty = true

		if !fits {
			// The required height depends on the displayed month, so allow
			// navigation even while the too-small message is shown.
			if ev.kind == evKey {
				switch ev.key {
				case keyCancel:
					return time.Time{}, time.Time{}, ErrCanceled
				case keyLeft, keyRight, keyUp, keyDown, keyPageUp, keyPageDown:
					r.navigate(ev.key)
				}
			}
			continue
		}

		switch ev.kind {
		case evResize:
			// The next iteration redraws with the new size.
		case evKey:
			switch ev.key {
			case keyCancel:
				return time.Time{}, time.Time{}, ErrCanceled
			case keyEnter:
				from, until := r.confirm()
				return from, until, nil
			case keySelect:
				r.selectCursor()
			default:
				r.navigate(ev.key)
			}
		case evMouse:
			switch act := r.cal.hit(ev.row, ev.col); act.kind {
			case actionPrevMonth:
				r.cal.shiftMonth(-1)
			case actionNextMonth:
				r.cal.shiftMonth(1)
			case actionSelect:
				r.clickDate(act.date)
			}
		}
	}
}
