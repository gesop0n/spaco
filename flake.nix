{
  description = "spaco dev environment";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    { nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          inherit system;
          config.allowUnfreePredicate = pkg: builtins.elem (nixpkgs.lib.getName pkg) [ "terraform" ];
        };

      in
      {
        formatter = pkgs.nixfmt-tree;

        devShells.default = pkgs.mkShell {
          packages = [
            # Frontend
            pkgs.nodejs_22
            pkgs.pnpm_11
            pkgs.google-cloud-sdk
            pkgs.terraform
            pkgs.actionlint

            # Backend
            pkgs.go_1_25
            pkgs.gopls
            pkgs.golangci-lint
            pkgs.sqlc
            pkgs.goose
            pkgs.postgresql

            # Nix
            pkgs.nixfmt-tree

            # ConnectRPC
            pkgs.buf
          ];
        };
      }
    );
}
