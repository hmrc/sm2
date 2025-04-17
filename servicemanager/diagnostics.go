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
	"strings"

	"sm2/version"
)

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
)

// Status constants
const (
	StatusRunning = "RUNNING"
	StatusOK      = "OK"
	StatusError   = "ERROR"
	StatusWarn    = "WARN"
	StatusInfo    = "INFO"
)

// Component name constants
const (
	CompOS        = "OS"
	CompJava      = "JAVA"
	CompGit       = "GIT"
	CompConfig    = "CONFIG"
	CompWorkspace = "WORKSPACE"
	CompVpnDns    = "VPN DNS"
	CompVpn       = "VPN"
)

func startStatus(component string) {
	printStatus(component, StatusRunning, "...")
}

// Helper function to print status with appropriate color
func printStatus(component, status, details string) {
	var colorCode string

	switch status {
	case StatusRunning:
		colorCode = ColorYellow
	case StatusOK:
		colorCode = ColorGreen
	case StatusError:
		colorCode = ColorRed
	case StatusWarn:
		colorCode = ColorYellow
	case StatusInfo:
		colorCode = ColorReset
	default:
		colorCode = ColorReset
	}

	// Format component name to be exactly 15 characters
	formattedComponent := component
	if len(component) > 15 {
		// Truncate if longer than 15 characters
		formattedComponent = component[:15] + ":"
	} else if len(component) < 15 {
		// Pad with spaces if shorter than 15 characters
		formattedComponent = component + ":" + strings.Repeat(" ", 14-len(component))
	}

	formattedStatus := fmt.Sprintf("%s%s%s", colorCode, status, ColorReset)

	fmt.Printf("%s%s (%s)\n", formattedComponent, formattedStatus, details)
}

// Helper function to update status for a running task
func updateStatus(component string, status string, details string) {
	// Move cursor up one line and clear the line
	fmt.Print("\033[1A\033[K")
	printStatus(component, status, details)
}

func RunDiagnostics(config ServiceManagerConfig) {
	version.PrintVersion()

	startStatus(CompOS)
	checkOS()

	startStatus(CompJava)
	checkJava()

	startStatus(CompGit)
	checkGit()

	startStatus(CompConfig)
	checkConfigRevision(config)

	startStatus(CompWorkspace)
	checkWorkspace(config)

	startStatus(CompVpn)
	checkNetwork(config)
}

func checkWorkspace(config ServiceManagerConfig) {
	stat, err := os.Stat(config.TmpDir)
	if err != nil {
		updateStatus(CompWorkspace, StatusError, err.Error())
		return
	}

	if !stat.IsDir() {
		updateStatus(CompWorkspace, StatusError, fmt.Sprintf("%s is not a directory", config.TmpDir))
		return
	}

	updateStatus(CompWorkspace, StatusOK, config.TmpDir)
}

func checkConfigRevision(config ServiceManagerConfig) {
	err := gitFetch(config.ConfigDir, "origin", "main")
	if err != nil {
		updateStatus(CompConfig, StatusWarn, "Unable to fetch latest remote version")
		return
	}

	localVersion, _ := gitShowShortRef(config.ConfigDir, "refs/heads/main")
	remoteVersion, _ := gitShowShortRef(config.ConfigDir, "refs/remotes/origin/main")

	if localVersion == "" || remoteVersion == "" {
		updateStatus(CompConfig, StatusError, fmt.Sprintf("Could not determine local (%s) or remote (%s) versions",
			localVersion, remoteVersion))
	} else if localVersion != remoteVersion {
		updateStatus(CompConfig, StatusWarn, fmt.Sprintf("Local version (%s) is not up to date with remote version (%s)",
			localVersion, remoteVersion))
	} else {
		updateStatus(CompConfig, StatusOK, fmt.Sprintf("Local version is up to date with remote version (%s)",
			localVersion))
	}
}

func checkJava() {
	cmd := exec.Command(javaPath(), "-version")
	out, err := cmd.CombinedOutput()

	if err != nil {
		updateStatus(CompJava, StatusError, fmt.Sprintf("%s", err))
		return
	}

	versionRegex := regexp.MustCompile(`(\d+\.\d+\.\d+)`)
	version := versionRegex.FindStringSubmatch(string(out))

	if version != nil {
		updateStatus(CompJava, StatusOK, version[1])
	} else {
		updateStatus(CompJava, StatusError, "Unable to find java version")
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

func checkGit() {
	version, err := gitVersion()

	if err != nil {
		updateStatus(CompGit, StatusError, fmt.Sprintf("Without git you can't run from source, err=%s", err))
		return
	}

	updateStatus(CompGit, StatusOK, version)
}

func checkOS() {
	switch runtime.GOOS {
	case "windows":
		updateStatus(CompOS, StatusWarn, "Windows is not fully supported")
	case "linux", "darwin":
		updateStatus(CompOS, StatusOK, fmt.Sprintf("%s, %s", runtime.GOOS, runtime.GOARCH))
	}
}

func checkNetwork(config ServiceManagerConfig) {
	artifactoryUrl, err := url.Parse(config.ArtifactoryPingUrl)
	if err != nil {
		updateStatus(CompVpn, StatusError, "Artifactory URL not valid")
		return
	}

	updateStatus(CompVpn, StatusOK, fmt.Sprintf("VPN check timeout %v", config.TimeoutShort))

	ip, err := net.LookupIP(artifactoryUrl.Host)

	if err != nil {
		updateStatus(CompVpnDns, StatusError, fmt.Sprintf("Failed to resolve IP of %s", artifactoryUrl.Host))
	} else {
		updateStatus(CompVpnDns, StatusOK, fmt.Sprintf("IP Address of %s resolves to %v", artifactoryUrl.Host, ip))
	}

	printStatus(CompVpn, StatusRunning, "...")
	client := &http.Client{}
	ok, err := checkVpn(client, config)

	if ok {
		updateStatus(CompVpn, StatusOK, fmt.Sprintf("%s responds to ping", artifactoryUrl))
	} else {
		updateStatus(CompVpn, StatusError, fmt.Sprintf("%s resolvable but not reachable - %v", artifactoryUrl, err))
	}
}
