package prompt

import (
	"fmt"
	"strings"

	"github.com/waynenilsen/crumbler/internal/models"
	"github.com/waynenilsen/crumbler/internal/phase"
	"github.com/waynenilsen/crumbler/internal/query"
	"github.com/waynenilsen/crumbler/internal/sprint"
)

// State represents the current state machine state.
type State string

const (
	// StateExit indicates the project is complete.
	StateExit State = "EXIT"

	// StateCreatePhase indicates a new phase should be created.
	StateCreatePhase State = "CREATE_PHASE"

	// StateCreatePhaseGoals indicates phase goals should be created.
	StateCreatePhaseGoals State = "CREATE_PHASE_GOALS"

	// StateCreateSprint indicates a new sprint should be created.
	StateCreateSprint State = "CREATE_SPRINT"

	// StateClosePhase indicates the current phase should be closed.
	StateClosePhase State = "CLOSE_PHASE"

	// StateCreateSprintGoals indicates sprint goals should be created.
	StateCreateSprintGoals State = "CREATE_SPRINT_GOALS"

	// StateCreateTickets indicates tickets should be created.
	StateCreateTickets State = "CREATE_TICKETS"

	// StateCloseSprint indicates the current sprint should be closed.
	StateCloseSprint State = "CLOSE_SPRINT"

	// StateCreateTicketGoals indicates ticket goals should be created.
	StateCreateTicketGoals State = "CREATE_TICKET_GOALS"

	// StateExecuteTicket indicates the current ticket should be executed.
	StateExecuteTicket State = "EXECUTE_TICKET"

	// StateMarkTicketDone indicates the current ticket should be marked done.
	StateMarkTicketDone State = "MARK_TICKET_DONE"

	// StateError indicates an error in state determination.
	StateError State = "ERROR"
)

// StateInstruction contains the instruction for a given state.
type StateInstruction struct {
	State       State
	Title       string
	Description string
	Steps       []string
	Commands    []string
	Notes       []string
}

// DetermineState determines the current state based on project context.
// This implements the flowchart decision logic.
func DetermineState(ctx *ProjectContext) (State, error) {
	// CHECK_PHASE: Does an open phase exist?
	if !ctx.HasOpenPhase() {
		// No open phase - check if roadmap is complete
		if ctx.IsRoadmapComplete() {
			return StateExit, nil
		}
		// Roadmap not complete - create next phase
		return StateCreatePhase, nil
	}

	// Open phase exists - check phase goals
	phaseGoalsExist, err := query.PhaseGoalsExist(ctx.CurrentPhase.Path)
	if err != nil {
		return StateError, err
	}
	if !phaseGoalsExist {
		return StateCreatePhaseGoals, nil
	}

	// CHECK_SPRINT: Does an open sprint exist?
	if !ctx.HasOpenSprint() {
		// No open sprint - check if phase goals met
		phaseGoalsMet, err := phase.ArePhaseGoalsMet(ctx.CurrentPhase.Path)
		if err != nil {
			return StateError, err
		}
		if phaseGoalsMet {
			return StateClosePhase, nil
		}
		// Phase goals not met - need to create sprint or goals
		sprintsExist, err := query.SprintsExist(ctx.CurrentPhase.Path)
		if err != nil {
			return StateError, err
		}
		if !sprintsExist {
			return StateCreateSprint, nil
		}
		// Sprints exist but not all closed - this shouldn't happen with open=false
		// but let's handle it by creating a sprint
		return StateCreateSprint, nil
	}

	// Open sprint exists - check sprint goals
	sprintGoalsExist, err := query.SprintGoalsExist(ctx.CurrentSprint.Path)
	if err != nil {
		return StateError, err
	}
	if !sprintGoalsExist {
		return StateCreateSprintGoals, nil
	}

	// CHECK_TICKETS: Do open tickets exist?
	if !ctx.HasOpenTicket() {
		// No open tickets - check if sprint goals met
		sprintGoalsMet, err := sprint.AreSprintGoalsMet(ctx.CurrentSprint.Path)
		if err != nil {
			return StateError, err
		}
		if sprintGoalsMet {
			return StateCloseSprint, nil
		}
		// Sprint goals not met - need to create tickets
		ticketsExist, err := query.TicketsExist(ctx.CurrentSprint.Path)
		if err != nil {
			return StateError, err
		}
		if !ticketsExist {
			return StateCreateTickets, nil
		}
		// Tickets exist but not all done - this shouldn't happen
		return StateCreateTickets, nil
	}

	// Open ticket exists - check ticket goals
	ticketGoalsExist, err := query.TicketGoalsExist(ctx.CurrentTicket.Path)
	if err != nil {
		return StateError, err
	}
	if !ticketGoalsExist {
		return StateCreateTicketGoals, nil
	}

	// Check if ticket is complete (all goals closed)
	if ctx.AllTicketGoalsClosed() {
		return StateMarkTicketDone, nil
	}

	// Ticket has open goals - execute ticket
	return StateExecuteTicket, nil
}

