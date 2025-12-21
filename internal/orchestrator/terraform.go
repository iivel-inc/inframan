package orchestrator

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// TerraformExecutor handles Terraform command execution
type TerraformExecutor struct {
	workDir string
}

// NewTerraformExecutor creates a new Terraform executor
func NewTerraformExecutor() (*TerraformExecutor, error) {
	workDir, err := GetTerraformDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get terraform directory: %w", err)
	}

	// Ensure the directory exists
	if err := EnsureDir(workDir); err != nil {
		return nil, err
	}

	return &TerraformExecutor{workDir: workDir}, nil
}

// SetupWorkdir creates the workdir and copies the config file
func (t *TerraformExecutor) SetupWorkdir(configPath string) error {
	// Read the source config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Write to workdir as config.tf.json
	targetPath := filepath.Join(t.workDir, ConfigFileName)
	if err := os.WriteFile(targetPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Init runs terraform init
func (t *TerraformExecutor) Init() error {
	cmd := exec.Command("terraform", "init")
	cmd.Dir = t.workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	// Pass through environment (includes AWS credentials)
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("terraform init failed: %w", err)
	}

	return nil
}

// Apply runs terraform apply
func (t *TerraformExecutor) Apply() error {
	cmd := exec.Command("terraform", "apply")
	cmd.Dir = t.workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	// Pass through environment (includes AWS credentials)
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("terraform apply failed: %w", err)
	}

	return nil
}

// Destroy runs terraform destroy
func (t *TerraformExecutor) Destroy() error {
	cmd := exec.Command("terraform", "destroy")
	cmd.Dir = t.workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("terraform destroy failed: %w", err)
	}

	return nil
}

// TerraformOutput represents the structure of terraform output -json
type TerraformOutput struct {
	PublicIP struct {
		Value string `json:"value"`
	} `json:"public_ip"`
}

// GetTargetIP retrieves the public IP from terraform output
func (t *TerraformExecutor) GetTargetIP() (string, error) {
	cmd := exec.Command("terraform", "output", "-json")
	cmd.Dir = t.workDir
	cmd.Env = os.Environ()

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("terraform output failed: %w", err)
	}

	var terraformOutput TerraformOutput
	if err := json.Unmarshal(output, &terraformOutput); err != nil {
		return "", fmt.Errorf("failed to parse terraform output: %w", err)
	}

	if terraformOutput.PublicIP.Value == "" {
		return "", fmt.Errorf("public_ip not found in terraform output")
	}

	return terraformOutput.PublicIP.Value, nil
}

// GetWorkDir returns the workdir path
func (t *TerraformExecutor) GetWorkDir() string {
	return t.workDir
}

// InstanceInfo contains information about a provisioned instance
type InstanceInfo struct {
	ProjectName string
	PublicIP    string
}

// GetOutputForProject retrieves terraform output for a specific project
func GetOutputForProject(projectName string) (*InstanceInfo, error) {
	terraformDir, err := GetTerraformDirForProject(projectName)
	if err != nil {
		return nil, fmt.Errorf("failed to get terraform directory: %w", err)
	}

	// Check if terraform directory exists
	if _, err := os.Stat(terraformDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("project %q does not exist", projectName)
	}

	cmd := exec.Command("terraform", "output", "-json")
	cmd.Dir = terraformDir
	cmd.Env = os.Environ()

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("terraform output failed for project %q: %w", projectName, err)
	}

	var terraformOutput TerraformOutput
	if err := json.Unmarshal(output, &terraformOutput); err != nil {
		return nil, fmt.Errorf("failed to parse terraform output: %w", err)
	}

	if terraformOutput.PublicIP.Value == "" {
		return nil, fmt.Errorf("public_ip not found in terraform output for project %q", projectName)
	}

	return &InstanceInfo{
		ProjectName: projectName,
		PublicIP:    terraformOutput.PublicIP.Value,
	}, nil
}

// GetAllInstances returns instance info for all projects
func GetAllInstances() ([]*InstanceInfo, error) {
	projects, err := GetAllProjectDirs()
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	if len(projects) == 0 {
		return nil, nil
	}

	var instances []*InstanceInfo
	for _, project := range projects {
		info, err := GetOutputForProject(project)
		if err != nil {
			// Skip projects with errors (might not have public_ip output)
			continue
		}
		instances = append(instances, info)
	}

	return instances, nil
}
