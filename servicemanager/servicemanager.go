package servicemanager

import (
	"fmt"
	"net/http"
	"os"
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

func (sm ServiceManager) PrintVerbose(s string, args ...interface{}) {
	if sm.Commands.Verbose {
		fmt.Printf(s, args...)
	}
}

// based on config, find the directory a service is installed into.
// TODO: rename this something less confusing. it doesnt really 'find' anything,
//       rather it guesses where it is...
func (sm *ServiceManager) findInstallDirOfService(serviceName string) (string, error) {
	if service, ok := sm.Services[serviceName]; ok {
		return path.Join(sm.Config.TmpDir, service.Binary.DestinationSubdir), nil
	}
	return "", fmt.Errorf("unknown service: %s", serviceName)
}

func (sm *ServiceManager) LoadConfig() {
	workspacePath, envIsSet := os.LookupEnv("WORKSPACE")
	if !envIsSet {
		// todo print example of how to set this up
		fmt.Println("Config issue! You need to set the WORKSPACE environment variable to poin to a folder service manager can install services to.")
		fmt.Println("add something like: export WORKSPACE=$HOME/.servicemanager to your .bashrc or .profile")
		fmt.Println("You'll need to make sure this directory exists, is writable and has sufficent space.")
		os.Exit(1)
	}

	configPath := path.Join(workspacePath, "service-manager-config")
	if sm.Commands.Config != "" {
		configPath = sm.Commands.Config
	}

	if stat, err := os.Stat(configPath); err != nil || !stat.IsDir() {
		fmt.Println("Config issue! Your $WORKSPACE folder needs a copy of service-manager-config.")
		fmt.Printf("This can be fixed by `cd %s` and cloning a copy of service-manager-config from github.\n", workspacePath)
		os.Exit(1)
	}

	// load repo details from config.json
	// @todo does this need to return an error if loader can return safe default?
	configJsonFileName := path.Join(configPath, "config.json")
	repoConfig, err := loadRepoConfig(configJsonFileName)
	if err != nil {
		repoConfig = DefaultArtifactoryUrls
	}

	sm.Config = ServiceManagerConfig{
		ArtifactoryRepoUrl: repoConfig.RepoUrl,
		ArtifactoryPingUrl: repoConfig.PingUrl,
		TmpDir:             path.Join(workspacePath, "install"),
		ConfigDir:          configPath,
	}

	// @speed consider lazy loading these rather than loading on startup
	serviceFilePath := path.Join(configPath, "services.json")
	services, err := loadServicesFromFile(serviceFilePath)
	if err != nil {
		fmt.Printf("Failed to load %s\n  %s\n", serviceFilePath, err)
		os.Exit(1)
	}
	sm.Services = *services

	profileFilePath := path.Join(configPath, "profiles.json")
	profiles, err := loadProfilesFromFile(profileFilePath)
	if err != nil {
		fmt.Printf("Failed to load %s\n %s\n", profileFilePath, err)
		os.Exit(1)
	}
	sm.Profiles = *profiles

	// ensure install dir exists
	err = os.MkdirAll(sm.Config.TmpDir, 0755)
	if err != nil {
		fmt.Printf("Failed to create the installation directory in %s, %s.\n", sm.Config.TmpDir, err)
		os.Exit(1)
	}
}
