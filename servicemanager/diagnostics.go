package servicemanager

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"runtime"

	"sm2/version"
)

func RunDiagnostics(config ServiceManagerConfig, noProgress bool) {
	version.PrintVersion()

	startStatus(CompOS, noProgress)
	checkOS(noProgress)

	startStatus(CompJava, noProgress)
	checkJava(noProgress)

	startStatus(CompGit, noProgress)
	checkGit(noProgress)

	startStatus(CompConfig, noProgress)
	checkConfigRevision(config, noProgress)

	startStatus(CompWorkspace, noProgress)
	checkWorkspace(config, noProgress)

	startStatus(CompVpn, noProgress)
	checkNetwork(config, noProgress)
}

func checkWorkspace(config ServiceManagerConfig, noProgress bool) {
	stat, err := os.Stat(config.TmpDir)
	if err != nil {
		updateStatus(CompWorkspace, StatusError, err.Error(), noProgress)
		return
	}

	if !stat.IsDir() {
		updateStatus(CompWorkspace, StatusError, fmt.Sprintf("%s is not a directory", config.TmpDir), noProgress)
		return
	}

	updateStatus(CompWorkspace, StatusOK, config.TmpDir, noProgress)
}

func checkConfigRevision(config ServiceManagerConfig, noProgress bool) {
	err := gitFetch(config.ConfigDir, "origin", "main")
	if err != nil {
		updateStatus(CompConfig, StatusWarn, "Unable to fetch latest remote version", noProgress)
		return
	}

	localVersion, _ := gitShowShortRef(config.ConfigDir, "refs/heads/main")
	remoteVersion, _ := gitShowShortRef(config.ConfigDir, "refs/remotes/origin/main")

	if localVersion == "" || remoteVersion == "" {
		updateStatus(CompConfig, StatusError, fmt.Sprintf("Could not determine local (%s) or remote (%s) versions",
			localVersion, remoteVersion), noProgress)
	} else if localVersion != remoteVersion {
		updateStatus(CompConfig, StatusWarn, fmt.Sprintf("Local version (%s) is not up to date with remote version (%s)",
			localVersion, remoteVersion), noProgress)
	} else {
		updateStatus(CompConfig, StatusOK, fmt.Sprintf("Local version is up to date with remote version (%s)",
			localVersion), noProgress)
	}
}

func checkJava(noProgress bool) {
	cmd := exec.Command(javaPath(), "-version")
	out, err := cmd.CombinedOutput()

	if err != nil {
		updateStatus(CompJava, StatusError, fmt.Sprintf("%s", err), noProgress)
		return
	}

	versionRegex := regexp.MustCompile(`(\d+\.\d+\.\d+)`)
	version := versionRegex.FindStringSubmatch(string(out))

	if version != nil {
		updateStatus(CompJava, StatusOK, version[1], noProgress)
	} else {
		updateStatus(CompJava, StatusError, "Unable to find java version", noProgress)
	}
}

func javaPath() string {
	javaHome, javaHomeDefined := os.LookupEnv("JAVA_HOME")
	if javaHomeDefined {
		return javaHome + "/bin/java"
	} else {
		return "java"
	}
}

func checkGit(noProgress bool) {
	version, err := gitVersion()

	if err != nil {
		updateStatus(CompGit, StatusError, fmt.Sprintf("Without git you can't run from source, err=%s", err), noProgress)
		return
	}

	updateStatus(CompGit, StatusOK, version, noProgress)
}

func checkOS(noProgress bool) {
	switch runtime.GOOS {
	case "windows":
		updateStatus(CompOS, StatusWarn, "Windows is not fully supported", noProgress)
	case "linux", "darwin":
		updateStatus(CompOS, StatusOK, fmt.Sprintf("%s, %s", runtime.GOOS, runtime.GOARCH), noProgress)
	}
}

func checkNetwork(config ServiceManagerConfig, noProgress bool) {
	artifactoryUrl, err := url.Parse(config.ArtifactoryPingUrl)
	if err != nil {
		updateStatus(CompVpn, StatusError, "Artifactory URL not valid", noProgress)
		return
	}

	updateStatus(CompVpn, StatusOK, fmt.Sprintf("VPN check timeout %v", config.TimeoutShort), noProgress)

	ip, err := net.LookupIP(artifactoryUrl.Host)

	if err != nil {
		updateStatus(CompVpnDns, StatusError, fmt.Sprintf("Failed to resolve IP of %s", artifactoryUrl.Host), noProgress)
	} else {
		updateStatus(CompVpnDns, StatusOK, fmt.Sprintf("IP Address of %s resolves to %v", artifactoryUrl.Host, ip), noProgress)
	}

	printStatus(CompVpn, StatusRunning, "...", noProgress)
	client := &http.Client{}
	ok, err := checkVpn(client, config)

	if ok {
		updateStatus(CompVpn, StatusOK, fmt.Sprintf("%s responds to ping", artifactoryUrl), noProgress)
	} else {
		updateStatus(CompVpn, StatusError, fmt.Sprintf("%s resolvable but not reachable - %v", artifactoryUrl, err), noProgress)
	}
}
