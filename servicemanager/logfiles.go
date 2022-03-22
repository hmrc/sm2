package servicemanager

import (
	"fmt"
	"io"
	"os"
	"path"
)

// clears exists logs and creates the folder if its missing
func initLogDir(serviceDir string) (string, error) {
	logPath := path.Join(serviceDir, "logs")

	// if logdir exists remove it
	if _, err := os.Stat(logPath); os.IsExist(err) {
		rmErr := os.RemoveAll(logPath)
		if rmErr != nil {
			return logPath, rmErr
		}
	}
	return logPath, os.MkdirAll(logPath, 0755)
}

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
