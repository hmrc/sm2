package servicemanager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
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

func getPort(u string) int {
	url, _ := url.Parse(u)
	port, _ := strconv.Atoi(url.Port())
	return port
}

// In order to keep the tables looking nice on everyones default terminals we cap the status at 80 chars
func TestStatusFitsIn80Chars(t *testing.T) {
	sb := bytes.NewBufferString("")
	statuses := []serviceStatus{
		serviceStatus{0, 1, "SHORT_ID", "1.2.3", "PASS"},
		serviceStatus{1386351, 10801, "A_REALLY_LONG_SERVICE_NAME_INDEED_IT_IS_VERY_VERY_VERY_LONG", "42.999.1", "PASS"},
	}
	printTable(statuses, sb)

	lines := strings.Split(sb.String(), "\n")
	for _, line := range lines {

		line = strings.ReplaceAll(line, "\033[;32m", "")
		line = strings.ReplaceAll(line, "\033[0m", "")
		fmt.Printf("%d: [%s]\n", len(line), line)
		if len(line) > 80+11 {
			t.Errorf("Status line is > 80 chars (%d)", len(line))
		}
	}
}

func TestBSTUptimeBug(t *testing.T) {
	startedStr := "2022-06-15T10:25:52.113165678+01:00"
	uptimeStr := "2022-06-15 10:17:10"

	jsonState := fmt.Sprintf(`{"Service":"SERVICE_CONFIGS","Artifact":"service-configs_2.13","Version":"0.117.0","Path":"/tmp","Started":"%s","Pid":1,"Port":8460,"Args":[]}`, startedStr)
	var state ledger.StateFile
	json.Unmarshal([]byte(jsonState), &state)

	bootTime, err := time.ParseInLocation("2006-01-02 15:04:05", uptimeStr, time.Local)
	if err != nil {
		t.Error(err)
	}

	if state.Started.Before(bootTime) {
		t.Errorf("started (%v) is BEFORE bootTime (%v), though this is not actually the case", state.Started, bootTime)
	}
}

func TestExcudeFailedStatusFromPreviousBoot(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer svr.Close()

	baseTime := time.Date(2020, time.Month(2), 1, 12, 0, 0, 0, time.UTC)
	uptime := func() time.Time {
		return baseTime.Add(time.Hour * 10)
	}
	mockStates := []ledger.StateFile{
		// failed service started prior to the last boot
		{
			Service: "FOO",
			Started: baseTime,
			Port:    0,
			Pid:     0,
		},
	}

	sm := ServiceManager{
		Client:   &http.Client{},
		Platform: platform.Platform{Uptime: uptime, PidLookup: mockPidLookup},
		Ledger: ledger.Ledger{
			FindAllStateFiles: func(_ string) ([]ledger.StateFile, error) {
				return mockStates, nil
			},
			ClearStateFile: func(_ string) error {
				return nil
			},
		},
	}

	result := sm.findStatuses()
	if len(result) != 0 {
		t.Errorf("results has %d items expected 0", len(result))
		for _, r := range result {
			fmt.Printf("%s %s\n, ", r.service, string(r.health))
		}
	}

}

func TestCustomPingUrls(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.EscapedPath() == "/pong" {
			w.WriteHeader(200)
		} else {
			w.WriteHeader(404)
		}
	}))
	defer svr.Close()

	port := getPort(svr.URL)

	states := []ledger.StateFile{
		// healthy service (pid & ping)
		{
			Service:        "FOO",
			Started:        time.Now().Add(time.Duration(-1) * time.Hour),
			Port:           port,
			Pid:            9999,
			HealthcheckUrl: fmt.Sprintf("http://localhost:%d/pong", port),
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

	if len(result) != 1 {
		t.Errorf("results has %d items expected 1", len(result))
	}

	if result[0].health != PASS {
		t.Errorf("%s health was not PASS, it was %s", result[0].service, result[0].health)
	}
}

func TestStatuses(t *testing.T) {
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
