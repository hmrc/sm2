# Service manager V2
SM2 (service-manager-2) is a tool to manager starting and stopping groups of scala/play microservices for local development and testing.

It's based on the the original [service-manager](https://github.com/hmrc/service-manager) but re-written in golang.

## Install from Binary
1. Run the following command in your terminal for your operating system/cpu:

**Linux**
```base
curl -L -O https://github.com/hmrc/sm2/releases/download/v1.0.0/sm2-1.0.0-linux-intel.zip && unzip sm2-1.0.0-linux-intel.zip && rm sm2-1.0.0-linux-intel.zip
```

**OSX/Apple (latest M1/M2 cpus)**

```base
curl -L -O https://github.com/hmrc/sm2/releases/download/v1.0.0/sm2-1.0.0-apple-arm64.zip && unzip sm2-1.0.0-apple-arm64.zip && rm sm2-1.0.0-apple-arm64.zip
```

**OSX/Apple (older Intel cpus)**

```base
curl -L -O https://github.com/hmrc/sm2/releases/download/v1.0.0/sm2-1.0.0-apple-intel.zip && unzip sm2-1.0.0-apple-intel.zip && rm sm2-1.0.0-apple-intel.zip
```

If everything has worked you should have an executable called `sm2`. 

2. Move the executable somewhere on your systems $PATH.
3. Run `sm2` from the terminal and follow the instructions to configure your system (if its not already set up)
4. For more information, see [USERGUIDE.md](USERGUIDE.md)

Alternatively you can download and decompress sm2 yourself from the [releases]() section

## Post install/setup checkers
You can check everything is setup correctly by running:
```base
sm2 --diagnostic
```


## Building/Developing Service-Manager-2 Locally
SM2 has no external dependencies other than go 1.16+. You can build it locally via:

```bash
go build
```
When building a new release, use the included makefile:
```bash
make build_all package
```

## Running tests 

To run all tests in all sub-directories, ensure you are in the project root and run:

```bash
make test
```

To run tests for the subdirectory you are currently in:
```bash
go test
```

## Build with Nix

To install from source:

```bash
nix-env -i -f default.nix
```

## For maintainers

nixpkgs is pinned. To update - edit branch in `update-nixpkgs.sh`, then regenerate `nixpkgs.json`:

```bash
bash update-nixpkgs.sh
```

To build with shell:
```bash
nix-shell --run 'go build'
```
to create `./sm2`

or to build locally:

```bash
nix-build
```
to create `./result/bin/sm2`
