package servicemanager

import (
	"net/http"
	"time"
)

// Tests vpn connectivity by attempting to open a connection
// to artifactory using a http client with a short timeout.
func (sm *ServiceManager) checkVpn() bool {
	shortTimeoutClient := &http.Client{
		Timeout: 4 * time.Second,
	}
	_, err := shortTimeoutClient.Head(sm.Config.ArtifactoryRepoUrl)
	return err == nil
}
