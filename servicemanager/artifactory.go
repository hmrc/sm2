package servicemanager

import (
	"archive/tar"
	"compress/gzip"
	"crypto/md5"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"sm2/version"
)

type MavenMetadata struct {
	Artifact string   `xml:"artifactId"`
	Group    string   `xml:"groupId"`
	Latest   string   `xml:"versioning>latest"`
	Release  string   `xml:"versioning>release"`
	Versions []string `xml:"versioning>versions>version"`
}

// checks if a given version is in the MavenMetadata versions.
func (m MavenMetadata) ContainsVersion(version string) bool {
	for _, v := range m.Versions {
		if v == version {
			return true
		}
	}
	return false
}

var scalaSuffix *regexp.Regexp = regexp.MustCompile(`_(2\.\d{2}|3|%%)$`)

var latestVersionScalaVersionSuffix *regexp.Regexp = regexp.MustCompile(`_%%$`)

var userAgent = fmt.Sprintf("sm2/%s (%s %s)", version.Version, runtime.GOOS, runtime.GOARCH)

func ParseMetadataXml(r io.Reader) (MavenMetadata, error) {
	metadata := MavenMetadata{}
	decoder := xml.NewDecoder(r)
	err := decoder.Decode(&metadata)
	return metadata, err
}

func convertVersionToComparableInt(version string) (int, error) {
	parts := strings.Split(version, ".")

	if len(parts) != 3 {
		return 0, fmt.Errorf("Invalid version format: %s", version)
	}

	part1, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}

	part2, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}

	part3, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, err
	}

	comparableVersion := (part1 * 1000000) + (part2 * 1000) + part3
	return comparableVersion, nil
}

const (
	ScalaVersion_3    = "_3"
	ScalaVersion_2_13 = "_2.13"
	ScalaVersion_2_12 = "_2.12"
	ScalaVersion_2_11 = "_2.11"
	ScalaVersion_Any  = "_%%"
)

func (sm *ServiceManager) GetLatestVersions(s ServiceBinary, suppliedScalaVersion string, suppliedServiceVersion string) (MavenMetadata, error) {
	scalaVersions := []string{ScalaVersion_3, ScalaVersion_2_13, ScalaVersion_2_12, ScalaVersion_2_11}

	// honours supplied Scala version
	if suppliedScalaVersion != "" {
		artifact := scalaSuffix.ReplaceAllLiteralString(s.Artifact, "_"+suppliedScalaVersion)
		metadata, err := sm.getLatestVersion(s.GroupId, artifact)

		return metadata, err
	}

	// looks for _%% to always use latest version
	if latestVersionScalaVersionSuffix.MatchString(s.Artifact) {
		var result MavenMetadata
		var latestVersion int

		for _, v := range scalaVersions {
			// tries all Scala versions to find which artifact contains the latest version
			artifact := strings.Replace(s.Artifact, ScalaVersion_Any, v, 1)
			metadata, err := sm.getLatestVersion(s.GroupId, artifact)

			if err != nil {
				continue
			}

			if suppliedServiceVersion != "" && metadata.ContainsVersion(suppliedServiceVersion) {
				return metadata, nil
			}

			comparableVersion, err := convertVersionToComparableInt(metadata.Latest)

			if err != nil {
				return MavenMetadata{}, fmt.Errorf("invalid latest version number: %s", metadata.Latest)
			}

			if comparableVersion > latestVersion {
				latestVersion = comparableVersion
				result = metadata
			}
		}

		return result, nil
	}

	// uses Scala version set in config and for non Scala services
	metadata, err := sm.getLatestVersion(s.GroupId, s.Artifact)
	if err != nil {
		return metadata, fmt.Errorf("failed to find maven-metadata.xml for %s", s.Artifact)
	}

	return metadata, err
}

