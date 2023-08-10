{
  description = "Docker Pod v0.2.0";

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
        version = "0.2.0";
        package-overlay = pkgs: packages:
          packages // {
            default = pkgs.writeShellScriptBin "entrypoint" ''
              	    ${packages.app}/bin/babashka-pod-docker
              	  '';
          };
      };
}
