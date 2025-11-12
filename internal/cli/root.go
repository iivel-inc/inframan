package cli

import (
	"fmt"
	"os"

	"github.com/iivel-inc/inframan/internal/commands"
	"github.com/spf13/cobra"
)

var (
	flakeRef   string
	projectDir string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "inframan",
	Short: "Orchestrator for nix, terranix, terraform, and colmena",
	Long: `Inframan is a CLI tool that orchestrates DevOps workflows by managing
the execution of nix builds, terranix configuration generation, terraform
operations, and colmena deployments.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Set default project directory to current working directory if not set
		if projectDir == "" {
			var err error
			projectDir, err = os.Getwd()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to get working directory: %v\n", err)
				os.Exit(1)
			}
		}

		// Set default flake reference if not set
		if flakeRef == "" {
			flakeRef = "."
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flakeRef, "flake", "", "Nix flake reference (default: current directory)")
	rootCmd.PersistentFlags().StringVar(&projectDir, "dir", "", "Project directory (default: current working directory)")

	// Add subcommands
	rootCmd.AddCommand(commands.NewApplyCommand())
	rootCmd.AddCommand(commands.NewDestroyCommand())
	rootCmd.AddCommand(commands.NewStatusCommand())
	rootCmd.AddCommand(commands.NewUICommand())
}

