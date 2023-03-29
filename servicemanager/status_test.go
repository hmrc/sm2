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

func TestFindStatusesSortsResultsAlphabetically(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer svr.Close()

	port := getPort(svr.URL)
	mockStates := []ledger.StateFile{
		{
			Service: "FOO",
			Started: time.Now().Add(time.Duration(-1) * time.Hour),
			Port:    port,
			Pid:     9999,
		},
		{
			Service: "BAR",
			Started: time.Now(),
			Port:    0,
			Pid:     7777,
		},
		{
			Service: "WORLD",
			Started: time.Now(),
			Port:    port,
			Pid:     6666,
		},
		{
			Service: "HELLO",
			Started: time.Now(),
			Port:    port,
			Pid:     6666,
		},
	}

	mockMongoStatus := serviceStatus{
		pid:     0,
		port:    27017,
		service: "MONGO",
		health:  FAIL,
	}

	mockServicePidLookup := func(s string) (bool, []int) {
		return false, []int{}
	}

	mockPortPidLookup := func() map[int]int {
		return map[int]int{}
	}

	mockGetTerminal := func() (int, int) {
		return 80, 25
	}

	sm := ServiceManager{
		Client: &http.Client{},
		Platform: platform.Platform{
			Uptime:             mockUptime,
			PidLookup:          mockPidLookup,
			PidLookupByService: mockServicePidLookup,
			PortPidLookup:      mockPortPidLookup,
			GetTerminalSize:    mockGetTerminal,
		},
		Ledger: ledger.Ledger{
			FindAllStateFiles: func(_ string) ([]ledger.StateFile, error) {
				return mockStates, nil
			},
		},
	}

	statuses := []serviceStatus{mockMongoStatus}

	//Sorting occurs within the `findStatuses() function
	statuses = append(statuses, sm.findStatuses()...)

	if statuses[0].service != "MONGO" {
		t.Errorf("First service was not MONGO, it was %s", statuses[0].service)
	}
	if statuses[1].service != "BAR" {
		t.Errorf("Second service was not BAR, it was %s", statuses[1].service)
	}
	if statuses[2].service != "FOO" {
		t.Errorf("Third service was not FOO, it was %s", statuses[2].service)
	}
	if statuses[3].service != "HELLO" {
		t.Errorf("Fourth service was not HELLO, it was %s", statuses[3].service)
	}
	if statuses[4].service != "WORLD" {
		t.Errorf("Fifth service was not WORLD, it was %s", statuses[4].service)
	}
}

func TestStatusWrapsServiceNames(t *testing.T) {
	sb := bytes.NewBufferString("")
	statuses := []serviceStatus{
		serviceStatus{0, 1, "SHORT_ID", "1.2.3", "PASS"},
		serviceStatus{123, 10801, "THE_SERVICE_IS_35_CHARS_DO_NOT_WRAP", "42.999.1", "PASS"},
		serviceStatus{2, 3, "SERVICE_IS_38_CHARS_STILL_CROP_IT_OKAY", "1.5", "PASS"},
		serviceStatus{3, 4, "SERVICE_IS_39_CHARS_SO_WRAP_OVERFLOW_OK", "2.8", "PASS"},
		serviceStatus{4, 5, "SERVICE_IS_54_CHARS_SO_DEFINITELY_WRAP_THE_OVERFLOW_OK", "3.1", "PASS"},
		serviceStatus{5, 6, "SERVICE_IS_73_CHARS_SO_DEFINITELY_CROP_THE_SECOND_LINE_SO_NO_3RD_OVERFLOW", "3.2", "PASS"},
		serviceStatus{6, 7, "SERVICE_IS_74_CHARS_SO_DEFINITELY_WRAP_THE_3RD_LINE_SO_WE_CAN_SEE_OVERFLOW", "3.3", "PASS"},
	}
	expectedOutput :=
		`+---------------------------------------+-----------+---------+-------+--------+
| Name                                  | Version   | PID     | Port  | Status |
+---------------------------------------+-----------+---------+-------+--------+
| SHORT_ID                              | 1.2.3     | 0       | 1     |  PASS  |
| THE_SERVICE_IS_35_CHARS_DO_NOT_WRAP   | 42.999.1  | 123     | 10801 |  PASS  |
| SERVICE_IS_38_CHARS_STILL_CROP_IT_OKAY| 1.5       | 2       | 3     |  PASS  |
| SERVICE_IS_39_CHARS_SO_WRAP_OVERFLOW_O| 2.8       | 3       | 4     |  PASS  |
| K                                     |           |         |       |        |
| SERVICE_IS_54_CHARS_SO_DEFINITELY_WRAP| 3.1       | 4       | 5     |  PASS  |
| _THE_OVERFLOW_OK                      |           |         |       |        |
| SERVICE_IS_73_CHARS_SO_DEFINITELY_CROP| 3.2       | 5       | 6     |  PASS  |
| _THE_SECOND_LINE_SO_NO_3RD_OVERFLOW   |           |         |       |        |
| SERVICE_IS_74_CHARS_SO_DEFINITELY_WRAP| 3.3       | 6       | 7     |  PASS  |
| _THE_3RD_LINE_SO_WE_CAN_SEE_OVERFLOW  |           |         |       |        |
+---------------------------------------+-----------+---------+-------+--------+`

	printTable(statuses, 80, 39, sb)

	println(sb.String())
	actualOutput := strings.TrimSuffix(sb.String(), "\n")
	actualLines := strings.Split(actualOutput, "\n")

	expectedLines := strings.Split(expectedOutput, "\n")

	if len(expectedLines) != len(actualLines) {
		t.Errorf("Actual lines was %d, but expected lines was %d", len(actualLines), len(expectedLines))
	}

	for i, line := range actualLines {
		// ignore control chracters
		line = strings.ReplaceAll(line, "\033[32m", "")
		line = strings.ReplaceAll(line, "\033[0m", "")
		if line != expectedLines[i] {
			t.Errorf("Line %d in actualLines was: \n%s\n, but in expectedLines was \n%s\n", i, line, expectedLines[i])
		}
	}
}

