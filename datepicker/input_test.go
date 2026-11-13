package datepicker

import (
	"reflect"
	"testing"
)

func TestParseOne(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     event
		consumed int
		needMore bool
	}{
		{name: "cancel q", input: "q", want: keyEvent(keyCancel), consumed: 1},
		{name: "cancel Q", input: "Q", want: keyEvent(keyCancel), consumed: 1},
		{name: "ctrl-c", input: "\x03", want: keyEvent(keyCancel), consumed: 1},
		{name: "enter cr", input: "\r", want: keyEvent(keyEnter), consumed: 1},
		{name: "enter lf", input: "\n", want: keyEvent(keyEnter), consumed: 1},
		{name: "left h", input: "h", want: keyEvent(keyLeft), consumed: 1},
		{name: "right l", input: "l", want: keyEvent(keyRight), consumed: 1},
		{name: "up k", input: "k", want: keyEvent(keyUp), consumed: 1},
		{name: "down j", input: "j", want: keyEvent(keyDown), consumed: 1},
		{name: "arrow up", input: "\x1b[A", want: keyEvent(keyUp), consumed: 3},
		{name: "arrow down", input: "\x1b[B", want: keyEvent(keyDown), consumed: 3},
		{name: "arrow right", input: "\x1b[C", want: keyEvent(keyRight), consumed: 3},
		{name: "arrow left", input: "\x1b[D", want: keyEvent(keyLeft), consumed: 3},
		{name: "application up", input: "\x1bOA", want: keyEvent(keyUp), consumed: 3},
		{name: "modified right", input: "\x1b[1;5C", want: keyEvent(keyRight), consumed: 6},
		{name: "page up", input: "\x1b[5~", want: keyEvent(keyPageUp), consumed: 4},
		{name: "page down", input: "\x1b[6~", want: keyEvent(keyPageDown), consumed: 4},
		{name: "mouse press", input: "\x1b[<0;12;5M", want: event{kind: evMouse, row: 5, col: 12}, consumed: 10},
		{name: "mouse release", input: "\x1b[<0;12;5m", want: event{kind: evIgnore}, consumed: 10},
		{name: "mouse right button", input: "\x1b[<2;12;5M", want: event{kind: evIgnore}, consumed: 10},
		{name: "mouse wheel", input: "\x1b[<64;12;5M", want: event{kind: evIgnore}, consumed: 11},
		{name: "x10 mouse", input: "\x1b[M\x71\x21\x21", want: event{kind: evIgnore}, consumed: 6},
		{name: "x10 left press", input: "\x1b[M\x20\x21\x21", want: event{kind: evMouse, row: 1, col: 1}, consumed: 6},
		{name: "x10 motion", input: "\x1b[M\x40\x21\x21", want: event{kind: evIgnore}, consumed: 6},
		{name: "x10 wheel up", input: "\x1b[M\x60\x21\x21", want: event{kind: evIgnore}, consumed: 6},
		{name: "x10 wheel down", input: "\x1b[M\x61\x21\x21", want: event{kind: evIgnore}, consumed: 6},
		{name: "x10 shift click", input: "\x1b[M\x24\x21\x21", want: event{kind: evIgnore}, consumed: 6},
		{name: "alt q is ignored", input: "\x1bq", want: event{kind: evIgnore}, consumed: 2},
		{name: "unknown csi", input: "\x1b[1;2H", want: event{kind: evIgnore}, consumed: 6},
		{name: "unknown byte", input: "x", want: event{kind: evIgnore}, consumed: 1},
		{name: "lone esc", input: "\x1b", needMore: true},
		{name: "csi start", input: "\x1b[", needMore: true},
		{name: "partial mouse", input: "\x1b[<0;1;2", needMore: true},
		{name: "mouse coords", input: "\x1b[<0;1;2M", want: event{kind: evMouse, row: 2, col: 1}, consumed: 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, consumed, needMore := parseOne([]byte(tt.input))
			if needMore != tt.needMore {
				t.Fatalf("needMore = %v, want %v", needMore, tt.needMore)
			}
			if needMore {
				return
			}
			if consumed != tt.consumed {
				t.Fatalf("consumed = %d, want %d", consumed, tt.consumed)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("event = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseOnePartialSequence(t *testing.T) {
	input := []byte("\x1b[<0;12;5M")
	for i := 1; i < len(input); i++ {
		_, consumed, needMore := parseOne(input[:i])
		if !needMore {
			t.Fatalf("prefix %q: needMore = false, consumed = %d", input[:i], consumed)
		}
	}
}

func TestX10CoordinatesAreNotKeys(t *testing.T) {
	input := []byte("\x1b[M\x71\x21\x21q")
	ev, consumed, needMore := parseOne(input)
	if needMore || ev.kind != evIgnore || consumed != 6 {
		t.Fatalf("first event = %+v consumed=%d needMore=%v, want ignore/6/false", ev, consumed, needMore)
	}

	ev, consumed, needMore = parseOne(input[consumed:])
	if needMore || consumed != 1 || ev.kind != evKey || ev.key != keyCancel {
		t.Fatalf("second event = %+v consumed=%d needMore=%v, want cancel", ev, consumed, needMore)
	}
}

func FuzzParseOne(f *testing.F) {
	seeds := []string{
		"q",
		"\x03",
		"\r",
		"\x1b",
		"\x1b[A",
		"\x1b[5~",
		"\x1b[<0;12;5M",
		"\x1b[M\x20\x21\x21",
		"\x1b[<0;12;5",
	}
	for _, seed := range seeds {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		_, consumed, needMore := parseOne(data)
		if consumed < 0 || consumed > len(data) {
			t.Fatalf("consumed = %d for %d bytes", consumed, len(data))
		}
		if !needMore && consumed == 0 {
			t.Fatal("complete parse consumed no bytes")
		}
	})
}
