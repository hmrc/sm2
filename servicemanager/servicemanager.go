package servicemanager

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

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
	TimeoutShort       time.Duration
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
	Repo        string   `json:"repo"`
	ExtraParams []string `json:"extra_params"`
}

type Healthcheck struct {
	Url      string `json:"url"`
	Response string `json:"response"`
}

const DEFAULT_SHORT_TIMEOUT = 20

func (sm ServiceManager) PrintVerbose(s string, args ...interface{}) {
	if sm.Commands.Verbose {
		fmt.Printf(s, args...)
	}
}

func (sm *ServiceManager) NewShortContext() context.Context {
	ttl := sm.Config.TimeoutShort
	if ttl == 0 {
		ttl = DEFAULT_SHORT_TIMEOUT * time.Second
	}
	ctx, _ := context.WithTimeout(context.Background(), ttl)
	return ctx
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

func (sm *ServiceManager) LoadConfig() error {
	workspacePath, envIsSet := os.LookupEnv("WORKSPACE")
	if !envIsSet {
		// todo print example of how to set this up
		return fmt.Errorf("Config issue! You need to set the WORKSPACE environment variable to point to a folder service manager can install services to.\n" +
			"add something like: export WORKSPACE=$HOME/.servicemanager to your .bashrc or .profile\n" +
			"You'll need to make sure this directory exists, is writable and has sufficent space.\n")
	}

	configPath := path.Join(workspacePath, "service-manager-config")
	if sm.Commands.Config != "" {
		configPath = sm.Commands.Config
	}

	if stat, err := os.Stat(configPath); err != nil || !stat.IsDir() {
		return fmt.Errorf("Config issue! Your $WORKSPACE folder needs a copy of service-manager-config.\n"+
			"This can be fixed by `cd %s` and cloning a copy of service-manager-config from github.\n", workspacePath)
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
		TimeoutShort:       DEFAULT_SHORT_TIMEOUT * time.Second,
	}

	// allow for short timout (vpn check etc) to be overriden in case of network weirdness
	if timeout, isSet := os.LookupEnv("SM_TIMEOUT"); isSet {
		if value, err := strconv.ParseInt(timeout, 10, 64); err == nil {
			sm.Config.TimeoutShort = time.Second * time.Duration(value)
		}
	}

	// @speed consider lazy loading these rather than loading on startup
	serviceFilePath := path.Join(configPath, "services.json")
	services, err := loadServicesFromFile(serviceFilePath)
	if err != nil {
		return fmt.Errorf("Failed to load %s\n  %s\n", serviceFilePath, err)
	}
	sm.Services = *services

	profileFilePath := path.Join(configPath, "profiles.json")
	profiles, err := loadProfilesFromFile(profileFilePath)
	if err != nil {
		return fmt.Errorf("Failed to load %s\n %s\n", profileFilePath, err)
	}
	sm.Profiles = *profiles

	// ensure install dir exists
	err = os.MkdirAll(sm.Config.TmpDir, 0755)
	if err != nil {
		return fmt.Errorf("Failed to create the installation directory in %s, %s.\n", sm.Config.TmpDir, err)
	}

	return nil
}
