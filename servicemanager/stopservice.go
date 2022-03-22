package servicemanager

import (
	"fmt"
	"os"
)

func stopPid(pid int) {
	osProc, err := os.FindProcess(pid)
	if err != nil {
		fmt.Printf("error finding pid %d\n", pid)
	}

	err = osProc.Kill()
	if err != nil {
		fmt.Printf("error stopping pid %d - %s\n", pid, err)
	}
}

func (sm ServiceManager) StopService(serviceName string) error {

	statues := sm.findStatuses()

	for _, status := range statues {
		if status.service == serviceName {
			fmt.Printf("Stopping %s\t(pid %d)\n", serviceName, status.pid)
			stopPid(status.pid)

			// clean up service.state
			if installDir, err := sm.findInstallDirOfService(serviceName); err == nil { // ok
				sm.Ledger.ClearStateFile(installDir)
			}
			return nil
		}
	}

	if serviceName == "*" {
		fmt.Println("The command --stop ALL is deprecated, use --stop-all instead.")
		sm.StopAll()
		return nil
	}

	fmt.Printf("Unable to find service %s\n", serviceName)
	return nil
}

func (sm ServiceManager) StopAll() {

	fmt.Printf("Stopping ALL services!\n")

	statues := sm.findStatuses()

	for _, status := range statues {
		fmt.Printf("Stopping %s\t(pid %d)\n", status.service, status.pid)
		stopPid(status.pid)

		// clean up service.state
		if installDir, err := sm.findInstallDirOfService(status.service); err == nil { // ok
			sm.Ledger.ClearStateFile(installDir)
		}
	}

}
