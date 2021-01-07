{pkgs ? import <nixpkgs> {}, port ? 8889}:
let
  redial_proxy = pkgs.callPackage ./package.nix {};
in
pkgs.writeShellScriptBin "redial_proxy" ''
${redial_proxy}/bin/redial_proxy -p ${builtins.toString port} $*
''
