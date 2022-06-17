package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sm2/servicemanager"
)

// structs for unmarshalling config.json into...
type SmConfig struct {
	Artifactory RepoConfig `json:"artifactory"`
}

type RepoConfig struct {
	Protocol     string            `json:"protocol"`
	Host         string            `json:"host"`
	RepoMappings map[string]string `json:"repoMappings"`
	Ping         string            `json:"ping"`
}

func (cfg RepoConfig) url() string {
	if repoPath, ok := cfg.RepoMappings["RELEASE"]; ok {
		return fmt.Sprintf("%s://%s/%s", cfg.Protocol, cfg.Host, repoPath)
	}
	return "https://artefacts.tax.service.gov.uk/artifactory/hmrc-releases"
}

func (cfg RepoConfig) ping() string {
	if cfg.Ping != "" {
		return fmt.Sprintf("%s://%s/%s", cfg.Protocol, cfg.Host, cfg.Ping)
	}
	return "https://artefacts.tax.service.gov.uk/artifactory/api/system/ping"
}

var defaultRepoConfig = RepoConfig{
	Protocol:     "https",
	Host:         "artefacts.tax.service.gov.uk",
	RepoMappings: map[string]string{"RELEASE": "artifactory/hmrc-releases"},
	Ping:         "artifactory/api/system/ping",
}

type Services map[string]servicemanager.Service
type Profiles map[string][]string

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

// loads config.json which contains repo urls etc
func loadRepoConfig(configFileName string) (*RepoConfig, error) {
	config := SmConfig{}

	file, err := os.Open(configFileName)
	if err != nil {
		// ignore the file not being there, use the default instead
		return &defaultRepoConfig, nil
	}

	defer file.Close()

	err = json.NewDecoder(file).Decode(&config)
	return &config.Artifactory, err
}
