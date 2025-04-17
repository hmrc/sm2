package servicemanager

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sm2/version"
	"strings"
)

func update(workspaceInstallPath string) error {
	currentVersion := version.Version
	latestVersion, err := getLatestVersion()
	if err != nil {
		return err
	}

	if currentVersion == latestVersion {
		// already on latest, short-circuit
		fmt.Printf("Already up to date.\n")
		return nil
	}

	installLocation, err := getInstallLocation()
	if err != nil {
		return err
	}

	fmt.Printf("Current Version: %s\n", currentVersion)
	fmt.Printf("Latest Version:  %s\n", latestVersion)
	fmt.Printf("OS:  %s\n", runtime.GOOS)
	fmt.Printf("CPU: %s\n", runtime.GOARCH)
	fmt.Printf("Current Install Location: %s\n", installLocation)

	err = downloadAndInstall(latestVersion, workspaceInstallPath, installLocation)
	if err != nil {
		return err
	}

	return nil
}

func getLatestVersion() (string, error) {
	// create a custom client that doesn't follow redirects
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// hit releases/latest which will redirect us
	resp, err := client.Get("https://github.com/hmrc/sm2/releases/latest")
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	// extract the redirect location, should look like https://github.com/hmrc/sm2/releases/tag/v0.0.0
	location := resp.Header.Get("location")

	// split the url and extract the version from the tag
	parts := strings.Split(location, "/")
	tag := parts[len(parts)-1]
	version := strings.TrimPrefix(tag, "v")

	return version, nil
}

func getInstallLocation() (string, error) {
	output, err := os.Executable()

	if err != nil {
		fmt.Println("Unable to determine the location of the sm2 binary currently installed.")
		return "", err
	}
	return filepath.EvalSymlinks(output)
}

func downloadAndInstall(versionToInstall string, workspaceInstallPath string, installLocation string) error {

	// we'll download and inflate the zip into $WORKSPACE/install
	downloadLocation := workspaceInstallPath + "/sm2"

	// convert `darwin` to `apple` for download url
	var osForUrl string
	switch runtime.GOOS {
	case "darwin":
		osForUrl = "apple"
	case "linux":
		osForUrl = "linux"
	default:
		log.Fatalf("unsupported OS: %s", runtime.GOOS)
	}

	// convert `amd64` to `intel` for download url
	var arch string
	switch runtime.GOARCH {
	case "amd64":
		arch = "intel"
	case "arm64":
		arch = "arm64"
	default:
		log.Fatalf("unsupported CPU architecture %s", runtime.GOARCH)
	}

	downloadUrl := fmt.Sprintf("https://github.com/hmrc/sm2/releases/download/v%s/sm2-%s-%s-%s.zip", versionToInstall, versionToInstall, osForUrl, arch)

	fmt.Printf("Downloading %s...\n", downloadUrl)

	resp, err := http.Get(downloadUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return err
	}

	for _, zipFile := range zipReader.File {
		if zipFile.Name == "sm2" {
			rc, err := zipFile.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			fileData, err := io.ReadAll(rc)
			if err != nil {
				return err
			}

			fmt.Printf("Unzipping into %s...\n", downloadLocation)

			err = os.WriteFile(downloadLocation, fileData, 0755)
			if err != nil {
				return err
			}
		}
	}

	fmt.Printf("Moving new sm2 binary from %s to %s...\n", downloadLocation, installLocation)

	// attempt to move without sudo first
	err = exec.Command("mv", downloadLocation, installLocation).Run()
	if err != nil {
		fmt.Printf("Moving the new sm2 binary to %s requires `sudo` - you may be prompted for your password...\n", installLocation)
		// failed so attempting with sudo
		err = exec.Command("sudo", "mv", downloadLocation, installLocation).Run()
		if err != nil {
			return err
		}
	}

	fmt.Printf("Successfully installed v%s!\n", versionToInstall)

	return nil
}
