package commands

import (
	"fmt"
	"os"

	"github.com/iivel-inc/inframan/internal/orchestrator"
	"github.com/spf13/cobra"
)

// NewInitCommand creates the init command
func NewInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize terraform workspace and pull remote state",
		Long: `Init sets up the terraform workspace without applying changes:
1. Copies the Terranix JSON config to .inframan/terraform/config.tf.json
2. Runs terraform init to connect to the backend and pull remote state

This is useful when you need terraform state (e.g., for deploy or ssh)
but haven't run 'infra' in this environment yet.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get INFRA_CONFIG_JSON from environment
			infraConfigJSON := os.Getenv("INFRA_CONFIG_JSON")
			if infraConfigJSON == "" {
				return fmt.Errorf("INFRA_CONFIG_JSON environment variable is not set")
			}

			// Verify the config file exists
			if _, err := os.Stat(infraConfigJSON); os.IsNotExist(err) {
				return fmt.Errorf("INFRA_CONFIG_JSON file does not exist: %s", infraConfigJSON)
			}

			// Copy config to workspace
			fmt.Println("Setting up terraform workspace...")
			terranixExec, err := orchestrator.NewTerranixExecutor()
			if err != nil {
				return fmt.Errorf("failed to create terranix executor: %w", err)
			}
			if _, err := terranixExec.BuildFromConfig(infraConfigJSON); err != nil {
				return fmt.Errorf("failed to setup workspace: %w", err)
			}

			// Run terraform init
			terraformExec, err := orchestrator.NewTerraformExecutor()
			if err != nil {
				return fmt.Errorf("failed to create terraform executor: %w", err)
			}

			fmt.Println("Initializing Terraform...")
			if err := terraformExec.Init(); err != nil {
				return fmt.Errorf("terraform init failed: %w", err)
			}

			fmt.Println("Workspace initialized successfully!")
			return nil
		},
	}

	return cmd
}
