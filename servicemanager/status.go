package servicemanager

import (
	"fmt"
	"net"
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
	printTable(statues)
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
		fmt.Printf("error reading state files: %s", err)
		return statuses
	}

	// for each state file
	for _, state := range states {
		// ignore services that were started before the os started
		if state.Started.Before(bootTime) {
			// clean up state file
			installDir, err := sm.findInstallDirOfService(state.Service)
			if err != nil {
				_ = sm.Ledger.ClearStateFile(installDir)
			}
			continue
		}

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
			// no pid
			status.health = FAIL
		}
		statuses = append(statuses, status)
	}

	return statuses
}

func printTable(statuses []serviceStatus) {

	border := fmt.Sprintf("+%s+%s+%s+%s+%s+\n", strings.Repeat("-", 36), strings.Repeat("-", 11), strings.Repeat("-", 9), strings.Repeat("-", 7), strings.Repeat("-", 8))

	fmt.Print(border)
	fmt.Printf("| %-35s| %-10s| %-8s| %-6s| %-7s|\n", "Name", "Version", "PID", "Port", "Status")
	fmt.Print(border)
	for _, status := range statuses {
		fmt.Printf("| %-35s", crop(status.service, 35))
		fmt.Printf("| %-10s", status.version)
		fmt.Printf("| %-8d", status.pid)
		fmt.Printf("| %-6d", status.port)
		switch status.health {
		case PASS:
			fmt.Printf("|  \033[1;32m%-6s\033[0m|\n", "PASS")
		case FAIL:
			fmt.Printf("|  \033[1;31m%-6s\033[0m|\n", "FAIL")
		case BOOT:
			fmt.Printf("|  \033[1;34m%-6s\033[0m|\n", "BOOT")
		}
	}
	fmt.Print(border)
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
