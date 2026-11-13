package datepicker

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestRunRangeConfirmsDefault(t *testing.T) {
	resetFlags()
	term := newFake(event{kind: evKey, key: keyEnter})
	state := newRangeState(fixedNow(), time.Time{})

	from, until, err := runRange(term, state, fixedNow)
	if err != nil {
		t.Fatalf("runRange returned error: %v", err)
	}
	if want := day(2026, time.November, 15); !from.Equal(want) {
		t.Fatalf("from = %v, want %v", from, want)
	}
	if want := day(2026, time.November, 15); !until.Equal(want) {
		t.Fatalf("until = %v, want %v", until, want)
	}
	if len(term.frames) == 0 {
		t.Fatal("no frame was drawn")
	}
}

func TestRunRangeClicks(t *testing.T) {
	resetFlags()
	term := newFake(
		event{kind: evMouse, row: 4, col: 9},  // November 10
		event{kind: evMouse, row: 5, col: 21}, // November 20
		event{kind: evKey, key: keyEnter},
	)
	state := newRangeState(fixedNow(), time.Time{})

	from, until, err := runRange(term, state, fixedNow)
	if err != nil {
		t.Fatalf("runRange returned error: %v", err)
	}
	if want := day(2026, time.November, 10); !from.Equal(want) {
		t.Fatalf("from = %v, want %v", from, want)
	}
	if want := day(2026, time.November, 20); !until.Equal(want) {
		t.Fatalf("until = %v, want %v", until, want)
	}
}

func TestRunRangeEnterUsesCursorWhenUntilMissing(t *testing.T) {
	resetFlags()
	term := newFake(
		event{kind: evMouse, row: 4, col: 9}, // From: November 10
		event{kind: evKey, key: keyRight},
		event{kind: evKey, key: keyRight}, // cursor: November 12
		event{kind: evKey, key: keyEnter},
	)
	state := newRangeState(fixedNow(), time.Time{})

	from, until, err := runRange(term, state, fixedNow)
	if err != nil {
		t.Fatalf("runRange returned error: %v", err)
	}
	if want := day(2026, time.November, 10); !from.Equal(want) {
		t.Fatalf("from = %v, want %v", from, want)
	}
	if want := day(2026, time.November, 12); !until.Equal(want) {
		t.Fatalf("until = %v, want %v", until, want)
	}
}

func TestRunRangeSpaceSelects(t *testing.T) {
	resetFlags()
	term := newFake(
		event{kind: evKey, key: keySelect}, // start a new range at the cursor (November 15)
		event{kind: evKey, key: keyRight},  // cursor: November 16
		event{kind: evKey, key: keySelect}, // Until: November 16
		event{kind: evKey, key: keyEnter},
	)
	state := newRangeState(fixedNow(), time.Time{})

	from, until, err := runRange(term, state, fixedNow)
	if err != nil {
		t.Fatalf("runRange returned error: %v", err)
	}
	if want := day(2026, time.November, 15); !from.Equal(want) {
		t.Fatalf("from = %v, want %v", from, want)
	}
	if want := day(2026, time.November, 16); !until.Equal(want) {
		t.Fatalf("until = %v, want %v", until, want)
	}
}

func TestRunRangeHeaderNavigationKeepsEndpoints(t *testing.T) {
	resetFlags()
	term := newFake(
		event{kind: evMouse, row: 1, col: 27}, // next month
		event{kind: evKey, key: keyEnter},
	)
	state := newRangeState(fixedNow(), time.Time{})

	from, until, err := runRange(term, state, fixedNow)
	if err != nil {
		t.Fatalf("runRange returned error: %v", err)
	}
	if want := day(2026, time.November, 15); !from.Equal(want) || !until.Equal(want) {
		t.Fatalf("range = %v..%v, want both %v", from, until, want)
	}
}

func TestRunRangeRecoversFromTooSmall(t *testing.T) {
	resetFlags()
	// August 2026 needs 8 calendar rows plus the status line, so 28x8 shows
	// the too-small message. PageUp moves the cursor to July (7 rows + status),
	// which fits; the endpoints stay unchanged.
	term := newFake(
		event{kind: evKey, key: keyPageUp},
		event{kind: evKey, key: keyEnter},
	)
	term.width, term.height = 28, 8
	state := newRangeState(day(2026, time.August, 15), day(2026, time.September, 15))

	from, until, err := runRange(term, state, fixedNow)
	if err != nil {
		t.Fatalf("runRange returned error: %v", err)
	}
	if want := day(2026, time.August, 15); !from.Equal(want) {
		t.Fatalf("from = %v, want %v", from, want)
	}
	if want := day(2026, time.September, 15); !until.Equal(want) {
		t.Fatalf("until = %v, want %v", until, want)
	}
	if len(term.frames) < 2 {
		t.Fatalf("expected the message frame and a calendar frame, got %d frames", len(term.frames))
	}
}

