package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/iivel-inc/inframan/internal/state"
)

type model struct {
	stateManager *state.Manager
	currentState *state.State
	err          error
	width        int
	height       int
}

type tickMsg time.Time

func NewProgram(projectDir string) *tea.Program {
	stateManager, err := state.NewManager(projectDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to initialize state manager: %v\n", err)
		os.Exit(1)
	}

	m := model{
		stateManager: stateManager,
		currentState: stateManager.GetState(),
	}

	return tea.NewProgram(m, tea.WithAltScreen())
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tick(),
		loadState(m.stateManager),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			// Reload state
			return m, loadState(m.stateManager)
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tickMsg:
		return m, tea.Batch(tick(), loadState(m.stateManager))
	case state.State:
		m.currentState = &msg
		return m, nil
	}
	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress 'q' to quit", m.err)
	}

	var b strings.Builder

	// Header
	b.WriteString("┌─ Inframan ─────────────────────────────────────────────┐\n")
	b.WriteString("│ Press 'q' to quit, 'r' to refresh                      │\n")
	b.WriteString("└───────────────────────────────────────────────────────┘\n\n")

	// Project info
	if m.currentState.Project != "" {
		b.WriteString(fmt.Sprintf("Project: %s\n", m.currentState.Project))
	} else {
		b.WriteString("Project: (none)\n")
	}

	if !m.currentState.LastApplied.IsZero() {
		b.WriteString(fmt.Sprintf("Last Applied: %s\n", m.currentState.LastApplied.Format(time.RFC3339)))
	} else {
		b.WriteString("Last Applied: Never\n")
	}

	b.WriteString(fmt.Sprintf("Colmena Applied: %v\n", m.currentState.ColmenaApplied))
	b.WriteString("\n")

	// Workflow steps
	if len(m.currentState.Workflow) > 0 {
		b.WriteString("Workflow Steps:\n")
		b.WriteString("───────────────\n")

		steps := []string{state.StepNixBuild, state.StepTerraform, state.StepColmena}
		for _, stepName := range steps {
			stepStatus, exists := m.currentState.Workflow[stepName]
			if !exists {
				b.WriteString(fmt.Sprintf("  %s: %s\n", stepName, state.StatusPending))
				continue
			}

			statusIcon := "○"
			switch stepStatus.Status {
			case state.StatusSuccess:
				statusIcon = "✓"
			case state.StatusFailed:
				statusIcon = "✗"
			case state.StatusRunning:
				statusIcon = "⟳"
			}

			b.WriteString(fmt.Sprintf("  %s %s: %s", statusIcon, stepName, stepStatus.Status))
			if !stepStatus.Timestamp.IsZero() {
				b.WriteString(fmt.Sprintf(" (%s)", stepStatus.Timestamp.Format("15:04:05")))
			}
			if stepStatus.Message != "" {
				b.WriteString(fmt.Sprintf(" - %s", stepStatus.Message))
			}
			b.WriteString("\n")

			if stepStatus.Error != "" {
				b.WriteString(fmt.Sprintf("    Error: %s\n", stepStatus.Error))
			}
		}
	} else {
		b.WriteString("No workflow steps executed yet.\n")
	}

	b.WriteString("\n")
	b.WriteString("Press 'q' to quit, 'r' to refresh\n")

	return b.String()
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func loadState(sm *state.Manager) tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(time.Time) tea.Msg {
		if err := sm.Load(); err != nil {
			return state.State{Project: "error"}
		}
		return *sm.GetState()
	})
}

