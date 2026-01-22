package prompt

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/waynenilsen/crumbler/internal/models"
	"github.com/waynenilsen/crumbler/internal/phase"
	"github.com/waynenilsen/crumbler/internal/roadmap"
	"github.com/waynenilsen/crumbler/internal/sprint"
	"github.com/waynenilsen/crumbler/internal/state"
	"github.com/waynenilsen/crumbler/internal/ticket"
)

// ContextFile represents a file that may be included in the prompt context.
type ContextFile struct {
	RelPath  string // Relative path from project root
	Contents string // File contents (may be empty)
	Empty    bool   // True if file exists but is empty
	Missing  bool   // True if file doesn't exist
}

// ProjectContext holds all the context information for generating a prompt.
type ProjectContext struct {
	ProjectRoot string // Absolute path to project root

	// Roadmap
	Roadmap       *ContextFile       // Contents of roadmap.md
	RoadmapParsed *models.Roadmap    // Parsed roadmap structure
	TotalPhases   int                // Total phases from roadmap
	ClosedPhases  int                // Number of closed phases

	// Current Phase
	CurrentPhase    *models.Phase   // Currently open phase (nil if none)
	PhaseReadme     *ContextFile    // Contents of phase README.md
	PhaseGoals      []models.Goal   // Goals for current phase
	PhaseIndex      int             // 1-based index of current phase
	NextPhaseDef    *models.PhaseDefinition // Next phase to create (if applicable)

	// Current Sprint
	CurrentSprint *models.Sprint // Currently open sprint (nil if none)
	SprintReadme  *ContextFile   // Contents of sprint README.md
	SprintPRD     *ContextFile   // Contents of sprint PRD.md
	SprintERD     *ContextFile   // Contents of sprint ERD.md
	SprintGoals   []models.Goal  // Goals for current sprint
	SprintIndex   int            // 1-based index of current sprint

	// Current Ticket (first open ticket in current sprint)
	CurrentTicket *models.Ticket // First open ticket (nil if none)
	TicketReadme  *ContextFile   // Contents of ticket README.md
	TicketGoals   []models.Goal  // Goals for current ticket
	TicketIndex   int            // 1-based index of current ticket

	// Open tickets list
	OpenTickets []models.Ticket // All open tickets in current sprint
}

// GatherContext collects all relevant context for prompt generation.
func GatherContext(projectRoot string) (*ProjectContext, error) {
	ctx := &ProjectContext{
		ProjectRoot: projectRoot,
	}

	// Load roadmap
	roadmapPath := state.RoadmapPath(projectRoot)
	ctx.Roadmap = ReadFileContext(projectRoot, roadmapPath)
	if !ctx.Roadmap.Missing {
		parsed, err := roadmap.LoadProjectRoadmap(projectRoot)
		if err == nil {
			ctx.RoadmapParsed = parsed
			ctx.TotalPhases = len(parsed.Phases)
		}
	}

	// Count closed phases
	phases, err := phase.ListPhases(projectRoot)
	if err == nil {
		for _, p := range phases {
			if p.Status == models.StatusClosed {
				ctx.ClosedPhases++
			}
		}
	}

	// Get current (open) phase
	openPhase, err := phase.GetOpenPhase(projectRoot)
	if err == nil && openPhase != nil {
		ctx.CurrentPhase = openPhase
		ctx.PhaseIndex = openPhase.Index
		ctx.PhaseGoals = openPhase.Goals

		// Load phase README
		phaseReadmePath := filepath.Join(openPhase.Path, models.ReadmeFile)
		ctx.PhaseReadme = ReadFileContext(projectRoot, phaseReadmePath)

		// Get current (open) sprint in this phase
		openSprint, err := sprint.GetOpenSprint(openPhase.Path)
		if err == nil && openSprint != nil {
			ctx.CurrentSprint = openSprint
			ctx.SprintIndex = openSprint.Index
			ctx.SprintGoals = openSprint.Goals

			// Load sprint files
			sprintReadmePath := filepath.Join(openSprint.Path, models.ReadmeFile)
			ctx.SprintReadme = ReadFileContext(projectRoot, sprintReadmePath)

			prdPath := filepath.Join(openSprint.Path, models.PRDFile)
			ctx.SprintPRD = ReadFileContext(projectRoot, prdPath)

			erdPath := filepath.Join(openSprint.Path, models.ERDFile)
			ctx.SprintERD = ReadFileContext(projectRoot, erdPath)

			// Get open tickets
			openTickets, err := ticket.GetOpenTickets(openSprint.Path)
			if err == nil {
				ctx.OpenTickets = openTickets
				if len(openTickets) > 0 {
					// Use first open ticket as current
					ctx.CurrentTicket = &openTickets[0]
					ctx.TicketIndex = openTickets[0].Index
					ctx.TicketGoals = openTickets[0].Goals

					// Load ticket README
					ticketReadmePath := filepath.Join(openTickets[0].Path, models.ReadmeFile)
					ctx.TicketReadme = ReadFileContext(projectRoot, ticketReadmePath)
				}
			}
		}
	}

	// If no open phase but roadmap has more phases, get next phase definition
	if ctx.CurrentPhase == nil && ctx.RoadmapParsed != nil {
		nextPhase, nextIndex, err := roadmap.GetNextPhaseFromRoadmap(ctx.RoadmapParsed, projectRoot)
		if err == nil && nextPhase != nil {
			ctx.NextPhaseDef = nextPhase
			ctx.PhaseIndex = nextIndex
		}
	}

	return ctx, nil
}

