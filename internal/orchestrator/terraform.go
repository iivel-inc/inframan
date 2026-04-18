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

// IsInitialized checks if terraform has been initialized in the workdir
func (t *TerraformExecutor) IsInitialized() bool {
	terraformDir := filepath.Join(t.workDir, ".terraform")
	_, err := os.Stat(terraformDir)
	return err == nil
}

// EnsureInit runs terraform init if not already initialized
// This is useful for commands that need terraform state (like output, destroy)
// but may be run in CI environments with remote backends where state isn't checked in
func (t *TerraformExecutor) EnsureInit() error {
	if t.IsInitialized() {
		return nil
	}
	fmt.Println("Initializing Terraform...")
	return t.Init()
}

// ensureInitInDir ensures terraform is initialized in the specified directory
// This is a helper for standalone functions that don't use TerraformExecutor
func ensureInitInDir(terraformDir string) error {
	dotTerraformDir := filepath.Join(terraformDir, ".terraform")
	if _, err := os.Stat(dotTerraformDir); err == nil {
		// Already initialized
		return nil
	}

	fmt.Printf("Initializing Terraform in %s...\n", terraformDir)
	cmd := exec.Command("terraform", "init")
	cmd.Dir = terraformDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("terraform init failed: %w", err)
	}
	return nil
}

// Apply runs terraform apply. If autoApprove is true, passes -auto-approve
// to skip the interactive confirmation (useful for CI).
func (t *TerraformExecutor) Apply(autoApprove bool) error {
	args := []string{"apply"}
	if autoApprove {
		args = append(args, "-auto-approve")
	}
	cmd := exec.Command("terraform", args...)
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

// Output runs terraform output with any additional args passed through
func (t *TerraformExecutor) Output(args ...string) error {
	cmdArgs := append([]string{"output"}, args...)
	cmd := exec.Command("terraform", cmdArgs...)
	cmd.Dir = t.workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("terraform output failed: %w", err)
	}

	return nil
}

// Import runs terraform import to bring an existing resource into state
func (t *TerraformExecutor) Import(address, id string) error {
	cmd := exec.Command("terraform", "import", address, id)
	cmd.Dir = t.workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("terraform import failed: %w", err)
	}

	return nil
}

// Destroy runs terraform destroy. If autoApprove is true, passes -auto-approve
// to skip the interactive confirmation (useful for CI).
func (t *TerraformExecutor) Destroy(autoApprove bool) error {
	args := []string{"destroy"}
	if autoApprove {
		args = append(args, "-auto-approve")
	}
	cmd := exec.Command("terraform", args...)
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
// Supports both single instance (public_ip/private_ip) and multiple instances (instances map)
type TerraformOutput struct {
	// Legacy single instance output
	PublicIP struct {
		Value string `json:"value"`
	} `json:"public_ip"`

	// Private IP for instances behind a bastion/proxy jump
	PrivateIP struct {
		Value string `json:"value"`
	} `json:"private_ip"`

	// Multiple named instances output: { "web-1": "1.2.3.4", "db-1": "5.6.7.8" }
	Instances struct {
		Value map[string]string `json:"value"`
	} `json:"instances"`
}

// GetTargetIP retrieves the public IP from terraform output
func (t *TerraformExecutor) GetTargetIP() (string, error) {
	// Ensure terraform is initialized (needed for remote backends in CI)
	if err := t.EnsureInit(); err != nil {
		return "", fmt.Errorf("failed to initialize terraform: %w", err)
	}

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

	// Prefer private_ip when explicitly requested or when proxy jump is configured
	if GetUsePrivateIP() || GetSSHProxyJump() != "" {
		if terraformOutput.PrivateIP.Value != "" {
			return terraformOutput.PrivateIP.Value, nil
		}
	}

	if terraformOutput.PublicIP.Value != "" {
		return terraformOutput.PublicIP.Value, nil
	}

	return "", fmt.Errorf("no IP found in terraform output (expected 'public_ip' or 'private_ip' with USE_PRIVATE_IP or proxy jump configured)")
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

	// Ensure terraform is initialized (needed for remote backends in CI)
	if err := ensureInitInDir(terraformDir); err != nil {
		return nil, fmt.Errorf("failed to initialize terraform for project %q: %w", projectName, err)
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

	// Fall back to legacy single instance
	// Prefer private_ip when proxy jump is configured (target is behind bastion)
	ip := terraformOutput.PublicIP.Value
	if GetSSHProxyJump() != "" && terraformOutput.PrivateIP.Value != "" {
		ip = terraformOutput.PrivateIP.Value
	}

	if ip != "" {
		instances = append(instances, &InstanceInfo{
			ProjectName:  projectName,
			InstanceName: "", // Empty for single instance
			PublicIP:     ip,
		})
		return instances, nil
	}

	return nil, fmt.Errorf("no instances found in terraform output for project %q (expected 'instances' map, 'public_ip', or 'private_ip' with proxy jump)", projectName)
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
