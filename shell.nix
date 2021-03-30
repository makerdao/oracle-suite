{ pkgs ? import <nixpkgs> {}
, oracle-suite ? pkgs.callPackage  ./default.nix {}
}: pkgs.mkShell {
  buildInputs = [ oracle-suite ];
}
