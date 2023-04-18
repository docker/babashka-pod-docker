{
  description = "Docker Pod";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/release-22.11";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:

    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [ go gotools golangci-lint gopls gopkgs go-outline ];
        };
        packages = {
          default = pkgs.buildGoModule {
            pname = "babashka-pod-docker";
            version = "0.0.1";
            src = ./.;
            vendorSha256 = "sha256-KUWqddPcv+hLStd7JEzQBUiGLPLYwfmyVoG1BtaHWXY=";
            postInstall = ''
              	        mv $out/bin/parser $out/bin/babashka-pod-docker
              	        '';

          };
        };
      });
}
