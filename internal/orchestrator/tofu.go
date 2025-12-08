package orchestrator

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const defaultWorkdir = ".runner-workdir"
const configFileName = "config.tf.json"

// TofuExecutor handles OpenTofu command execution
type TofuExecutor struct {
	workDir string
}

// NewTofuExecutor creates a new OpenTofu executor
func NewTofuExecutor() (*TofuExecutor, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	workDir := filepath.Join(cwd, defaultWorkdir)
	return &TofuExecutor{workDir: workDir}, nil
}

// SetupWorkdir creates the workdir and copies the config file
func (t *TofuExecutor) SetupWorkdir(configPath string) error {
	// Create workdir if it doesn't exist
	if err := os.MkdirAll(t.workDir, 0755); err != nil {
		return fmt.Errorf("failed to create workdir: %w", err)
	}

	// Read the source config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Write to workdir as config.tf.json
	targetPath := filepath.Join(t.workDir, configFileName)
	if err := os.WriteFile(targetPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Init runs tofu init
func (t *TofuExecutor) Init() error {
	cmd := exec.Command("tofu", "init")
	cmd.Dir = t.workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	// Pass through environment (includes AWS credentials)
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tofu init failed: %w", err)
	}

	return nil
}

// Apply runs tofu apply
func (t *TofuExecutor) Apply() error {
	cmd := exec.Command("tofu", "apply")
	cmd.Dir = t.workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	// Pass through environment (includes AWS credentials)
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tofu apply failed: %w", err)
	}

	return nil
}

// Destroy runs tofu destroy
func (t *TofuExecutor) Destroy() error {
	cmd := exec.Command("tofu", "destroy")
	cmd.Dir = t.workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tofu destroy failed: %w", err)
	}

	return nil
}

// TofuOutput represents the structure of tofu output -json
type TofuOutput struct {
	PublicIP struct {
		Value string `json:"value"`
	} `json:"public_ip"`
}

// GetTargetIP retrieves the public IP from tofu output
func (t *TofuExecutor) GetTargetIP() (string, error) {
	cmd := exec.Command("tofu", "output", "-json")
	cmd.Dir = t.workDir
	cmd.Env = os.Environ()

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("tofu output failed: %w", err)
	}

	var tofuOutput TofuOutput
	if err := json.Unmarshal(output, &tofuOutput); err != nil {
		return "", fmt.Errorf("failed to parse tofu output: %w", err)
	}

	if tofuOutput.PublicIP.Value == "" {
		return "", fmt.Errorf("public_ip not found in tofu output")
	}

	return tofuOutput.PublicIP.Value, nil
}

// GetWorkDir returns the workdir path
func (t *TofuExecutor) GetWorkDir() string {
	return t.workDir
}

