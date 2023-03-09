package platform

import (
	"os"
	"syscall"
	"unsafe"
)

type TermSize struct {
	Rows    uint16
	Cols    uint16
	Xpixels uint16
	Ypixels uint16
}

func GetTerminalSize() TermSize {
	ts := TermSize{}

	syscall.Syscall6(
		syscall.SYS_IOCTL,
		os.Stdin.Fd(),
		syscall.TIOCGWINSZ,
		uintptr(unsafe.Pointer(&ts)),
		0, 0, 0,
	)

	return ts
}
