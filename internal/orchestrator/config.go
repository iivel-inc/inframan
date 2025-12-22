package orchestrator

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	// InframanDir is the root directory for all inframan-generated files
	InframanDir = ".inframan"

	// TerraformSubdir is the subdirectory for terraform state and config
	TerraformSubdir = "terraform"

	// ColmenaSubdir is the subdirectory for colmena hive files
	ColmenaSubdir = "colmena"

	// ConfigFileName is the name of the terraform config file
	ConfigFileName = "config.tf.json"

	// HiveFileName is the name of the colmena hive file
	HiveFileName = "hive.nix"

	// DefaultProjectName is used when PROJECT_NAME is not set
	DefaultProjectName = "default"
)

// GetProjectName returns the project name from environment or default
func GetProjectName() string {
	projectName := os.Getenv("PROJECT_NAME")
	if projectName == "" {
		return DefaultProjectName
	}
	return projectName
}

// GetSSHKeyPath returns the SSH key path from environment, or empty string if not set
func GetSSHKeyPath() string {
	return os.Getenv("SSH_KEY_PATH")
}

// GetInframanDir returns the absolute path to the .inframan directory
func GetInframanDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}
	return filepath.Join(cwd, InframanDir), nil
}

// GetProjectDir returns the absolute path to the project-specific directory
// Structure: .inframan/<project-name>/
func GetProjectDir() (string, error) {
	inframanDir, err := GetInframanDir()
	if err != nil {
		return "", err
	}
	projectName := GetProjectName()
	return filepath.Join(inframanDir, projectName), nil
}

// GetTerraformDir returns the absolute path to the terraform subdirectory
// Structure: .inframan/<project-name>/terraform/
func GetTerraformDir() (string, error) {
	projectDir, err := GetProjectDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(projectDir, TerraformSubdir), nil
}

// GetColmenaDir returns the absolute path to the colmena subdirectory
// Structure: .inframan/<project-name>/colmena/
func GetColmenaDir() (string, error) {
	projectDir, err := GetProjectDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(projectDir, ColmenaSubdir), nil
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}
	return nil
}

// InitInframanDirs creates the .inframan directory structure
func InitInframanDirs() error {
	terraformDir, err := GetTerraformDir()
	if err != nil {
		return err
	}
	if err := EnsureDir(terraformDir); err != nil {
		return err
	}

	colmenaDir, err := GetColmenaDir()
	if err != nil {
		return err
	}
	if err := EnsureDir(colmenaDir); err != nil {
		return err
	}

	return nil
}

// GetAllProjectDirs returns all project directories under .inframan/
// Each project directory is expected to contain a terraform/ subdirectory
func GetAllProjectDirs() ([]string, error) {
	inframanDir, err := GetInframanDir()
	if err != nil {
		return nil, err
	}

	// Check if .inframan directory exists
	if _, err := os.Stat(inframanDir); os.IsNotExist(err) {
		return nil, nil // No projects yet
	}

	entries, err := os.ReadDir(inframanDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read inframan directory: %w", err)
	}

	var projects []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if this project has been initialized (works with any backend type)
		// We check for .terraform/ directory (created by terraform init) or config.tf.json
		terraformDir := filepath.Join(inframanDir, entry.Name(), TerraformSubdir)
		terraformInitDir := filepath.Join(terraformDir, ".terraform")
		configPath := filepath.Join(terraformDir, ConfigFileName)

		// Project is valid if terraform has been initialized OR config exists
		_, initErr := os.Stat(terraformInitDir)
		_, configErr := os.Stat(configPath)

		if initErr == nil || configErr == nil {
			projects = append(projects, entry.Name())
		}
	}

	return projects, nil
}

// GetTerraformDirForProject returns the terraform directory for a specific project
func GetTerraformDirForProject(projectName string) (string, error) {
	inframanDir, err := GetInframanDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(inframanDir, projectName, TerraformSubdir), nil
}