// GetStateInstruction returns the instruction for the given state.
func GetStateInstruction(state State, ctx *ProjectContext) *StateInstruction {
	switch state {
	case StateExit:
		return getExitInstruction(ctx)
	case StateCreatePhase:
		return getCreatePhaseInstruction(ctx)
	case StateCreatePhaseGoals:
		return getCreatePhaseGoalsInstruction(ctx)
	case StateCreateSprint:
		return getCreateSprintInstruction(ctx)
	case StateClosePhase:
		return getClosePhaseInstruction(ctx)
	case StateCreateSprintGoals:
		return getCreateSprintGoalsInstruction(ctx)
	case StateCreateTickets:
		return getCreateTicketsInstruction(ctx)
	case StateCloseSprint:
		return getCloseSprintInstruction(ctx)
	case StateCreateTicketGoals:
		return getCreateTicketGoalsInstruction(ctx)
	case StateExecuteTicket:
		return getExecuteTicketInstruction(ctx)
	case StateMarkTicketDone:
		return getMarkTicketDoneInstruction(ctx)
	default:
		return &StateInstruction{
			State:       StateError,
			Title:       "Error",
			Description: "Unknown state encountered.",
		}
	}
}

func getExitInstruction(ctx *ProjectContext) *StateInstruction {
	return &StateInstruction{
		State:       StateExit,
		Title:       "Project Complete!",
		Description: "All phases in the roadmap have been completed. The project is done.",
		Steps: []string{
			"Review the completed work",
			"Verify all goals were achieved",
			"Consider any final documentation updates",
		},
		Notes: []string{
			"Congratulations! The roadmap has been fully executed.",
		},
	}
}

func getCreatePhaseInstruction(ctx *ProjectContext) *StateInstruction {
	var phaseName string
	var phaseGoals []string
	if ctx.NextPhaseDef != nil {
		phaseName = ctx.NextPhaseDef.Name
		phaseGoals = ctx.NextPhaseDef.Goals
	}

	steps := []string{
		"Run: crumbler phase create",
	}

	if phaseName != "" {
		steps = append(steps, fmt.Sprintf("Populate README.md with phase description for: %s", phaseName))
	} else {
		steps = append(steps, "Populate README.md with phase description")
	}

	var notes []string
	notes = append(notes, "IMPORTANT: This phase should be planned to have MULTIPLE sprints (typically 3-5+ sprints)")
	notes = append(notes, "Phase goals should be HIGH-LEVEL objectives, NOT ticket-level implementation details")
	notes = append(notes, "DO NOT create a phase that maps 1-1 to a single sprint - that defeats the purpose!")
	if len(phaseGoals) > 0 {
		notes = append(notes, "")
		notes = append(notes, "Roadmap goals for this phase:")
		for i, g := range phaseGoals {
			notes = append(notes, fmt.Sprintf("  %d. %s", i+1, g))
		}
	}

	return &StateInstruction{
		State:       StateCreatePhase,
		Title:       "Create Next Phase",
		Description: fmt.Sprintf("Create phase %d from the roadmap.", ctx.PhaseIndex),
		Steps:       steps,
		Commands: []string{
			"crumbler phase create",
		},
		Notes: notes,
	}
}

