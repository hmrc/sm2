package servicemanager

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	. "sm2/testing"
)

const mavenMetadata string = `
<metadata>
   <groupId>foo.bar</groupId>
   <artifactId>foo_2.12</artifactId>
   <versioning>
     <latest>2.33.0</latest>
     <release>2.32.0</release>
     <versions>
       <version>2.0.22</version>
       <version>2.0.23</version>
       <version>2.0.24</version>
       <version>2.0.25</version>
       <version>2.0.26</version>
       <version>2.0.27</version>
       <version>2.0.28</version>
       <version>2.32.0</version>
       <version>2.33.0</version>
     </versions>
     <lastUpdated>20160131090159</lastUpdated>
   </versioning>
</metadata>
`

const mavenMetadata213 string = `
<metadata>
   <groupId>foo.bar</groupId>
   <artifactId>foo_2.13</artifactId>
   <versioning>
     <latest>3.44.0</latest>
     <release>3.44.0</release>
     <versions>
       <version>3.44.0</version>
     </versions>
     <lastUpdated>20160131070159</lastUpdated>
   </versioning>
</metadata>
`

func TestGetLatestVersionGetsScala213IfPresent(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/foo/bar/foo_2.13/maven-metadata.xml" {
			fmt.Fprint(w, mavenMetadata213)
		} else if r.URL.Path == "/foo/bar/foo_2.12/maven-metadata.xml" {
			fmt.Fprint(w, mavenMetadata)
		} else {
			w.WriteHeader(404)
		}
	}))
	defer svr.Close()

	sm := ServiceManager{
		Client: &http.Client{},
		Config: ServiceManagerConfig{
			ArtifactoryRepoUrl: svr.URL,
		},
	}

	sb := ServiceBinary{
		GroupId:  "foo/bar/",
		Artifact: "foo_2.12",
	}
	meta, err := sm.GetLatestVersions(sb)

	AssertNotErr(t, err)

	if meta.Latest != "3.44.0" {
		t.Errorf("latest version was not 3.44.0, it was %s", meta.Latest)
	}

	if meta.Release != "3.44.0" {
		t.Errorf("release version was not 3.44.0, it was %s", meta.Latest)
	}
}

func TestGetLatestVersionGetsScala213IfMissing(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/foo/bar/foo_2.12/maven-metadata.xml" {
			fmt.Fprint(w, mavenMetadata)
		} else {
			w.WriteHeader(404)
		}
	}))
	defer svr.Close()

	sm := ServiceManager{
		Client: &http.Client{},
		Config: ServiceManagerConfig{
			ArtifactoryRepoUrl: svr.URL,
		},
	}

	sb := ServiceBinary{
		GroupId:  "foo/bar/",
		Artifact: "foo_2.12",
	}
	meta, err := sm.GetLatestVersions(sb)

	AssertNotErr(t, err)

	if meta.Latest != "2.33.0" {
		t.Errorf("latest version was not 2.32.0, it was %s", meta.Latest)
	}

	if meta.Release != "2.32.0" {
		t.Errorf("release version was not 2.32.0, it was %s", meta.Latest)
	}
}

func TestGetLatestVersionGetsJavaService(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/foo/bar/foo/maven-metadata.xml" {
			fmt.Fprint(w, mavenMetadata)
		} else {
			w.WriteHeader(404)
		}
	}))
	defer svr.Close()

	sm := ServiceManager{
		Client: &http.Client{},
		Config: ServiceManagerConfig{
			ArtifactoryRepoUrl: svr.URL,
		},
	}

	sb := ServiceBinary{
		GroupId:  "foo/bar/",
		Artifact: "foo",
	}
	meta, err := sm.GetLatestVersions(sb)

	AssertNotErr(t, err)

	if meta.Latest != "2.33.0" {
		t.Errorf("latest version was not 2.32.0, it was %s", meta.Latest)
	}

	if meta.Release != "2.32.0" {
		t.Errorf("release version was not 2.32.0, it was %s", meta.Latest)
	}
}

func TestGetLatestVersion(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, mavenMetadata)
	}))
	defer svr.Close()

	sm := ServiceManager{
		Client: &http.Client{},
		Config: ServiceManagerConfig{
			ArtifactoryRepoUrl: svr.URL,
		},
	}

	sb := ServiceBinary{
		GroupId:  "foo/bar/",
		Artifact: "foo_2.12",
	}
	meta, err := sm.GetLatestVersions(sb)

	AssertNotErr(t, err)

	if meta.Latest != "2.33.0" {
		t.Errorf("latest version was not 2.32.0, it was %s", meta.Latest)
	}

	if meta.Release != "2.32.0" {
		t.Errorf("release version was not 2.32.0, it was %s", meta.Latest)
	}
}

func TestParseValidMetadata(t *testing.T) {
	metadata, err := ParseMetadataXml(strings.NewReader(mavenMetadata))

	AssertNotErr(t, err)

	if metadata.Latest != "2.33.0" || metadata.Release != "2.32.0" {
		t.Errorf("latest [%s] and/or release [%s] data is invalid", metadata.Latest, metadata.Release)
	}

	if metadata.Artifact != "foo_2.12" {
		t.Errorf("metadata artifact was not foo_2.12 it was %s", metadata.Artifact)
	}

	if metadata.Group != "foo.bar" {
		t.Errorf("metadata group was not foo.bar it was %s", metadata.Group)
	}
}

func TestDownloadAndDecompress(t *testing.T) {

	outdir, err := ioutil.TempDir(os.TempDir(), "test-downloader*")
	AssertNotErr(t, err)
	defer os.RemoveAll(outdir)

	// discard progres
	discarder := make(chan Progress)
	go func(c chan Progress) {
		for range c {
		}
	}(discarder)
	defer close(discarder)

	progress := ProgressTracker{update: discarder}

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("../testing/testdata/playtest-1.0.0.tgz")
		if err != nil {
			fmt.Printf("test data missing: %s", err)
			w.WriteHeader(404)
			return
		}
		defer f.Close()
		io.Copy(w, f)
	}))
	defer svr.Close()

	sm := ServiceManager{
		Client: &http.Client{},
		Config: ServiceManagerConfig{
			ArtifactoryRepoUrl: svr.URL,
		},
	}

	// download the mock tgz
	serviceDir, err := sm.downloadAndDecompress(svr.URL, outdir, &progress)

	AssertNotErr(t, err)

	if serviceDir != path.Join(outdir, "playtest-1.0.0") {
		t.Errorf("service dir was not what we expected: %s", serviceDir)
	}

	AssertDirExists(t, path.Join(serviceDir, "bin"))
	AssertFileExists(t, path.Join(serviceDir, "bin", "playtest"))

	AssertDirExists(t, path.Join(serviceDir, "lib"))
	AssertFileExists(t, path.Join(serviceDir, "lib", "foo.jar"))

	if progress.totalRead == 0 {
		t.Errorf("progress tracker read 0 bytes, expected > 0")
	}
}
