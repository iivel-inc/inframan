{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";
    terranix.url = "https://flakehub.com/f/terranix/terranix/*";
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
      packages = forAllDevSystems ({pkgs, system, ...}: {
        default = pkgs.buildGoModule {
          pname = "inframan";
          version = "0.1.0";
          src = ./.;
          vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="; # Will be updated on first build
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
          packages = with pkgs; [
            go
            terraform
            colmena
            nix
          ];
        };
      });
      # Colmena configuration (can be extended by projects using this flake)
      colmena = {
        meta = {
          nixpkgs = import nixpkgs { system = "x86_64-linux"; };
        };
      };
      # Colmena hive output for direct flake evaluation
      colmenaHive = colmena.lib.makeHive self.outputs.colmena;
    };
}

