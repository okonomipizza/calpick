package datepicker

import (
	"sync/atomic"
	"time"
)

const (
	enterScreen  = "\x1b[?1049h"
	leaveScreen  = "\x1b[?1049l"
	hideCursor   = "\x1b[?25l"
	showCursor   = "\x1b[?25h"
	enableMouse  = "\x1b[?1000h\x1b[?1006h"
	disableMouse = "\x1b[?1006l\x1b[?1000l"
)

var (
	resizeRequested    atomic.Bool
	interruptRequested atomic.Bool
)

// terminal abstracts the platform-specific console.
type terminal interface {
	Start() error
	Stop()
	Size() (width, height int, err error)
	Draw(frame string) error
	PollEvent(timeout time.Duration) (event, bool, error)
}
