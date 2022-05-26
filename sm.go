package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	"sm2/cli"
	"sm2/ledger"
	"sm2/platform"
	"sm2/servicemanager"
)

func main() {

	cmds, err := cli.Parse(os.Args[1:])
	if err != nil {
		fmt.Printf("Invalid option: %s\n", err)
		os.Exit(1)
	}

	workspacePath, envIsSet := os.LookupEnv("WORKSPACE")
	if !envIsSet {
		// todo print example of how to set this up
		fmt.Println("Config issue! You need to set the WORKSPACE environment variable to poin to a folder service manager can install services to.")
		fmt.Println("add something like: export WORKSPACE=$HOME/.servicemanager to your .bashrc or .profile")
		fmt.Println("You'll need to make sure this directory exists, is writable and has sufficent space.")
		os.Exit(1)
	}

	configPath := path.Join(workspacePath, "service-manager-config")
	if cmds.Config != "" {
		configPath = cmds.Config
	}

	if stat, err := os.Stat(configPath); err != nil || !stat.IsDir() {
		fmt.Println("Config issue! Your $WORKSPACE folder needs a copy of service-manager-config.")
		fmt.Printf("This can be fixed by `cd %s` and cloning a copy of service-manager-config from github.\n", workspacePath)
		os.Exit(1)
	}

	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	config := servicemanager.ServiceManagerConfig{
		ArtifactoryRepoUrl: "https://artefacts.tax.service.gov.uk/artifactory/hmrc-releases",
		ArtifactoryPingUrl: "https://artefacts.tax.service.gov.uk/artifactory/api/system/ping",
		TmpDir:             path.Join(workspacePath, "install"),
		ConfigDir:          configPath,
	}

	serviceManager := servicemanager.ServiceManager{
		Client:   client,
		Config:   config,
		Commands: *cmds,

		Platform: platform.DetectPlatform(),
		Ledger:   ledger.NewLedger(),
	}

	// @speed consider lazy loading these rather than loading on startup
	serviceFilePath := path.Join(configPath, "services.json")
	services, err := loadServicesFromFile(serviceFilePath)
	if err != nil {
		fmt.Printf("Failed to load %s\n  %s\n", serviceFilePath, err)
		os.Exit(1)
	}

	serviceManager.Services = *services

	profileFilePath := path.Join(configPath, "profiles.json")
	profiles, err := loadProfilesFromFile(profileFilePath)
	if err != nil {
		fmt.Printf("Failed to load %s\n %s\n", profileFilePath, err)
		os.Exit(1)
	}

	serviceManager.Profiles = *profiles

	// ensure install dir exists
	err = os.MkdirAll(config.TmpDir, 0755)
	if err != nil {
		fmt.Printf("Failed to create the installation directory in %s, %s.\n", config.TmpDir, err)
		os.Exit(2)
	}

	serviceManager.Run()
}

type Services map[string]servicemanager.Service
type Profiles map[string][]string
type ServiceMappings map[string]string

func loadServicesFromFile(serviceFile string) (*Services, error) {
	services := make(Services, 1600)

	file, err := os.Open(serviceFile)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	err = json.NewDecoder(file).Decode(&services)
	if err != nil {
		return nil, err
	}

	for k, v := range services {
		// add ID into service
		v.Id = k
		services[k] = v
	}

	return &services, nil
}

// @speed do we need to cache the whole thing? we only ever look up 1 profile
//  maybe just load, find the profile and discard the rest?
func loadProfilesFromFile(profileFileName string) (*Profiles, error) {
	profiles := make(Profiles, 600)

	file, err := os.Open(profileFileName)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	err = json.NewDecoder(file).Decode(&profiles)
	return &profiles, err
}

// not used! find out if anyone still need this etc
func loadServiceMappingsFromFile(mappingsFileName string) (*ServiceMappings, error) {
	mappings := ServiceMappings{}

	file, err := os.Open(mappingsFileName)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	err = json.NewDecoder(file).Decode(&mappings)
	return &mappings, err
}
