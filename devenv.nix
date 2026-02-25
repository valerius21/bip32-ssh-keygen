{ pkgs, lib, config, inputs, ... }:
{
  packages = [
    pkgs.gopls
    pkgs.golangci-lint
    pkgs.delve
    pkgs.act
    pkgs.goreleaser
  ];

  languages.go.enable = true;
}