func TestRunRangeShowsPendingUntil(t *testing.T) {
	resetFlags()
	term := newFake(
		event{kind: evMouse, row: 4, col: 9}, // From: November 10
		event{kind: evKey, key: keyEnter},
	)
	state := newRangeState(fixedNow(), time.Time{})

	if _, _, err := runRange(term, state, fixedNow); err != nil {
		t.Fatalf("runRange returned error: %v", err)
	}

	for _, frame := range term.frames {
		if strings.Contains(frame, "2026-11-10 -> --") {
			return
		}
	}
	t.Fatalf("no frame showed the pending Until status: %q", term.frames)
}

func TestRunRangeThirdClickStartsOver(t *testing.T) {
	resetFlags()
	term := newFake(
		event{kind: evMouse, row: 4, col: 9},  // November 10: From
		event{kind: evMouse, row: 5, col: 21}, // November 20: Until
		event{kind: evMouse, row: 6, col: 13}, // November 25: new From
		event{kind: evKey, key: keyEnter},
	)
	state := newRangeState(fixedNow(), time.Time{})

	from, until, err := runRange(term, state, fixedNow)
	if err != nil {
		t.Fatalf("runRange returned error: %v", err)
	}
	if want := day(2026, time.November, 25); !from.Equal(want) || !until.Equal(want) {
		t.Fatalf("range = %v..%v, want both %v", from, until, want)
	}
}

func TestRunRangeTooSmallWidth(t *testing.T) {
	resetFlags()
	term := newFake(event{kind: evKey, key: keyCancel})
	term.width = 27
	state := newRangeState(fixedNow(), time.Time{})

	if _, _, err := runRange(term, state, fixedNow); !errors.Is(err, ErrCanceled) {
		t.Fatalf("error = %v, want ErrCanceled", err)
	}
	if len(term.frames) != 1 {
		t.Fatalf("expected one message frame, got %d", len(term.frames))
	}
}

func TestRunRangeFitsFourWeekMonth(t *testing.T) {
	resetFlags()
	term := newFake(event{kind: evKey, key: keyEnter})
	term.width, term.height = 28, 7
	state := newRangeState(day(2026, time.February, 10), time.Time{})

	from, until, err := runRange(term, state, fixedNow)
	if err != nil {
		t.Fatalf("runRange returned error: %v", err)
	}
	if want := day(2026, time.February, 10); !from.Equal(want) || !until.Equal(want) {
		t.Fatalf("range = %v..%v, want both %v", from, until, want)
	}
	if len(term.frames) != 1 || strings.Contains(term.frames[0], "too small") {
		t.Fatalf("expected a calendar frame, got %q", term.frames)
	}
}

func TestRunRangeTooSmallCancel(t *testing.T) {
	resetFlags()
	term := newFake(event{kind: evKey, key: keyCancel})
	term.width, term.height = 28, 5
	state := newRangeState(fixedNow(), time.Time{})

	if _, _, err := runRange(term, state, fixedNow); !errors.Is(err, ErrCanceled) {
		t.Fatalf("error = %v, want ErrCanceled", err)
	}
	if len(term.frames) != 1 {
		t.Fatalf("expected one message frame, got %d", len(term.frames))
	}
}

func TestRunRangeCancel(t *testing.T) {
	resetFlags()
	term := newFake(event{kind: evKey, key: keyCancel})
	state := newRangeState(fixedNow(), time.Time{})

	if _, _, err := runRange(term, state, fixedNow); !errors.Is(err, ErrCanceled) {
		t.Fatalf("error = %v, want ErrCanceled", err)
	}
}

func TestRunRangeInterruptFlag(t *testing.T) {
	resetFlags()
	interruptRequested.Store(true)

	if _, _, err := runRange(newFake(), newRangeState(fixedNow(), time.Time{}), fixedNow); !errors.Is(err, ErrCanceled) {
		t.Fatalf("error = %v, want ErrCanceled", err)
	}
}
