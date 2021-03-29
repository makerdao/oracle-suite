{ buildGoModule
}: buildGoModule rec {
  pname = "oracle-suite";
  version = "dev";
  src = ./.;
  vendorSha256 = "07fzxllv6zbh38z28l6pi0iv1ynndrcb6lb77hv26ilprx3gzfvm";
  subPackages = [ "cmd/gofer" ];
}
