package servicemanager

import (
	"context"
	"io"
	"net/http"
	"time"
)

// Tests vpn connectivity by attempting to open a connection
// to artifactory using a http client with a short timeout.
func checkVpn(client *http.Client, config ServiceManagerConfig) bool {

	// TODO: move short timeout to config
	shortTimeout := 4 * time.Second
	ctx, _ := context.WithTimeout(context.Background(), shortTimeout)

	req, err := http.NewRequestWithContext(ctx, "HEAD", config.ArtifactoryPingUrl, nil)
	if err != nil {
		println(err)
		return false
	}

	resp, err := client.Do(req)
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