func getCreatePhaseGoalsInstruction(ctx *ProjectContext) *StateInstruction {
	var suggestedGoals []string
	if ctx.RoadmapParsed != nil && ctx.PhaseIndex > 0 && ctx.PhaseIndex <= len(ctx.RoadmapParsed.Phases) {
		phaseDef := ctx.RoadmapParsed.Phases[ctx.PhaseIndex-1]
		suggestedGoals = phaseDef.Goals
	}

	steps := []string{
		"Review the phase README.md and roadmap to understand objectives",
		"Create goals that represent measurable outcomes for this phase",
	}

	for i, g := range suggestedGoals {
		steps = append(steps, fmt.Sprintf("Run: crumbler phase goal create \"%s\"", g))
		if i == 0 {
			steps[len(steps)-1] += " (from roadmap)"
		}
	}

	if len(suggestedGoals) == 0 {
		steps = append(steps, "Run: crumbler phase goal create \"<goal-name>\" for each goal")
	}

	return &StateInstruction{
		State:       StateCreatePhaseGoals,
		Title:       "Create Phase Goals",
		Description: fmt.Sprintf("Create goals for phase %s.", ctx.CurrentPhase.ID),
		Steps:       steps,
		Commands: []string{
			"crumbler phase goal create \"<goal-name>\"",
		},
		Notes: []string{
			"IMPORTANT: Phase goals should be HIGH-LEVEL objectives that require MULTIPLE sprints to achieve",
			"DO NOT create ticket-level implementation goals here (e.g., 'Create user model', 'Add API endpoint')",
			"Good phase goals: 'Implement user authentication system', 'Build payment processing', 'Create admin dashboard'",
			"Bad phase goals: 'Create database schema', 'Write login function', 'Add button to UI'",
			"Goals should be specific and measurable",
			"Use the roadmap as guidance for what goals to create",
			"Each phase goal should span multiple sprints, not map 1-1 to a single sprint",
		},
	}
}

func getCreateSprintInstruction(ctx *ProjectContext) *StateInstruction {
	// Check if this is the first sprint of the phase
	isFirstSprint := false
	if ctx.CurrentPhase != nil {
		sprintsExist, err := query.SprintsExist(ctx.CurrentPhase.Path)
		if err == nil && !sprintsExist {
			isFirstSprint = true
		}
	}

	notes := []string{
		"IMPORTANT: This sprint should be planned to have MULTIPLE tickets (typically 3-10+ tickets)",
		"DO NOT create a sprint with 1 ticket that maps 1-1 to sprint goals - break it down!",
	}

	// Add tech debt paydown rule for first sprint
	if isFirstSprint {
		notes = append(notes, "")
		notes = append(notes, "🚨 CRITICAL: This is the FIRST sprint of a new phase. You MUST create a TECH DEBT PAYDOWN sprint.",
			"A comprehensive Tech Debt Paydown Guide has been included in the CONTEXT section above.",
			"Read that guide thoroughly - it contains detailed information about:",
			"  - Common technical debt categories and how to identify them",
			"  - Prioritization frameworks for tech debt",
			"  - How to plan sprint goals and create tickets",
			"  - Tools and resources for addressing tech debt",
			"  - Patterns and anti-patterns to follow",
			"Use the guide to identify specific tech debt items in the codebase and create a focused sprint plan.",
			"The PRD.md and ERD.md should detail the specific tech debt items to address in this sprint.",
			"After this tech debt sprint, subsequent sprints in the phase can focus on feature development.")
	}

	notes = append(notes,
		"The sprint should work toward completing the phase goals",
		"PRD.md should detail what features/changes will be built",
		"ERD.md (Engineering Requirements Document) should detail the technical implementation approach, architecture, and engineering requirements",
		"Sprint goals should be achievable within the sprint timeframe, but require multiple tickets to complete")

	return &StateInstruction{
		State:       StateCreateSprint,
		Title:       "Create Sprint",
		Description: fmt.Sprintf("Create a new sprint in phase %s.", ctx.CurrentPhase.ID),
		Steps: []string{
			"Run: crumbler sprint create",
			"Populate sprint README.md with sprint objectives",
			"Populate PRD.md with product requirements for this sprint (see PRD Guide in CONTEXT section above for comprehensive guidance on writing effective PRDs)",
			"Populate ERD.md (Engineering Requirements Document) with technical design and implementation requirements (see ERD Guide in CONTEXT section above for comprehensive guidance. IMPORTANT: Write the PRD FIRST, then write the ERD based on the PRD requirements)",
		},
		Commands: []string{
			"crumbler sprint create",
		},
		Notes: notes,
	}
}

