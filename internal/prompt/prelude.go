package prompt

import (
	"fmt"
	"strings"
)

// GeneratePrelude generates the prelude section of the prompt.
func GeneratePrelude(ctx *ProjectContext, state State) string {
	var sb strings.Builder

	// Header
	sb.WriteString("╔═══════════════════════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║                           CRUMBLER AGENT PROMPT                               ║\n")
	sb.WriteString("╚═══════════════════════════════════════════════════════════════════════════════╝\n\n")

	// State indicator
	sb.WriteString(fmt.Sprintf("STATE: %s\n", state))
	sb.WriteString(fmt.Sprintf("POSITION: %s\n\n", ctx.GetCurrentPosition()))

	// What is crumbler
	sb.WriteString("## What is Crumbler?\n\n")
	sb.WriteString(CrumblerDescription)
	sb.WriteString("\n\n")

	// State explanation
	sb.WriteString("## State Management\n\n")
	sb.WriteString(StateExplanation)
	sb.WriteString("\n\n")

	// Goals
	sb.WriteString("## Goals\n\n")
	sb.WriteString(GoalExplanation)
	sb.WriteString("\n\n")

	// Hierarchy rules
	sb.WriteString("## Hierarchy Rules\n\n")
	sb.WriteString(HierarchyRules)
	sb.WriteString("\n\n")

	// How to work
	sb.WriteString("## How to Work\n\n")
	sb.WriteString(AgentWorkflowInstructions)
	sb.WriteString("\n\n")

	// Progress summary
	if ctx.RoadmapParsed != nil && ctx.TotalPhases > 0 {
		sb.WriteString("## Progress\n\n")
		sb.WriteString(fmt.Sprintf("Roadmap: %d/%d phases completed\n", ctx.ClosedPhases, ctx.TotalPhases))
		if ctx.CurrentPhase != nil {
			sb.WriteString(fmt.Sprintf("Current Phase: %s (%s)\n", ctx.CurrentPhase.ID, ctx.CurrentPhase.Status))
		}
		if ctx.CurrentSprint != nil {
			sb.WriteString(fmt.Sprintf("Current Sprint: %s (%s)\n", ctx.CurrentSprint.ID, ctx.CurrentSprint.Status))
		}
		if ctx.CurrentTicket != nil {
			sb.WriteString(fmt.Sprintf("Current Ticket: %s (%s)\n", ctx.CurrentTicket.ID, ctx.CurrentTicket.Status))
		}
		if len(ctx.OpenTickets) > 1 {
			sb.WriteString(fmt.Sprintf("Open Tickets in Sprint: %d\n", len(ctx.OpenTickets)))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// GeneratePreludeMinimal generates a minimal prelude (just state and position).
func GeneratePreludeMinimal(ctx *ProjectContext, state State) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("STATE: %s\n", state))
	sb.WriteString(fmt.Sprintf("POSITION: %s\n\n", ctx.GetCurrentPosition()))

	return sb.String()
}
