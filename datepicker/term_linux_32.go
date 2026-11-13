//go:build linux && (386 || arm || mips || mipsle)

package datepicker

import (
	"syscall"
	"time"
)

func waitReadable(fd uintptr, timeout time.Duration) (bool, error) {
	var readSet syscall.FdSet
	if fd/32 >= uintptr(len(readSet.Bits)) {
		return false, errFDTooLarge
	}
	readSet.Bits[fd/32] |= int32(1) << (fd % 32)

	tv := syscall.NsecToTimeval(timeout.Nanoseconds())
	n, err := syscall.Select(int(fd)+1, &readSet, nil, nil, &tv)
	if err != nil {
		if err == syscall.EINTR {
			return false, nil
		}
		return false, err
	}
	return n > 0, nil
}
