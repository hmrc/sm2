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
	Uptime             func() time.Time
	PidLookup          func() map[int]int
	PidLookupByService func(string) (bool, []int)
	PortPidLookup      func() map[int]int
	GetTerminalSize    func() (int, int)
}

func DetectPlatform() Platform {
	switch runtime.GOOS {
	case "darwin":
		return Platform{uptimeDarwin, processLookupUnix, portPidLookup, processLookupByServiceName}
	case "linux":
		return Platform{uptimeLinux, processLookupUnix, portPidLookup, processLookupByServiceName}
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

	uptime, err := time.ParseInLocation("2006-01-02 15:04:05", strings.Trim(string(output), "\n"), time.Local)
	if err != nil {
		fmt.Printf("failed to parse time %s\n", err)
		return time.Unix(0, 0)
	}

	return uptime
}

// OSX doesnt support the -s flag on uptime!
// so we use sysctl and get the epoc seconds
func uptimeDarwin() time.Time {
	rx := regexp.MustCompile(`sec = (\d+)`)
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

// Looks at the arguments of a process for a arg matching `service.manager.serviceName=$SERVICE`.
// When starting from source its possible to have multiple processes (sbt bash script, sbt itself and the server)
// all with this argument. To avoid making it overly-specific to sbt all pids are returned.
func processLookupByServiceName(service string) (bool, []int) {

	pids := []int{}
	cmd := exec.Command("ps", "-eo", "pid,args")

	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Failed to get process list. Unable to list running services.\n%s\n", err)
		return false, pids
	}

	lookFor := fmt.Sprintf("service.manager.serviceName=%s", service)
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {

		split := strings.SplitN(strings.Trim(scanner.Text(), " "), " ", 2)

		if len(split) == 2 {
			if strings.Contains(split[1], lookFor) {
				if pid, err := strconv.Atoi(split[0]); err == nil {
					pids = append(pids, pid)
				}
			}
		}
	}

	return len(pids) > 0, pids
}

// Returns a map of all the open TCP listening ports and their Pid
func portPidLookup() map[int]int {

	rx := regexp.MustCompile(`.+\:(\d+)\D*`)
	portPid := map[int]int{}

	cmd := exec.Command("lsof", "-iTCP", "-sTCP:LISTEN", "-n", "-P", "-T")

	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Failed to list running ports.\n%s\n", err)
		return portPid
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		split := strings.Fields(scanner.Text())

		if len(split) == 9 {
			if matches := rx.FindStringSubmatch(string(split[8])); matches != nil {
				port, _ := strconv.Atoi(matches[1]) // no err check since regex ensures its a number
				pid, _ := strconv.Atoi(split[1])    // no err check since regex ensures its a number
				portPid[port] = pid
			}
		}
	}

	return portPid
}
