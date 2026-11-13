package datepicker

import (
	"errors"
	"testing"
	"time"
)

// fakeTerminal feeds scripted events to run and records drawn frames.
type fakeTerminal struct {
	width  int
	height int
	events []event
	frames []string
}

func (f *fakeTerminal) Start() error { return nil }
func (f *fakeTerminal) Stop()        {}

func (f *fakeTerminal) Size() (int, int, error) {
	return f.width, f.height, nil
}

func (f *fakeTerminal) Draw(frame string) error {
	f.frames = append(f.frames, frame)
	return nil
}

func (f *fakeTerminal) PollEvent(time.Duration) (event, bool, error) {
	if len(f.events) == 0 {
		return event{}, false, nil
	}
	ev := f.events[0]
	f.events = f.events[1:]
	return ev, true, nil
}

func fixedNow() time.Time {
	return time.Date(2026, time.November, 15, 12, 0, 0, 0, time.UTC)
}

func newFake(events ...event) *fakeTerminal {
	return &fakeTerminal{width: 80, height: 24, events: events}
}

func resetFlags() {
	interruptRequested.Store(false)
	resizeRequested.Store(false)
}

func TestRunSelectsClickedDate(t *testing.T) {
	resetFlags()
	term := newFake(event{kind: evMouse, row: 5, col: 1})

	got, err := run(term, newCalendar(fixedNow()), fixedNow)
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	if want := time.Date(2026, time.November, 15, 0, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Fatalf("selected %v, want %v", got, want)
	}
	if len(term.frames) == 0 {
		t.Fatal("no frame was drawn")
	}
}

func TestRunNextMonthThenSelect(t *testing.T) {
	resetFlags()
	term := newFake(
		event{kind: evMouse, row: 1, col: 27},
		event{kind: evMouse, row: 3, col: 9},
	)

	got, err := run(term, newCalendar(fixedNow()), fixedNow)
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	if want := time.Date(2026, time.December, 1, 0, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Fatalf("selected %v, want %v", got, want)
	}
	if len(term.frames) < 2 {
		t.Fatalf("expected a redraw after changing month, got %d frames", len(term.frames))
	}
}

func TestRunEnterSelectsCursor(t *testing.T) {
	resetFlags()
	term := newFake(
		event{kind: evKey, key: keyRight},
		event{kind: evKey, key: keyEnter},
	)

	got, err := run(term, newCalendar(fixedNow()), fixedNow)
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	if want := time.Date(2026, time.November, 16, 0, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Fatalf("selected %v, want %v", got, want)
	}
}

func TestRunCancel(t *testing.T) {
	resetFlags()
	term := newFake(event{kind: evKey, key: keyCancel})

	_, err := run(term, newCalendar(fixedNow()), fixedNow)
	if !errors.Is(err, ErrCanceled) {
		t.Fatalf("error = %v, want ErrCanceled", err)
	}
}

func TestRunTooSmall(t *testing.T) {
	resetFlags()
	term := newFake(event{kind: evKey, key: keyCancel})
	term.width, term.height = 20, 5

	_, err := run(term, newCalendar(fixedNow()), fixedNow)
	if !errors.Is(err, ErrCanceled) {
		t.Fatalf("error = %v, want ErrCanceled", err)
	}
	if len(term.frames) != 1 {
		t.Fatalf("expected one message frame, got %d", len(term.frames))
	}
}

func TestRunRejectsExactCalendarWidth(t *testing.T) {
	resetFlags()
	term := newFake(event{kind: evKey, key: keyCancel})
	term.width = calendarWidth

	if _, err := run(term, newCalendar(fixedNow()), fixedNow); !errors.Is(err, ErrCanceled) {
		t.Fatalf("error = %v, want ErrCanceled", err)
	}
	if len(term.frames) != 1 {
		t.Fatalf("expected one message frame, got %d", len(term.frames))
	}
}

func TestRunRejectsTooFewRows(t *testing.T) {
	resetFlags()
	term := newFake(event{kind: evKey, key: keyCancel})
	term.height = calendarMinHeight - 1

	if _, err := run(term, newCalendar(fixedNow()), fixedNow); !errors.Is(err, ErrCanceled) {
		t.Fatalf("error = %v, want ErrCanceled", err)
	}
	if len(term.frames) != 1 {
		t.Fatalf("expected one message frame, got %d", len(term.frames))
	}
}

func TestRunInterruptFlag(t *testing.T) {
	resetFlags()
	interruptRequested.Store(true)

	_, err := run(newFake(), newCalendar(fixedNow()), fixedNow)
	if !errors.Is(err, ErrCanceled) {
		t.Fatalf("error = %v, want ErrCanceled", err)
	}
}

func TestFrameUsesAbsolutePositioning(t *testing.T) {
	got := frame([]string{"one", "two"})
	want := "\x1b[1;1Hone\x1b[K\x1b[2;1Htwo\x1b[K\x1b[J"
	if got != want {
		t.Fatalf("frame = %q, want %q", got, want)
	}
}
