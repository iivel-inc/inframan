package orchestrator

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
// Supports both single instance (public_ip) and multiple instances (instances map)
type TerraformOutput struct {
	// Legacy single instance output
	PublicIP struct {
		Value string `json:"value"`
	} `json:"public_ip"`

	// Multiple named instances output: { "web-1": "1.2.3.4", "db-1": "5.6.7.8" }
	Instances struct {
		Value map[string]string `json:"value"`
	} `json:"instances"`
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
	ProjectName  string
	InstanceName string // Empty for single-instance projects (legacy public_ip)
	PublicIP     string
}

// FullName returns the full identifier for the instance (project/instance or just project)
func (i *InstanceInfo) FullName() string {
	if i.InstanceName == "" {
		return i.ProjectName
	}
	return fmt.Sprintf("%s/%s", i.ProjectName, i.InstanceName)
}

// GetInstancesForProject retrieves all instances for a specific project
func GetInstancesForProject(projectName string) ([]*InstanceInfo, error) {
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

	var instances []*InstanceInfo

	// Check for multiple instances first (instances map)
	if len(terraformOutput.Instances.Value) > 0 {
		for name, ip := range terraformOutput.Instances.Value {
			instances = append(instances, &InstanceInfo{
				ProjectName:  projectName,
				InstanceName: name,
				PublicIP:     ip,
			})
		}
		return instances, nil
	}

	// Fall back to legacy single instance (public_ip)
	if terraformOutput.PublicIP.Value != "" {
		instances = append(instances, &InstanceInfo{
			ProjectName:  projectName,
			InstanceName: "", // Empty for single instance
			PublicIP:     terraformOutput.PublicIP.Value,
		})
		return instances, nil
	}

	return nil, fmt.Errorf("no instances found in terraform output for project %q (expected 'instances' map or 'public_ip')", projectName)
}

// GetInstance retrieves a specific instance by project and optional instance name
func GetInstance(projectName, instanceName string) (*InstanceInfo, error) {
	instances, err := GetInstancesForProject(projectName)
	if err != nil {
		return nil, err
	}

	// If no instance name specified
	if instanceName == "" {
		if len(instances) == 1 {
			return instances[0], nil
		}
		return nil, fmt.Errorf("project %q has %d instances, specify one: %s", projectName, len(instances), formatInstanceNames(instances))
	}

	// Find the specific instance
	for _, inst := range instances {
		if inst.InstanceName == instanceName {
			return inst, nil
		}
	}

	return nil, fmt.Errorf("instance %q not found in project %q, available: %s", instanceName, projectName, formatInstanceNames(instances))
}

// formatInstanceNames returns a comma-separated list of instance names
func formatInstanceNames(instances []*InstanceInfo) string {
	names := make([]string, len(instances))
	for i, inst := range instances {
		if inst.InstanceName == "" {
			names[i] = "(default)"
		} else {
			names[i] = inst.InstanceName
		}
	}
	return fmt.Sprintf("[%s]", strings.Join(names, ", "))
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

	var allInstances []*InstanceInfo
	for _, project := range projects {
		instances, err := GetInstancesForProject(project)
		if err != nil {
			// Skip projects with errors
			continue
		}
		allInstances = append(allInstances, instances...)
	}

	return allInstances, nil
}
