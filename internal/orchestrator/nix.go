package orchestrator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// NixExecutor handles nix command execution
type NixExecutor struct {
	flakeRef string
}

// NewNixExecutor creates a new nix executor
func NewNixExecutor(flakeRef string) *NixExecutor {
	return &NixExecutor{flakeRef: flakeRef}
}

// BuildTerranix builds a terranix configuration for a project
// Returns the path to the generated JSON file
func (n *NixExecutor) BuildTerranix(project string) (string, error) {
	// Build the nix command: nix build <flake>#terranix.<project> --no-link --print-out-paths
	flakeAttr := fmt.Sprintf("%s#terranix.%s", n.flakeRef, project)

	cmd := exec.Command("nix", "build", flakeAttr, "--no-link", "--print-out-paths")
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("nix build failed: %w", err)
	}

	// The output is the path to the built derivation
	outputPath := strings.TrimSpace(string(output))

	// The terranix output should be a JSON file
	// Check if it's a directory or file
	info, err := os.Stat(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to stat nix output: %w", err)
	}

	if info.IsDir() {
		// If it's a directory, look for a JSON file inside
		jsonPath := filepath.Join(outputPath, "config.tf.json")
		if _, err := os.Stat(jsonPath); err == nil {
			return jsonPath, nil
		}
		// Try other common names
		jsonPath = filepath.Join(outputPath, "terraform.tf.json")
		if _, err := os.Stat(jsonPath); err == nil {
			return jsonPath, nil
		}
		return "", fmt.Errorf("no JSON file found in nix output directory")
	}

	// If it's a file, assume it's the JSON
	return outputPath, nil
}

// CopyTerranixConfig copies the generated terranix config to the project directory
func (n *NixExecutor) CopyTerranixConfig(sourcePath, projectDir string) (string, error) {
	targetPath := filepath.Join(projectDir, "config.tf.json")

	// Read source file
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to read terranix config: %w", err)
	}

	// Write to target
	if err := os.WriteFile(targetPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write terranix config: %w", err)
	}

	return targetPath, nil
}

