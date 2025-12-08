{
  description = "Example usage of inframan for Nix-Go-GitOps workflow";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    inframan.url = "path:..";  # In real usage: github:your-org/inframan
  };

  outputs = { self, nixpkgs, inframan, ... }:
    let
      system = "x86_64-linux";
    in
    {
      # Create the runner using inframan's lib.mkRunner
      packages.${system}.default = inframan.lib.mkRunner {
        inherit system;
        infraConfig = ./infrastructure.nix;  # Terranix configuration
        machineConfig = ./machine.nix;       # NixOS module for deployment
      };

      # Convenience alias
      apps.${system}.default = {
        type = "app";
        program = "${self.packages.${system}.default}/bin/runner";
      };
    };
}

