# CLAUDE.md

Inframan is a Go CLI tool (Cobra framework) that orchestrates Terranix (Infrastructure as Nix) and Colmena (NixOS deployment). It provisions cloud infrastructure via Terraform and deploys NixOS configurations in a unified workflow. Distributed as a Nix flake.

## Architecture

```
cmd/inframan/main.go → internal/cli/root.go → internal/commands/*.go → internal/orchestrator/*.go
```

- **Commands** (`internal/commands/`): One file per CLI command, each exports `NewXxxCommand()` returning `*cobra.Command`
- **Orchestrator** (`internal/orchestrator/`): Core logic -- executor structs wrapping external CLIs via `os/exec`
- **Config** (`internal/orchestrator/config.go`): All env var getters and directory path helpers
- **Nix** (`flake.nix`): `lib.mkRunner` creates shell wrappers setting env vars and exec'ing the Go binary

Configuration flows: Nix (`flake.nix` `lib.mkRunner`) → environment variables → Go binary. All generated files live under `.inframan/<project-name>/{terraform,colmena}/`.

## Key Patterns

### Adding a New Command
1. Create `internal/commands/<name>.go` with `New<Name>Command()` returning `*cobra.Command`
2. Register in `internal/cli/root.go` `init()` with `rootCmd.AddCommand()`
3. Pattern: validate env vars → create executor(s) → perform operations → print status
4. Reference: `internal/commands/init.go` (simple), `internal/commands/ssh.go` (complex with flags)

### Adding a New Environment Variable
1. Add getter in `internal/orchestrator/config.go` (follow `GetSSHKeyPath` for optional strings, `GetBuildOnTarget` for booleans)
2. Add parameter to `lib.mkRunner` in `flake.nix` with `? null` or `? defaultValue`
3. Add conditional export in the `writeShellApplication` text block
4. Use the getter in relevant commands
5. Document in README.md env vars table and root command Long description

### Executor Pattern
All external tools are wrapped by executor structs in `internal/orchestrator/`:
- `TerraformExecutor`: `Init()`, `Apply()`, `Destroy()`, `Output()`, `Import()`, `GetTargetIP()`, `EnsureInit()`
- `TerranixExecutor`: `Build()`, `BuildFromConfig()`
- `ColmenaExecutor`: `GenerateHive()`, `Apply()`, `ValidateHive()`

Each executor: creates workdir on init, runs commands via `os/exec` with stdin/stdout/stderr piped through, passes `os.Environ()`, returns wrapped errors.

### Error Handling
- Always wrap: `fmt.Errorf("context: %w", err)`
- Commands return error from `RunE` (Cobra handles display)
- Validate env vars at start of command, fail fast with clear message

### IP Resolution
- Default: `public_ip` from terraform output
- When `USE_PRIVATE_IP=true` or `SSH_PROXY_JUMP` is set: prefer `private_ip`
- Multi-instance: `instances` map output takes precedence over `public_ip`

## Code Conventions
- Go: standard `go fmt`, exported functions have doc comments
- Nix: 2-space indent
- Commits: `type: short summary` (feat, fix, docs, refactor, test, chore)
- Branches: `feature/`, `fix/`, `docs/`, `refactor/` prefixes

## Build & Run
- `nix build` or `go build -o inframan ./cmd/inframan`
- `nix develop` for dev shell with Go, Terraform, Colmena
- Consumer usage: `nix run .#projectName -- <command>`
- When vendorHash changes: run `nix build`, copy the `got: sha256-...` from error, update `flake.nix`

## Files to Never Commit
- `.inframan/` directory (terraform state, generated hive.nix)
- `*.tfstate`, `*.tfstate.backup`, `terraform.tfvars`
- `*.pem`, `*.key` (SSH private keys)
- `result` (Nix build symlink)

## Important Implementation Details
- `ssh` command uses `syscall.Exec()` to replace the process (not `os/exec`), giving SSH full terminal control
- `output` command has `DisableFlagParsing: true` -- all flags pass through to terraform
- `deploy` re-copies `INFRA_CONFIG_JSON` to workspace before terraform init (needed for remote backends on different machines)
- `hive.nix` is generated fresh on every deploy with current IPs and SSH options
- `ColmenaExecutor.Apply()` sets `NIX_SSHOPTS` env var for `nix-copy-closure` SSH options
- `EnsureInit()` is idempotent -- skips if `.terraform/` directory already exists
- `GetAllInstances()` silently skips projects with errors (e.g., missing state)

## Documentation
Comprehensive docs are in `docs/`. See `docs/README.md` for the index.
