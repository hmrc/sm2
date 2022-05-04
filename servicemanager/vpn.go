package servicemanager

import (
	"io"
	"net/http"
	"time"
)

// Tests vpn connectivity by attempting to open a connection
// to artifactory using a http client with a short timeout.
func checkVpn(config ServiceManagerConfig) bool {
	shortTimeoutClient := &http.Client{
		Timeout: 4 * time.Second,
	}
	resp, err := shortTimeoutClient.Head(config.ArtifactoryPingUrl)
	if err != nil {
		return false
	}

	defer resp.Body.Close()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	return resp.StatusCode == 200

}
