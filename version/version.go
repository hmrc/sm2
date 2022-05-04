package version

import "fmt"

// set during build using
// go build -ldflags="-X 'sm2/version.Version=1.0.0'"
var (
	Version string = "non-release-version"
	Build   string = "non-release-build"
)

func PrintVersion() {
	fmt.Printf("version: %s\n  build: %s\n", Version, Build)
}
