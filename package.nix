{ buildGoModule
, lib
}:

buildGoModule {
  name = "redial_proxy";
  version = "0.0.1";
  vendorHash = "sha256-PR3zEwn010xcxyqGVYR1H904jMdapTRdatmeXfefbFs=";
  src = ./.;
  meta = with lib; {
    description = "SOCKS5 compatible proxy that retries 3 times all requests that gives any error that have the word route. It is a workaround to a internet problem I am having.";
    homepage = "https://github.com/lucasew/redial_proxy";
    platforms = platforms.linux;
  };
}
