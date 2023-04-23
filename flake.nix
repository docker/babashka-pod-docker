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
        packages = rec {
          default = pkgs.buildGoModule {
            pname = "babashka-pod-docker";
            version = "0.0.1";
            src = ./.;
	    # uncomment this and re-run to find the new vendor sha when module deps change
	    #   note that you'll get inconsistent vendor deps if this gets out of sync because
	    #   this sha defines the input for the mod deps derivation
            # vendorSha256 = nixpkgs.lib.fakeSha256;
	    vendorSha256 = "sha256-jCjNhi0eqEBNPts/xmbwugs0T6HUw1ESip/li4/J6YY=";
	    CGO_ENABLED = 0;
          };

  	  docker = pkgs.dockerTools.buildImage {
	    name = "docker-pod";
	    tag = "latest";
	    config = {
	      Cmd = ["${default}/bin/babashka-pod-docker"];
	    };
	  };

	  default-linux = default.overrideAttrs (old: old // {GOOS = "linux"; GOARCH = "arm64";});
	  docker-arm64 = pkgs.dockerTools.buildImage {
	    name = "docker-pod";
	    tag = "latest";
	    config = {
	      Cmd = ["${default-linux}/bin/linux_arm64/babashka-pod-docker"];
	    };
	  };
        };
      });
}
