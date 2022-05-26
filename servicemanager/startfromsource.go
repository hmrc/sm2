package servicemanager

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"time"

	"sm2/ledger"
)

func (sm *ServiceManager) StartFromSource(serviceName string) error {

	service, ok := sm.Services[serviceName]
	if !ok {
		return fmt.Errorf("%s is not a valid service", serviceName)
	}

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

	return sm.Ledger.SaveStateFile(installDir, state)
}

func (sm *ServiceManager) installFromGit(installDir string, gitUrl string, service Service) (ledger.InstallFile, error) {

	// TODO work out if we should update or clone
	if sm.Commands.Clean {
		removeSrcDir(installDir)
	}

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
		Version:  "src",
		Path:     srcDir,
		Md5Sum:   "TODO",
		Created:  time.Now(),
	}

	return installFile, nil
}

// shallow-clones a gitrepo into $installDir/src
func gitClone(gitUrl string, installDir string) (string, error) {
	cmd := exec.Command("git", "clone", "--depth", "1", gitUrl, "src")
	cmd.Dir = installDir

	stdout, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to clone %s into %s.\n", gitUrl, installDir)
		fmt.Println(string(stdout))
		return "", err
	}

	return path.Join(installDir, "src"), nil
}

func (sm ServiceManager) sbtBuildAndRun(srcDir string, service Service) (ledger.StateFile, error) {
	state := ledger.StateFile{}
	port := sm.findPort(service)
	args := []string{"run", fmt.Sprintf("-Dhttp.port=%d", port)}
	args = append(args, sm.generateArgs(service, "src", srcDir)...)

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

	state = ledger.StateFile{
		Service:  service.Id,
		Artifact: service.Binary.Artifact,
		Version:  "src",
		Path:     srcDir,
		Md5Sum:   "TODO",
		Started:  time.Now(),
		Pid:      cmd.Process.Pid,
		Args:     args,
	}

	return state, nil
}

func removeSrcDir(installDir string) error {
	srcPath := path.Join(installDir, "src")
	if Exists(srcPath) {
		return os.RemoveAll(srcPath)
	}
	return nil
}
