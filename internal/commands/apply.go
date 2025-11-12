package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/iivel-inc/inframan/internal/orchestrator"
	"github.com/iivel-inc/inframan/internal/state"
	"github.com/spf13/cobra"
)

// NewApplyCommand creates the apply command
func NewApplyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply [project]",
		Short: "Apply a project's configuration",
		Long: `Apply orchestrates the full workflow:
1. Build terranix configuration using nix
2. Copy config to project directory
3. Run terraform init and apply
4. Save terraform output
5. Run colmena apply`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project := args[0]

			// Get flags from parent
			flakeRef, _ := cmd.Root().PersistentFlags().GetString("flake")
			projectDir, _ := cmd.Root().PersistentFlags().GetString("dir")
			if projectDir == "" {
				var err error
				projectDir, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get working directory: %w", err)
				}
			}
			if flakeRef == "" {
				flakeRef = "."
			}

			// Initialize state manager
			stateManager, err := state.NewManager(projectDir)
			if err != nil {
				return fmt.Errorf("failed to initialize state manager: %w", err)
			}
			stateManager.SetProject(project)

			// Step 1: Build terranix config
			fmt.Printf("Building terranix configuration for project '%s'...\n", project)
			stateManager.UpdateStepStatus(state.StepNixBuild, state.StatusRunning, "Building terranix config", nil)

			nixExec := orchestrator.NewNixExecutor(flakeRef)
			terranixPath, err := nixExec.BuildTerranix(project)
			if err != nil {
				stateManager.UpdateStepStatus(state.StepNixBuild, state.StatusFailed, "", err)
				stateManager.Save()
				return fmt.Errorf("failed to build terranix config: %w", err)
			}

			// Step 2: Copy config to project directory
			projectDeployDir := filepath.Join(projectDir, "deployments", project)
			if err := os.MkdirAll(projectDeployDir, 0755); err != nil {
				stateManager.UpdateStepStatus(state.StepNixBuild, state.StatusFailed, "", err)
				stateManager.Save()
				return fmt.Errorf("failed to create project directory: %w", err)
			}

			configPath, err := nixExec.CopyTerranixConfig(terranixPath, projectDeployDir)
			if err != nil {
				stateManager.UpdateStepStatus(state.StepNixBuild, state.StatusFailed, "", err)
				stateManager.Save()
				return fmt.Errorf("failed to copy terranix config: %w", err)
			}

			stateManager.SetTerranixConfig(configPath)
			stateManager.UpdateStepStatus(state.StepNixBuild, state.StatusSuccess, "Terranix config built and copied", nil)
			stateManager.Save()

			// Step 3: Terraform init and apply
			fmt.Printf("Running terraform init and apply...\n")
			stateManager.UpdateStepStatus(state.StepTerraform, state.StatusRunning, "Running terraform", nil)
			stateManager.Save()

			tfExec := orchestrator.NewTerraformExecutor(projectDeployDir)
			if err := tfExec.Init(); err != nil {
				stateManager.UpdateStepStatus(state.StepTerraform, state.StatusFailed, "", err)
				stateManager.Save()
				return fmt.Errorf("terraform init failed: %w", err)
			}

			if err := tfExec.Apply(); err != nil {
				stateManager.UpdateStepStatus(state.StepTerraform, state.StatusFailed, "", err)
				stateManager.Save()
				return fmt.Errorf("terraform apply failed: %w", err)
			}

			// Step 4: Save terraform output
			outputPath, err := tfExec.Output()
			if err != nil {
				stateManager.UpdateStepStatus(state.StepTerraform, state.StatusFailed, "", err)
				stateManager.Save()
				return fmt.Errorf("failed to save terraform output: %w", err)
			}

			stateManager.SetTerraformState(filepath.Join(projectDeployDir, ".terraform"))
			stateManager.UpdateStepStatus(state.StepTerraform, state.StatusSuccess, fmt.Sprintf("Terraform applied, output saved to %s", outputPath), nil)
			stateManager.Save()

			// Step 5: Colmena apply
			fmt.Printf("Running colmena apply...\n")
			stateManager.UpdateStepStatus(state.StepColmena, state.StatusRunning, "Running colmena apply", nil)
			stateManager.Save()

			colmenaExec := orchestrator.NewColmenaExecutor(projectDir)
			if err := colmenaExec.Apply(project); err != nil {
				stateManager.UpdateStepStatus(state.StepColmena, state.StatusFailed, "", err)
				stateManager.SetColmenaApplied(false)
				stateManager.Save()
				return fmt.Errorf("colmena apply failed: %w", err)
			}

			stateManager.UpdateStepStatus(state.StepColmena, state.StatusSuccess, "Colmena applied successfully", nil)
			stateManager.SetColmenaApplied(true)
			stateManager.SetLastApplied()
			stateManager.Save()

			fmt.Printf("Successfully applied project '%s'\n", project)
			return nil
		},
	}

	return cmd
}

