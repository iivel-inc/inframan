package orchestrator

import (
	"fmt"
	"os"
	"os/exec"
)

// ColmenaExecutor handles colmena command execution
type ColmenaExecutor struct {
	workDir string
}

// NewColmenaExecutor creates a new colmena executor
func NewColmenaExecutor(workDir string) *ColmenaExecutor {
	return &ColmenaExecutor{workDir: workDir}
}

// Apply runs colmena apply for a specific project
func (c *ColmenaExecutor) Apply(project string) error {
	tag := fmt.Sprintf("@project-%s", project)

	cmd := exec.Command("nix", "run", ".", "#colmena", "apply", "--", "--on", tag, "--impure")
	cmd.Dir = c.workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("colmena apply failed: %w", err)
	}

	return nil
}

// Destroy runs colmena destroy for a specific project
func (c *ColmenaExecutor) Destroy(project string) error {
	tag := fmt.Sprintf("@project-%s", project)

	cmd := exec.Command("nix", "run", ".", "#colmena", "destroy", "--", "--on", tag, "--impure")
	cmd.Dir = c.workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("colmena destroy failed: %w", err)
	}

	return nil
}

