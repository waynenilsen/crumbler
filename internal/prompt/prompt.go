// Package prompt generates AI agent prompts for crumbler v2.
// The prompt tells the agent what the current crumb is and what to do with it.
package prompt

import (
	"strings"

	"github.com/waynenilsen/crumbler/internal/crumb"
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
		return formatDonePrompt(config), nil
	}

	// Get current crumb
	current, err := crumb.GetCurrent(root)
	if err != nil {
		return "", err
	}

	if current == nil {
		return formatDonePrompt(config), nil
	}

	// Build the full prompt
	return formatPrompt(root, current, config)
}

// formatPrompt builds the complete prompt string.
func formatPrompt(root string, current *crumb.Crumb, config *Config) (string, error) {
	var sb strings.Builder

	// Preamble
	if !config.NoPreamble {
		sb.WriteString(formatPreamble(config.Minimal))
		sb.WriteString("\n")
	}

	// Context
	if !config.NoContext {
		sb.WriteString(formatContext(root, current))
		sb.WriteString("\n")
	}

	// Postamble
	if !config.NoPostamble {
		sb.WriteString(formatPostamble(config.Minimal))
	}

	return sb.String(), nil
}

// formatDonePrompt generates the prompt when all work is complete.
func formatDonePrompt(config *Config) string {
	var sb strings.Builder

	sb.WriteString("# DONE\n\n")
	sb.WriteString("All crumbs have been completed. The project is done.\n")

	if !config.NoPostamble && !config.Minimal {
		sb.WriteString("\nNo further action is required.\n")
	}

	return sb.String()
}
