package orchestrator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// TerraformExecutor handles terraform command execution
type TerraformExecutor struct {
	workDir string
}

// NewTerraformExecutor creates a new terraform executor
func NewTerraformExecutor(workDir string) *TerraformExecutor {
	return &TerraformExecutor{workDir: workDir}
}

// Init runs terraform init
func (t *TerraformExecutor) Init() error {
	cmd := exec.Command("terraform", "init")
	cmd.Dir = t.workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("terraform init failed: %w", err)
	}

	return nil
}

// Apply runs terraform apply
func (t *TerraformExecutor) Apply() error {
	cmd := exec.Command("terraform", "apply", "-auto-approve")
	cmd.Dir = t.workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("terraform apply failed: %w", err)
	}

	return nil
}

// Destroy runs terraform destroy
func (t *TerraformExecutor) Destroy() error {
	cmd := exec.Command("terraform", "destroy", "-auto-approve")
	cmd.Dir = t.workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("terraform destroy failed: %w", err)
	}

	return nil
}

// Output saves terraform output to tf-output.json
func (t *TerraformExecutor) Output() (string, error) {
	outputPath := filepath.Join(t.workDir, "tf-output.json")

	cmd := exec.Command("terraform", "output", "-json")
	cmd.Dir = t.workDir

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("terraform output failed: %w", err)
	}

	if err := os.WriteFile(outputPath, output, 0644); err != nil {
		return "", fmt.Errorf("failed to write terraform output: %w", err)
	}

	return outputPath, nil
}

