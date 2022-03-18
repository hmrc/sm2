package servicemanager

import (
	"archive/tar"
	"compress/gzip"
	"crypto/md5"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
)

type MavenMetadata struct {
	Latest  string `xml:"versioning>latest"`
	Release string `xml:"versioning>release"`
}

func ParseMetadataXml(r io.Reader) (MavenMetadata, error) {
	metadata := MavenMetadata{}
	decoder := xml.NewDecoder(r)
	err := decoder.Decode(&metadata)
	return metadata, err
}

// Connects to artifactory and parses maven metadata to get the latest release
func (sm ServiceManager) GetLatestVersions(s ServiceBinary) (MavenMetadata, error) {

	// build url
	url := sm.Config.ArtifactoryRepoUrl + path.Join("/", s.GroupId, s.Artifact, "maven-metadata.xml")

	// download metadata
	resp, err := sm.Client.Get(url)

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
func (sm ServiceManager) downloadAndDecompress(url string, outdir string, progressTracker *ProgressTracker) (string, error) {

	// ensure base dir and logs dir exist
	if err := os.MkdirAll(outdir, 0755); err != nil {
		return "", err
	}

	resp, err := sm.Client.Get(url)
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

	progressTracker.contentLength = int(resp.ContentLength)
	tee := io.TeeReader(resp.Body, progressTracker) // split off to progress tracker
	body := io.TeeReader(tee, md5Hasher)            // split off to calculate the checksum

	gz, err := gzip.NewReader(body)
	if err != nil {
		return "", err
	}
	defer gz.Close()

	// used to determin the serviceDir
	dirsSeen := map[string]uint8{}

	tarReader := tar.NewReader(gz)
	for true {
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
			defer outfile.Close()

			_, err = io.Copy(outfile, tarReader)
			if err != nil {
				log.Fatal(err)
			}
			// fix up the permissions
			outfile.Chmod(header.FileInfo().Mode())
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
