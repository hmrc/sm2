package servicemanager

import (
	"fmt"
	"io"
	"os"
	"path"
)

func (sm *ServiceManager) PrintLogsForService(serviceName string) {

	installDir, err := sm.findInstallDirOfService(serviceName)
	if err != nil {
		fmt.Printf("Couldn't find the logs for %s\n", serviceName)
		return
	}

	installFile, err := sm.Ledger.LoadInstallFile(installDir)
	if err != nil {
		fmt.Printf("Unable to find installation of service in %s\n\t%s", installDir, err)
		return
	}

	logDir := path.Join(installFile.Path, "logs")

	if !Exists(logDir) {
		fmt.Printf("Couldn't find the logs for %s\n", serviceName)
		return
	}

	pathToLog := path.Join(logDir, "stdout.log")

	file, err := os.Open(pathToLog)
	if err != nil {
		fmt.Printf("Failed to open logfile %s: %s\n", pathToLog, err)
		return
	}

	defer file.Close()

	io.Copy(os.Stdout, file)
}
