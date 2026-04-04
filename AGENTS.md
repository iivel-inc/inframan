# AGENTS.md

Guide for AI agents working with this repository. Start here to find what you need.

## What This Project Is

Inframan is a Go CLI that orchestrates Terranix (Nix → Terraform JSON) and Colmena (NixOS deployment). Users define infrastructure and machine config in Nix, then run `nix run .#<project> -- <command>` to provision and deploy.

## Where to Find Things

### Entry Points

| What | Where |
|------|-------|
| Go binary entry point | `cmd/inframan/main.go` |
| CLI command registration | `internal/cli/root.go` |
| Nix flake (build, mkRunner, devShell) | `flake.nix` |

### CLI Commands

Each command is a single file exporting `New<Name>Command()`:

| Command | File | Purpose |
|---------|------|---------|
| `init` | `internal/commands/init.go` | Initialize terraform workspace |
| `infra` | `internal/commands/infra.go` | Provision infrastructure |
| `deploy` | `internal/commands/deploy.go` | Deploy NixOS via Colmena |
| `destroy` | `internal/commands/destroy.go` | Tear down infrastructure |
| `ssh` | `internal/commands/ssh.go` | SSH to instances |
| `import` | `internal/commands/import.go` | Import existing resources |
| `output` | `internal/commands/output.go` | Show terraform outputs |

### Core Logic

| What | Where |
|------|-------|
| Environment variables and directory paths | `internal/orchestrator/config.go` |
| Terraform operations (init, apply, destroy, IP resolution) | `internal/orchestrator/terraform.go` |
| Terranix operations (build config, copy config) | `internal/orchestrator/terranix.go` |
| Colmena operations (generate hive.nix, apply) | `internal/orchestrator/colmena.go` |

### Configuration

| What | Where |
|------|-------|
| All env var getters and defaults | `internal/orchestrator/config.go` |
| `lib.mkRunner` parameter definitions | `flake.nix` (lines 20-40 comments, line 40 function signature) |
| Env var → mkRunner mapping | `flake.nix` (lines 57-98, the export blocks) |

### Examples and Templates

| What | Where |
|------|-------|
| Terranix infrastructure config | `example/infrastructure.nix` |
| NixOS machine config | `example/machine.nix` |
| Multi-project flake | `example/flake.nix` |
| Multi-account guide | `example/MULTI_ACCOUNT_GUIDE.md` |

### Documentation

| Topic | Where |
|-------|-------|
| Full docs index | `docs/README.md` |
| Architecture and data flow | `docs/architecture.md` |
| All CLI commands reference | `docs/commands.md` |
| All configuration options | `docs/configuration.md` |
| Nix/flake integration details | `docs/nix-integration.md` |
| Writing Terranix configs | `docs/infrastructure-templates.md` |
| Writing NixOS modules | `docs/machine-configuration.md` |
| Multi-project setup | `docs/multi-project.md` |
| SSH and networking | `docs/ssh-and-networking.md` |
| Development how-tos | `docs/development.md` |
| Troubleshooting | `docs/troubleshooting.md` |

### AI Agent Config

| What | Where |
|------|-------|
| Claude Code rules | `CLAUDE.md` |
| Claude Code skills | `.claude/skills/` |
| Claude Code settings | `.claude/settings.json` |

## Key Concepts to Understand

### Config Flow

```
Nix (flake.nix lib.mkRunner) → environment variables → Go binary reads os.Getenv()
```

The Go binary never reads config files directly. All configuration comes through env vars set by the Nix shell wrapper.

### Generated Files

Everything inframan generates lives under `.inframan/<project-name>/`:
- `terraform/config.tf.json` -- Terranix output copied here
- `terraform/.terraform/` -- Terraform plugin cache
- `terraform/terraform.tfstate` -- Local state (if not using remote backend)
- `colmena/hive.nix` -- Generated fresh on every `deploy`

### How to Add Things

- **New CLI command**: Create file in `internal/commands/`, register in `internal/cli/root.go` `init()`. See `docs/development.md`.
- **New env var / config option**: Add getter in `internal/orchestrator/config.go`, add parameter + export in `flake.nix`. See `docs/development.md`.
- **New orchestrator capability**: Add method to existing executor or create new one in `internal/orchestrator/`.

### Important Gotchas

- `ssh` command uses `syscall.Exec()` (process replacement), not `os/exec`
- `output` command has `DisableFlagParsing: true` -- all args pass through to terraform
- `deploy` re-copies `INFRA_CONFIG_JSON` before init (remote backends need this)
- In Terranix `.nix` files, Terraform `${...}` must be escaped: `"\${var.name}"`
- `.inframan/` is gitignored -- never commit terraform state or generated hive files

## Build and Run

```bash
nix develop              # Dev shell with Go, Terraform, Colmena
go build ./cmd/inframan  # Build the binary
nix build                # Build via Nix
go test ./...            # Run tests
```
