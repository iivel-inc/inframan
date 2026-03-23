package cli

import (
	"github.com/iivel-inc/inframan/internal/commands"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "inframan",
	Short: "Nix-Go-GitOps Runner for Terranix and Colmena",
	Long: `Inframan is a CLI tool that bridges Terranix (Infrastructure as Code)
and Colmena (NixOS Deployment).

Environment Variables:
  INFRA_CONFIG_JSON  - Path to the Terranix-generated JSON file
  NIXOS_MODULE_PATH  - Path to the NixOS configuration module
  PROJECT_NAME       - Project name for organizing .inframan/<project>/ folders (default: "default")

Commands:
  init    - Initialize terraform workspace and pull remote state
  infra   - Build and apply infrastructure using Terraform
  deploy  - Deploy NixOS configuration using Colmena
  destroy - Destroy infrastructure using Terraform
  ssh     - SSH to an instance by project name
  import  - Import existing resources into Terraform state
  output  - Show Terraform output values`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(commands.NewInitCommand())
	rootCmd.AddCommand(commands.NewInfraCommand())
	rootCmd.AddCommand(commands.NewDeployCommand())
	rootCmd.AddCommand(commands.NewDestroyCommand())
	rootCmd.AddCommand(commands.NewSSHCommand())
	rootCmd.AddCommand(commands.NewImportCommand())
	rootCmd.AddCommand(commands.NewOutputCommand())
}
