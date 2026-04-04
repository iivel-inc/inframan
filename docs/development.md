# Development Guide

How to work on inframan itself -- the Go CLI, the Nix flake, and the project conventions.

## Development Environment

```bash
git clone https://github.com/iivel-inc/inframan
cd inframan
nix develop
```

This provides Go, Terraform, Colmena, and Nix in your shell.

## Building

```bash
# Via Nix (produces result/bin/inframan)
nix build

# Via Go directly
go build -o inframan ./cmd/inframan

# Run directly for testing
go run ./cmd/inframan -- help
```

## Project Structure

```
cmd/inframan/main.go              Entry point
internal/
├── cli/root.go                   Cobra root command + subcommand registration
├── commands/                     One file per CLI command
│   ├── init.go                   NewInitCommand()
│   ├── infra.go                  NewInfraCommand()
│   ├── deploy.go                 NewDeployCommand()
│   ├── destroy.go                NewDestroyCommand()
│   ├── ssh.go                    NewSSHCommand()
│   ├── import.go                 NewImportCommand()
│   └── output.go                 NewOutputCommand()
└── orchestrator/                 Core logic
    ├── config.go                 Env var getters, directory paths, constants
    ├── terraform.go              TerraformExecutor + instance discovery
    ├── terranix.go               TerranixExecutor
    └── colmena.go                ColmenaExecutor + hive generation
```

## How to Add a New Command

1. **Create the command file** at `internal/commands/<name>.go`:

```go
package commands

import (
    "fmt"
    "github.com/iivel-inc/inframan/internal/orchestrator"
    "github.com/spf13/cobra"
)

func NewStatusCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "status",
        Short: "Show project status",
        RunE: func(cmd *cobra.Command, args []string) error {
            // 1. Validate env vars
            // 2. Create executor(s)
            // 3. Perform operations
            // 4. Print status
            fmt.Println("Status: OK")
            return nil
        },
    }

    // Add flags if needed:
    // cmd.Flags().StringVarP(&someVar, "flag", "f", "default", "description")

    return cmd
}
```

2. **Register in `internal/cli/root.go`**:

```go
func init() {
    // ... existing commands
    rootCmd.AddCommand(commands.NewStatusCommand())
}
```

3. **Build and test**:

```bash
go build -o inframan ./cmd/inframan
./inframan status
```

### Patterns to Follow

- Look at `internal/commands/init.go` for a simple command (validate env, create executor, run)
- Look at `internal/commands/ssh.go` for a complex command (flags, argument parsing, instance discovery)
- Look at `internal/commands/output.go` for a passthrough command (`DisableFlagParsing: true`)

## How to Add a New Environment Variable

1. **Add getter in `internal/orchestrator/config.go`**:

```go
// GetMyNewOption returns the value from environment, or default
func GetMyNewOption() string {
    v := os.Getenv("MY_NEW_OPTION")
    if v == "" {
        return "default-value"
    }
    return v
}
```

For booleans, follow the `GetBuildOnTarget()` pattern:

```go
func GetMyFlag() bool {
    v := os.Getenv("MY_FLAG")
    if v == "false" || v == "0" {
        return false
    }
    return true  // default true
}
```

2. **Add parameter to `lib.mkRunner` in `flake.nix`**:

```nix
lib.mkRunner = { system, ..., myNewOption ? null }:
  let
    # ...
    myNewOptionExport = if myNewOption != null
      then ''export MY_NEW_OPTION="${myNewOption}"''
      else "";
  in
  pkgs.writeShellApplication {
    # ...
    text = ''
      # ... existing exports
      ${myNewOptionExport}
      exec ${inframanBin}/bin/inframan "$@"
    '';
  };
```

3. **Use the getter** in the relevant command or orchestrator function.

4. **Document** in the root command's `Long` description (`internal/cli/root.go`) and the README.

## Executor Pattern

All external tool interaction goes through executor structs:

```go
type MyExecutor struct {
    workDir string
}

func NewMyExecutor() (*MyExecutor, error) {
    workDir, err := GetSomeDir()
    if err != nil {
        return nil, fmt.Errorf("failed to get directory: %w", err)
    }
    if err := EnsureDir(workDir); err != nil {
        return nil, err
    }
    return &MyExecutor{workDir: workDir}, nil
}

func (e *MyExecutor) Run() error {
    cmd := exec.Command("my-tool", "arg1", "arg2")
    cmd.Dir = e.workDir
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Stdin = os.Stdin
    cmd.Env = os.Environ()

    if err := cmd.Run(); err != nil {
        return fmt.Errorf("my-tool failed: %w", err)
    }
    return nil
}
```

Key conventions:
- Always pipe `Stdout`, `Stderr`, `Stdin` through for interactive commands
- Always pass `os.Environ()` so AWS creds and other env vars flow through
- Always wrap errors with `fmt.Errorf("context: %w", err)`

## Error Handling

- Wrap all errors with context: `fmt.Errorf("failed to X: %w", err)`
- Commands return errors from `RunE` -- Cobra handles display
- Validate required env vars at the start of each command, fail fast with a clear message
- Check file existence before operations (`os.Stat`)

## Code Conventions

- **Go**: Standard `go fmt` formatting, exported functions have doc comments
- **Nix**: 2-space indentation
- **Commits**: `type: short summary` (types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`)
- **Branches**: `feature/`, `fix/`, `docs/`, `refactor/`, `test/` prefixes

## Testing

```bash
go test ./...
go test -cover ./...
go test ./internal/orchestrator
```

Use table-driven tests:

```go
func TestGetProjectName(t *testing.T) {
    tests := []struct {
        name    string
        envVar  string
        want    string
    }{
        {"default", "", "default"},
        {"custom", "prod", "prod"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            os.Setenv("PROJECT_NAME", tt.envVar)
            defer os.Unsetenv("PROJECT_NAME")
            if got := orchestrator.GetProjectName(); got != tt.want {
                t.Errorf("GetProjectName() = %q, want %q", got, tt.want)
            }
        })
    }
}
```

## Updating Dependencies

When Go dependencies change:

```bash
go mod tidy
nix build  # Will fail with hash mismatch
# Copy the "got: sha256-..." hash from the error
# Update vendorHash in flake.nix
nix build  # Should succeed
```

## Files to Never Commit

- `.inframan/` -- Terraform state, generated hive files
- `*.tfstate`, `*.tfstate.backup`
- `terraform.tfvars` -- Credentials
- `*.pem`, `*.key` -- SSH private keys
- `result` -- Nix build output symlink
