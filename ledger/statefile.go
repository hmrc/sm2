package ledger

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"time"
)

const stateFileName = ".state"
const proxyStateFileName = ".proxy_state"

type StateFile struct {
	Service        string
	Artifact       string
	Version        string
	Path           string
	Started        time.Time
	Pid            int
	Port           int
	Args           []string
	HealthcheckUrl string
}

type ProxyStateFile struct {
	Started    time.Time
	Pid        int
	ProxyPaths map[string]string
}

func saveStateFile(installDir string, ledger StateFile) error {
	file, err := os.Create(path.Join(installDir, stateFileName))
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(ledger)
}

func loadStateFile(installDir string) (StateFile, error) {
	state := StateFile{}

	file, err := os.OpenFile(path.Join(installDir, stateFileName), os.O_RDONLY, 0755)
	if err != nil {
		return state, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&state)
	return state, err
}

func clearStateFile(installDir string) error {
	return clearFile(installDir, stateFileName)
}

func findAll(baseDir string) ([]StateFile, error) {
	files, err := ioutil.ReadDir(baseDir)

	if err != nil {
		return nil, err
	}

	matches := []StateFile{}
	for _, file := range files {
		if file.IsDir() {
			if state, err := loadStateFile(path.Join(baseDir, file.Name())); err == nil {
				matches = append(matches, state)
			}
		}
	}
	return matches, nil
}

func saveProxyStateFile(installDir string, ledger ProxyStateFile) error {
	file, err := os.Create(path.Join(installDir, proxyStateFileName))
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(ledger)
}

func loadProxyStateFile(installDir string) ProxyStateFile {
	state := ProxyStateFile{}

	file, err := os.OpenFile(path.Join(installDir, proxyStateFileName), os.O_RDONLY, 0755)
	if err != nil {
		return state
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&state)
	return state
}

func clearProxyStateFile(installDir string) error {
	return clearFile(installDir, proxyStateFileName)
}

func clearFile(installDir string, fileName string) error {
	err := os.Remove(path.Join(installDir, fileName))
	if err != nil {
		if os.IsNotExist(err) {
			// don't care if it doesn't exist
			return nil
		}
		return err
	}
	return nil
}
