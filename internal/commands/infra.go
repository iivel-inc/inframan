package commands

import (
	"fmt"
	"os"

	"github.com/iivel-inc/inframan/internal/orchestrator"
	"github.com/spf13/cobra"
)

// NewInfraCommand creates the infra command
func NewInfraCommand() *cobra.Command {
	var autoApprove bool

	cmd := &cobra.Command{
		Use:   "infra",
		Short: "Apply infrastructure using Terranix and Terraform",
		Long: `Infra orchestrates infrastructure provisioning:
1. Reads the Terranix JSON config from INFRA_CONFIG_JSON env var
2. Copies config to .inframan/terraform/config.tf.json
3. Runs terraform init and terraform apply
4. Passes through AWS credentials from environment

Use --auto-approve (-y) to skip the interactive approval prompt
(typically required in CI environments).`,
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

			// Create terranix executor to copy config
			terranixExec, err := orchestrator.NewTerranixExecutor()
			if err != nil {
				return fmt.Errorf("failed to create terranix executor: %w", err)
			}

			// Setup workdir and copy config
			fmt.Println("Setting up infrastructure workspace...")
			if _, err := terranixExec.BuildFromConfig(infraConfigJSON); err != nil {
				return fmt.Errorf("failed to setup workdir: %w", err)
			}

			// Create terraform executor
			terraformExec, err := orchestrator.NewTerraformExecutor()
			if err != nil {
				return fmt.Errorf("failed to create terraform executor: %w", err)
			}

			// Run terraform init
			fmt.Println("Initializing Terraform...")
			if err := terraformExec.Init(); err != nil {
				return fmt.Errorf("terraform init failed: %w", err)
			}

			// Run terraform apply
			fmt.Println("Applying infrastructure...")
			if err := terraformExec.Apply(autoApprove); err != nil {
				return fmt.Errorf("terraform apply failed: %w", err)
			}

			fmt.Println("Infrastructure applied successfully!")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&autoApprove, "auto-approve", "y", false, "Skip interactive approval of terraform apply")

	return cmd
}
