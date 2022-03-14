package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
)

var (
	baseDir      string = "/tmp/assets"
	port         int    = 9032
	client       http.Client
	mu           sync.Mutex // guards downloading
	versionRegex = regexp.MustCompile("\\d+\\.\\d+\\.\\d+")
)

/*
 * Proof of concept of how assets frontend might work.
 * basically when the request comes in, work out what version it is,
 * download the matching zip of the assets (which is only ~1 Meg). The zip is saved
 * and the files are served direct from the zip.
 * This could run from sm2, maybe by having it launch another copy of itself with the flags etc
 * or we could run it as a normal service, in which case a jvm based solution might make
 * more sense...
 */
func main() {

	flag.IntVar(&port, "http.port", 9032, "which port to run assets frontend on")
	flag.StringVar(&baseDir, "dir", "", "where to cache the assets")

	fmt.Printf("starting on demand file server on %d serving files from %s", port, baseDir)

	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("localhost:%d", port), nil))
}

// @improvement scan maven-metadata.xml to get a list of possible assets
// and use that to validate the request and prove the assets exist?
func handler(w http.ResponseWriter, req *http.Request) {

	// assume that the first part of the path is the asset version number
	parts := strings.Split(req.URL.Path, "/")
	if len(parts) < 2 || parts[1] == "" {
		w.WriteHeader(400)
		return
	}

	// validate request
	asset := parts[1]
	if !versionRegex.MatchString(asset) {
		w.WriteHeader(400)
		return
	}
	fmt.Printf("requested: %s\n", asset)

	// Check if the assets have already been downloaded..
	// mutex here is to ensure we only download it once.
	mu.Lock()
	err := downloadIfMissing(baseDir, asset)
	mu.Unlock()

	if err != nil {
		w.WriteHeader(500)
		return
	}

	// serve the data from the zip
	zf, err := zip.OpenReader(path.Join(baseDir, asset, asset+".zip"))
	if err != nil {
		println(err.Error())
		w.WriteHeader(500)
		return
	}
	defer zf.Close()

	filename := strings.SplitN(req.URL.Path, "/", 3)[2]
	println("opening " + filename)

	file, err := zf.Open(filename)
	if err != nil {
		w.WriteHeader(404)
		return
	}
	defer file.Close()

	// serve the file
	w.Header().Set("Content-Type", mime.TypeByExtension(path.Ext(req.URL.Path)))
	io.Copy(w, file)
}

func downloadIfMissing(baseDir, asset string) error {
	if _, err := os.Stat(path.Join(baseDir, asset)); os.IsNotExist(err) {

		log.Printf("attempting to download %s...\n", asset)
		// otherwise download it

		dlResp, err := client.Get(fmt.Sprintf("https://artefacts.tax.service.gov.uk/artifactory/hmrc-releases/uk/gov/hmrc/assets-frontend/%s/assets-frontend-%s.zip", asset, asset))

		if err != nil {
			log.Printf("failed to get assets: %s from artifactory", asset)
			return err
		}
		defer dlResp.Body.Close()

		// stub dir on 404's to prevent repeated requests
		if dlResp.StatusCode == 404 {
			log.Printf("assets: %s not found, will not try again!\n", asset)
			os.Mkdir(path.Join(baseDir, asset), 0755)
		}

		// abort on other errors
		if dlResp.StatusCode != 200 {
			log.Printf("failed to download assets: %s, %s\n", asset, dlResp.Status)
			return fmt.Errorf("failed to download")
		}

		// unpack the assets
		os.MkdirAll(path.Join(baseDir, asset), 0755)
		outFile, err := os.Create(path.Join(baseDir, asset, asset+".zip"))
		if err != nil {
			println(err.Error())
			return err
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, dlResp.Body)
		if err != nil {
			println(err.Error())
			return err
		}
	}
	return nil
}
