//go:build linux && (amd64 || arm64 || loong64 || mips64 || mips64le || ppc64 || ppc64le || riscv64 || s390x)

package datepicker

import (
	"syscall"
	"time"
)

func waitReadable(fd uintptr, timeout time.Duration) (bool, error) {
	var readSet syscall.FdSet
	if fd/64 >= uintptr(len(readSet.Bits)) {
		return false, errFDTooLarge
	}
	readSet.Bits[fd/64] |= int64(1) << (fd % 64)

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