// Connects to artifactory and parses maven metadata to get the latest release
func (sm *ServiceManager) getLatestVersion(group string, artifact string) (MavenMetadata, error) {

	// build url
	url := sm.Config.ArtifactoryRepoUrl + path.Join("/", group, artifact, "maven-metadata.xml")

	// download metadata
	ctx := sm.NewShortContext()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return MavenMetadata{}, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := sm.Client.Do(req)
	if err != nil {
		return MavenMetadata{}, err
	}

	defer resp.Body.Close()

	// parse metadata
	if resp.StatusCode != 200 {
		return MavenMetadata{}, fmt.Errorf("failed to find maven-metadata.xml at %s", url)
	}
	return ParseMetadataXml(resp.Body)
}

// downloads a url and attempt to decompress it to a folder
// assumes the target is a .tgz file
// this could return the install(service) dir, would remove need to look it up later
func (sm *ServiceManager) downloadAndDecompress(url string, outdir string, progressWriter *ProgressWriter) (string, error) {

	// ensure base dir and logs dir exist
	if err := os.MkdirAll(outdir, 0755); err != nil {
		return "", err
	}

	// use default timeout. limiting by ctx works if its < client's timeout but not longer...
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	// overrider header so we can track usage in artifactory
	req.Header.Set("User-Agent", userAgent)

	resp, err := sm.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	//TODO: follow redirect, more status codes etc
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("http GET %s failed with status %s, expected 200", url, resp.Status)
	}

	md5Hasher := md5.New()
	expectedHash, hasMd5 := resp.Header["X-Checksum-Md5"]

	progressWriter.contentLength = int(resp.ContentLength)
	tee := io.TeeReader(resp.Body, progressWriter) // split off to progress tracker
	body := io.TeeReader(tee, md5Hasher)           // split off to calculate the checksum

	gz, err := gzip.NewReader(body)
	if err != nil {
		return "", err
	}
	defer gz.Close()

	// used to determin the serviceDir
	dirsSeen := map[string]uint8{}

	tarReader := tar.NewReader(gz)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		switch header.Typeflag {

		case tar.TypeDir:
			// TODO: track dirs created so we can determin where exactly the app is
			if err := os.MkdirAll(path.Join(outdir, header.Name), 0755); err != nil {
				log.Fatalf("ExtractTarGz: Mkdir() failed: %s", err.Error())
			}

		case tar.TypeReg:
			// create folder if required
			dir, _ := path.Split(header.Name)
			if err := os.MkdirAll(path.Join(outdir, dir), 0755); err != nil {
				log.Fatalf("Failed to created dir %s: %s", dir, err)
			}

			rootDir := strings.SplitN(path.Clean(dir), "/", 2)[0]
			dirsSeen[rootDir] = 1

			// write the file
			outfile, err := os.Create(path.Join(outdir, header.Name))
			if err != nil {
				log.Fatalf("\nfailed to write to file %s\n%s", path.Join(outdir, header.Name), err)
			}

			_, err = io.Copy(outfile, tarReader)
			if err != nil {
				log.Fatal(err)
			}
			// fix up the permissions
			outfile.Chmod(header.FileInfo().Mode())
			outfile.Close()
		}
	}

	// check checksum and fail if it doesnt match
	if hasMd5 {
		actualHash := fmt.Sprintf("%x", md5Hasher.Sum(nil))
		if actualHash != expectedHash[0] {
			return "", fmt.Errorf("md5 did not match, %s != %s", actualHash, expectedHash[0])
		}
		// todo: do we need to return the hash? once validated its not much use tbh!
	}

	// based on the directories we've had to make, figure out which one the service is in
	// we're assuming theres only one, this could be better
	var serviceDir string

	delete(dirsSeen, ".")
	for k := range dirsSeen {
		// TODO: regex it or something? maybe inc the count every times its seen and go with the largest?
		//       if we know what the bin dir is (from services.json) we could use that too
		serviceDir = path.Join(outdir, k)
	}

	return serviceDir, nil
}
