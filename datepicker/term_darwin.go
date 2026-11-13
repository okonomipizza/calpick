//go:build darwin

package datepicker

import (
	"syscall"
	"time"
	"unsafe"
)

func isTerminal(fd uintptr) bool {
	var termios syscall.Termios
	return ioctl(fd, syscall.TIOCGETA, unsafe.Pointer(&termios)) == nil
}

func rawMode(fd uintptr) (func(), error) {
	var old syscall.Termios
	if err := ioctl(fd, syscall.TIOCGETA, unsafe.Pointer(&old)); err != nil {
		return nil, err
	}

	raw := old
	raw.Iflag &^= syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK | syscall.ISTRIP |
		syscall.INLCR | syscall.IGNCR | syscall.ICRNL | syscall.IXON
	raw.Oflag &^= syscall.OPOST
	raw.Lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	raw.Cflag &^= syscall.CSIZE | syscall.PARENB
	raw.Cflag |= syscall.CS8
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0

	if err := ioctl(fd, syscall.TIOCSETAF, unsafe.Pointer(&raw)); err != nil {
		return nil, err
	}
	return func() {
		_ = ioctl(fd, syscall.TIOCSETAF, unsafe.Pointer(&old))
	}, nil
}

func terminalSize(fd uintptr) (int, int, error) {
	var size struct {
		Rows, Cols, Xpixel, Ypixel uint16
	}
	if err := ioctl(fd, syscall.TIOCGWINSZ, unsafe.Pointer(&size)); err != nil {
		return 0, 0, err
	}
	return int(size.Cols), int(size.Rows), nil
}

func waitReadable(fd uintptr, timeout time.Duration) (bool, error) {
	var readSet syscall.FdSet
	if fd/32 >= uintptr(len(readSet.Bits)) {
		return false, errFDTooLarge
	}
	readSet.Bits[fd/32] |= int32(1) << (fd % 32)
	tv := syscall.NsecToTimeval(timeout.Nanoseconds())

	if err := syscall.Select(int(fd)+1, &readSet, nil, nil, &tv); err != nil {
		if err == syscall.EINTR {
			return false, nil
		}
		return false, err
	}

	// Select does not report the count on Darwin, so check the set again.
	return readSet.Bits[fd/32]&(int32(1)<<(fd%32)) != 0, nil
}
