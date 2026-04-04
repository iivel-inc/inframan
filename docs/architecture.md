# Architecture

Inframan orchestrates three external tools through a Go CLI, connecting Nix-based configuration with infrastructure provisioning and NixOS deployment.

## Three-Layer Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                   Nix Configuration Layer                    │
│  flake.nix (lib.mkRunner) + infrastructure.nix + machine.nix│
└──────────────────────┬──────────────────────────────────────┘
                       │ Environment variables
┌──────────────────────▼──────────────────────────────────────┐
│                  Go CLI Orchestration Layer                  │
│  cmd/inframan → internal/cli → internal/commands            │
│                        → internal/orchestrator              │
└──────────────────────┬──────────────────────────────────────┘
                       │ os/exec + syscall.Exec
┌──────────────────────▼──────────────────────────────────────┐
│                    External Tools Layer                      │
│  Terraform · Terranix · Colmena · SSH                       │
└─────────────────────────────────────────────────────────────┘
```

## Data Flow

```
infrastructure.nix ──(Nix eval)──▶ config.tf.json ──(terraform)──▶ AWS Resources
                                                                        │
                                                                   terraform output
                                                                        │
machine.nix ──(colmena)──▶ NixOS Deployment ◀── target IP ◀────────────┘
```

1. **Nix evaluation time**: `terranix.lib.terranixConfiguration` compiles `infrastructure.nix` into a JSON file (Terraform config) stored in the Nix store
2. **`inframan infra`**: Copies the JSON to `.inframan/<project>/terraform/config.tf.json`, runs `terraform init` + `terraform apply`
3. **`inframan deploy`**: Reads `terraform output -json` to get the instance IP, generates an ephemeral `hive.nix` injecting that IP, runs `colmena apply`

## The `lib.mkRunner` Bridge

The Nix `lib.mkRunner` function creates a shell wrapper (`writeShellApplication`) that:

1. Sets environment variables (`INFRA_CONFIG_JSON`, `NIXOS_MODULE_PATH`, `PROJECT_NAME`, etc.)
2. Adds `terraform`, `colmena`, and `nix` to `PATH` via `runtimeInputs`
3. Execs the Go binary with all arguments passed through

This bridges Nix (compile-time config) and Go (runtime orchestration) via the environment.

## Project Isolation

Each project gets its own isolated directory under `.inframan/`:

```
.inframan/
├── production/
│   ├── terraform/          # Terraform state, config.tf.json, .terraform/
│   └── colmena/            # Generated hive.nix
├── staging/
│   ├── terraform/
│   └── colmena/
└── default/                # When projectName is not specified
    ├── terraform/
    └── colmena/
```

Projects never share state. Each has independent Terraform state and Colmena configuration.

## Go Project Structure

```
cmd/inframan/main.go              Entry point, calls cli.Execute()
internal/
├── cli/root.go                   Cobra root command, registers all subcommands
├── commands/                     One file per command
│   ├── init.go                   NewInitCommand()
│   ├── infra.go                  NewInfraCommand()
│   ├── deploy.go                 NewDeployCommand()
│   ├── destroy.go                NewDestroyCommand()
│   ├── ssh.go                    NewSSHCommand()
│   ├── import.go                 NewImportCommand()
│   └── output.go                 NewOutputCommand()
└── orchestrator/                 Core logic, one file per concern
    ├── config.go                 Environment variables, directory paths, constants
    ├── terraform.go              TerraformExecutor: init, apply, destroy, output, import, IP resolution
    ├── terranix.go               TerranixExecutor: build from Nix, copy pre-built config
    └── colmena.go                ColmenaExecutor: generate hive.nix, apply deployment
```

## Executor Pattern

All external tool interaction goes through executor structs in `internal/orchestrator/`:

- **TerraformExecutor** -- Wraps `terraform` CLI. Manages working directory, passes environment (including AWS creds). Key methods: `Init()`, `Apply()`, `Destroy()`, `Output()`, `Import()`, `GetTargetIP()`, `EnsureInit()`.
- **TerranixExecutor** -- Wraps `terranix` CLI or copies pre-built config. Key methods: `Build()`, `BuildFromConfig()`.
- **ColmenaExecutor** -- Wraps `colmena` CLI. Generates dynamic `hive.nix` files. Key methods: `GenerateHive()`, `Apply()`, `ValidateHive()`.

Each executor:
1. Creates its working directory on initialization
2. Runs external commands via `os/exec` with `Stdin`/`Stdout`/`Stderr` piped through
3. Passes the full environment (`os.Environ()`) to child processes
4. Returns wrapped errors with context

## IP Resolution

When deploying or connecting via SSH, inframan needs the target IP. The resolution logic:

1. Parse `terraform output -json`
2. Check for `instances` map (multi-instance projects) -- returns `map[string]string` of name-to-IP
3. If `USE_PRIVATE_IP=true` or `SSH_PROXY_JUMP` is set, prefer `private_ip`
4. Otherwise use `public_ip`
5. Error if no IP found

## SSH Process Replacement

The `ssh` command uses `syscall.Exec()` (not `os/exec`) to replace the inframan process with the SSH binary. This gives SSH full terminal control -- TTY allocation, signal handling, piped input, and interactive sessions all work correctly.
