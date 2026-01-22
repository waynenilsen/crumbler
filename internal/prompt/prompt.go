package prompt

import (
	"fmt"
	"strings"
)

// PromptConfig configures prompt generation.
type PromptConfig struct {
	IncludePrelude  bool // Include the prelude section
	IncludePostlude bool // Include the postlude section
	IncludeContext  bool // Include file contents in context section
	Minimal         bool // Use minimal versions of prelude/postlude
}

// DefaultConfig returns the default prompt configuration.
func DefaultConfig() *PromptConfig {
	return &PromptConfig{
		IncludePrelude:  true,
		IncludePostlude: true,
		IncludeContext:  true,
		Minimal:         false,
	}
}

// GeneratePrompt generates a complete prompt for the AI agent.
func GeneratePrompt(projectRoot string, config *PromptConfig) (string, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Gather context
	ctx, err := GatherContext(projectRoot)
	if err != nil {
		return "", fmt.Errorf("failed to gather context: %w", err)
	}

	// Determine state
	state, err := DetermineState(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to determine state: %w", err)
	}

	// Get instruction for this state
	instruction := GetStateInstruction(state, ctx)

	// Build prompt
	var sb strings.Builder

	// Prelude
	if config.IncludePrelude {
		if config.Minimal {
			sb.WriteString(GeneratePreludeMinimal(ctx, state))
		} else {
			sb.WriteString(GeneratePrelude(ctx, state))
		}
	}

	// Context section (file contents)
	if config.IncludeContext {
		sb.WriteString(generateContextSection(ctx, state))
	}

	// Instruction section
	sb.WriteString(generateInstructionSection(instruction))

	// Postlude
	if config.IncludePostlude {
		if config.Minimal {
			sb.WriteString(GeneratePostludeMinimal(state))
		} else {
			sb.WriteString(GeneratePostlude(ctx, state, instruction))
		}
	}

	return sb.String(), nil
}

// generateContextSection generates the context section with file contents.
func generateContextSection(ctx *ProjectContext, state State) string {
	var sb strings.Builder

	sb.WriteString("═══════════════════════════════════════════════════════════════════════════════\n")
	sb.WriteString("                              CONTEXT                                          \n")
	sb.WriteString("═══════════════════════════════════════════════════════════════════════════════\n\n")

	// Roadmap (always relevant)
	if ctx.Roadmap != nil && !ctx.Roadmap.Missing {
		sb.WriteString(formatFileContent(ctx.Roadmap))
		sb.WriteString("\n")
	}

	// Phase context (if phase exists or being created)
	if ctx.CurrentPhase != nil {
		// Phase goals
		if len(ctx.PhaseGoals) > 0 {
			sb.WriteString("### Phase Goals\n\n")
			sb.WriteString(FormatGoalsList(ctx.PhaseGoals))
			sb.WriteString("\n\n")
		}

		// Phase README
		if ctx.PhaseReadme != nil && !ctx.PhaseReadme.Missing {
			sb.WriteString(formatFileContent(ctx.PhaseReadme))
			sb.WriteString("\n")
		}
	}

	// Sprint context (if sprint exists)
	if ctx.CurrentSprint != nil {
		// Sprint goals
		if len(ctx.SprintGoals) > 0 {
			sb.WriteString("### Sprint Goals\n\n")
			sb.WriteString(FormatGoalsList(ctx.SprintGoals))
			sb.WriteString("\n\n")
		}

		// Sprint files
		if ctx.SprintReadme != nil && !ctx.SprintReadme.Missing {
			sb.WriteString(formatFileContent(ctx.SprintReadme))
			sb.WriteString("\n")
		}
		if ctx.SprintPRD != nil && !ctx.SprintPRD.Missing {
			sb.WriteString(formatFileContent(ctx.SprintPRD))
			sb.WriteString("\n")
		}
		if ctx.SprintERD != nil && !ctx.SprintERD.Missing {
			sb.WriteString(formatFileContent(ctx.SprintERD))
			sb.WriteString("\n")
		}
	}

	// Ticket context (if ticket exists)
	if ctx.CurrentTicket != nil {
		// Ticket goals
		if len(ctx.TicketGoals) > 0 {
			sb.WriteString("### Ticket Goals\n\n")
			sb.WriteString(FormatGoalsList(ctx.TicketGoals))
			sb.WriteString("\n\n")
		}

		// Ticket README
		if ctx.TicketReadme != nil && !ctx.TicketReadme.Missing {
			sb.WriteString(formatFileContent(ctx.TicketReadme))
			sb.WriteString("\n")
		}
	}

	// List all open tickets if multiple
	if len(ctx.OpenTickets) > 1 {
		sb.WriteString("### Open Tickets in Sprint\n\n")
		for _, t := range ctx.OpenTickets {
			status := "OPEN"
			sb.WriteString(fmt.Sprintf("- %s (%s)\n", t.ID, status))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// formatFileContent formats a context file for display.
func formatFileContent(cf *ContextFile) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("### %s\n\n", cf.RelPath))

	if cf.Missing {
		sb.WriteString("(file does not exist)\n")
	} else if cf.Empty {
		sb.WriteString("(file is empty - YOU MUST POPULATE THIS)\n")
	} else {
		// Truncate very long content
		content := cf.Contents
		const maxLen = 4000
		if len(content) > maxLen {
			content = content[:maxLen] + "\n... (truncated)"
		}
		sb.WriteString("```\n")
		sb.WriteString(content)
		sb.WriteString("\n```\n")
	}

	return sb.String()
}

// generateInstructionSection generates the instruction section.
func generateInstructionSection(instruction *StateInstruction) string {
	var sb strings.Builder

	sb.WriteString("═══════════════════════════════════════════════════════════════════════════════\n")
	sb.WriteString("                             INSTRUCTION                                       \n")
	sb.WriteString("═══════════════════════════════════════════════════════════════════════════════\n\n")

	// Title and description
	sb.WriteString(fmt.Sprintf("## %s\n\n", instruction.Title))
	sb.WriteString(instruction.Description)
	sb.WriteString("\n\n")

	// Steps
	if len(instruction.Steps) > 0 {
		sb.WriteString("### Steps\n\n")
		for i, step := range instruction.Steps {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, step))
		}
		sb.WriteString("\n")
	}

	// Notes
	if len(instruction.Notes) > 0 {
		sb.WriteString("### Notes\n\n")
		for _, note := range instruction.Notes {
			sb.WriteString(fmt.Sprintf("- %s\n", note))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// GetState returns just the current state without generating the full prompt.
func GetState(projectRoot string) (State, error) {
	ctx, err := GatherContext(projectRoot)
	if err != nil {
		return StateError, fmt.Errorf("failed to gather context: %w", err)
	}

	return DetermineState(ctx)
}

// GetStateString returns the current state as a string.
func GetStateString(projectRoot string) (string, error) {
	state, err := GetState(projectRoot)
	if err != nil {
		return "", err
	}
	return string(state), nil
}
