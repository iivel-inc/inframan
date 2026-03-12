package commands

import (
	"fmt"

	"github.com/iivel-inc/inframan/internal/orchestrator"
	"github.com/spf13/cobra"
)

// NewOutputCommand creates the output command
func NewOutputCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "output [name]",
		Short: "Show Terraform output values",
		Long: `Output displays the values of Terraform outputs from the current project's state:

1. Ensures terraform is initialized in the project's terraform directory
2. Runs terraform output with any additional arguments passed through

Examples:
  inframan output              # Show all outputs
  inframan output -json        # Show all outputs as JSON
  inframan output public_ip    # Show a specific output value`,
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create terraform executor
			terraformExec, err := orchestrator.NewTerraformExecutor()
			if err != nil {
				return fmt.Errorf("failed to create terraform executor: %w", err)
			}

			// Ensure terraform is initialized
			if err := terraformExec.EnsureInit(); err != nil {
				return fmt.Errorf("failed to initialize terraform: %w", err)
			}

			// Run terraform output with all args passed through
			if err := terraformExec.Output(args...); err != nil {
				return fmt.Errorf("terraform output failed: %w", err)
			}

			return nil
		},
	}

	return cmd
}
