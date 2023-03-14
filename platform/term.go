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

func GetTerminalSize() (int, int) {
	ts := TermSize{}

	_, _, err := syscall.Syscall6(
		syscall.SYS_IOCTL,
		os.Stdin.Fd(),
		syscall.TIOCGWINSZ,
		uintptr(unsafe.Pointer(&ts)),
		0, 0, 0,
	)

	if err != 0 {
		return 80, 25
	}

	return int(ts.Cols), int(ts.Rows)
}
