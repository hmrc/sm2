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

func TestGenerateArgs(t *testing.T) {
	sm := ServiceManager{}
	sm.Commands.Port = 6666
	sm.Commands.ExtraArgs = map[string][]string{
		"FOO": {"-user1", "-user2"},
	}

	foo := Service{
		Id:          "FOO",
		DefaultPort: 9999,
		Binary: ServiceBinary{
			Cmd: []string{"./bin/foo", "-cmd1", "-cmd2"},
		},
	}

	expectedArgs := []string{
		"-cmd1",
		"-cmd2",
		"-Dservice.manager.serviceName=FOO",
		"-Dservice.manager.runFrom=1.0.1",
		"-Duser.home=/tmp/foo",
		"-user1",
		"-user2",
	}
	args := sm.generateArgs(foo, "1.0.1", "/tmp/foo/foo-1.0.1", foo.Binary.Cmd[1:])

	for i, arg := range args {
		if expectedArgs[i] != arg {
			t.Errorf("arg %s != %s", arg, expectedArgs[i])
			return
		}
	}
}

func TestFindPort(t *testing.T) {
	sm := ServiceManager{}
	foo := Service{
		Id:          "FOO",
		DefaultPort: 9999,
		Binary: ServiceBinary{
			Cmd: []string{"./bin/foo", "-cmd1", "-cmd2"},
		},
	}

	// test it uses the default port
	if p := sm.findPort(foo); p != 9999 {
		t.Errorf("port %d was not default port %d", p, 9999)
	}

	// test you can override default via --port
	sm.Commands.Port = 6666
	if p := sm.findPort(foo); p != 6666 {
		t.Errorf("port %d was not override port %d", p, 6666)
	}
}

func TestWhatVersionToRun(t *testing.T) {
	foo := Service{
		Id:          "FOO",
		DefaultPort: 9999,
		Binary: ServiceBinary{
			Artifact: "foo_2.12",
			GroupId:  "org.foo",
			Cmd:      []string{"./bin/foo", "-cmd1", "-cmd2"},
		},
	}

	latest := MavenMetadata{
		Artifact: "foo_2.12",
		Group:    "org.foo",
		Latest:   "2.0.0",
		Release:  "2.0.0",
	}
	latestFunc := func(b ServiceBinary, s string, v string) (MavenMetadata, error) {
		return latest, nil
	}
	caseServiceOnly := ServiceAndVersion{"FOO", "", ""}
	caseServiceAndVersion := ServiceAndVersion{"FOO", "1.66.0", ""}
	caseServiceAndScalaAndVersion := ServiceAndVersion{"FOO", "1.12.0", "2.11"}
	caseServiceAndScala := ServiceAndVersion{"FOO", "", "2.12"}

	group, artifact, version, err := whatVersionToRun(foo, caseServiceOnly, false, latestFunc)
	AssertNotErr(t, err)
	if group != "org.foo" || artifact != "foo_2.12" || version != latest.Latest {
		t.Errorf("expected org.foo:foo_2.12:2.0.0, got %s %s %s", group, artifact, version)
	}

	group, artifact, version, err = whatVersionToRun(foo, caseServiceAndVersion, false, latestFunc)
	AssertNotErr(t, err)
	if group != "org.foo" || artifact != "foo_2.12" || version != caseServiceAndVersion.version {
		t.Errorf("expected org.foo:foo_2.12:1.66.0, got %s %s %s", group, artifact, version)
	}

	group, artifact, version, err = whatVersionToRun(foo, caseServiceAndScalaAndVersion, false, latestFunc)
	AssertNotErr(t, err)
	if group != "org.foo" || artifact != "foo_2.11" || version != caseServiceAndScalaAndVersion.version {
		t.Errorf("expected org.foo:foo_2.11:1.12.0, got %s %s %s", group, artifact, version)
	}

	group, artifact, version, err = whatVersionToRun(foo, caseServiceAndScala, false, latestFunc)
	AssertNotErr(t, err)
	if group != "org.foo" || artifact != "foo_2.12" || version != latest.Latest {
		t.Errorf("expected org.foo:foo_2.12:2.0.0, got %s %s %s", group, artifact, version)
	}

}

func TestFindHealthcheckUrl(t *testing.T) {
	customCheck := Service{
		Id:          "FOO",
		DefaultPort: 9999,
		Healthcheck: Healthcheck{
			Url: "http://localhost:${port}/foo/ping/pong",
		},
	}

	defaultCheck := Service{
		Id: "BAR",
	}

	if url := findHealthcheckUrl(customCheck, 9999); url != "http://localhost:9999/foo/ping/pong" {
		t.Errorf("wrong custom healthcheck: %s", url)
	}

	if url := findHealthcheckUrl(defaultCheck, 8888); url != "http://localhost:8888/ping/ping" {
		t.Errorf("wrong default healthcheck, %s", url)
	}

}
