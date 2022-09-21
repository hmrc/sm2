# ########################################################## #
# Makefile for Golang Project
# Includes cross-compiling, installation, cleanup
# ########################################################## #

# Check for required command tools to build or stop immediately

ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

BINARY := sm2
VERSION := 1.0.2
BUILD := `git rev-parse HEAD`

# Setup linker flags option for build that interoperate with variable names in src code
# -s / -w omit debug symbols (for intel/arm) same as running strip cli tool
# -X writes the version/build into the binary so its available in `sm2 --version`
LDFLAGS=-ldflags "-s -w -X sm2/version.Version=$(VERSION) -X sm2/version.Build=$(BUILD)"

default: build

all: clean build_all package

build:
	go build ${LDFLAGS} -o ${BINARY}

build_all:
	@echo building all versions...
	$(shell export CGO_ENABLED=0; export GOOS=linux;  export GOARCH=amd64; go build $(LDFLAGS) -o build/$(BINARY)-$(VERSION)-linux-intel/$(BINARY))
	$(shell export CGO_ENABLED=0; export GOOS=darwin; export GOARCH=amd64; go build $(LDFLAGS) -o build/$(BINARY)-$(VERSION)-apple-intel/$(BINARY))
	$(shell export CGO_ENABLED=0; export GOOS=darwin; export GOARCH=arm64; go build $(LDFLAGS) -o build/$(BINARY)-$(VERSION)-apple-arm64/$(BINARY))

package:
	@echo compressing releases
	@find ${ROOT_DIR}/build -name '${BINARY}[-?][a-zA-Z0-9]*[-?][a-zA-Z0-9]*' -exec zip -j {}.zip {}/$(BINARY) \;

# Remove only what we've created
clean:
	rm -rf ./build

test:
	go test ./...

.PHONY: check clean install build_all all
