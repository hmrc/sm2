package servicemanager

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
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

type Services map[string]Service
type Profiles map[string][]string

func loadServices(configPath string) (*Services, error) {
	var services Services
	
	// first try to load from services directory
	servicesDir := path.Join(configPath, "services")
	if stat, err := os.Stat(servicesDir); err == nil && stat.IsDir() {
		// load services from directory
		dirServices, err := loadServicesFromDirectory(servicesDir)
		if err != nil {
			return nil, fmt.Errorf("failed to load services from directory: %w", err)
		}
		services = dirServices
	} else {
		// fallback to services.json
		serviceFile := path.Join(configPath, "services.json")
		fileServices, err := loadServicesFromFile(serviceFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load services.json: %w", err)
		}
		services = fileServices
	}
	
	for k, v := range services {
		// add ID into service
		v.Id = k
		services[k] = v
	}
	
	return &services, nil
}

func loadServicesFromDirectory(servicesDir string) (Services, error) {
	services := make(Services)
	
	// use filepath.Walk to recursively process all subdirectories
	err := filepath.Walk(servicesDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// skip directories themselves, but process their contents
		// only process .json files
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".json") {
			fileServices, err := loadServicesFromFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to load %s: %w", filePath, err)
			}
			
			// check for key clashes
			for key, service := range fileServices {
				if _, exists := services[key]; exists {
					fmt.Printf("WARN: Service '%s' from file '%s' will overwrite existing definition\n", 
						key, filePath)
				}
				services[key] = service
			}
		}
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("error walking the services directory: %w", err)
	}
	
	return services, nil
}

func loadServicesFromFile(filePath string) (Services, error) {
	services := make(Services)
	
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	err = json.NewDecoder(file).Decode(&services)
	if err != nil {
		return nil, err
	}
	
	return services, nil
}

// @speed do we need to cache the whole thing? we only ever look up 1 profile
//
//	maybe just load, find the profile and discard the rest?
func loadProfiles(configPath string) (*Profiles, error) {
	profiles := make(Profiles)

	profileFile := path.Join(configPath, "profiles.json")
	file, err := os.Open(profileFile)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	err = json.NewDecoder(file).Decode(&profiles)
	return &profiles, err
}

// loads config.json which contains repo urls etc
func loadRepoConfig(configPath string) (ArtifactoryUrls, error) {

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
	configFileName := path.Join(configPath, "config.json")
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
