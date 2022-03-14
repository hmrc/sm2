package servicemanager

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"sm2/ledger"
	"sm2/platform"
)

func mockPidLookup() map[int]int {
	return map[int]int{
		9999: 9999,
		7777: 7777,
	}
}

func mockUptime() time.Time {
	return time.Now().Add(time.Duration(-2) * time.Hour)
}

func TestFindStatuses(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer svr.Close()

	port := getPort(svr.URL)

	states := []ledger.StateFile{
		// healthy service (pid & ping)
		{
			Service: "FOO",
			Started: time.Now().Add(time.Duration(-1) * time.Hour),
			Port:    port,
			Pid:     9999,
		},
		// booting service (pid, no ping)
		{
			Service: "BAR",
			Started: time.Now(),
			Port:    0,
			Pid:     7777,
		},
		// dead service (no pid)
		{
			Service: "BAZ",
			Started: time.Now(),
			Port:    port,
			Pid:     6666,
		},
	}

	sm := ServiceManager{
		Client:   &http.Client{},
		Platform: platform.Platform{Uptime: mockUptime, PidLookup: mockPidLookup},
		Ledger: ledger.Ledger{
			FindAllStateFiles: func(_ string) ([]ledger.StateFile, error) {
				return states, nil
			},
		},
	}

	result := sm.findStatuses()

	if len(result) != 3 {
		t.Errorf("results has %d items expected 3", len(result))
	}

	if result[0].health != PASS {
		t.Errorf("%s health was not PASS, it was %s", result[0].service, result[0].health)
	}
	if result[1].health != BOOT {
		t.Errorf("%s health was not BOOT, it was %s", result[1].service, result[1].health)
	}
	if result[2].health != FAIL {
		t.Errorf("%s health was not FAIL, it was %s", result[2].service, result[2].health)
	}
}

func getPort(u string) int {
	url, _ := url.Parse(u)
	port, _ := strconv.Atoi(url.Port())
	return port
}
