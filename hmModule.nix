{lib, config, pkgs, ...}:
let
  binary = "${(pkgs.callPackage ./package.nix {})}/bin/redial_proxy";
in
with lib;
with lib.types;
{
  options.services.redial_proxy = {
    enable = mkEnableOption "Enable redial proxy";
    port = mkOption {
      type = int;
      description = "Port where the proxy will listen";
      default = 8889;
    };
  };
  config = mkIf config.services.redial_proxy.enable {
    systemd.user.services.redial_proxy = {
      Unit = {
        Description = "Redial proxy";
      };
      Service = {
        Type = "exec";
        ExecStart = "${binary} -p ${builtins.toString config.services.redial_proxy.port}";
        Restart = "on-failure";
      };
      Install = {
        WantedBy = [
          "default.target"
        ];
      };
    };
  };
}
