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
	ProxyPaths  []string      `json:"proxyPaths"`
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

const DEFAULT_WORKSPACE = ".sm2"

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

	// use the default workspace path if one isn't set
	if !envIsSet {
		defaultWorkspacePath, err := createDefaultWorkspace()
		if err != nil {
			return fmt.Errorf("Failed to create the default workspace in %v. Check this is writable or override the default workspace path by setting a WORKSPACE environment variable. %v", workspacePath, err)
		}
		workspacePath = defaultWorkspacePath
	}

	// ensure workspace isn't a relative path like ./sm2 or ../../../sm2 or some weirdness
	if !path.IsAbs(workspacePath) {
		return fmt.Errorf("Config issue! Your WORKSPACE environment variable must be an absolute path:\ni.e. starting with a '/' like '/home/user/.servicemanager'\n")
	}

	// check service-manager-config is present
	configPath := path.Join(workspacePath, "service-manager-config")
	if sm.Commands.Config != "" {
		configPath = sm.Commands.Config
	}

	if stat, err := os.Stat(configPath); err != nil || !stat.IsDir() {
		return fmt.Errorf("Setup incomplete! No copy of service-manager-config found in your workspace (%s).\n"+
			"This can be fixed by `cd %s` and cloning a copy of service-manager-config from github.\n", workspacePath, workspacePath)
	}

	// load repo details from config.json
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

func createDefaultWorkspace() (string, error) {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("Failed to lookup users home dir! %v", err)
	}
	workspacePath := path.Join(homeDir, DEFAULT_WORKSPACE)

	// check if folder exists
	if Exists(workspacePath) {
		return workspacePath, nil
	}

	// create it if it doesn't
	fmt.Printf("Creating default workspace in %s...\n", workspacePath)
	return workspacePath, os.MkdirAll(workspacePath, 0755)
}
