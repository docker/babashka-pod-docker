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
      };
}
