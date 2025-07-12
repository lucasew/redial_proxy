{ buildGoModule
, lib
}:

buildGoModule {
  name = "redial_proxy";
  version = "0.0.1";
  vendorHash = "sha256-L9RFOQD+iMK1x4Nt3vzZgl6AOc/YN3oszJdi6R56yLk=";
  src = ./.;
  meta = with lib; {
    description = "SOCKS5 compatible proxy that retries 3 times all requests that gives any error that have the word route. It is a workaround to a internet problem I am having.";
    homepage = "https://github.com/lucasew/redial_proxy";
    platforms = platforms.linux;
  };
}
