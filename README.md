# Service manager V2
Work-in-progress rewrite of the original service manager but in go.

## Install with Nix
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
