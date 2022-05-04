package servicemanager

import (
	"os"
	"strings"
)

func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Pads or crop a string with spaces until it matches the given width
func pad(s string, width int) string {
	if len(s) <= width {
		return s + strings.Repeat(" ", width-len(s))
	} else {
		return s[:width]
	}
}

func crop(s string, width int) string {
	if len(s) <= width {
		return s
	}
	return s[:width]
}
