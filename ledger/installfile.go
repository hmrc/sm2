package ledger

import (
	"encoding/json"
	"os"
	"path"
	"time"
)

const installFileName = ".install"

type InstallFile struct {
	Service  string
	Artifact string
	Version  string
	Path     string
	Md5Sum   string
	Created  time.Time
}

func saveInstallFile(installDir string, install InstallFile) error {
	file, err := os.Create(path.Join(installDir, installFileName))
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	return encoder.Encode(install)
}

func loadInstallFile(installDir string) (InstallFile, error) {
	install := InstallFile{}
	file, err := os.OpenFile(path.Join(installDir, installFileName), os.O_RDONLY, 0755)
	if err != nil {
		return install, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&install)
	return install, err
}
