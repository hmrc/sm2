package servicemanager

import (
	"fmt"
	"io/ioutil"
	"path"
)

func (sm *ServiceManager) showDebug(serviceName string) {

	_, ok := sm.Services[serviceName]
	if !ok {
		fmt.Printf("Service %s is not in config!\n", serviceName)
		return
	}

	installDir, err := sm.findInstallDirOfService(serviceName)
	if err != nil {
		fmt.Println("Unable to find install dir!")
		return
	}

	// check install file
	fmt.Println("Checking .install file...")
	installFile, err := sm.Ledger.LoadInstallFile(installDir)
	if err != nil {
		fmt.Printf("No install file found, or it was not readable. This suggest the service was not installed.\n [%s]\n", err)
		return
	}

	if !Exists(installFile.Path) {
		fmt.Printf("installation seems to be missing at %s!\n", installFile.Path)
		return
	}

	fmt.Printf("%s: version %s\n Installed at %s on %s\n", installFile.Service, installFile.Version, installFile.Path, installFile.Created)

	// check state file
	fmt.Println("Checking .state file...")
	stateFile, err := sm.Ledger.LoadStateFile(installDir)
	if err != nil {
		fmt.Printf("No state file found, or it was not readable. This suggests the service was not started.\n [%s]\n", err)
		return
	}

	// print out the interesting bits of state
	fmt.Printf("The .state file says %s version %s was started on %s with PID %d\n", stateFile.Service, stateFile.Version, stateFile.Started, stateFile.Pid)

	// check if it was started prior to the last reboot
	if stateFile.Started.Before(sm.Platform.Uptime()) {
		fmt.Printf("This service was started prior to the system's last reboot (%v). It will be excluded from the status page.\n", sm.Platform.Uptime())
	}

	fmt.Printf("It was run with the following args:\n")
	for _, arg := range stateFile.Args {
		fmt.Printf("\t- %s\n", arg)
	}
	// check pid
	fmt.Printf("Checking pid: %d is running...\n", stateFile.Pid)
	procs := sm.Platform.PidLookup()
	_, pidFound := procs[stateFile.Pid]
	if pidFound {
		fmt.Printf("Pid %d exists, service is probably running...", stateFile.Pid)
	} else {
		fmt.Printf("Pid not found, service is not running (or if it is, it wasnt started by servicemanager)\n")
	}

	// ping service
	fmt.Printf("pinging service on port %d...\n", stateFile.Port)
	if sm.CheckHealth(stateFile.HealthcheckUrl) {
		fmt.Printf("Service responded to ping on [%s], its alive.\n", stateFile.HealthcheckUrl)
		if !pidFound {
			fmt.Printf("It looks like %s was started by something other than service-manager.", stateFile.Service)
		}
	} else {
		fmt.Printf("Service did not respond on [%s]... check the log files\n", stateFile.HealthcheckUrl)
	}
	// show what logs we have
	logDir := path.Join(installFile.Path, "logs")
	files, err := ioutil.ReadDir(logDir)
	if err != nil {
		fmt.Printf("unable to read log dir: %s\n%s\n", logDir, err)
		return
	}

	fmt.Printf("Log files in %s:\n", logDir)
	for _, file := range files {
		fmt.Printf("\t%20s  %d\n", file.Name(), file.Size())
	}

}
