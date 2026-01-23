// Package prompt generates AI agent prompts for crumbler v2.
// The prompt tells the agent what the current crumb is and what to do with it.
package prompt

import (
	"fmt"
	"strings"

	"github.com/waynenilsen/crumbler/internal/crumb"
)

// State represents the current state of the crumbler project.
type State string

const (
	// StateDecompose indicates the current crumb should be broken down into sub-crumbs.
	StateDecompose State = "DECOMPOSE"

	// StateExecute indicates the current crumb should be executed (work done).
	StateExecute State = "EXECUTE"

	// StateDone indicates all work is complete (no crumbs remain).
	StateDone State = "DONE"
)

// Config controls prompt generation options.
type Config struct {
	// NoPreamble skips the preamble section.
	NoPreamble bool

	// NoPostamble skips the postamble section.
	NoPostamble bool

	// NoContext skips the context section (README contents).
	NoContext bool

	// Minimal uses minimal preamble/postamble.
	Minimal bool

	// StateOnly outputs only the state name.
	StateOnly bool
}

// GeneratePrompt generates the AI agent prompt for the current project state.
func GeneratePrompt(root string, config *Config) (string, error) {
	if config == nil {
		config = &Config{}
	}

	// Check if project is done
	done, err := crumb.IsDone(root)
	if err != nil {
		return "", err
	}

	if done {
		if config.StateOnly {
			return string(StateDone), nil
		}
		return formatDonePrompt(config), nil
	}

	// Get current crumb
	current, err := crumb.GetCurrent(root)
	if err != nil {
		return "", err
	}

	if current == nil {
		if config.StateOnly {
			return string(StateDone), nil
		}
		return formatDonePrompt(config), nil
	}

	// Determine state based on README content
	readme, err := current.GetReadme()
	if err != nil {
		return "", fmt.Errorf("failed to read README: %w", err)
	}

	// If README is empty, we're in DECOMPOSE state (need to plan the work)
	// If README has content, we're in EXECUTE state (do the work)
	state := StateExecute
	if strings.TrimSpace(readme) == "" {
		state = StateDecompose
	}

	if config.StateOnly {
		return string(state), nil
	}

	// Build the full prompt
	return formatPrompt(root, current, readme, state, config)
}

// GetState returns just the current state without generating the full prompt.
func GetState(root string) (State, error) {
	done, err := crumb.IsDone(root)
	if err != nil {
		return "", err
	}
	if done {
		return StateDone, nil
	}

	current, err := crumb.GetCurrent(root)
	if err != nil {
		return "", err
	}
	if current == nil {
		return StateDone, nil
	}

	readme, err := current.GetReadme()
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(readme) == "" {
		return StateDecompose, nil
	}
	return StateExecute, nil
}

// formatPrompt builds the complete prompt string.
func formatPrompt(root string, current *crumb.Crumb, readme string, state State, config *Config) (string, error) {
	var sb strings.Builder

	// Preamble
	if !config.NoPreamble {
		sb.WriteString(formatPreamble(config.Minimal))
		sb.WriteString("\n")
	}

	// Context
	if !config.NoContext {
		sb.WriteString(formatContext(root, current, readme))
		sb.WriteString("\n")
	}

	// State-specific instructions
	sb.WriteString(formatInstructions(state, current))
	sb.WriteString("\n")

	// Postamble
	if !config.NoPostamble {
		sb.WriteString(formatPostamble(config.Minimal))
	}

	return sb.String(), nil
}

// formatDonePrompt generates the prompt when all work is complete.
func formatDonePrompt(config *Config) string {
	var sb strings.Builder

	sb.WriteString("STATE: DONE\n\n")
	sb.WriteString("All crumbs have been completed. The project is done.\n")

	if !config.NoPostamble && !config.Minimal {
		sb.WriteString("\nNo further action is required.\n")
	}

	return sb.String()
}
