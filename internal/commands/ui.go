package commands

import (
	"fmt"
	"os"

	"github.com/iivel-inc/inframan/internal/ui"
	"github.com/spf13/cobra"
)

// NewUICommand creates the UI command
func NewUICommand() *cobra.Command {
	var projectDir string

	cmd := &cobra.Command{
		Use:   "ui",
		Short: "Launch TUI mode",
		Long:  "Launch the terminal user interface for managing inframan workflows",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectDir, _ = cmd.Root().PersistentFlags().GetString("dir")
			if projectDir == "" {
				var err error
				projectDir, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get working directory: %w", err)
				}
			}

			program := ui.NewProgram(projectDir)
			if _, err := program.Run(); err != nil {
				return fmt.Errorf("TUI error: %w", err)
			}

			return nil
		},
	}

	return cmd
}

