package prompt

import (
	"strings"
)

// GeneratePostlude generates the postlude section of the prompt.
func GeneratePostlude(ctx *ProjectContext, state State, instruction *StateInstruction) string {
	var sb strings.Builder

	sb.WriteString("═══════════════════════════════════════════════════════════════════════════════\n")
	sb.WriteString("                              NEXT STEPS                                       \n")
	sb.WriteString("═══════════════════════════════════════════════════════════════════════════════\n\n")

	if state == StateExit {
		sb.WriteString("The project is complete! No further action required.\n\n")
		sb.WriteString("You may want to:\n")
		sb.WriteString("- Review the completed work\n")
		sb.WriteString("- Run final tests\n")
		sb.WriteString("- Create a summary of what was accomplished\n")
		return sb.String()
	}

	// Commands to run
	sb.WriteString("## Commands to Run\n\n")
	if len(instruction.Commands) > 0 {
		for _, cmd := range instruction.Commands {
			sb.WriteString("```\n")
			sb.WriteString(cmd)
			sb.WriteString("\n```\n\n")
		}
	}

	// After completing this step
	sb.WriteString("## After Completing This Step\n\n")
	sb.WriteString("Run this command to get your next instruction:\n\n")
	sb.WriteString("```\n")
	sb.WriteString("crumbler get-next-prompt\n")
	sb.WriteString("```\n\n")

	// Useful commands reference
	sb.WriteString("## Quick Reference\n\n")
	sb.WriteString("Status commands:\n")
	sb.WriteString("- `crumbler status` - Show current project status\n")
	sb.WriteString("- `crumbler phase list` - List all phases\n")
	sb.WriteString("- `crumbler sprint list` - List sprints in current phase\n")
	sb.WriteString("- `crumbler ticket list` - List tickets in current sprint\n\n")

	sb.WriteString("Help commands:\n")
	sb.WriteString("- `crumbler help` - General help\n")
	sb.WriteString("- `crumbler help phase` - Phase command help\n")
	sb.WriteString("- `crumbler help sprint` - Sprint command help\n")
	sb.WriteString("- `crumbler help ticket` - Ticket command help\n")

	return sb.String()
}

// GeneratePostludeMinimal generates a minimal postlude (just the next command).
func GeneratePostludeMinimal(state State) string {
	var sb strings.Builder

	if state == StateExit {
		sb.WriteString("Project complete. No further action required.\n")
		return sb.String()
	}

	sb.WriteString("After completing this step, run:\n")
	sb.WriteString("  crumbler get-next-prompt\n")

	return sb.String()
}
