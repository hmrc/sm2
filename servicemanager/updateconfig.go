package servicemanager

import (
	"fmt"
	"path"
)

func updateConfig(configDir string) error {

	if !Exists(path.Join(configDir, ".git")) {
		return fmt.Errorf("config directory does not appear to be a .git repository")
	}

	branch, err := gitCurrentBranch(configDir)

	if err != nil {
		return err
	}

	if branch != "main" {
		return fmt.Errorf("Unable to update config!\nExpected `main` branch to be checked out, instead found `%s`.\nTo fix this, please run `git checkout -b main` in %s", branch, configDir)
	}

	return gitPull(configDir)
}
