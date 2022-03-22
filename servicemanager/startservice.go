package servicemanager

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"time"

	"sm2/ledger"
)

// startService attempts to start a version of a service, if the version is not specified
// service manager will get the latest vesion from artifactory.
func (sm ServiceManager) StartService(serviceName string, requestedVersion string) error {

	// look-up the service
	service, ok := sm.Services[serviceName]
	if !ok {
		return fmt.Errorf("%s is not a valid service", serviceName)
	}

	// check if its already running and exit if it is
	if sm.CheckHealth(service.DefaultPort) {
		sm.UiUpdates <- Progress{service: serviceName, percent: 100, state: "Already running"}
		return fmt.Errorf("Already Running")
	}

	offline := sm.Commands.Offline
	installDir, _ := sm.findInstallDirOfService(serviceName)
	versionToInstall := requestedVersion
	group := service.Binary.GroupId
	artifact := service.Binary.Artifact

	// look up the latest version if its not supplied
	if requestedVersion == "" && !offline {
		metadata, err := sm.GetLatestVersions(service.Binary)
		if err != nil {
			sm.UiUpdates <- Progress{service: serviceName, percent: 0, state: "Failed"}
			return fmt.Errorf("No version found")
		}
		group = metadata.Group
		artifact = metadata.Artifact
		versionToInstall = metadata.Latest
	}

	// install requested version of service if required
	isInstalled := false
	installFile, err := sm.Ledger.LoadInstallFile(installDir)
	if err == nil {
		isInstalled = verifyInstall(installFile, service.Id, versionToInstall, offline)
	}

	if !isInstalled || sm.Commands.Clean {

		// if we're offline and its not installed, there's not much we can do!
		if offline {
			sm.UiUpdates <- Progress{service: serviceName, percent: 0, state: "Failed"}
			return fmt.Errorf("Unavailable")
		}

		sm.UiUpdates <- Progress{service: serviceName, state: "Installing..."}

		var err error
		installFile, err = sm.installService(installDir, service.Id, group, artifact, versionToInstall)
		if err != nil {
			return err
		}
	}

	// re-init log dirs
	_, err = initLogDir(installFile.Path)
	if err != nil {
		return err
	}

	// start the service
	port := sm.findPort(service)
	args := sm.generateArgs(service, versionToInstall, installFile.Path)
	sm.UiUpdates <- Progress{service: serviceName, percent: 100, state: "Starting..."}
	state, err := run(service, installFile, args, port)
	if err != nil {
		return err
	}

	return sm.Ledger.SaveStateFile(installDir, state)
}

func (sm ServiceManager) installService(installDir string, serviceId string, group string, artifact string, version string) (ledger.InstallFile, error) {

	var installFile ledger.InstallFile

	err := removeExistingVersions(installDir)
	if err != nil {
		return installFile, err
	}

	sm.UiUpdates <- Progress{service: serviceId, percent: 0, state: "Init"}

	groupPath := strings.ReplaceAll(group, ".", "/")
	filename := fmt.Sprintf("%s-%s.tgz", url.PathEscape(artifact), url.PathEscape(version))
	downloadUrl := sm.Config.ArtifactoryRepoUrl + path.Join("/", groupPath, url.PathEscape(artifact), url.PathEscape(version), filename)

	progressTracker := ProgressTracker{
		service: serviceId,
		update:  sm.UiUpdates,
	}

	serviceDir, err := sm.downloadAndDecompress(downloadUrl, installDir, &progressTracker)
	if err != nil {
		return installFile, fmt.Errorf("failed to find service directory in %s: %s", installDir, err)
	}

	installFile = ledger.InstallFile{
		Service:  serviceId,
		Artifact: artifact,
		Version:  version,
		Path:     serviceDir,
		Md5Sum:   "TODO",
		Created:  time.Now(),
	}

	err = sm.Ledger.SaveInstallFile(installDir, installFile)
	return installFile, err
}

// Given a service (config) some args and an installFile (code) run the service.
func run(service Service, installFile ledger.InstallFile, args []string, port int) (ledger.StateFile, error) {

	serviceDir := installFile.Path
	version := installFile.Version

	removeRunningPid(serviceDir)

	logFile, err := os.Create(path.Join(serviceDir, "logs", "stdout.log"))
	if err != nil {
		return ledger.StateFile{}, err
	}

	// patch the port number onto the arg list
	args = append(args, fmt.Sprintf("-Dhttp.port=%d", port))

	// this is a bit of a hack to get the old config working with the new installation
	_, runCmd := path.Split(service.Binary.Cmd[0])
	cmd := exec.Command(path.Join(serviceDir, "bin", runCmd), args...)
	cmd.Dir = serviceDir
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	err = cmd.Start()
	if err != nil {
		return ledger.StateFile{}, err
	}

	state := ledger.StateFile{
		Service:  service.Id,
		Artifact: service.Binary.Artifact,
		Version:  version,
		Path:     serviceDir,
		Md5Sum:   "TODO",
		Started:  time.Now(),
		Pid:      cmd.Process.Pid,
		Port:     port,
		Args:     args,
	}

	return state, nil
}

func verifyInstall(installFile ledger.InstallFile, service string, version string, offline bool) bool {

	// verify its the right one
	if installFile.Service != service {
		return false
	}

	// check version (or not, if --offline)
	if installFile.Version != version && !offline {
		// wrong version means a reinstall
		return false
	}

	// verify its actually where it says it is
	if _, err := os.Stat(installFile.Path); os.IsNotExist(err) {
		return false
	}

	// TODO: verify hashes etc...
	return true
}

func (sm ServiceManager) findPort(service Service) int {
	portNumber := service.DefaultPort
	if sm.Commands.Port > 0 {
		portNumber = sm.Commands.Port
	}
	return portNumber
}

func (sm ServiceManager) generateArgs(service Service, version string, serviceDir string) []string {

	args := service.Binary.Cmd[1:]

	// add service-manager generated args
	smArgs := []string{
		fmt.Sprintf("-Dservice.manager.serviceName=%s", service.Id),
		fmt.Sprintf("-Dservice.manager.runFrom=%s", version),
		fmt.Sprintf("-Duser.home=%s", path.Join(serviceDir, "..")),
	}
	args = append(args, smArgs...)

	// add user supplied args
	if userArgs, ok := sm.Commands.ExtraArgs[service.Id]; ok {
		args = append(args, userArgs...)
	}

	return args
}

// killing a process doesn't cleanup the RUNNING_PID preventing it being rerun
func removeRunningPid(serviceDir string) {
	pidPath := path.Join(serviceDir, "RUNNING_PID")
	if _, err := os.Stat(pidPath); err == nil {
		os.Remove(pidPath)
	}
}

// cleans up previous installs
// @improvement could keep n previous versions?
func removeExistingVersions(installDir string) error {
	if !path.IsAbs(installDir) {
		// since we're removing a whole dir here, lets be careful that no-one has put ../../../ in the config etc
		panic("removeExistingVersions was passed a non-absoulte path. This shouldn't happen!")
	}
	if err := os.RemoveAll(installDir); err != nil {
		return err
	}
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return err
	}
	return nil
}
