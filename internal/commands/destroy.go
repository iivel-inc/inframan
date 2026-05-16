package commands

import (
	"fmt"

	"github.com/iivel-inc/inframan/internal/orchestrator"
	"github.com/spf13/cobra"
)

// NewDestroyCommand creates the destroy command
func NewDestroyCommand() *cobra.Command {
	var autoApprove bool

	cmd := &cobra.Command{
		Use:   "destroy",
		Short: "Destroy infrastructure using Terraform",
		Long: `Destroy tears down infrastructure provisioned by inframan:
1. Runs terraform destroy in the project's terraform directory
2. Removes all resources tracked in the terraform state
3. Passes through AWS credentials from environment

This is the reverse of 'inframan infra' and will destroy all resources
that were created during infrastructure provisioning.

Use --auto-approve (-y) to skip the interactive approval prompt
(typically required in CI environments).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create terraform executor
			terraformExec, err := orchestrator.NewTerraformExecutor()
			if err != nil {
				return fmt.Errorf("failed to create terraform executor: %w", err)
			}

			// Ensure terraform is initialized (needed for remote backends in CI)
			if err := terraformExec.EnsureInit(); err != nil {
				return fmt.Errorf("failed to initialize terraform: %w", err)
			}

			// Run terraform destroy
			fmt.Println("Destroying infrastructure...")
			if err := terraformExec.Destroy(autoApprove); err != nil {
				return fmt.Errorf("terraform destroy failed: %w", err)
			}

			fmt.Println("Infrastructure destroyed successfully!")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&autoApprove, "auto-approve", "y", false, "Skip interactive approval of terraform destroy")

	return cmd
}