func TestStatusExpandsServiceName(t *testing.T) {
	sb := bytes.NewBufferString("")
	statuses := []serviceStatus{
		serviceStatus{0, 1, "SHORT_ID", "1.2.3", "PASS"},
		serviceStatus{6, 7, "SERVICE_IS_VERY_LONG_LIKE_REALLY_REALLY_LONG_BUT_WERE_OK", "3.3", "PASS"},
	}
	expectedOutput := `+----------------------------------------------------------+-----------+---------+-------+--------+
| Name                                                     | Version   | PID     | Port  | Status |
+----------------------------------------------------------+-----------+---------+-------+--------+
| SHORT_ID                                                 | 1.2.3     | 0       | 1     |  PASS  |
| SERVICE_IS_VERY_LONG_LIKE_REALLY_REALLY_LONG_BUT_WERE_OK | 3.3       | 6       | 7     |  PASS  |
+----------------------------------------------------------+-----------+---------+-------+--------+`

	longestServiceName := getLongestString([]string{"SHORT_ID", "SERVICE_IS_VERY_LONG_LIKE_REALLY_REALLY_LONG_BUT_WERE_OK"})
	printTable(statuses, 256, longestServiceName, sb)

	actualOutput := strings.TrimSuffix(sb.String(), "\n")
	actualLines := strings.Split(actualOutput, "\n")

	expectedLines := strings.Split(expectedOutput, "\n")

	if len(expectedLines) != len(actualLines) {
		t.Errorf("Actual lines was %d, but expected lines was %d", len(actualLines), len(expectedLines))
	}

	for i, line := range actualLines {
		// ignore control chracters
		line = strings.ReplaceAll(line, "\033[32m", "")
		line = strings.ReplaceAll(line, "\033[0m", "")
		if line != expectedLines[i] {
			t.Errorf("Line %d in actualLines was: \n%s, but in expectedLines was \n%s", i, line, expectedLines[i])
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

func TestExcludeFailedStatusFromPreviousBoot(t *testing.T) {
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

	if result[0].health != BOOT {
		t.Errorf("%s health was not BOOT, it was %s", result[0].service, result[0].health)
	}
	if result[1].health != FAIL {
		t.Errorf("%s health was not FAIL, it was %s", result[1].service, result[1].health)
	}
	if result[2].health != PASS {
		t.Errorf("%s health was not PASS, it was %s", result[2].service, result[2].health)
	}
}

func TestVerifyIsRunning(t *testing.T) {
	services := []ServiceAndVersion{
		{"FOO", "1.0.0", "2.12"},
		{"BAZ", "2.0.0", "2.12"},
		{"BAR", "3.0.0", "2.12"},
	}

	statuses := []serviceStatus{
		{0, 0, "FOO", "1.0.0", PASS},
		{0, 0, "BAZ", "1.0.0", PASS},
		{0, 0, "BAR", "1.0.0", PASS},
	}

	output := bytes.NewBufferString("")

	// happy case
	if verifyIsRunning(services, statuses, output) == false {
		t.Errorf("verifyIsRunning returned false when it should be true!")
	}

	// unhappy cases
	if verifyIsRunning(services, statuses[:1], output) {
		t.Errorf("verifyIsRunning returned a false positive")
	}

	if verifyIsRunning(services, []serviceStatus{}, output) {
		t.Errorf("verifyIsRunning returned a false positive with an empty status list")
	}

	statuses[0].health = FAIL
	if verifyIsRunning(services, statuses, output) {
		t.Errorf("verifyIsRunning returned true when status was FAILED")
	}
}

func TestStatusWrapsProxyPaths(t *testing.T) {
	sb := bytes.NewBufferString("")
	state := ledger.ProxyState{
		Pid: 101,
		ProxyPaths: map[string]string{
			"/1/SHORT_PATH": "localhost:4000",
			"/2/THE_PATH____IS_60_CHARS_DO_NOT_WRAP______________________":                                                                 "localhost:4001",
			"/3/PATH____IS_61_CHARS_WRAP_OVERFLOW_OK____________________OK":                                                                "localhost:4002",
			"/4/PATH____IS_124_CHARS_SO_DEFINITELY_WRAP_THE_OVERFLOW_TO_3_LINES___________________________________________________WRAPPED": "localhost:4003",
		},
	}
	expectedOutput :=
		`+-------------------------------------------------------------+----------------+
| Proxy Path                                                  | Service Path   |
+-------------------------------------------------------------+----------------+
| /1/SHORT_PATH                                               | localhost:4000 |
| /2/THE_PATH____IS_60_CHARS_DO_NOT_WRAP______________________| localhost:4001 |
| /3/PATH____IS_61_CHARS_WRAP_OVERFLOW_OK____________________O| localhost:4002 |
|K                                                            |                |
| /4/PATH____IS_124_CHARS_SO_DEFINITELY_WRAP_THE_OVERFLOW_TO_3| localhost:4003 |
|_LINES___________________________________________________WRA |                |
|PPED                                                         |                |
+-------------------------------------------------------------+----------------+`

	printProxyTable(state, 80, sb)

	println(sb.String())
	actualOutput := strings.TrimSuffix(sb.String(), "\n")
	actualLines := strings.Split(actualOutput, "\n")

	expectedLines := strings.Split(expectedOutput, "\n")

	if len(expectedLines) != len(actualLines) {
		t.Errorf("Actual lines was %d, but expected lines was %d", len(actualLines), len(expectedLines))
	}

	for i, line := range actualLines {
		if line != expectedLines[i] {
			t.Errorf("Line %d in actualLines was: \n%s\n, but in expectedLines was \n%s\n", i, line, expectedLines[i])
		}
	}
}

func TestStatusExpandsProxyPaths(t *testing.T) {
	sb := bytes.NewBufferString("")
	state := ledger.ProxyState{
		Pid: 101,
		ProxyPaths: map[string]string{
			"/1/SHORT_PATH": "localhost:4000",
			"/2/PATH_IS_VERY_LONG_LIKE_REALLY_REALLY_LONG_BUT_WERE_OK": "localhost:4001",
		},
	}
	expectedOutput := `+----------------------------------------------------------+----------------+
| Proxy Path                                               | Service Path   |
+----------------------------------------------------------+----------------+
| /1/SHORT_PATH                                            | localhost:4000 |
| /2/PATH_IS_VERY_LONG_LIKE_REALLY_REALLY_LONG_BUT_WERE_OK | localhost:4001 |
+----------------------------------------------------------+----------------+`

	printProxyTable(state, 256, sb)

	actualOutput := strings.TrimSuffix(sb.String(), "\n")
	actualLines := strings.Split(actualOutput, "\n")

	expectedLines := strings.Split(expectedOutput, "\n")

	if len(expectedLines) != len(actualLines) {
		t.Errorf("Actual lines was %d, but expected lines was %d", len(actualLines), len(expectedLines))
	}

	for i, line := range actualLines {
		if line != expectedLines[i] {
			t.Errorf("Line %d in actualLines was: \n%s, but in expectedLines was \n%s", i, line, expectedLines[i])
		}
	}
}
