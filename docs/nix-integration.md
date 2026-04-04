# Nix Integration

Inframan is deeply integrated with the Nix ecosystem. The Go binary handles runtime orchestration, but the Nix flake handles configuration compilation, dependency management, and project composition.

## How `flake.nix` Works

The main `flake.nix` in the inframan repository defines:

### Inputs

```nix
inputs = {
  nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";
  terranix.url = "github:terranix/terranix";
  colmena.url = "github:zhaofengli/colmena";
};
```

- **nixpkgs** -- Standard Nix packages (provides Go, Terraform, etc.)
- **terranix** -- Compiles Nix attribute sets to Terraform JSON
- **colmena** -- NixOS deployment tool

### Outputs

| Output | Description |
|--------|-------------|
| `lib.mkRunner` | Factory function to create project-specific runners |
| `packages.<system>.default` | The Go binary built with `buildGoModule` |
| `apps.<system>.default` | Direct app pointing to the binary |
| `devShells.<system>.default` | Development shell with Go, Terraform, Colmena, Nix |

## `lib.mkRunner` in Detail

`mkRunner` is the primary API for consuming inframan in downstream projects. It creates a shell wrapper script (`writeShellApplication`) that:

1. **Compiles infrastructure config at Nix eval time:**
   ```nix
   terranixConfig = terranix.lib.terranixConfiguration {
     inherit system;
     modules = [ infraConfig ];
   };
   ```
   This produces a Nix store path (e.g., `/nix/store/abc...-config.tf.json`) containing the compiled Terraform JSON.

2. **Sets environment variables:**
   ```bash
   export INFRA_CONFIG_JSON="/nix/store/abc...-config.tf.json"
   export NIXOS_MODULE_PATH="/nix/store/xyz.../machine.nix"
   export PROJECT_NAME="production"
   export TARGET_SYSTEM="x86_64-linux"
   export BUILD_ON_TARGET="true"
   # SSH vars set conditionally (only if non-null)
   ```

3. **Bundles runtime dependencies:**
   ```nix
   runtimeInputs = [ pkgs.terraform colmena.packages.${system}.colmena pkgs.nix ];
   ```
   Terraform, Colmena, and Nix are added to `PATH`.

4. **Execs the Go binary:**
   ```bash
   exec ${inframanBin}/bin/inframan "$@"
   ```

### Conditional Exports

Optional parameters (`sshKeyPath`, `sshConfigPath`, `sshProxyJump`, `usePrivateIp`) are only exported when non-null/non-default. This means the Go binary sees an empty string (via `os.Getenv`) for unset options, which it treats as "not configured."

## Consumer Flake Pattern

Downstream projects reference inframan as a flake input and use `lib.mkRunner`:

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";
    inframan.url = "github:iivel-inc/inframan";
  };

  outputs = { self, nixpkgs, inframan, ... }:
    let
      system = "x86_64-linux";
    in
    {
      # Single project
      apps.${system}.default = {
        type = "app";
        program = "${inframan.lib.mkRunner {
          inherit system;
          infraConfig = ./infrastructure.nix;
          machineConfig = ./machine.nix;
        }}/bin/runner";
      };

      # Multiple projects
      apps.${system}.prod = {
        type = "app";
        program = "${inframan.lib.mkRunner {
          inherit system;
          infraConfig = ./infra-prod.nix;
          machineConfig = ./machine-prod.nix;
          projectName = "production";
          sshKeyPath = "./keys/prod.pem";
        }}/bin/runner";
      };

      apps.${system}.staging = {
        type = "app";
        program = "${inframan.lib.mkRunner {
          inherit system;
          infraConfig = ./infra-staging.nix;
          machineConfig = ./machine-staging.nix;
          projectName = "staging";
        }}/bin/runner";
      };
    };
}
```

Then run with:

```bash
nix run .#prod -- infra
nix run .#staging -- deploy
```

## Cross-Architecture

The `system` parameter controls where the runner builds. The `targetSystem` parameter controls the target NixOS architecture in the generated `hive.nix`:

```nix
# Build runner on macOS, deploy to ARM Linux
inframan.lib.mkRunner {
  system = "aarch64-darwin";      # Runner runs on macOS
  targetSystem = "aarch64-linux"; # Target is ARM Linux
  infraConfig = ./infrastructure.nix;
  machineConfig = ./machine.nix;
}
```

## The Go Binary Build

The Go binary is built with `buildGoModule`:

```nix
packages.default = pkgs.buildGoModule {
  pname = "inframan";
  version = "0.1.0";
  src = ./.;
  vendorHash = "sha256-eKeUhS2puz6ALb+cQKl7+DGvm9Cl+miZAHX0imf9wdg=";
  subPackages = [ "cmd/inframan" ];
};
```

When Go dependencies change, `vendorHash` must be updated. To get the new hash:

```bash
nix build 2>&1 | grep "got:"
# Use the hash from the error output
```

## Development Shell

```bash
nix develop
```

Provides: Go, Terraform, Colmena, Nix. Allows building and testing without the full flake wrapper:

```bash
go build -o inframan ./cmd/inframan
export INFRA_CONFIG_JSON=./config.tf.json
export NIXOS_MODULE_PATH=./machine.nix
./inframan infra
```
