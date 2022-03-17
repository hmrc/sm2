# ########################################################## #
# Makefile for Golang Project
# Includes cross-compiling, installation, cleanup
# ########################################################## #

# Check for required command tools to build or stop immediately

ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

BINARY := sm2
VERSION := 0.4.0
BUILD := `git rev-parse HEAD`

# Setup linker flags option for build that interoperate with variable names in src code
LDFLAGS=-ldflags "-s -w -X sm2/version.Version=$(VERSION) -X sm2/version.Build=$(BUILD)"

default: build

all: clean build_all package

build:
	go build ${LDFLAGS} -o ${BINARY}

build_all:
	@echo building all versions...
	$(shell export GOOS=linux;  export GOARCH=amd64; go build $(LDFLAGS) -o build/$(BINARY)-$(VERSION)-linux-intel)
	$(shell export GOOS=darwin; export GOARCH=amd64; go build $(LDFLAGS) -o build/$(BINARY)-$(VERSION)-apple-intel)
	$(shell export GOOS=darwin; export GOARCH=arm64; go build $(LDFLAGS) -o build/$(BINARY)-$(VERSION)-apple-arm64)

package:
	tar czf build/$(BINARY)-$(VERSION).tgz build/*linux* build/*apple*

# Remove only what we've created
clean:
	find ${ROOT_DIR} -name '${BINARY}[-?][a-zA-Z0-9]*[-?][a-zA-Z0-9]*' -delete

.PHONY: check clean install build_all all
