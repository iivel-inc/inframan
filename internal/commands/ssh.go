package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
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
		Use:   "ssh [project[/instance]]",
		Short: "SSH to an instance by project name",
		Long: `SSH connects to a provisioned instance using its project and instance name.

For single-instance projects, use just the project name.
For multi-instance projects, use project/instance-name syntax.

Examples:
  # List all available instances
  inframan ssh --list

  # Connect to a single-instance project
  inframan ssh account1

  # Connect to a specific instance in a multi-instance project
  inframan ssh production/web-1
  inframan ssh production/db-1

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

			target := args[0]
			return connectToInstance(target, user, identityFile)
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
		fmt.Printf("  %-30s %s\n", inst.FullName(), inst.PublicIP)
	}
	fmt.Println()
	fmt.Println("Connect with: inframan ssh <project[/instance]>")

	return nil
}

// parseTarget parses a target string into project and instance name
// Examples: "account1" -> ("account1", ""), "production/web-1" -> ("production", "web-1")
func parseTarget(target string) (projectName, instanceName string) {
	parts := strings.SplitN(target, "/", 2)
	projectName = parts[0]
	if len(parts) > 1 {
		instanceName = parts[1]
	}
	return
}

// connectToInstance establishes an SSH connection to the specified instance
func connectToInstance(target, user, identityFile string) error {
	// Parse target into project and instance name
	projectName, instanceName := parseTarget(target)

	// Get instance info
	info, err := orchestrator.GetInstance(projectName, instanceName)
	if err != nil {
		return fmt.Errorf("failed to get instance info: %w", err)
	}

	fmt.Printf("Connecting to %s (%s) as %s...\n", info.FullName(), info.PublicIP, user)

	// Build SSH command arguments
	sshArgs := []string{"ssh"}

	// Add identity file if specified, or fall back to SSH_KEY_PATH env var
	if identityFile != "" {
		sshArgs = append(sshArgs, "-i", identityFile)
	} else if sshKeyPath := orchestrator.GetSSHKeyPath(); sshKeyPath != "" {
		sshArgs = append(sshArgs, "-i", sshKeyPath)
	}

	// Add common SSH options for convenience
	sshArgs = append(sshArgs,
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=ERROR",
	)

	// Add target
	sshTarget := fmt.Sprintf("%s@%s", user, info.PublicIP)
	sshArgs = append(sshArgs, sshTarget)

	// Find ssh binary
	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		return fmt.Errorf("ssh not found in PATH: %w", err)
	}

	// Replace the current process with ssh (exec)
	// This gives full terminal control to ssh
	return syscall.Exec(sshPath, sshArgs, os.Environ())
}
