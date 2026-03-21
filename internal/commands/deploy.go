package commands

import (
	"fmt"
	"os"

	"github.com/iivel-inc/inframan/internal/orchestrator"
	"github.com/spf13/cobra"
)

// NewDeployCommand creates the deploy command
func NewDeployCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy NixOS configuration using Colmena",
		Long: `Deploy orchestrates NixOS deployment:
1. Fetches infrastructure state from Terraform
2. Parses target IP from terraform output
3. Generates ephemeral hive.nix with injected IP
4. Runs colmena apply to deploy to the target`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get NIXOS_MODULE_PATH from environment
			nixosModulePath := os.Getenv("NIXOS_MODULE_PATH")
			if nixosModulePath == "" {
				return fmt.Errorf("NIXOS_MODULE_PATH environment variable is not set")
			}

			// Verify the module file exists
			if _, err := os.Stat(nixosModulePath); os.IsNotExist(err) {
				return fmt.Errorf("NIXOS_MODULE_PATH file does not exist: %s", nixosModulePath)
			}

			// Copy Terranix config to workspace (needed for terraform init with remote backends)
			infraConfigJSON := os.Getenv("INFRA_CONFIG_JSON")
			if infraConfigJSON == "" {
				return fmt.Errorf("INFRA_CONFIG_JSON environment variable is not set")
			}
			terranixExec, err := orchestrator.NewTerranixExecutor()
			if err != nil {
				return fmt.Errorf("failed to create terranix executor: %w", err)
			}
			if _, err := terranixExec.BuildFromConfig(infraConfigJSON); err != nil {
				return fmt.Errorf("failed to setup terraform workspace: %w", err)
			}

			// Create terraform executor to get output
			terraformExec, err := orchestrator.NewTerraformExecutor()
			if err != nil {
				return fmt.Errorf("failed to create terraform executor: %w", err)
			}

			// Get target IP from terraform output
			fmt.Println("Fetching infrastructure state...")
			targetIP, err := terraformExec.GetTargetIP()
			if err != nil {
				return fmt.Errorf("failed to get target IP: %w", err)
			}
			fmt.Printf("Target IP: %s\n", targetIP)

			// Create colmena executor
			colmenaExec, err := orchestrator.NewColmenaExecutor()
			if err != nil {
				return fmt.Errorf("failed to create colmena executor: %w", err)
			}

			// Generate dynamic hive.nix
			fmt.Println("Generating Colmena hive configuration...")
			hivePath, err := colmenaExec.GenerateHive(nixosModulePath, targetIP)
			if err != nil {
				return fmt.Errorf("failed to generate hive: %w", err)
			}
			fmt.Printf("Generated hive at: %s\n", hivePath)

			// Run colmena apply
			fmt.Println("Deploying with Colmena...")
			if err := colmenaExec.Apply(hivePath); err != nil {
				return fmt.Errorf("colmena apply failed: %w", err)
			}

			fmt.Println("Deployment completed successfully!")
			return nil
		},
	}

	return cmd
}
