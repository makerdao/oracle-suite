{ buildGoModule }: {
  gofer = buildGoModule {
    pname = "gofer";
    version = "dev";
    src = ./.;
    vendorSha256 = "07fzxllv6zbh38z28l6pi0iv1ynndrcb6lb77hv26ilprx3gzfvm";
    subPackages = [ "cmd/gofer" ];
    postConfigure = "export CGO_ENABLED=0";
    postInstall = "cp ./gofer.json $out";
  };
  spire = buildGoModule {
    pname = "spire";
    version = "dev";
    src = ./.;
    vendorSha256 = "07fzxllv6zbh38z28l6pi0iv1ynndrcb6lb77hv26ilprx3gzfvm";
    subPackages = [ "cmd/spire" ];
    postConfigure = "export CGO_ENABLED=0";
    postInstall = "cp ./spire.json $out";
  };
}
