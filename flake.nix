{
  description = "Docker Pod";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/release-22.11";
    flake-utils.url = "github:numtide/flake-utils";
    gomod2nix.url = "github:nix-community/gomod2nix";
    devshell = {
      url = "github:numtide/devshell";
      inputs.flake-utils.follows = "flake-utils";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, flake-utils, gomod2nix, devshell }:

    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs
          {
            inherit system;
            overlays = [ gomod2nix.overlays.default devshell.overlays.default ];
          };
      in
      {
        devShells.default = pkgs.devshell.mkShell {
          packages = with pkgs; [ go gotools golangci-lint gopls gopkgs go-outline gomod2nix.packages.${system}.default 
	                          (clojure.override { jdk = temurin-bin; })
				  clojure-lsp temurin-bin neovim];
          commands = [
            {
              name = "update-gomod2nix";
              help = "update gomod2nix.toml";
              command = "gomod2nix";
            }
          ];
        };
        packages = rec {
          default = pkgs.buildGoApplication {
            pname = "babashka-pod-docker";
            version = "0.0.1";
            src = ./.;
            pwd = ./.;
            CGO_ENABLED = 0;
            modules = ./gomod2nix.toml;
          };

          docker = pkgs.dockerTools.buildImage {
            name = "docker-pod";
            tag = "latest";
            config = {
              Cmd = [ "${default}/bin/babashka-pod-docker" ];
            };
          };

          default-linux = default.overrideAttrs (old: old // { GOOS = "linux"; GOARCH = "arm64"; });
          docker-arm64 = pkgs.dockerTools.buildImage {
            name = "docker-pod";
            tag = "latest";
            config = {
              Cmd = [ "${default-linux}/bin/linux_arm64/babashka-pod-docker" ];
            };
          };
        };
      });
}
