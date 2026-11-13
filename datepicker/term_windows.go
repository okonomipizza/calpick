//go:build windows

package datepicker

import (
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"
)

const (
	enableProcessedInput = 0x0001
	enableLineInput      = 0x0002
	enableEchoInput      = 0x0004
	enableWindowInput    = 0x0008
	enableMouseInput     = 0x0010
	enableQuickEditMode  = 0x0040
	enableExtendedFlags  = 0x0080
	enableVTProcessing   = 0x0004

	keyEventType          = 0x0001
	mouseEventType        = 0x0002
	windowBufferSizeType  = 0x0004
	leftMostButtonPressed = 0x0001
	doubleClickMouseFlag  = 0x0002
	vkPrior               = 0x21
	vkNext                = 0x22
	vkLeft                = 0x25
	vkUp                  = 0x26
	vkRight               = 0x27
	vkDown                = 0x28
	vkReturn              = 0x0D
	vkEscape              = 0x1B

	waitObject0 = 0x00000000
	waitTimeout = 0x00000102
)

var (
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode             = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode             = kernel32.NewProc("SetConsoleMode")
	procReadConsoleInputW          = kernel32.NewProc("ReadConsoleInputW")
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
	procWaitForSingleObject        = kernel32.NewProc("WaitForSingleObject")
)

type coord struct {
	x int16
	y int16
}

type smallRect struct {
	left   int16
	top    int16
	right  int16
	bottom int16
}

type consoleScreenBufferInfo struct {
	size              coord
	cursorPosition    coord
	attributes        uint16
	window            smallRect
	maximumWindowSize coord
}

// inputRecord mirrors the Windows INPUT_RECORD: a 16-bit type followed by a
// 16-byte union. Keeping the union 4-byte aligned lets the event structs be
// cast from it safely.
type inputRecord struct {
	eventType uint16
	_         uint16
	event     [4]uint32
}

type keyEventRecord struct {
	keyDown         int32
	repeatCount     uint16
	virtualKeyCode  uint16
	virtualScanCode uint16
	unicodeChar     uint16
	controlKeyState uint32
}

type mouseEventRecord struct {
	x               int16
	y               int16
	buttonState     uint32
	controlKeyState uint32
	eventFlags      uint32
}

type winTerminal struct {
	in            *os.File
	out           *os.File
	ownedIn       bool
	ownedOut      bool
	oldInputMode  uint32
	oldOutputMode uint32
	started       bool

	records [16]inputRecord
	count   int
	next    int
}

func openTerminal() (terminal, error) {
	in, inErr := os.OpenFile("CONIN$", os.O_RDWR, 0)
	out, outErr := os.OpenFile("CONOUT$", os.O_RDWR, 0)
	if inErr == nil && outErr == nil {
		return &winTerminal{in: in, out: out, ownedIn: true, ownedOut: true}, nil
	}
	if in != nil {
		in.Close()
	}
	if out != nil {
		out.Close()
	}
	// Fall back to the standard streams; Start rejects them if they are not
	// consoles (for example when stdout is redirected).
	return &winTerminal{in: os.Stdin, out: os.Stdout}, nil
}

func (t *winTerminal) Start() error {
	inHandle := syscall.Handle(t.in.Fd())
	outHandle := syscall.Handle(t.out.Fd())

	inputMode, err := getConsoleMode(inHandle)
	if err != nil {
		return fmt.Errorf("datepicker: input is not a console: %w", err)
	}
	outputMode, err := getConsoleMode(outHandle)
	if err != nil {
		return fmt.Errorf("datepicker: output is not a console: %w", err)
	}

	newInputMode := inputMode &^ (enableProcessedInput | enableLineInput | enableEchoInput | enableQuickEditMode)
	newInputMode |= enableExtendedFlags | enableMouseInput | enableWindowInput
	if err := setConsoleMode(inHandle, newInputMode); err != nil {
		return fmt.Errorf("datepicker: cannot configure console input: %w", err)
	}
	if err := setConsoleMode(outHandle, outputMode|enableVTProcessing); err != nil {
		_ = setConsoleMode(inHandle, inputMode)
		return fmt.Errorf("datepicker: console does not support virtual terminal processing: %w", err)
	}

	t.oldInputMode = inputMode
	t.oldOutputMode = outputMode
	t.started = true

	if _, err := t.out.WriteString(enterScreen + hideCursor); err != nil {
		t.Stop()
		return err
	}
	return nil
}

func (t *winTerminal) Stop() {
	if t.started {
		t.out.WriteString(showCursor + leaveScreen)
		_ = setConsoleMode(syscall.Handle(t.in.Fd()), t.oldInputMode)
		_ = setConsoleMode(syscall.Handle(t.out.Fd()), t.oldOutputMode)
		t.started = false
	}
	if t.ownedIn && t.in != nil {
		t.in.Close()
		t.ownedIn = false
	}
	if t.ownedOut && t.out != nil {
		t.out.Close()
		t.ownedOut = false
	}
}

