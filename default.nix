{ buildGoModule }: {
  gofer = buildGoModule {
    pname = "gofer";
    version = "dev";
    src = ./.;
    vendorSha256 = "0m3npjwwyp789myxlzb1gd68mmzhn7fsylah2zbr1pmvc5g8vp0f";
    subPackages = [ "cmd/gofer" ];
    postConfigure = "export CGO_ENABLED=0";
    postInstall = "cp ./gofer.json $out";
  };
  spire = buildGoModule {
    pname = "spire";
    version = "dev";
    src = ./.;
    vendorSha256 = "0m3npjwwyp789myxlzb1gd68mmzhn7fsylah2zbr1pmvc5g8vp0f";
    subPackages = [ "cmd/spire" ];
    postConfigure = "export CGO_ENABLED=0";
    postInstall = "cp ./spire.json $out";
  };
  keeman = buildGoModule {
    pname = "keeman";
    version = "dev";
    src = ./.;
    vendorSha256 = "0m3npjwwyp789myxlzb1gd68mmzhn7fsylah2zbr1pmvc5g8vp0f";
    subPackages = [ "cmd/keeman" ];
    postConfigure = "export CGO_ENABLED=0";
  };
}
