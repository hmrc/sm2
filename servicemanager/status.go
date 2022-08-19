package servicemanager

import (
	"fmt"
	"io"
	"net"
	"os"
	"sort"
	"strings"
	"time"
)

type health string

const (
	PASS health = "PASS"
	FAIL health = "FAIL"
	BOOT health = "BOOT"
)

type serviceStatus struct {
	pid     int
	port    int
	service string
	version string
	health  health
}

func (sm *ServiceManager) PrintStatus() {
	statues := []serviceStatus{sm.CheckMongo()}
	statues = append(statues, sm.findStatuses()...)
	printTable(statues, os.Stdout)
	printHelpIfRequired(statues)
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
				if time.Since(state.Started).Seconds() > 30 {
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
			if numberOfLines > 1 && s == splitServiceName[len(splitServiceName) - 1] && len(s) < 4 {
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
					fmt.Fprintf(out, "|  \033[1;32m%-6s\033[0m|\n", "PASS")
				case FAIL:
					fmt.Fprintf(out, "|  \033[1;31m%-6s\033[0m|\n", "FAIL")
				case BOOT:
					fmt.Fprintf(out, "|  \033[1;34m%-6s\033[0m|\n", "BOOT")
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
	resp, err := sm.Client.Get(url)
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
