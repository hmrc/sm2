package servicemanager

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"sm2/version"
)

func RunDiagnostics(config ServiceManagerConfig) {

	version.PrintVersion()

	checkNetwork(config)
	checkWorkspace(config)
	checkConfigRevision(config)
	checkJava()
	checkGit()
	checkOS()
}

func checkWorkspace(config ServiceManagerConfig) {
	stat, err := os.Stat(config.TmpDir)
	if err != nil {
		fmt.Printf("WORKSPACE:\t NOT OK (%s)\n", err)
		return
	}

	if !stat.IsDir() {
		fmt.Printf("WORKSPACE:\t NOT OK (%s is not a directory)\n", config.TmpDir)
		return
	}

	fmt.Printf("WORKSPACE:\t OK (%s)\n", config.TmpDir)
}

func checkConfigRevision(config ServiceManagerConfig) {
	cmd := exec.Command("git", "log", "--pretty=format:%h,%ar,%s", "-1")
	cmd.Dir = config.ConfigDir

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("CONFIG:\t\t NOT OK: %s\n", err)
	}

	fmt.Printf("CONFIG:\t\t OK (%s)\n", string(out))
}

func checkJava() {
	cmd := exec.Command(javaPath(), "-version")

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("JAVA:\t\t NOT OK: %s\n", err)
		return
	}

	versionRegex := regexp.MustCompile(`(\d+\.\d+\.\d+)`)
	version := versionRegex.FindStringSubmatch(string(out))
	if version != nil {
		fmt.Printf("JAVA:\t\t OK (%s)\n", version[1])
	} else {
		fmt.Print("JAVA:\t\t NOT OK: unable to find java version\n")
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
	cmd := exec.Command("git", "--version")

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("GIT:\t\t NOT OK: %s\n", err)
		fmt.Printf("\t\t without git you can't run from source: %s\n", err)

		return
	}
	fmt.Printf("GIT:\t\t OK (%s)\n", strings.Trim(string(out), "\n "))
}

func checkOS() {
	switch runtime.GOOS {
	case "windows":
		fmt.Print("OS:\t\t WARN: windows is not fully supported\n")

	case "linux", "darwin":
		fmt.Printf("OS:\t\t OK (%s, %s)\n", runtime.GOOS, runtime.GOARCH)
	}
}

func checkNetwork(config ServiceManagerConfig) {
	artifactoryUrl, err := url.Parse(config.ArtifactoryRepoUrl)
	if err != nil {
		fmt.Print("VPN:\t\t artifactory url not valid!\n")
		return
	}

	_, err = net.LookupHost(artifactoryUrl.Host)
	if err != nil {
		fmt.Print("VPN:\t\t NOT OK\n")
		fmt.Printf("\t\t %s is not resolvable, check VPN\n", artifactoryUrl)
		return
	}

	if !checkVpn(config) {
		fmt.Print("VPN:\t\t NOT OK\n")
		fmt.Printf("\t\t %s resolvable but not reachable\n", artifactoryUrl)
	} else {
		fmt.Printf("VPN:\t\t OK (%s resolvable)\n", artifactoryUrl)
	}

}
