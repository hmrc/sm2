package servicemanager

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"sm2/ledger"
)

const SOURCE = "source"

func (sm *ServiceManager) StartFromSource(serviceName string) error {

	service, ok := sm.Services[serviceName]
	if !ok {
		return fmt.Errorf("%s is not a valid service", serviceName)
	}

	// TODO: check its not already running

	installDir, _ := sm.findInstallDirOfService(serviceName)

	sm.progress.update(serviceName, 0, "Cloning...")
	installFile, err := sm.installFromGit(installDir, service.Source.Repo, service)
	if err != nil {
		return err
	}

	err = sm.Ledger.SaveInstallFile(installDir, installFile)
	if err != nil {
		return err
	}

	// sbt run the service, redirect output to logs

	sm.progress.update(serviceName, 100, "Starting...")
	state, err := sm.sbtBuildAndRun(installFile.Path, service)
	if err != nil {
		return err
	}

	err = sm.Ledger.SaveStateFile(installDir, state)
	sm.pauseTillHealthy(state.HealthcheckUrl)
	return err
}

func (sm *ServiceManager) installFromGit(installDir string, gitUrl string, service Service) (ledger.InstallFile, error) {

	// TODO work out if we can just git pull instead
	removeExistingVersions(installDir)

	srcDir, err := gitClone(gitUrl, installDir)
	if err != nil {
		return ledger.InstallFile{}, err
	}

	// make logs dir inside the src dir
	_, err = initLogDir(srcDir)
	if err != nil {
		return ledger.InstallFile{}, err
	}

	installFile := ledger.InstallFile{
		Service:  service.Id,
		Artifact: service.Binary.Artifact,
		Version:  SOURCE,
		Path:     srcDir,
		Created:  time.Now(),
	}

	return installFile, nil
}

func (sm ServiceManager) sbtBuildAndRun(srcDir string, service Service) (ledger.StateFile, error) {
	state := ledger.StateFile{}
	port := sm.findPort(service)

	sbtStartCmds := "start " + fmt.Sprintf("start -Dhttp.port=%d ", port) + strings.Join(sm.generateArgs(service, "src", srcDir, append(service.Binary.Cmd[1:], service.Source.ExtraParams...)), " ")
	args := []string{"-mem", "2048", sbtStartCmds}

	cmd := exec.Command("sbt", args...)
	cmd.Dir = srcDir

	logFile, err := os.Create(path.Join(srcDir, "logs", "stdout.log"))
	if err != nil {
		return state, fmt.Errorf("unable to create stdout.log %s", err)
	}

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	err = cmd.Start()
	if err != nil {
		return state, err
	}

	healthcheckUrl := findHealthcheckUrl(service, state.Port)
	state = ledger.StateFile{
		Service:        service.Id,
		Artifact:       service.Binary.Artifact,
		Version:        SOURCE,
		Path:           srcDir,
		Started:        time.Now(),
		Pid:            cmd.Process.Pid,
		Port:           port,
		Args:           args,
		HealthcheckUrl: healthcheckUrl,
	}
	return state, nil
}
