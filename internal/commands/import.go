package commands

import (
	"fmt"

	"github.com/iivel-inc/inframan/internal/orchestrator"
	"github.com/spf13/cobra"
)

// NewImportCommand creates the import command
func NewImportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import [address] [id]",
		Short: "Import existing infrastructure into Terraform state",
		Long: `Import brings an existing resource into the Terraform state so it can
be managed by inframan going forward:

1. Ensures terraform is initialized in the project's terraform directory
2. Runs terraform import with the given resource address and ID
3. Passes through AWS credentials from environment

Example:
  inframan import aws_instance.web i-1234567890abcdef0`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			address := args[0]
			id := args[1]

			// Create terraform executor
			terraformExec, err := orchestrator.NewTerraformExecutor()
			if err != nil {
				return fmt.Errorf("failed to create terraform executor: %w", err)
			}

			// Ensure terraform is initialized
			if err := terraformExec.EnsureInit(); err != nil {
				return fmt.Errorf("failed to initialize terraform: %w", err)
			}

			// Run terraform import
			fmt.Printf("Importing %s as %s...\n", id, address)
			if err := terraformExec.Import(address, id); err != nil {
				return fmt.Errorf("terraform import failed: %w", err)
			}

			fmt.Println("Resource imported successfully!")
			return nil
		},
	}

	return cmd
}
