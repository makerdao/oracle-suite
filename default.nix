{ pkgs ? import <nixpkgs> { }, buildGoModule ? pkgs.buildGo116Module }:
buildGoModule {
  pname = "oracle-suite";
  version = pkgs.lib.fileContents ./version;
  src = ./.;
  vendorSha256 = "13g3s50kzaksbd64zsgxsxlyxqz5agkapp6iq0yf4darzj6fd7ny";
  subPackages = [ "cmd/..." ];
  postConfigure = "export CGO_ENABLED=0";
  postInstall = "cp ./gofer.json ./spire.json $out";
}
