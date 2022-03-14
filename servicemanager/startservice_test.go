package servicemanager

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	. "sm2/testing"
)

/*
func TestInstallService(t *testing.T) {

	installDir, err := ioutil.TempDir(os.TempDir(), "test-install*")
	AssertNotErr(t, err)

	service := Service{
		Id:          "FOO",
		DefaultPort: 9999,
		Binary: ServiceBinary{
			Artifact:          "foo_2.12",
			GroupId:           "foo.bar",
			DestinationSubdir: "foo",
			Cmd:               []string{"bin/foo"},
		},
	}

	version := "1.0.1"

	sm := ServiceManager{}

	sm.installService(installDir, service, version)
	// Work in progrss...
}
*/

func TestRemoveExistingVersion(t *testing.T) {

	baseDir, err := ioutil.TempDir(os.TempDir(), "test-removeExisting*")
	AssertNotErr(t, err)
	installDir := path.Join(baseDir, "foo")
	serviceDir := path.Join(installDir, "foo-1.0.1")
	AssertNotErr(t, os.MkdirAll(serviceDir, 0755))

	defer os.RemoveAll(baseDir)

	AssertDirExists(t, serviceDir)

	err = removeExistingVersions(installDir)
	AssertNotErr(t, err)

	AssertDirNotExists(t, serviceDir)
	AssertDirExists(t, installDir)
}

func TestRemoveRunningPid(t *testing.T) {
	baseDir, err := ioutil.TempDir(os.TempDir(), "test-removeRunningPid*")
	AssertNotErr(t, err)
	pidPath := path.Join(baseDir, "RUNNING_PID")
	AssertNotErr(t, os.WriteFile(pidPath, []byte{}, 0755))

	defer os.RemoveAll(baseDir)

	removeRunningPid(baseDir)

	AssertFileNotExists(t, pidPath)
	AssertDirExists(t, baseDir)
}
