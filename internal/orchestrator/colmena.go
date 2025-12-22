package orchestrator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ColmenaExecutor handles colmena command execution
type ColmenaExecutor struct {
	workDir string
}

// NewColmenaExecutor creates a new colmena executor
func NewColmenaExecutor() (*ColmenaExecutor, error) {
	workDir, err := GetColmenaDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get colmena directory: %w", err)
	}

	// Ensure the directory exists
	if err := EnsureDir(workDir); err != nil {
		return nil, err
	}

	return &ColmenaExecutor{workDir: workDir}, nil
}

// hiveTemplate is the template for generating a dynamic hive.nix
const hiveTemplate = `{
  meta = {
    nixpkgs = import <nixpkgs> { system = "x86_64-linux"; };
  };

  # Define the node
  target-node = { ... }: {
    imports = [ (import %s) ]; # Import the user's module
    deployment.targetHost = "%s"; # Injected IP
    deployment.targetUser = "root";
    deployment.buildOnTarget = true; # Build on remote instance, not locally
  };
}
`

// GenerateHive creates an ephemeral hive.nix with the target IP injected
func (c *ColmenaExecutor) GenerateHive(modulePath, targetIP string) (string, error) {
	// Ensure workdir exists
	if err := os.MkdirAll(c.workDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create workdir: %w", err)
	}

	// Convert module path to absolute path for Nix
	absModulePath, err := filepath.Abs(modulePath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Generate the hive content
	// Escape the path for Nix (wrap in quotes)
	nixPath := fmt.Sprintf("\"%s\"", absModulePath)
	hiveContent := fmt.Sprintf(hiveTemplate, nixPath, targetIP)

	// Write to hive.nix
	hivePath := filepath.Join(c.workDir, HiveFileName)
	if err := os.WriteFile(hivePath, []byte(hiveContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write hive.nix: %w", err)
	}

	return hivePath, nil
}

// Apply runs colmena apply with the generated hive
func (c *ColmenaExecutor) Apply(hivePath string) error {
	args := []string{"apply", "--on", "target-node", "-f", hivePath}

	// Add SSH key option if SSH_KEY_PATH is set
	if sshKeyPath := GetSSHKeyPath(); sshKeyPath != "" {
		args = append(args, "--ssh-option", fmt.Sprintf("IdentityFile=%s", sshKeyPath))
		// Also disable strict host key checking for new hosts
		args = append(args, "--ssh-option", "StrictHostKeyChecking=accept-new")
	}

	cmd := exec.Command("colmena", args...)
	cmd.Dir = c.workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("colmena apply failed: %w", err)
	}

	return nil
}

// ApplyWithTag runs colmena apply for a specific tag (legacy support)
func (c *ColmenaExecutor) ApplyWithTag(project string) error {
	tag := fmt.Sprintf("@project-%s", project)

	cmd := exec.Command("colmena", "apply", "--on", tag)
	cmd.Dir = c.workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("colmena apply failed: %w", err)
	}

	return nil
}

// Destroy runs colmena reboot (colmena doesn't have destroy, this is a placeholder)
func (c *ColmenaExecutor) Destroy(hivePath string) error {
	// Note: Colmena doesn't have a destroy command
	// Infrastructure destruction should be done via terraform destroy
	return fmt.Errorf("colmena destroy is not supported; use 'terraform destroy' in .inframan/terraform instead")
}

// GetHivePath returns the path to the generated hive.nix
func (c *ColmenaExecutor) GetHivePath() string {
	return filepath.Join(c.workDir, HiveFileName)
}

// ValidateHive checks if the hive.nix is valid by running colmena eval
func (c *ColmenaExecutor) ValidateHive(hivePath string) error {
	cmd := exec.Command("colmena", "eval", "-f", hivePath, "-E", "{ nodes, ... }: nodes")
	cmd.Dir = c.workDir
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("hive validation failed: %w\n%s", err, strings.TrimSpace(string(output)))
	}

	return nil
}
