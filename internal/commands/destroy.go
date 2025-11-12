package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/iivel-inc/inframan/internal/orchestrator"
	"github.com/iivel-inc/inframan/internal/state"
	"github.com/spf13/cobra"
)

// NewDestroyCommand creates the destroy command
func NewDestroyCommand() *cobra.Command {
	var projectDir string

	cmd := &cobra.Command{
		Use:   "destroy [project]",
		Short: "Destroy a project's configuration",
		Long: `Destroy reverses the apply workflow:
1. Run terraform destroy
2. Run colmena destroy`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project := args[0]

			// Get flags from parent
			projectDir, _ = cmd.Root().PersistentFlags().GetString("dir")
			if projectDir == "" {
				var err error
				projectDir, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get working directory: %w", err)
				}
			}

			// Initialize state manager
			stateManager, err := state.NewManager(projectDir)
			if err != nil {
				return fmt.Errorf("failed to initialize state manager: %w", err)
			}

			projectDeployDir := filepath.Join(projectDir, "deployments", project)

			// Step 1: Terraform destroy
			fmt.Printf("Running terraform destroy...\n")
			stateManager.UpdateStepStatus(state.StepTerraform, state.StatusRunning, "Running terraform destroy", nil)
			stateManager.Save()

			tfExec := orchestrator.NewTerraformExecutor(projectDeployDir)
			if err := tfExec.Destroy(); err != nil {
				stateManager.UpdateStepStatus(state.StepTerraform, state.StatusFailed, "", err)
				stateManager.Save()
				return fmt.Errorf("terraform destroy failed: %w", err)
			}

			stateManager.UpdateStepStatus(state.StepTerraform, state.StatusSuccess, "Terraform destroyed", nil)
			stateManager.Save()

			// Step 2: Colmena destroy
			fmt.Printf("Running colmena destroy...\n")
			stateManager.UpdateStepStatus(state.StepColmena, state.StatusRunning, "Running colmena destroy", nil)
			stateManager.Save()

			colmenaExec := orchestrator.NewColmenaExecutor(projectDir)
			if err := colmenaExec.Destroy(project); err != nil {
				stateManager.UpdateStepStatus(state.StepColmena, state.StatusFailed, "", err)
				stateManager.SetColmenaApplied(false)
				stateManager.Save()
				return fmt.Errorf("colmena destroy failed: %w", err)
			}

			stateManager.UpdateStepStatus(state.StepColmena, state.StatusSuccess, "Colmena destroyed", nil)
			stateManager.SetColmenaApplied(false)
			stateManager.Save()

			fmt.Printf("Successfully destroyed project '%s'\n", project)
			return nil
		},
	}

	return cmd
}

