package servicemanager

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"

	"sm2/ledger"
)

// startService attempts to start a version of a service, if the version is not specified
// service manager will get the latest vesion from artifactory.
func (sm *ServiceManager) StartService(serviceAndVersion ServiceAndVersion) error {

	offline := sm.Commands.Offline

	// look-up the service
	service, ok := sm.Services[serviceAndVersion.service]
	if !ok {
		return fmt.Errorf("%s is not a valid service", serviceAndVersion.service)
	}

	// check if its already running and exit if it is
	// TODO: check PID too
	port := sm.findPort(service)
	healthcheckUrl := findHealthcheckUrl(service, port)
	if sm.CheckHealth(healthcheckUrl) {
		sm.progress.update(serviceAndVersion.service, 100, "Already running")
		return fmt.Errorf("Already running")
	}

	// check if we're on the VPN (if required)
	if !sm.Commands.NoVpnCheck {
		vpnOk, _ := checkVpn(sm.Client, sm.Config)
		if !offline && !vpnOk {
			sm.progress.update(serviceAndVersion.service, 0, "No VPN")
			return fmt.Errorf("Check VPN connection, couldn't reach artifactory.")
		}
	}

	// work out what we will install, where...
	installDir, _ := sm.findInstallDirOfService(serviceAndVersion.service)
	group, artifact, versionToInstall, err := whatVersionToRun(service, serviceAndVersion, offline, sm.GetLatestVersions)
	if err != nil {
		sm.progress.update(serviceAndVersion.service, 0, "Failed")
		return err
	}
	isInstalled := false
	installFile, err := sm.Ledger.LoadInstallFile(installDir)
	if err == nil {
		isInstalled = verifyInstall(installFile, service.Id, versionToInstall, offline)
	}

	// and if required, install it...
	if !isInstalled || sm.Commands.Clean {

		// if we're offline and its not installed, there's not much we can do!
		if offline {
			sm.progress.update(serviceAndVersion.service, 0, "Failed")
			return fmt.Errorf("Not available offline")
		}

		sm.progress.update(serviceAndVersion.service, 0, "Install")

		var err error
		installFile, err = sm.installService(installDir, service.Id, group, artifact, versionToInstall)
		if err != nil {
			return err
		}
	}

	// clean and recreate log dirs...
	_, err = initLogDir(installFile.Path)
	if err != nil {
		sm.progress.update(serviceAndVersion.service, 0, "Failed")
		return err
	}

	// start the service...
	args := sm.generateArgs(service, versionToInstall, installFile.Path, service.Binary.Cmd[1:])
	sm.progress.update(serviceAndVersion.service, 100, "Starting...")
	state, err := run(service, installFile, args, port)
	if err != nil {
		sm.progress.update(serviceAndVersion.service, 0, "Failed")
		return err
	}
	state.HealthcheckUrl = healthcheckUrl
	// and finally, we record out success
	err = sm.Ledger.SaveStateFile(installDir, state)
	sm.pauseTillHealthy(healthcheckUrl)
	return err
}

