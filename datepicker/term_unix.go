//go:build linux || darwin

package datepicker

import (
	"errors"
	"io"
	"os"
	"syscall"
	"time"
	"unsafe"
)

type unixTerminal struct {
	in           *os.File
	out          *os.File
	owned        bool // in was opened by us and must be closed
	started      bool
	restore      func()
	buf          []byte
	partialSince time.Time
}

var errFDTooLarge = errors.New("datepicker: file descriptor is out of range for select")

func openTerminal() (terminal, error) {
	if f, err := os.OpenFile("/dev/tty", os.O_RDWR, 0); err == nil {
		if isTerminal(f.Fd()) {
			return &unixTerminal{in: f, out: f, owned: true}, nil
		}
		f.Close()
	}
	// Keep stdout clean for the result by drawing on stderr, but only when
	// both standard streams are terminals.
	if isTerminal(os.Stdin.Fd()) && isTerminal(os.Stderr.Fd()) {
		return &unixTerminal{in: os.Stdin, out: os.Stderr}, nil
	}
	return nil, errors.New("datepicker: no terminal available; run from an interactive terminal")
}

func (t *unixTerminal) Start() error {
	restore, err := rawMode(t.in.Fd())
	if err != nil {
		return err
	}
	t.restore = restore
	// Mark the terminal started before writing so a partially written enter
	// sequence is still undone by Stop.
	t.started = true
	if _, err := t.out.WriteString(enterScreen + hideCursor + enableMouse); err != nil {
		restore()
		t.restore = nil
		return err
	}
	return nil
}

func (t *unixTerminal) Stop() {
	if t.started && t.out != nil {
		t.out.WriteString(disableMouse + showCursor + leaveScreen)
		t.started = false
	}
	if t.restore != nil {
		t.restore()
		t.restore = nil
	}
	if t.owned && t.in != nil {
		t.in.Close()
		t.owned = false
	}
}

func (t *unixTerminal) Size() (int, int, error) {
	return terminalSize(t.in.Fd())
}

func (t *unixTerminal) Draw(frame string) error {
	_, err := t.out.WriteString(frame)
	return err
}

func (t *unixTerminal) PollEvent(timeout time.Duration) (event, bool, error) {
	// Guard against unbounded buffering if a terminal emits a malformed,
	// never-terminating sequence.
	if len(t.buf) > 64 {
		t.buf = nil
		t.partialSince = time.Time{}
	}
	deadline := time.Now().Add(timeout)

	for {
		if len(t.buf) == 0 {
			remaining := time.Until(deadline)
			if remaining <= 0 {
				return event{}, false, nil
			}
			readable, err := waitReadable(t.in.Fd(), remaining)
			if err != nil {
				return event{}, false, err
			}
			if !readable {
				return event{}, false, nil
			}
			if err := t.fill(); err != nil {
				return event{}, false, err
			}
			continue
		}

		ev, consumed, needMore := parseOne(t.buf)
		if consumed > 0 {
			t.buf = t.buf[consumed:]
			t.partialSince = time.Time{}
			return ev, true, nil
		}
		if !needMore {
			t.buf = t.buf[1:]
			return event{kind: evIgnore}, true, nil
		}

		if len(t.buf) == 1 && t.buf[0] == escByte {
			// A lone ESC is the cancel key, but wait briefly for a sequence.
			readable, err := waitReadable(t.in.Fd(), 50*time.Millisecond)
			if err != nil {
				return event{}, false, err
			}
			if !readable {
				t.buf = nil
				return keyEvent(keyCancel), true, nil
			}
		} else {
			if t.partialSince.IsZero() {
				t.partialSince = time.Now()
			}
			if time.Since(t.partialSince) > 500*time.Millisecond {
				// Read anything already queued before giving up: the tail may
				// complete the sequence even though the head is stale.
				if readable, err := waitReadable(t.in.Fd(), 0); err == nil && readable {
					if err := t.fill(); err != nil {
						return event{}, false, err
					}
					continue
				}
				// A real terminal sends sequences in one write. Do not let a
				// stale partial sequence swallow the next keystrokes forever.
				t.buf = nil
				t.partialSince = time.Time{}
				return event{kind: evIgnore}, true, nil
			}
			remaining := time.Until(deadline)
			if remaining <= 0 {
				// Keep the partial sequence for the next poll: splitting a
				// sequence across reads must not turn its tail into keys.
				return event{}, false, nil
			}
			readable, err := waitReadable(t.in.Fd(), remaining)
			if err != nil {
				return event{}, false, err
			}
			if !readable {
				return event{}, false, nil
			}
		}

		if err := t.fill(); err != nil {
			return event{}, false, err
		}
	}
}

func (t *unixTerminal) fill() error {
	var buf [256]byte
	n, err := t.in.Read(buf[:])
	if n > 0 {
		t.buf = append(t.buf, buf[:n]...)
	}
	if n == 0 && err == nil {
		// A hung-up terminal may report readable with an empty read.
		return io.EOF
	}
	if err != nil && n == 0 {
		// A signal interrupts Read; let the caller re-check its flags.
		if err == syscall.EINTR {
			return nil
		}
		return err
	}
	return nil
}

func ioctl(fd, request uintptr, arg unsafe.Pointer) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, request, uintptr(arg))
	if errno != 0 {
		return errno
	}
	return nil
}
