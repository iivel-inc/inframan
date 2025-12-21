package commands

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/iivel-inc/inframan/internal/orchestrator"
	"github.com/spf13/cobra"
)

// NewSSHCommand creates the ssh command
func NewSSHCommand() *cobra.Command {
	var user string
	var identityFile string
	var listInstances bool

	cmd := &cobra.Command{
		Use:   "ssh [instance-name]",
		Short: "SSH to an instance by project name",
		Long: `SSH connects to a provisioned instance using its project name.

The instance name corresponds to the project name used during provisioning
(e.g., account1, account2, or "default" if PROJECT_NAME was not set).

Examples:
  # List all available instances
  inframan ssh --list

  # Connect to an instance
  inframan ssh account1

  # Connect with a specific user
  inframan ssh account1 --user nixos

  # Connect with a specific identity file
  inframan ssh account1 --identity ~/.ssh/id_ed25519`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Handle --list flag
			if listInstances {
				return listAllInstances()
			}

			// If no arguments, show available instances and prompt
			if len(args) == 0 {
				return listAllInstances()
			}

			instanceName := args[0]
			return connectToInstance(instanceName, user, identityFile)
		},
	}

	cmd.Flags().StringVarP(&user, "user", "u", "root", "SSH user")
	cmd.Flags().StringVarP(&identityFile, "identity", "i", "", "Path to SSH identity file")
	cmd.Flags().BoolVarP(&listInstances, "list", "l", false, "List all available instances")

	return cmd
}

// listAllInstances displays all available instances
func listAllInstances() error {
	instances, err := orchestrator.GetAllInstances()
	if err != nil {
		return fmt.Errorf("failed to get instances: %w", err)
	}

	if len(instances) == 0 {
		fmt.Println("No instances found.")
		fmt.Println("Run 'inframan infra' to provision infrastructure first.")
		return nil
	}

	fmt.Println("Available instances:")
	fmt.Println()
	for _, inst := range instances {
		fmt.Printf("  %-20s %s\n", inst.ProjectName, inst.PublicIP)
	}
	fmt.Println()
	fmt.Println("Connect with: inframan ssh <instance-name>")

	return nil
}

// connectToInstance establishes an SSH connection to the specified instance
func connectToInstance(instanceName, user, identityFile string) error {
	// Get instance info
	info, err := orchestrator.GetOutputForProject(instanceName)
	if err != nil {
		return fmt.Errorf("failed to get instance info: %w", err)
	}

	fmt.Printf("Connecting to %s (%s) as %s...\n", instanceName, info.PublicIP, user)

	// Build SSH command arguments
	sshArgs := []string{"ssh"}

	// Add identity file if specified
	if identityFile != "" {
		sshArgs = append(sshArgs, "-i", identityFile)
	}

	// Add common SSH options for convenience
	sshArgs = append(sshArgs,
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=ERROR",
	)

	// Add target
	target := fmt.Sprintf("%s@%s", user, info.PublicIP)
	sshArgs = append(sshArgs, target)

	// Find ssh binary
	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		return fmt.Errorf("ssh not found in PATH: %w", err)
	}

	// Replace the current process with ssh (exec)
	// This gives full terminal control to ssh
	return syscall.Exec(sshPath, sshArgs, os.Environ())
}

