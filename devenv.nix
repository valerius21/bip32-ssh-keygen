{ pkgs, lib, config, inputs, ... }:

{
  # https://devenv.sh/packages/
  packages = [
    pkgs.gopls
    pkgs.golangci-lint
    pkgs.delve
  ];

  # https://devenv.sh/languages/
  languages.go.enable = true;

  # https://devenv.sh/pre-commit-hooks/
  # pre-commit.hooks.shellcheck.enable = true;

  # https://devenv.sh/processes/
  # processes.ping.exec = "ping localhost";

  # See full reference at https://devenv.sh/reference/options/
}
