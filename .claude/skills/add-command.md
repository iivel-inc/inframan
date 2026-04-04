---
name: add-command
description: Step-by-step guide for adding a new inframan CLI command
trigger: When the user asks to add, create, or implement a new CLI command or subcommand
---

# Adding a New Inframan CLI Command

## Steps

### 1. Create the command file

Create `internal/commands/<name>.go`:

```go
package commands

import (
    "fmt"
    "os"

    "github.com/iivel-inc/inframan/internal/orchestrator"
    "github.com/spf13/cobra"
)

// New<Name>Command creates the <name> command
func New<Name>Command() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "<name>",
        Short: "Brief description",
        Long:  `Detailed description of what the command does.`,
        RunE: func(cmd *cobra.Command, args []string) error {
            // 1. Validate required environment variables
            // 2. Create executor(s) from internal/orchestrator/
            // 3. Perform operations
            // 4. Print status messages
            return nil
        },
    }

    // Add flags if needed:
    // var flagVar string
    // cmd.Flags().StringVarP(&flagVar, "flag-name", "f", "default", "description")

    return cmd
}
```

### 2. Register the command

In `internal/cli/root.go`, add to the `init()` function:

```go
rootCmd.AddCommand(commands.New<Name>Command())
```

### 3. Add orchestrator logic (if needed)

If the command needs new orchestrator functions, add them to the appropriate file in `internal/orchestrator/`:
- `terraform.go` for Terraform operations
- `colmena.go` for Colmena operations
- `terranix.go` for Terranix operations
- Or create a new executor file following the existing pattern

### 4. Build and test

```bash
go build -o inframan ./cmd/inframan
./inframan <name> --help
```

## Reference Commands

- **Simple command** (env var validation + executor): `internal/commands/init.go`
- **Command with flags** (StringVarP, BoolVarP, argument parsing): `internal/commands/ssh.go`
- **Passthrough command** (DisableFlagParsing): `internal/commands/output.go`
- **Two-arg command** (cobra.ExactArgs): `internal/commands/import.go`

## Conventions

- Function name: `New<Name>Command()` returning `*cobra.Command`
- Use `RunE` (not `Run`) to return errors -- Cobra handles display
- Validate env vars at the top of RunE, fail fast
- Wrap errors with context: `fmt.Errorf("context: %w", err)`
- Print progress messages with `fmt.Println()`
- Pipe stdin/stdout/stderr through when running interactive external commands
