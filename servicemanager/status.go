package servicemanager

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sm2/ledger"
	"sort"
	"strings"
	"time"
)

type health string

const (
	PASS          health  = "PASS"
	FAIL          health  = "FAIL"
	BOOT          health  = "BOOT"
	GRACE_RELEASE float64 = 30
	GRACE_SOURCE  float64 = 60
)

type serviceStatus struct {
	pid     int
	port    int
	service string
	version string
	health  health
}

func (sm *ServiceManager) PrintStatus() {
	statuses := []serviceStatus{sm.CheckMongo()}
	statuses = append(statuses, sm.findStatuses()...)
	unmanaged := []serviceStatus{}
	proxyState := sm.Ledger.LoadProxyState(sm.Config.TmpDir)

	termWidth, _ := sm.Platform.GetTerminalSize()
	if sm.Commands.FormatPlain || termWidth < 80 {
		printPlainText(statuses, os.Stdout)
		if proxyState.Pid > 0 {
			printProxyPlainText(proxyState, os.Stdout)
		}
	} else {
		if !sm.Commands.NoPortCheck {
			unmanaged = sm.findUnmanagedServices(statuses)
		}

		longestServiceName := getLongestServiceName(append(statuses, unmanaged...))
		printTable(statuses, termWidth, longestServiceName, os.Stdout)
		printHelpIfRequired(statuses, sm.Commands.DelaySeconds)

		if len(unmanaged) > 0 {
			fmt.Print("\n\033[34mAlso, the following processes are running which occupy ports of services\n")
			fmt.Print("that are defined in service manager config:\n\n")
			fmt.Print("These might include entirely separate processes running on your machine,\n")
			fmt.Print("or they could be services running from inside your IDE or by other means.\n")
			fmt.Print("Please note: You will not be able to manage these services using sm2.\n")
			printUnmanagedTable(unmanaged, termWidth, longestServiceName, os.Stdout)
			fmt.Print("\033[0m\n")
		}
		if proxyState.Pid > 0 {
			printProxyTable(proxyState, termWidth, os.Stdout)
		}
	}
}

func (sm *ServiceManager) findUnmanagedServices(knownStatuses []serviceStatus) []serviceStatus {
	statuses := []serviceStatus{}

	portLookup := map[int]string{}
	for _, s := range sm.Services {
		portLookup[s.DefaultPort] = s.Id
	}

	knownPorts := map[int]string{}
	for _, s := range knownStatuses {
		knownPorts[s.port] = ""
	}

	for port, pid := range sm.Platform.PortPidLookup() {
		if _, ok := knownPorts[port]; ok {
			continue
		}

		if service, ok := portLookup[port]; ok {
			status := serviceStatus{
				pid:     pid,
				port:    port,
				service: service,
			}

			statuses = append(statuses, status)
		}
	}

	return statuses
}

func (sm *ServiceManager) findStatuses() []serviceStatus {

	statuses := []serviceStatus{}

	// get how long system has been up so we can exclude services that we stopped due to reboot
	bootTime := sm.Platform.Uptime()

	// get a set of all pids
	pids := sm.Platform.PidLookup()

	// find all the state files in the base dir...
	states, err := sm.Ledger.FindAllStateFiles(sm.Config.TmpDir)
	if err != nil {
		fmt.Printf("Unable to read state files in %s: %s\n", sm.Config.TmpDir, err)
		return statuses
	}

	// for each state file
	for _, state := range states {

		status := serviceStatus{
			pid:     state.Pid,
			port:    state.Port,
			service: state.Service,
			version: state.Version,
			health:  BOOT,
		}

		if _, ok := pids[state.Pid]; ok {
			url := state.HealthcheckUrl
			if url == "" {
				url = defaultHealthcheckUrl(state.Port)
			}
			if sm.CheckHealth(url) {
				status.health = PASS
			} else {
				// if boot grace period has passed, it fails
				grace := GRACE_RELEASE
				if status.version == SOURCE {
					println("grace from source")
					grace = GRACE_SOURCE
				}
				if time.Since(state.Started).Seconds() > grace {
					status.health = FAIL
				}
			}
		} else {
			// service is not running...

			// ignore services that were started before the os started
			if state.Started.Before(bootTime) {
				// clean up state file
				installDir, err := sm.findInstallDirOfService(state.Service)
				if err == nil {
					err = sm.Ledger.ClearStateFile(installDir)
					if err != nil {
						fmt.Printf("Error clearing %s state file: %s", state.Service, err)
					}
				}
				continue
			}
			status.health = FAIL
		}
		statuses = append(statuses, status)
	}

	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].service < statuses[j].service
	})

	return statuses
}

