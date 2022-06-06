package servicemanager

import (
	"fmt"
	"net/http"
	"path"

	"sm2/cli"
	"sm2/ledger"
	"sm2/platform"
)

type ServiceManager struct {
	Client   *http.Client
	Services map[string]Service
	Profiles map[string][]string
	Config   ServiceManagerConfig
	Commands cli.UserOption
	progress ProgressRenderer
	Platform platform.Platform
	Ledger   ledger.Ledger
}

type ServiceManagerConfig struct {
	TmpDir             string
	VpnTestHostname    string
	ArtifactoryRepoUrl string
	ArtifactoryPingUrl string
	ConfigDir          string
}

type Service struct {
	Id          string
	Name        string        `json:"name"`
	DefaultPort int           `json:"defaultPort"`
	Template    string        `json:"template"`
	Frontend    bool          `json:"frontend"`
	Source      Source        `json:"sources"`
	Binary      ServiceBinary `json:"binary"`
	Location    string        `json:"location"`
	Healthcheck Healthcheck   `json:"healthcheck"`
}

type ServiceBinary struct {
	Artifact          string   `json:"artifact"`
	GroupId           string   `json:"groupId"`
	DestinationSubdir string   `json:"destinationSubdir"`
	Cmd               []string `json:"cmd"`
}

type Source struct {
	Repo string `json:"repo"`
}

type Healthcheck struct {
	Url      string `json:"url"`
	Response string `json:"response"`
}

func (sm ServiceManager) PrintVerbose(s string, args ...string) {
	if sm.Commands.Verbose {
		fmt.Printf(s, args)
	}
}

/*
func (sm ServiceManager) whereIsServiceInstalled(serviceName, version string) (string, error) {

	// lookup service
	service, ok := sm.Services[serviceName]
	if !ok {
		return "", fmt.Errorf("unknown service: %s", serviceName)
	}

	// lookup .install file
	installDir := path.Join(sm.Config.TmpDir, service.Binary.DestinationSubdir)
	installFile, err := sm.Ledger.LoadInstallFile(installDir)
	if err != nil {
		return "", fmt.Errorf("no .install found in %s", installFile)
	}

	// verify its the right one
	if version != "" && installFile.Version != version {
		return "", fmt.Errorf("wrong version installed")
	}

	// and that it actually exists
	if _, err := os.Stat(installFile.Path); os.IsNotExist(err) {
		return "", err
	}

	// return the path
	return installFile.Path, nil
}
*/

// based on config, find the directory a service is installed into.
// TODO: rename this something less confusing. it doesnt really 'find' anything,
//       rather it guesses where it is...
func (sm *ServiceManager) findInstallDirOfService(serviceName string) (string, error) {
	if service, ok := sm.Services[serviceName]; ok {
		return path.Join(sm.Config.TmpDir, service.Binary.DestinationSubdir), nil
	}
	return "", fmt.Errorf("unknown service: %s", serviceName)
}
