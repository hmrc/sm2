let
  bootstrap = import <nixpkgs> {};
  nixpkgs   = bootstrap.fetchFromGitHub (bootstrap.lib.importJSON ./nixpkgs.json);
  pkgs      = import nixpkgs { };
in
  pkgs.stdenv.mkDerivation {
    name = "sm2";

    src = ./.;

    buildInputs = [ pkgs.go ];

    buildPhase = ''
      ${pkgs.go}/bin/go build
    '';


    installPhase = ''
      mkdir -p $out/bin
      cp -p sm2 $out/bin/sm2
    '';

    fixupPhase = ''
      chmod +x $out/bin/sm2
    '';
  }