// ReadFileContext reads a file and returns a ContextFile structure.
func ReadFileContext(projectRoot, absPath string) *ContextFile {
	relPath := getRelPath(projectRoot, absPath)

	content, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &ContextFile{
				RelPath: relPath,
				Missing: true,
			}
		}
		// Other error - treat as missing
		return &ContextFile{
			RelPath: relPath,
			Missing: true,
		}
	}

	trimmed := strings.TrimSpace(string(content))
	return &ContextFile{
		RelPath:  relPath,
		Contents: trimmed,
		Empty:    trimmed == "",
	}
}

// getRelPath returns the path relative to projectRoot.
func getRelPath(projectRoot, absPath string) string {
	rel, err := filepath.Rel(projectRoot, absPath)
	if err != nil {
		return absPath
	}
	return rel
}

// HasOpenPhase returns true if there's an open phase.
func (ctx *ProjectContext) HasOpenPhase() bool {
	return ctx.CurrentPhase != nil
}

// HasOpenSprint returns true if there's an open sprint.
func (ctx *ProjectContext) HasOpenSprint() bool {
	return ctx.CurrentSprint != nil
}

// HasOpenTicket returns true if there's an open ticket.
func (ctx *ProjectContext) HasOpenTicket() bool {
	return ctx.CurrentTicket != nil
}

// IsRoadmapComplete returns true if all phases from the roadmap are closed.
func (ctx *ProjectContext) IsRoadmapComplete() bool {
	if ctx.RoadmapParsed == nil || ctx.TotalPhases == 0 {
		return false
	}
	return ctx.ClosedPhases >= ctx.TotalPhases
}

// PhaseGoalsExist returns true if the current phase has goals.
func (ctx *ProjectContext) PhaseGoalsExist() bool {
	return len(ctx.PhaseGoals) > 0
}

// SprintGoalsExist returns true if the current sprint has goals.
func (ctx *ProjectContext) SprintGoalsExist() bool {
	return len(ctx.SprintGoals) > 0
}

// TicketGoalsExist returns true if the current ticket has goals.
func (ctx *ProjectContext) TicketGoalsExist() bool {
	return len(ctx.TicketGoals) > 0
}

// AllPhaseGoalsClosed returns true if all phase goals are closed.
func (ctx *ProjectContext) AllPhaseGoalsClosed() bool {
	if len(ctx.PhaseGoals) == 0 {
		return false
	}
	for _, g := range ctx.PhaseGoals {
		if g.Status != models.StatusClosed {
			return false
		}
	}
	return true
}

// AllSprintGoalsClosed returns true if all sprint goals are closed.
func (ctx *ProjectContext) AllSprintGoalsClosed() bool {
	if len(ctx.SprintGoals) == 0 {
		return false
	}
	for _, g := range ctx.SprintGoals {
		if g.Status != models.StatusClosed {
			return false
		}
	}
	return true
}

// AllTicketGoalsClosed returns true if all ticket goals are closed.
func (ctx *ProjectContext) AllTicketGoalsClosed() bool {
	if len(ctx.TicketGoals) == 0 {
		return true // No goals = vacuously true for tickets
	}
	for _, g := range ctx.TicketGoals {
		if g.Status != models.StatusClosed {
			return false
		}
	}
	return true
}

// GetCurrentPosition returns a human-readable string of current position.
func (ctx *ProjectContext) GetCurrentPosition() string {
	var parts []string

	if ctx.CurrentPhase != nil {
		parts = append(parts, "Phase "+ctx.CurrentPhase.ID)
	}
	if ctx.CurrentSprint != nil {
		parts = append(parts, "Sprint "+ctx.CurrentSprint.ID)
	}
	if ctx.CurrentTicket != nil {
		parts = append(parts, "Ticket "+ctx.CurrentTicket.ID)
	}

	if len(parts) == 0 {
		return "No active entities"
	}
	return strings.Join(parts, " > ")
}
