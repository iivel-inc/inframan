{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";
    terranix.url = "github:terranix/terranix";
    colmena.url = "github:zhaofengli/colmena";
  };
  outputs = {self, nixpkgs, terranix, colmena, ...}@inputs:
    let
      lib = inputs.nixpkgs.lib;
      allDevSystems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAllDevSystems = f: lib.genAttrs allDevSystems (system: f rec {
        pkgs = import nixpkgs {
          config.allowUnfree = true;
          inherit system;
        };
        inherit system;
      });
    in
    {
      # Library function to create a runner for a project
      # Parameters:
      #   - system: The system architecture (e.g., "x86_64-linux")
      #   - infraConfig: Path to the Terranix infrastructure configuration
      #   - machineConfig: Path to the NixOS machine configuration
      #   - projectName: (Optional) Name for the project, used to organize .inframan/<projectName>/ folders
      #                  Defaults to "default" if not specified
      lib.mkRunner = { system, infraConfig, machineConfig, projectName ? "default" }:
        let
          pkgs = import nixpkgs {
            config.allowUnfree = true;
            inherit system;
          };

          # Generate the Terranix JSON configuration
          terranixConfig = terranix.lib.terranixConfiguration {
            inherit system;
            modules = [ infraConfig ];
          };

          # The inframan Go binary
          inframanBin = self.packages.${system}.default;
        in
        pkgs.writeShellApplication {
          name = "runner";
          runtimeInputs = [
            pkgs.terraform
            colmena.packages.${system}.colmena
            pkgs.nix
          ];
          text = ''
            # Export environment variables for the Go tool
            export INFRA_CONFIG_JSON="${terranixConfig}"
            export NIXOS_MODULE_PATH="${machineConfig}"
            export PROJECT_NAME="${projectName}"

            # Run the inframan binary with all arguments
            exec ${inframanBin}/bin/inframan "$@"
          '';
        };

      packages = forAllDevSystems ({pkgs, system, ...}: {
        default = pkgs.buildGoModule {
          pname = "inframan";
          version = "0.1.0";
          src = ./.;
          vendorHash = "sha256-eKeUhS2puz6ALb+cQKl7+DGvm9Cl+miZAHX0imf9wdg=";
          buildInputs = [ pkgs.go ];
          subPackages = [ "cmd/inframan" ];
        };
      });

      apps = forAllDevSystems ({pkgs, system, ...}: {
        default = {
          type = "app";
          program = "${self.packages.${system}.default}/bin/inframan";
        };
        inframan = {
          type = "app";
          program = "${self.packages.${system}.default}/bin/inframan";
        };
      });

      devShells = forAllDevSystems ({pkgs, system, ...}: {
        default = pkgs.mkShell {
          inputsFrom = [ self.packages.${system}.default ];
          packages = [
            pkgs.go
            pkgs.terraform
            colmena.packages.${system}.colmena
            pkgs.nix
          ];
        };
      });
    };
}