func (t *winTerminal) Size() (int, int, error) {
	var info consoleScreenBufferInfo
	if err := getConsoleScreenBufferInfo(syscall.Handle(t.out.Fd()), &info); err != nil {
		return 0, 0, err
	}
	width := int(info.window.right-info.window.left) + 1
	height := int(info.window.bottom-info.window.top) + 1
	return width, height, nil
}

func (t *winTerminal) Draw(frame string) error {
	_, err := t.out.WriteString(frame)
	return err
}

func (t *winTerminal) PollEvent(timeout time.Duration) (event, bool, error) {
	deadline := time.Now().Add(timeout)
	handle := syscall.Handle(t.in.Fd())

	for {
		// Drain the previous batch first: ReadConsoleInputW can return many
		// records at once, and later ones must not be discarded.
		for t.next < t.count {
			record := &t.records[t.next]
			t.next++
			if ev, ok := t.parseRecord(record); ok {
				return ev, true, nil
			}
		}

		remaining := time.Until(deadline)
		if remaining < 0 {
			return event{}, false, nil
		}
		millis := uint32(remaining / time.Millisecond)

		result, _, callErr := procWaitForSingleObject.Call(uintptr(handle), uintptr(millis))
		switch result {
		case waitObject0:
			// Input records are available.
		case waitTimeout:
			return event{}, false, nil
		default:
			return event{}, false, winError("WaitForSingleObject", callErr)
		}

		var count uint32
		ok, _, callErr := procReadConsoleInputW.Call(
			uintptr(handle),
			uintptr(unsafe.Pointer(&t.records[0])),
			uintptr(len(t.records)),
			uintptr(unsafe.Pointer(&count)),
		)
		if ok == 0 {
			return event{}, false, winError("ReadConsoleInputW", callErr)
		}
		t.count = int(count)
		t.next = 0
	}
}

func (t *winTerminal) parseRecord(record *inputRecord) (event, bool) {
	switch record.eventType {
	case keyEventType:
		key := (*keyEventRecord)(unsafe.Pointer(&record.event[0]))
		if key.keyDown == 0 {
			return event{}, false
		}
		return t.parseKey(key)
	case mouseEventType:
		mouse := (*mouseEventRecord)(unsafe.Pointer(&record.event[0]))
		return t.parseMouse(mouse)
	case windowBufferSizeType:
		return event{kind: evResize}, true
	default:
		return event{}, false
	}
}

func (t *winTerminal) parseKey(key *keyEventRecord) (event, bool) {
	switch key.virtualKeyCode {
	case vkLeft:
		return keyEvent(keyLeft), true
	case vkUp:
		return keyEvent(keyUp), true
	case vkRight:
		return keyEvent(keyRight), true
	case vkDown:
		return keyEvent(keyDown), true
	case vkPrior:
		return keyEvent(keyPageUp), true
	case vkNext:
		return keyEvent(keyPageDown), true
	case vkReturn:
		return keyEvent(keyEnter), true
	case vkEscape:
		return keyEvent(keyCancel), true
	}

	if key.unicodeChar != 0 {
		if mapped, ok := runeKey(rune(key.unicodeChar)); ok {
			return keyEvent(mapped), true
		}
	}
	return event{}, false
}

func (t *winTerminal) parseMouse(mouse *mouseEventRecord) (event, bool) {
	// Only button transitions and double-clicks; ignore motion and wheels.
	if mouse.eventFlags != 0 && mouse.eventFlags != doubleClickMouseFlag {
		return event{}, false
	}
	if mouse.buttonState&leftMostButtonPressed == 0 {
		return event{}, false
	}

	var info consoleScreenBufferInfo
	if err := getConsoleScreenBufferInfo(syscall.Handle(t.out.Fd()), &info); err != nil {
		return event{}, false
	}
	col := int(mouse.x) - int(info.window.left) + 1
	row := int(mouse.y) - int(info.window.top) + 1
	if col < 1 || row < 1 {
		return event{}, false
	}
	return event{kind: evMouse, row: row, col: col}, true
}

func getConsoleMode(handle syscall.Handle) (uint32, error) {
	var mode uint32
	ok, _, callErr := procGetConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode)))
	if ok == 0 {
		return 0, winError("GetConsoleMode", callErr)
	}
	return mode, nil
}

func setConsoleMode(handle syscall.Handle, mode uint32) error {
	ok, _, callErr := procSetConsoleMode.Call(uintptr(handle), uintptr(mode))
	if ok == 0 {
		return winError("SetConsoleMode", callErr)
	}
	return nil
}

func getConsoleScreenBufferInfo(handle syscall.Handle, info *consoleScreenBufferInfo) error {
	ok, _, callErr := procGetConsoleScreenBufferInfo.Call(uintptr(handle), uintptr(unsafe.Pointer(info)))
	if ok == 0 {
		return winError("GetConsoleScreenBufferInfo", callErr)
	}
	return nil
}

// winError keeps a zero Errno from surfacing as "The operation completed
// successfully." when a Win32 call fails without setting the last error.
func winError(operation string, err error) error {
	if errno, ok := err.(syscall.Errno); ok && errno != 0 {
		return fmt.Errorf("datepicker: %s: %w", operation, err)
	}
	return fmt.Errorf("datepicker: %s failed", operation)
}
