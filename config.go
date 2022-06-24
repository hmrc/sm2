package main

import (
	"encoding/json"
	"fmt"

	"os"
	"sm2/servicemanager"
)

type ArtifactoryUrls struct {
	PingUrl string
	RepoUrl string
}

// @todo set the defaults at build time maybe, the same way we do the version?
var DefaultArtifactoryUrls = ArtifactoryUrls{
	RepoUrl: "https://artefacts.tax.service.gov.uk/artifactory/hmrc-releases",
	PingUrl: "https://artefacts.tax.service.gov.uk/artifactory/api/system/ping",
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
func loadRepoConfig(configFileName string) (ArtifactoryUrls, error) {

	// structs for unmarshalling config.json into...
	type repoConfig struct {
		Protocol     string            `json:"protocol"`
		Host         string            `json:"host"`
		RepoMappings map[string]string `json:"repoMappings"`
		Ping         string            `json:"ping"`
	}

	type smConfig struct {
		Artifactory repoConfig `json:"artifactory"`
	}

	urls := DefaultArtifactoryUrls

	file, err := os.Open(configFileName)
	if err != nil {
		// if the file is missing for some reason, just carry on with the (hopefully ok) defaults
		return urls, nil
	}

	defer file.Close()

	config := smConfig{}
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		return urls, err
	}

	// otherwise set the urls based on the config.json content
	if repoPath, ok := config.Artifactory.RepoMappings["RELEASE"]; ok {
		urls.RepoUrl = fmt.Sprintf("%s://%s/%s", config.Artifactory.Protocol, config.Artifactory.Host, repoPath)
	}

	if config.Artifactory.Ping != "" {
		urls.PingUrl = fmt.Sprintf("%s://%s/%s", config.Artifactory.Protocol, config.Artifactory.Host, config.Artifactory.Ping)
	}

	return urls, nil
}
