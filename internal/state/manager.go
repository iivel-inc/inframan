package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const defaultStateDir = ".inframan"
const defaultStateFile = "state.json"

// Manager handles state persistence
type Manager struct {
	statePath string
	state     *State
}

// NewManager creates a new state manager
func NewManager(projectRoot string) (*Manager, error) {
	stateDir := filepath.Join(projectRoot, defaultStateDir)
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create state directory: %w", err)
	}

	statePath := filepath.Join(stateDir, defaultStateFile)
	manager := &Manager{
		statePath: statePath,
		state:     &State{Workflow: make(map[string]StepStatus)},
	}

	// Load existing state if it exists
	if err := manager.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	return manager, nil
}

// Load reads the state from disk
func (m *Manager) Load() error {
	data, err := os.ReadFile(m.statePath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, m.state); err != nil {
		return fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return nil
}

// Save writes the state to disk
func (m *Manager) Save() error {
	data, err := json.MarshalIndent(m.state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(m.statePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// GetState returns the current state
func (m *Manager) GetState() *State {
	return m.state
}

// SetProject sets the project name
func (m *Manager) SetProject(project string) {
	m.state.Project = project
}

// UpdateStepStatus updates the status of a workflow step
func (m *Manager) UpdateStepStatus(step, status, message string, err error) {
	if m.state.Workflow == nil {
		m.state.Workflow = make(map[string]StepStatus)
	}

	stepStatus := StepStatus{
		Status:    status,
		Timestamp: time.Now(),
		Message:   message,
	}

	if err != nil {
		stepStatus.Error = err.Error()
	}

	m.state.Workflow[step] = stepStatus
}

// SetLastApplied updates the last applied timestamp
func (m *Manager) SetLastApplied() {
	m.state.LastApplied = time.Now()
}

// SetTerranixConfig sets the terranix config path
func (m *Manager) SetTerranixConfig(path string) {
	m.state.TerranixConfig = path
}

// SetTerraformState sets the terraform state path
func (m *Manager) SetTerraformState(path string) {
	m.state.TerraformState = path
}

// SetColmenaApplied sets whether colmena has been applied
func (m *Manager) SetColmenaApplied(applied bool) {
	m.state.ColmenaApplied = applied
}

