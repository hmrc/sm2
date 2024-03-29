#!/usr/bin/env nix-shell
#!nix-shell -i bash -p bash nix curl jq
#
# Updates the nixpkgs.json to the latest channel release
#
# from https://gist.github.com/zimbatm/de5350245874361762b6a4dfe5366530
set -euo pipefail

cd "$(dirname "$0")" || exit 1

branch=nixos-21.11

owner=NixOS
repo=nixpkgs
rev=$(curl -sfL https://api.github.com/repos/$owner/$repo/git/refs/heads/$branch | jq -r .object.sha)
url=https://github.com/$owner/$repo/archive/$rev.tar.gz

release_sha256=$(nix-prefetch-url --unpack "$url")

cat <<NIXPKGS | tee nixpkgs.json
{
  "owner": "$owner",
  "repo": "$repo",
  "rev": "$rev",
  "sha256": "$release_sha256"
}
NIXPKGS
