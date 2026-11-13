//go:build linux

package datepicker

import (
	"syscall"
	"unsafe"
)

func isTerminal(fd uintptr) bool {
	var termios syscall.Termios
	return ioctl(fd, syscall.TCGETS, unsafe.Pointer(&termios)) == nil
}

func rawMode(fd uintptr) (func(), error) {
	var old syscall.Termios
	if err := ioctl(fd, syscall.TCGETS, unsafe.Pointer(&old)); err != nil {
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

	if err := ioctl(fd, syscall.TCSETS, unsafe.Pointer(&raw)); err != nil {
		return nil, err
	}
	return func() {
		_ = ioctl(fd, syscall.TCSETS, unsafe.Pointer(&old))
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
