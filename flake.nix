{
  description = "Docker Pod";

  inputs = {
    platform-engineering.url = "github:slimslenderslacks/nix-modules";
    nixpkgs.url = "github:NixOS/nixpkgs/release-22.11";
  };

  outputs = { nixpkgs, ... }@inputs:
    inputs.platform-engineering.golang-project
      {
        inherit nixpkgs;
        dir = ./.;
        name = "babashka-pod-docker";
        version = "0.0.1";
        custom-packages = pkgs: packages:
          packages // {
            default = pkgs.writeShellScriptBin "entrypoint" ''
              	    ${packages.app}/bin/babashka-pod-docker
              	  '';
          };
      };
}
