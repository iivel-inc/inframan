package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/iivel-inc/inframan/internal/state"
	"github.com/spf13/cobra"
)

// NewStatusCommand creates the status command
func NewStatusCommand() *cobra.Command {
	var projectDir string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current state",
		Long:  "Display the current state of the inframan project",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectDir, _ = cmd.Root().PersistentFlags().GetString("dir")
			if projectDir == "" {
				var err error
				projectDir, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get working directory: %w", err)
				}
			}

			stateManager, err := state.NewManager(projectDir)
			if err != nil {
				return fmt.Errorf("failed to initialize state manager: %w", err)
			}

			currentState := stateManager.GetState()

			if jsonOutput {
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(currentState)
			}

			// Human-readable output
			fmt.Println("Inframan Status")
			fmt.Println("==============")
			fmt.Printf("Project: %s\n", currentState.Project)

			if !currentState.LastApplied.IsZero() {
				fmt.Printf("Last Applied: %s\n", currentState.LastApplied.Format(time.RFC3339))
			} else {
				fmt.Println("Last Applied: Never")
			}

			if currentState.TerranixConfig != "" {
				fmt.Printf("Terranix Config: %s\n", currentState.TerranixConfig)
			}

			if currentState.TerraformState != "" {
				fmt.Printf("Terraform State: %s\n", currentState.TerraformState)
			}

			fmt.Printf("Colmena Applied: %v\n", currentState.ColmenaApplied)

			if len(currentState.Workflow) > 0 {
				fmt.Println("\nWorkflow Steps:")
				for step, stepStatus := range currentState.Workflow {
					fmt.Printf("  %s: %s", step, stepStatus.Status)
					if !stepStatus.Timestamp.IsZero() {
						fmt.Printf(" (%s)", stepStatus.Timestamp.Format(time.RFC3339))
					}
					if stepStatus.Message != "" {
						fmt.Printf(" - %s", stepStatus.Message)
					}
					if stepStatus.Error != "" {
						fmt.Printf("\n    Error: %s", stepStatus.Error)
					}
					fmt.Println()
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output status as JSON")

	return cmd
}

