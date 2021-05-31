{ pkgs ? import <nixpkgs> { }, oracle-suite ? pkgs.callPackage ./default.nix { buildGoModule = pkgs.buildGo116Module; }
}:
pkgs.mkShell { buildInputs = [ pkgs.jq oracle-suite ]; }
