# Service manager V2
Work-in-progress rewrite of the original service manager but in go.

## Install from Binary

1. Download the latest binary for your os/arch for the github release page:
https://github.com/hmrc/sm2/releases
2. Unpack the sm2 binary and put it somewhere in your systems $PATH
3. Run `sm2` and follow the instructions to configure your system.
4. For more information, see [USERGUIDE.md](USERGUIDE.md)

## Build locally
SM2 has no external dependencies other than go 1.16+. You can build it locally via:

```bash
go build
```
When building a new release, use the included makefile:
```bash
make build_all package
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