func getClosePhaseInstruction(ctx *ProjectContext) *StateInstruction {
	return &StateInstruction{
		State:       StateClosePhase,
		Title:       "Close Phase",
		Description: fmt.Sprintf("Close phase %s - all sprints and goals are complete.", ctx.CurrentPhase.ID),
		Steps: []string{
			"Verify all phase goals have been achieved",
			"Run: crumbler phase close",
		},
		Commands: []string{
			"crumbler phase close",
		},
		Notes: []string{
			"All sprints must be closed before closing the phase",
			"All phase goals must be closed before closing the phase",
		},
	}
}

func getCreateSprintGoalsInstruction(ctx *ProjectContext) *StateInstruction {
	return &StateInstruction{
		State:       StateCreateSprintGoals,
		Title:       "Create Sprint Goals",
		Description: fmt.Sprintf("Create goals for sprint %s.", ctx.CurrentSprint.ID),
		Steps: []string{
			"Review the sprint PRD.md and ERD.md (Engineering Requirements Document) to understand deliverables",
			"Create goals that represent measurable outcomes for this sprint",
			"Run: crumbler sprint goal create \"<goal-name>\" for each goal",
		},
		Commands: []string{
			"crumbler sprint goal create \"<goal-name>\"",
		},
		Notes: []string{
			"IMPORTANT: Sprint goals should require MULTIPLE tickets to achieve",
			"DO NOT create sprint goals that map 1-1 to a single ticket",
			"Good sprint goals: 'Users can register and log in', 'Payment processing works end-to-end'",
			"Bad sprint goals: 'Create user model', 'Add login button' (these are ticket-level)",
			"Sprint goals should be achievable within the sprint",
			"Goals should map to the requirements in PRD.md",
			"Each sprint goal should decompose into 2-5+ tickets",
		},
	}
}

func getCreateTicketsInstruction(ctx *ProjectContext) *StateInstruction {
	return &StateInstruction{
		State:       StateCreateTickets,
		Title:       "Create Tickets",
		Description: fmt.Sprintf("Decompose sprint %s into tickets.", ctx.CurrentSprint.ID),
		Steps: []string{
			"Review the sprint PRD.md and ERD.md (Engineering Requirements Document) to understand the work",
			"Break down the work into discrete, implementable tickets",
			"Create MULTIPLE tickets (typically 3-10+ tickets per sprint)",
			"For each ticket:",
			"  1. Run: crumbler ticket create",
			"  2. Populate the ticket README.md with:",
			"     - Clear description of what needs to be done",
			"     - Acceptance criteria",
			"     - Technical notes",
		},
		Commands: []string{
			"crumbler ticket create",
		},
		Notes: []string{
			"IMPORTANT: Create MULTIPLE tickets - DO NOT create just 1 ticket per sprint goal!",
			"Each ticket should be a single, focused unit of work",
			"Tickets should work toward the sprint goals",
			"Order tickets by dependency if applicable",
			"Example: If sprint goal is 'User registration', create tickets like:",
			"  - Ticket 1: Database schema for users",
			"  - Ticket 2: Registration API endpoint",
			"  - Ticket 3: Input validation",
			"  - Ticket 4: Error handling",
			"  - Ticket 5: Tests",
		},
	}
}

