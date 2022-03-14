package platform

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Platform struct {
	Uptime    func() time.Time
	PidLookup func() map[int]int
}

func DetectPlatform() Platform {
	switch runtime.GOOS {
	case "darwin":
		return Platform{uptimeDarwin, processLookupUnix}
	case "linux":
		return Platform{uptimeLinux, processLookupUnix}
	case "windows":
		log.Fatal("windows is not supported yet!")
	default:
		log.Fatalf("unsupported OS: %s", runtime.GOOS)
	}
	return Platform{}
}

type ServiceProcess struct {
	Pid     int
	Name    string
	Version string
	Port    int
}

func uptimeLinux() time.Time {
	cmd := exec.Command("uptime", "-s")
	output, err := cmd.Output()
	if err != nil {
		return time.Unix(0, 0)
	}

	uptime, err := time.Parse("2006-01-02 15:04:05", strings.Trim(string(output), "\n"))
	if err != nil {
		fmt.Printf("failed to parse time %s\n", err)
		return time.Unix(0, 0)
	}

	return uptime
}

// OSX doesnt support the -s flag on uptime!
// so we use sysctl and get the epoc seconds
func uptimeDarwin() time.Time {
	rx := regexp.MustCompile("sec = (\\d+)")
	cmd := exec.Command("sysctl", "-n", "kern.boottime")
	output, err := cmd.Output()
	if err != nil {
		return time.Unix(0, 0)
	}

	if matches := rx.FindStringSubmatch(string(output)); matches != nil {
		secs, _ := strconv.Atoi(matches[1]) // no err check since regex ensures its a number
		return time.Unix(int64(secs), 0)
	}

	fmt.Printf("failed to parse time %s\n", err)
	return time.Unix(0, 0)
}

// Returns an unfiltered map of all the process ids running on the system
func processLookupUnix() map[int]int {
	lookup := map[int]int{}
	cmd := exec.Command("ps", "-eo", "pid")

	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Failed to get process list. Unable to list running services.\n%s\n", err)
		return lookup
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		t := strings.Trim(scanner.Text(), "\n ")
		if pid, err := strconv.Atoi(t); err == nil {
			lookup[pid] = pid
		}
	}

	return lookup
}
