package servicemanager

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// Tests vpn connectivity by attempting to open a connection
// to artifactory using a http client with a short timeout.
func checkVpn(client *http.Client, config ServiceManagerConfig) (bool, error) {

	ctx, cancel := context.WithTimeout(context.Background(), config.TimeoutShort)

	// cleanup the context
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", config.ArtifactoryPingUrl, nil)
	if err != nil {
		return false, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}

	defer resp.Body.Close()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	if resp.StatusCode != 200 {
		return false, fmt.Errorf("vpn check failed, http status %d", resp.StatusCode)
	}

	return true, nil

}
