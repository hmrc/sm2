package servicemanager

import (
	"net/http"
	"time"
)

// Tests vpn connectivity by attempting to open a connection
// to artifactory using a http client with a short timeout.
func checkVpn(config ServiceManagerConfig) bool {
	shortTimeoutClient := &http.Client{
		Timeout: 4 * time.Second,
	}
	_, err := shortTimeoutClient.Head(config.ArtifactoryRepoUrl)
	return err == nil
}
