//go:build linux || darwin

package datepicker

import (
	"os"
	"testing"
	"time"
)

func newPipeTerminal(t *testing.T) (*unixTerminal, *os.File) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	t.Cleanup(func() {
		r.Close()
		w.Close()
	})
	return &unixTerminal{in: r}, w
}

func TestPollEventRetainsSplitSequence(t *testing.T) {
	term, w := newPipeTerminal(t)

	if _, err := w.WriteString("\x1b[<0;1;"); err != nil {
		t.Fatalf("write: %v", err)
	}
	ev, ok, err := term.PollEvent(20 * time.Millisecond)
	if err != nil {
		t.Fatalf("first poll: %v", err)
	}
	if ok {
		t.Fatalf("first poll returned %+v, want timeout", ev)
	}
	if len(term.buf) == 0 {
		t.Fatal("partial sequence was discarded")
	}

	if _, err := w.WriteString("5M"); err != nil {
		t.Fatalf("write: %v", err)
	}
	ev, ok, err = term.PollEvent(200 * time.Millisecond)
	if err != nil {
		t.Fatalf("second poll: %v", err)
	}
	if !ok {
		t.Fatal("second poll timed out")
	}
	if want := (event{kind: evMouse, row: 5, col: 1}); ev != want {
		t.Fatalf("event = %+v, want %+v", ev, want)
	}
}

func TestPollEventLoneEscapeCancels(t *testing.T) {
	term, w := newPipeTerminal(t)

	if _, err := w.WriteString("\x1b"); err != nil {
		t.Fatalf("write: %v", err)
	}
	ev, ok, err := term.PollEvent(200 * time.Millisecond)
	if err != nil {
		t.Fatalf("poll: %v", err)
	}
	if !ok {
		t.Fatal("poll timed out")
	}
	if want := keyEvent(keyCancel); ev != want {
		t.Fatalf("event = %+v, want %+v", ev, want)
	}
}

func TestPollEventDropsStalePartial(t *testing.T) {
	term, _ := newPipeTerminal(t)
	term.buf = []byte("\x1b[<0;1;")
	term.partialSince = time.Now().Add(-time.Second)

	ev, ok, err := term.PollEvent(10 * time.Millisecond)
	if err != nil {
		t.Fatalf("poll: %v", err)
	}
	if !ok {
		t.Fatal("poll timed out")
	}
	if ev.kind != evIgnore {
		t.Fatalf("event = %+v, want ignore", ev)
	}
	if len(term.buf) != 0 {
		t.Fatalf("buffer still holds %d bytes", len(term.buf))
	}
}