func getCloseSprintInstruction(ctx *ProjectContext) *StateInstruction {
	return &StateInstruction{
		State:       StateCloseSprint,
		Title:       "Close Sprint",
		Description: fmt.Sprintf("Close sprint %s - all tickets and goals are complete.", ctx.CurrentSprint.ID),
		Steps: []string{
			"Verify all sprint goals have been achieved",
			"Run: crumbler sprint close",
		},
		Commands: []string{
			"crumbler sprint close",
		},
		Notes: []string{
			"All tickets must be done before closing the sprint",
			"All sprint goals must be closed before closing the sprint",
		},
	}
}

func getCreateTicketGoalsInstruction(ctx *ProjectContext) *StateInstruction {
	return &StateInstruction{
		State:       StateCreateTicketGoals,
		Title:       "Create Ticket Goals",
		Description: fmt.Sprintf("Create goals for ticket %s.", ctx.CurrentTicket.ID),
		Steps: []string{
			"Review the ticket README.md to understand the work",
			"Create goals that represent verifiable completion criteria",
			fmt.Sprintf("Run: crumbler ticket goal create %s \"<goal-name>\" for each goal", ctx.CurrentTicket.ID),
		},
		Commands: []string{
			fmt.Sprintf("crumbler ticket goal create %s \"<goal-name>\"", ctx.CurrentTicket.ID),
		},
		Notes: []string{
			"Ticket goals should be checkable/verifiable",
			"Examples: 'Implement X function', 'Add tests for Y', 'Update documentation for Z'",
		},
	}
}

func getExecuteTicketInstruction(ctx *ProjectContext) *StateInstruction {
	var openGoals []string
	for _, g := range ctx.TicketGoals {
		if g.Status == models.StatusOpen {
			openGoals = append(openGoals, fmt.Sprintf("- [ ] %s: %s", g.ID, g.Name))
		}
	}

	steps := []string{
		"Read the ticket README.md for context",
		"Work on the open goals listed below",
		"As you complete each goal, close it:",
	}

	for _, g := range ctx.TicketGoals {
		if g.Status == models.StatusOpen {
			steps = append(steps, fmt.Sprintf("  crumbler ticket goal close %s %s", ctx.CurrentTicket.ID, g.ID))
		}
	}

	return &StateInstruction{
		State:       StateExecuteTicket,
		Title:       "Execute Ticket",
		Description: fmt.Sprintf("Execute ticket %s - complete the open goals.", ctx.CurrentTicket.ID),
		Steps:       steps,
		Commands: []string{
			fmt.Sprintf("crumbler ticket goal close %s <goal-id>", ctx.CurrentTicket.ID),
		},
		Notes: append([]string{"Open goals:"}, openGoals...),
	}
}

func getMarkTicketDoneInstruction(ctx *ProjectContext) *StateInstruction {
	return &StateInstruction{
		State:       StateMarkTicketDone,
		Title:       "Mark Ticket Done",
		Description: fmt.Sprintf("Mark ticket %s as done - all goals are complete.", ctx.CurrentTicket.ID),
		Steps: []string{
			"Verify the ticket work is complete",
			fmt.Sprintf("Run: crumbler ticket done %s", ctx.CurrentTicket.ID),
		},
		Commands: []string{
			fmt.Sprintf("crumbler ticket done %s", ctx.CurrentTicket.ID),
		},
		Notes: []string{
			"All ticket goals must be closed before marking done",
		},
	}
}

// FormatGoalsList formats goals as a checklist string.
func FormatGoalsList(goals []models.Goal) string {
	if len(goals) == 0 {
		return "(no goals)"
	}

	var lines []string
	for _, g := range goals {
		var checkbox string
		if g.Status == models.StatusClosed {
			checkbox = "[x]"
		} else {
			checkbox = "[ ]"
		}
		lines = append(lines, fmt.Sprintf("- %s %s: %q (%s)", checkbox, g.ID, g.Name, g.Status))
	}
	return strings.Join(lines, "\n")
}