func (sm *ServiceManager) pauseTillHealthy(healthcheckUrl string) {
	if sm.Commands.DelaySeconds > 0 {
		count := 0
		for count < sm.Commands.DelaySeconds*2 && !sm.CheckHealth(healthcheckUrl) {
			count++
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func (sm *ServiceManager) installService(installDir string, serviceId string, group string, artifact string, version string) (ledger.InstallFile, error) {

	var installFile ledger.InstallFile

	err := removeExistingVersions(installDir)
	if err != nil {
		return installFile, err
	}

	sm.progress.update(serviceId, 0.0, "Init")

	groupPath := strings.ReplaceAll(group, ".", "/")
	filename := fmt.Sprintf("%s-%s.tgz", url.PathEscape(artifact), url.PathEscape(version))
	downloadUrl := sm.Config.ArtifactoryRepoUrl + path.Join("/", groupPath, url.PathEscape(artifact), url.PathEscape(version), filename)

	progressWriter := ProgressWriter{
		service:  serviceId,
		renderer: &sm.progress,
	}

	serviceDir, err := sm.downloadAndDecompress(downloadUrl, installDir, &progressWriter)
	if err != nil {
		return installFile, fmt.Errorf("failed %s", err)
	}

	installFile = ledger.InstallFile{
		Service:  serviceId,
		Artifact: artifact,
		Version:  version,
		Path:     serviceDir,
		Created:  time.Now(),
	}

	err = sm.Ledger.SaveInstallFile(installDir, installFile)
	return installFile, err
}

// Given a service (config) some args and an installFile (code) run the service.
func run(service Service, installFile ledger.InstallFile, args []string, port int) (ledger.StateFile, error) {

	serviceDir := installFile.Path
	version := installFile.Version

	// @TODO: check if pid is already running
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

func (sm *ServiceManager) findPort(service Service) int {
	portNumber := service.DefaultPort
	if sm.Commands.Port > 0 {
		portNumber = sm.Commands.Port
	}
	return portNumber
}

func defaultHealthcheckUrl(port int) string {
	return fmt.Sprintf("http://localhost:%d/ping/ping", port)
}

func findHealthcheckUrl(service Service, port int) string {
	if service.Healthcheck.Url != "" {
		return strings.Replace(service.Healthcheck.Url, "${port}", fmt.Sprint(port), 1)
	}
	return defaultHealthcheckUrl(port)
}

func whatVersionToRun(service Service, serviceAndVersion ServiceAndVersion, offline bool, getLatest func(ServiceBinary, string) (MavenMetadata, error)) (string, string, string, error) {
	versionToInstall := serviceAndVersion.version
	group := service.Binary.GroupId
	artifact := service.Binary.Artifact

	// override scala version if required
	if serviceAndVersion.scalaVersion != "" {
		artifact = scalaSuffix.ReplaceAllLiteralString(artifact, "_"+serviceAndVersion.scalaVersion)
	}

	if versionToInstall == "" && !offline {
		metadata, err := getLatest(service.Binary, serviceAndVersion.scalaVersion)
		if err != nil {
			return "", "", "", err
		}
		group = metadata.Group
		artifact = metadata.Artifact
		versionToInstall = metadata.Latest
	}

	return group, artifact, versionToInstall, nil
}

// builds an array of arguments from service config, user supplied args and some sm defaults
func (sm *ServiceManager) generateArgs(service Service, version string, serviceDir string, serviceArgs []string) []string {

	args := serviceArgs
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

// run as a go routine to start services off a queue
func (sm *ServiceManager) startServiceWorker(tasks chan ServiceAndVersion, wg *sync.WaitGroup) {

	for task := range tasks {

		var err error
		if sm.Commands.FromSource {
			err = sm.StartFromSource(task.service)
		} else {
			err = sm.StartService(task)
		}

		if err != nil {
			sm.progress.update(task.service, 100, "Failed")
			sm.progress.error(task.service, err)
		} else {
			sm.progress.update(task.service, 100, "Done")
		}
		wg.Done()
	}

}

// Starts a bunch of services at once, but not all at once...
// the serviceWorkers run in concurrently, starting services as they arrive on the
// channel. The renderer also runs concurrently, drawing input as it gets it.
// A wait group is used to keep the app waiting for everything to finish downloading.
func (sm *ServiceManager) asyncStart(services []ServiceAndVersion) {

	// fire up the progress bar renderer
	sm.progress.noProgress = sm.Commands.NoProgress
	sm.progress.getTerminalSize = sm.Platform.GetTerminalSize
	go sm.progress.renderLoop()
	sm.progress.init(services)
	taskQueue := make(chan ServiceAndVersion, len(services))

	if len(services) == 1 {
		sm.Commands.Workers = 1 // only need 1 worker if starting single service
		fmt.Printf("Starting %d service on %d worker\n", len(services), sm.Commands.Workers)
	} else if sm.Commands.Workers == 1 { // set explicitly with --workers 1
		fmt.Printf("Starting %d services on %d worker\n", len(services), sm.Commands.Workers)
	} else {
		fmt.Printf("Starting %d services on %d workers\n", len(services), sm.Commands.Workers)
	}

	// start up a number of workers (controlled by --workers param)
	wg := sync.WaitGroup{}
	for i := 0; i < sm.Commands.Workers; i++ {
		go sm.startServiceWorker(taskQueue, &wg)
	}

	for _, sv := range services {
		wg.Add(1)
		taskQueue <- sv
	}

	wg.Wait()
	// @hack @hack waits a ms in the hope the renderloop finishes.
	// this could be way better, wait groups, or force a final paint or something??
	time.Sleep(time.Millisecond)

	if sm.Commands.Wait > 0 {
		fmt.Printf("Waiting %d secs for all services to start.", sm.Commands.Wait)
		sm.Await(services, sm.Commands.Wait)
	}

	// if anything has failed to start, report why
	if len(sm.progress.errors) > 0 && !sm.progress.noProgress {
		fmt.Println("\n\033[1;31mSome services failed to start:\033[0m")
		for k, v := range sm.progress.errors {
			fmt.Printf("  %s: %s\n", k, v.Error())
		}
	}

}
