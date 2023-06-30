package servicemanager

import (
	"fmt"
	"os"
)

func (sm *ServiceManager) StopService(serviceName string) error {

	// @improve just load the state file instead and kill the listed pid?
	statuses := sm.findStatuses()

	for _, status := range statuses {
		if status.service == serviceName {
			sm.stop(status)
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

func (sm *ServiceManager) StopAll() {

	fmt.Printf("Stopping ALL services!\n")

	statuses := sm.findStatuses()

	for _, status := range statuses {
		sm.stop(status)
	}

}

func (sm *ServiceManager) stop(status serviceStatus) {
	serviceName := status.service

	// services running from source will have been forked from the original sbt process
	// to stop them we need to look them up by service name and stop all the associated pids
	if status.version == SOURCE {
		if found, pids := sm.Platform.PidLookupByService(serviceName); found {
			fmt.Printf("Stopping %-40s (running from source)\n", serviceName)
			for _, pid := range pids {
				stopPid(pid)
			}
		} else {
			fmt.Printf("Unable to find pid for service started from source %s.\n", serviceName)
			return
		}
	} else {
		// run from release, kill the pid in the .state file
		fmt.Printf("Stopping %-40s(pid %-7d).\n", serviceName, status.pid)
		stopPid(status.pid)
	}

	// clean up service.state
	if installDir, err := sm.findInstallDirOfService(serviceName); err == nil { // ok
		sm.Ledger.ClearStateFile(installDir)
	}

}

func stopPid(pid int) {
	osProc, err := os.FindProcess(pid)
	if err != nil {
		fmt.Printf("PID %d does not exists.\n", pid)
	}

	err = osProc.Kill()
	if err != nil {
		fmt.Printf("Unable to stop pid %d, %s.\n", pid, err)
	}
}
