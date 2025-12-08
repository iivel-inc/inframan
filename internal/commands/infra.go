package commands

import (
	"fmt"
	"os"

	"github.com/iivel-inc/inframan/internal/orchestrator"
	"github.com/spf13/cobra"
)

// NewInfraCommand creates the infra command
func NewInfraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "infra",
		Short: "Apply infrastructure using Terranix and OpenTofu",
		Long: `Infra orchestrates infrastructure provisioning:
1. Reads the Terranix JSON config from INFRA_CONFIG_JSON env var
2. Copies config to .runner-workdir/config.tf.json
3. Runs tofu init and tofu apply
4. Passes through AWS credentials from environment`,
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

			// Create tofu executor
			tofuExec, err := orchestrator.NewTofuExecutor()
			if err != nil {
				return fmt.Errorf("failed to create tofu executor: %w", err)
			}

			// Setup workdir and copy config
			fmt.Println("Setting up infrastructure workspace...")
			if err := tofuExec.SetupWorkdir(infraConfigJSON); err != nil {
				return fmt.Errorf("failed to setup workdir: %w", err)
			}

			// Run tofu init
			fmt.Println("Initializing OpenTofu...")
			if err := tofuExec.Init(); err != nil {
				return fmt.Errorf("tofu init failed: %w", err)
			}

			// Run tofu apply
			fmt.Println("Applying infrastructure...")
			if err := tofuExec.Apply(); err != nil {
				return fmt.Errorf("tofu apply failed: %w", err)
			}

			fmt.Println("Infrastructure applied successfully!")
			return nil
		},
	}

	return cmd
}

