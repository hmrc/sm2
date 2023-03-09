package servicemanager

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
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
	other := sm.findUnmanagedServices(statuses)

	if sm.Commands.FormatPlain {
		printPlainText(statuses, os.Stdout)
	} else {
		printTable(statuses, os.Stdout)
		printHelpIfRequired(statuses)

		if len(other) > 0 {
			fmt.Print("\n\033[34mAlso, it looks like the following services are running outside of sm2:\n\n")
			fmt.Print("These might include services running from inside your IDE or by other means.\n")
			fmt.Print("Please note: You will not be able to manage these services using sm2.\n")
			printUnmanagedTable(other, os.Stdout)
			fmt.Print("\033[0m\n")
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
				if err != nil {
					_ = sm.Ledger.ClearStateFile(installDir)
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

func printPlainText(statuses []serviceStatus, out io.Writer) {
	for _, status := range statuses {
		fmt.Fprintf(out, "%s\t%s\t%d\t%d\t%s\n", status.service, status.version, status.port, status.pid, status.health)
	}
}

func printTable(statuses []serviceStatus, out io.Writer) {

	border := fmt.Sprintf("+%s+%s+%s+%s+%s+\n", strings.Repeat("-", 36), strings.Repeat("-", 11), strings.Repeat("-", 9), strings.Repeat("-", 7), strings.Repeat("-", 8))

	fmt.Fprint(out, border)
	fmt.Fprintf(out, "| %-35s| %-10s| %-8s| %-6s| %-7s|\n", "Name", "Version", "PID", "Port", "Status")
	fmt.Fprint(out, border)

	const chunkSize = 35 //max size of service name before we wrap to next line

	for _, status := range statuses {
		serviceName := status.service

		if len(serviceName) > chunkSize {
			serviceName = addDelimiter(status.service, ",", chunkSize)
		}

		splitServiceName := strings.Split(serviceName, ",")
		numberOfLines := len(splitServiceName)

		for i, s := range splitServiceName {

			//Don't show final line of service name, if overflow < 4 chars.
			if numberOfLines > 1 && s == splitServiceName[len(splitServiceName)-1] && len(s) < 4 {
				break
			} else {
				fmt.Fprintf(out, "| %-35s", s)
			}
			//Only print the version/pid/port/status if first line of wrapped string
			if i == 0 {
				fmt.Fprintf(out, "| %-10s", status.version)
				fmt.Fprintf(out, "| %-8d", status.pid)
				fmt.Fprintf(out, "| %-6d", status.port)
				switch status.health {
				case PASS:
					fmt.Fprintf(out, "|  \033[32m%-6s\033[0m|\n", "PASS")
				case FAIL:
					fmt.Fprintf(out, "|  \033[31m%-6s\033[0m|\n", "FAIL")
				case BOOT:
					fmt.Fprintf(out, "|  \033[34m%-6s\033[0m|\n", "BOOT")
				}
			} else {
				fmt.Fprintf(out, "| %-10s", "")
				fmt.Fprintf(out, "| %-8s", "")
				fmt.Fprintf(out, "| %-6s", "")
				fmt.Fprintf(out, "|  %-6s|\n", "")

			}
		}
	}
	fmt.Fprint(out, border)
}

func printUnmanagedTable(statuses []serviceStatus, out io.Writer) {

	border := fmt.Sprintf("+%s+%s+%s+\n", strings.Repeat("-", 7), strings.Repeat("-", 9), strings.Repeat("-", 57))

	fmt.Fprint(out, border)
	fmt.Fprintf(out, "| %-6s| %-8s| %-56s|\n", "Port", "PID", "Reserved by")
	fmt.Fprint(out, border)

	const chunkSize = 35 //max size of service name before we wrap to next line

	for _, status := range statuses {
		serviceName := status.service

		if len(serviceName) > chunkSize {
			serviceName = addDelimiter(status.service, ",", chunkSize)
		}

		splitServiceName := strings.Split(serviceName, ",")

		for _, s := range splitServiceName {
			fmt.Fprintf(out, "| %-6d", status.port)
			//Only print the pid/port if first line of wrapped string
			fmt.Fprintf(out, "| %-8d", status.pid)
			fmt.Fprintf(out, "| %-56s|\n", s)
		}
	}
	fmt.Fprint(out, border)
}

func printHelpIfRequired(statuses []serviceStatus) {
	for _, status := range statuses {
		if status.health == FAIL && status.service != "MONGO" {
			fmt.Print("\n\033[1;31mOne or more services have failed to start.\033[0m\n")
			fmt.Print("You can check the logs of the fail service(s) or see at which point the service failed to start using:\n")
			fmt.Print("  sm2 --logs  SERVICE_NAME\n")
			fmt.Print("  sm2 --debug SERVICE_NAME\n\n")
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