func (sm *ServiceManager) cleanupFailedServices() {
	statuses := sm.findStatuses()

	// for each service status
	for _, status := range statuses {
		if status.health == FAIL && status.service != "MONGO" {
			// clean up state file
			installDir, err := sm.findInstallDirOfService(status.service)
			if err == nil {
				err = sm.Ledger.ClearStateFile(installDir)
				if err != nil {
					fmt.Printf("Error clearing %s state file: %s\n", status.service, err)
				} else {
					fmt.Printf("Cleaned up %s\n", status.service)
				}
			}
		}
	}
}

func printPlainText(statuses []serviceStatus, out io.Writer) {
	for _, status := range statuses {
		fmt.Fprintf(out, "%s\t%s\t%d\t%d\t%s\n", status.service, status.version, status.port, status.pid, status.health)
	}
}

func printProxyPlainText(proxyState ledger.ProxyState, out *os.File) {
	fmt.Fprintf(out, "Reverse Proxy Running with PID %d", proxyState.Pid)
	for path, port := range proxyState.ProxyPaths {
		fmt.Fprintf(out, "%s\t%s\n", path, port)
	}
}

const (
	widthVersion     = 11
	widthPid         = 9
	widthPort        = 7
	widthStatus      = 8
	widthServicePath = 16
)

func getLongestServiceName(statuses []serviceStatus) int {
	var serviceNames []string
	for _, s := range append(statuses) {
		serviceNames = append(serviceNames, s.service)
	}
	return getLongestString(serviceNames)
}

func getLongestProxyPath(paths map[string]string) int {
	proxyPath := make([]string, 0, len(paths))
	for k, _ := range paths {
		proxyPath = append(proxyPath, k)
	}
	return getLongestString(proxyPath)
}

func getLongestString(strings []string) int {
	longestString := 35
	for _, s := range strings {
		if len(s)+2 > longestString {
			longestString = len(s) + 2
		}
	}
	return longestString
}

func printTable(statuses []serviceStatus, maxWidth int, longestServiceName int, out io.Writer) {
	// We want it to be at least 35 cols wide, and at most as long as the longest service name or
	// the width of the terminal - the space we've given to the other columns.
	widthName := maxWidth - (widthVersion + widthPid + widthPort + widthStatus + 6)
	if longestServiceName < widthName {
		widthName = longestServiceName
	}

	// Draw the border & header.
	border := fmt.Sprintf("+%s+%s+%s+%s+%s+\n", strings.Repeat("-", widthName), strings.Repeat("-", widthVersion), strings.Repeat("-", widthPid), strings.Repeat("-", widthPort), strings.Repeat("-", widthStatus))

	fmt.Fprint(out, border)
	fmt.Fprintf(out, "|%s|%s|%s|%s|%s|\n", pad(" Name", widthName), pad(" Version", widthVersion), pad(" PID", widthPid), pad(" Port", widthPort), pad(" Status", widthStatus))
	fmt.Fprint(out, border)

	for _, status := range statuses {

		// Handle word-wrapping.
		splitServiceName := partition(status.service, widthName-1)

		// Draw the first line complete with ports and pids.
		fmt.Fprintf(out, "| %s", pad(splitServiceName[0], widthName-1))
		fmt.Fprintf(out, "| %s", pad(status.version, widthVersion-1))
		fmt.Fprintf(out, "| %s", pad(fmt.Sprintf("%d", status.pid), widthPid-1))
		fmt.Fprintf(out, "| %s", pad(fmt.Sprintf("%d", status.port), widthPort-1))
		switch status.health {
		case PASS:
			fmt.Fprintf(out, "|  \033[32m%-6s\033[0m|\n", "PASS")
		case FAIL:
			fmt.Fprintf(out, "|  \033[31m%-6s\033[0m|\n", "FAIL")
		case BOOT:
			fmt.Fprintf(out, "|  \033[34m%-6s\033[0m|\n", "BOOT")
		}

		// Draw the subsequent lines if the name wraps, we leave non-name fields empty so they're not repeated.
		for _, s := range splitServiceName[1:] {
			fmt.Fprintf(out, "| %s|%s|%s|%s|%s|\n", pad(s, widthName-1), pad("", widthVersion), pad("", widthPid), pad("", widthPort), pad("", widthStatus))
		}
	}
	fmt.Fprint(out, border)
}

