package state

import "time"

// State represents the overall state of an inframan project
type State struct {
	Project        string                 `json:"project"`
	LastApplied    time.Time              `json:"lastApplied"`
	TerranixConfig string                 `json:"terranixConfig,omitempty"`
	TerraformState string                 `json:"terraformState,omitempty"`
	ColmenaApplied bool                   `json:"colmenaApplied"`
	Workflow       map[string]StepStatus  `json:"workflow"`
}

// StepStatus represents the status of a workflow step
type StepStatus struct {
	Status    string    `json:"status"` // "success", "failed", "pending", "running"
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// WorkflowStep names
const (
	StepNixBuild   = "nixBuild"
	StepTerraform  = "terraform"
	StepColmena    = "colmena"
)

// Step status values
const (
	StatusPending = "pending"
	StatusRunning = "running"
	StatusSuccess = "success"
	StatusFailed  = "failed"
)

