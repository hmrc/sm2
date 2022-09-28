package servicemanager

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "sm2/testing"
)

func TestCheckVpn(t *testing.T) {

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))

	defer svr.Close()

	client := &http.Client{}
	config := ServiceManagerConfig{
		ArtifactoryPingUrl: svr.URL,
		TimeoutShort:       4 * time.Second,
	}

	res, err := checkVpn(client, config)

	AssertNotErr(t, err)

	if res == false {
		t.Errorf("vpn check didn't error but failed")
	}
}

func TestCheckVpnFails(t *testing.T) {

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))

	defer svr.Close()

	client := &http.Client{}
	config := ServiceManagerConfig{
		ArtifactoryPingUrl: "http://254.254.254.254/ping",
		TimeoutShort:       1 * time.Millisecond,
	}

	res, err := checkVpn(client, config)

	if err == nil {
		t.Errorf("expected vpn failure, it didnt")
	}
	if res == true {
		t.Errorf("vpn check didn't error but failed")
	}

}
