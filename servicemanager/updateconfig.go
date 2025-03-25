package servicemanager

import (
	"fmt"
	"path"
)

/*
Attempts to git pull the service-manager-config in $WORKSPACE

	Fails if not on main, bound to the --update-config cmd
*/
func updateConfig(configRepo string) error {

	// check we're looking at a real .git directory
	if !Exists(path.Join(configRepo, ".git")) {
		return fmt.Errorf("%s does not appear to be a .git repository", configRepo)
	}

	// get the branch and ensure we're on main
	branch, err := gitCurrentBranch(configRepo)
	if err != nil {
		return err
	}

	// fail if we're not on main. they might have uncommited changes etc so leave it up to the user to figure out
	if branch != "main" {
		return fmt.Errorf("Unable to update config!\nExpected main branch to be checked out, instead found `%s`.\nTo fix this, please run git checkout -b main in %s", branch, configRepo)
	}

	// pull the changes
	fmt.Printf("Config repo located at: %s\n", configRepo)
	fmt.Print("Pulling down the latest config... ")
	err = gitPull(configRepo)
	if err != nil {
		fmt.Printf("Update failed!\n")
		return err
	}

	// show the latest version/commit
	latestCommit, err := gitLastCommit(configRepo)
	if err != nil {
		return err
	}
	fmt.Printf("Done!\n\nLatest commit [%s]\n\n", latestCommit)
	return nil
}
