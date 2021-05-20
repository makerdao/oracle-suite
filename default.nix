{ buildGoModule }:
buildGoModule {
  pname = "oracle-suite";
  version = "dev20210521";
  src = ./.;
  vendorSha256 = "11xjwx8gn2iwca66783fmddnjs3z5pnn0rk87lpbpas9p6pipakj";
  subPackages = [ "cmd/..." ];
  postConfigure = "export CGO_ENABLED=0";
  postInstall = "cp ./gofer.json ./spire.json $out";
}
