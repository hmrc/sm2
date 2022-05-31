package servicemanager

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
)

// shallow-clones a gitrepo into $repoDir/src
func gitClone(gitUrl string, repoDir string) (string, error) {
	cmd := exec.Command("git", "clone", "--depth", "1", gitUrl, "src")
	cmd.Dir = repoDir

	stdout, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to clone %s into %s.\n", gitUrl, repoDir)
		fmt.Println(string(stdout))
		return "", err
	}

	return path.Join(repoDir, "src"), nil
}

// returns the current branch name
func gitCurrentBranch(repoDir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repoDir

	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.Trim(string(stdout), "\n "), nil
}

// returns the latest commit id & message
func gitLastCommit(repoDir string) (string, error) {
	cmd := exec.Command("git", "log", "--pretty=format:%h,%ar,%s", "-1")
	cmd.Dir = repoDir

	out, err := cmd.CombinedOutput()
	return string(out), err
}

func gitPull(repoDir string) error {
	cmd := exec.Command("git", "pull", "--quiet")
	cmd.Dir = repoDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// returns the version of the git cli tool
func gitVersion() (string, error) {
	cmd := exec.Command("git", "--version")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.Trim(string(out), "\n "), nil
}
