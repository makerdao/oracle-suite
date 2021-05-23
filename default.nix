{ pkgs ? import <nixpkgs> { }, buildGoModule ? pkgs.buildGo116Module }:
let
  rev = pkgs.stdenv.mkDerivation {
    name = "rev";
    buildInputs = [ pkgs.git ];
    src = ./.;
    buildPhase = "true";
    installPhase = ''
      echo "$(git rev-parse --short HEAD 2>/dev/null || find . -type f -print0 | sort -z | xargs -0 sha1sum | sha1sum | sed -r 's/\s+//g')" > $out
    '';
  };
  ver = "${pkgs.lib.removeSuffix "\n" (builtins.readFile "${rev}")}";
in buildGoModule {
  pname = "oracle-suite";
  version = "dev_${ver}";
  src = ./.;
  vendorSha256 = "11xjwx8gn2iwca66783fmddnjs3z5pnn0rk87lpbpas9p6pipakj";
  subPackages = [ "cmd/..." ];
  postConfigure = "export CGO_ENABLED=0";
  postInstall = "cp ./gofer.json ./spire.json $out";
}
