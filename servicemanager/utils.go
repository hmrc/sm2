package servicemanager

import (
	"fmt"
	"os"
	"path"
	"strings"
)

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false
}

// based on config, find the directory a service is installed into.
// TODO: rename this something less confusing. it doesnt really 'find' anything,
//       rather it guesses where it is...
func (sm ServiceManager) findInstallDirOfService(serviceName string) (string, error) {
	if service, ok := sm.Services[serviceName]; ok {
		return path.Join(sm.Config.TmpDir, service.Binary.DestinationSubdir), nil
	} else {
		return "", fmt.Errorf("Unknown service: %s", serviceName)
	}
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