func printProxyTable(status ledger.ProxyState, maxWidth int, out io.Writer) {
	longestProxyPath := getLongestProxyPath(status.ProxyPaths)
	widthProxyPath := maxWidth - (widthServicePath + 3)
	if longestProxyPath < widthProxyPath {
		widthProxyPath = longestProxyPath
	}

	// Draw the border & header.
	border := fmt.Sprintf("+%s+%s+\n", strings.Repeat("-", widthProxyPath), strings.Repeat("-", widthServicePath))

	fmt.Fprint(out, border)
	fmt.Fprintf(out, "|%s|%s|\n", pad(" Proxy Path", widthProxyPath), pad(" Service Path", widthServicePath))
	fmt.Fprint(out, border)

	var sortedProxyPaths []string
	for k, _ := range status.ProxyPaths {
		sortedProxyPaths = append(sortedProxyPaths, k)
	}
	sort.Strings(sortedProxyPaths)
	for _, k := range sortedProxyPaths {
		splitServiceName := partition(k, widthProxyPath-1)
		fmt.Fprintf(out, "| %s", pad(splitServiceName[0], widthProxyPath-1))
		fmt.Fprintf(out, "| %s|\n", pad(status.ProxyPaths[k], widthServicePath-1))
		for _, s := range splitServiceName[1:] {
			fmt.Fprintf(out, "|%s|%s|\n", pad(s, widthProxyPath), pad("", widthServicePath))
		}
	}
	fmt.Fprint(out, border)
}

func printUnmanagedTable(statuses []serviceStatus, maxWidth int, longestServiceName int, out io.Writer) {
	// We want it to be at least 35 cols wide, and at most as long as the longest service name or
	// the width of the terminal - the space we've given to the other columns.
	widthName := maxWidth - (widthPid + widthPort + 6)
	if longestServiceName < widthName {
		widthName = longestServiceName + widthVersion + widthStatus + 2
	}

	border := fmt.Sprintf("+%s+%s+%s+\n", strings.Repeat("-", widthPid), strings.Repeat("-", widthPort), strings.Repeat("-", widthName))

	fmt.Fprint(out, border)
	fmt.Fprintf(out, "|%s|%s|%s|\n", pad(" PID", widthPid), pad(" Port", widthPort), pad(" Reserved by", widthName))
	fmt.Fprint(out, border)

	for _, status := range statuses {
		fmt.Fprintf(out, "| %s", pad(fmt.Sprintf("%d", status.pid), widthPid-1))
		fmt.Fprintf(out, "| %s", pad(fmt.Sprintf("%d", status.port), widthPort-1))
		fmt.Fprintf(out, "| %s|\n", pad(status.service, widthName-1))
	}
	fmt.Fprint(out, border)
}

func printHelpIfRequired(statuses []serviceStatus, delay int) {
	for _, status := range statuses {
		if status.health == FAIL && status.service != "MONGO" {
			fmt.Print("\n\033[1;31mOne or more services have failed to start.\033[0m\n")
			fmt.Print("You can check the logs of the fail service(s) or see at which point the service failed to start using:\n")
			fmt.Print("  sm2 -logs  SERVICE_NAME\n")
			fmt.Print("  sm2 -debug SERVICE_NAME\n\n")
			fmt.Print("Alternatively, you can remove them from this list by using:\n")
			fmt.Print("  sm2 -prune\n\n")

			if delay == 0 && len(statuses) >= 10 { // not already using --delay-seconds
				fmt.Println("Note: If you're starting a profile that contains a lot of services,")
				fmt.Println("try using `--delay-seconds 5` to add a 5 second delay after starting")
				fmt.Println("each service. This will help prevent your CPU getting overloaded,")
				fmt.Println("which can cause the services to take too long to respond to the healthcheck.")
				fmt.Println("See `sm2 --help` for more information.")
			}

			return
		}
	}
}

// returns true if the service ping endpoint responds
func (sm *ServiceManager) CheckHealth(url string) bool {
	ctx := sm.NewShortContext()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)

	resp, err := sm.Client.Do(req)
	return err == nil && resp.StatusCode == 200
}

// v.basic mongo check that just sees if the port is open
// @improve send minimal bytes to start a real connection and get version
func (sm ServiceManager) CheckMongo() serviceStatus {
	mongoStatus := serviceStatus{
		pid:     0,
		port:    27017,
		service: "MONGO",
		health:  FAIL,
	}

	conn, err := net.DialTimeout("tcp", net.JoinHostPort("localhost", "27017"), time.Duration(50)*time.Millisecond)

	if err == nil && conn != nil {
		mongoStatus.health = PASS
		conn.Close()
	}

	return mongoStatus
}

func (sm *ServiceManager) VerifyAllServicesAreRunning(services []ServiceAndVersion) bool {
	statuses := sm.findStatuses()
	return verifyIsRunning(services, statuses, os.Stdout)
}

// For a given list of services check if they're running (i.e. in PASS state)
// Print out the results and return a bool to indicate everything is ok
func verifyIsRunning(services []ServiceAndVersion, statuses []serviceStatus, out io.Writer) bool {

	allOk := true
	for _, service := range services {
		found := false
		for _, status := range statuses {
			if service.service == status.service && status.health == PASS {
				found = true
				break
			}
		}
		if found {
			fmt.Printf("%s\tOK\n", service.service)
		} else {
			allOk = false
			fmt.Printf("%s\tMISSING\n", service.service)
		}
	}

	return allOk
}
