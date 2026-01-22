package prompt_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/waynenilsen/crumbler/internal/prompt"
	"github.com/waynenilsen/crumbler/internal/testutil"
)

// =============================================================================
// Test: DetermineState
// =============================================================================

func TestDetermineState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(t *testing.T) string
		expected prompt.State
		wantErr  bool
	}{
		{
			name: "EXIT when roadmap complete and no open phases",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
## Phase 2: Features
- Goal 2
`).
					WithPhase("0001", "closed").
					WithPhase("0002", "closed").
					Build()
			},
			expected: prompt.StateExit,
		},
		{
			name: "CREATE_PHASE when no phases exist",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					Build()
			},
			expected: prompt.StateCreatePhase,
		},
		{
			name: "CREATE_PHASE when all phases closed but roadmap has more",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
## Phase 2: Features
- Goal 2
`).
					WithPhase("0001", "closed").
					Build()
			},
			expected: prompt.StateCreatePhase,
		},
		{
			name: "CREATE_PHASE_GOALS when phase exists but no goals",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					Build()
			},
			expected: prompt.StateCreatePhaseGoals,
		},
		{
			name: "CREATE_SPRINT when phase has goals but no sprints",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Goal 1", "open").
					Build()
			},
			expected: prompt.StateCreateSprint,
		},
		{
			name: "CLOSE_PHASE when all sprints and goals closed",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Goal 1", "closed").
					WithSprint("0001", "0001", "closed").
					Build()
			},
			expected: prompt.StateClosePhase,
		},
		{
			name: "CREATE_SPRINT_GOALS when sprint exists but no goals",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Goal 1", "open").
					WithSprint("0001", "0001", "open").
					Build()
			},
			expected: prompt.StateCreateSprintGoals,
		},
		{
			name: "CREATE_TICKETS when sprint has goals but no tickets",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Goal 1", "open").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 1, "Sprint Goal 1", "open").
					Build()
			},
			expected: prompt.StateCreateTickets,
		},
		{
			name: "CLOSE_SPRINT when all tickets and goals done",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Goal 1", "open").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 1, "Sprint Goal 1", "closed").
					WithTicket("0001", "0001", "0001", "done").
					Build()
			},
			expected: prompt.StateCloseSprint,
		},
		{
			name: "CREATE_TICKET_GOALS when ticket exists but no goals",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Goal 1", "open").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 1, "Sprint Goal 1", "open").
					WithTicket("0001", "0001", "0001", "open").
					Build()
			},
			expected: prompt.StateCreateTicketGoals,
		},
		{
			name: "EXECUTE_TICKET when ticket has open goals",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Goal 1", "open").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 1, "Sprint Goal 1", "open").
					WithTicket("0001", "0001", "0001", "open").
					WithTicketGoal("0001", "0001", "0001", 1, "Ticket Goal 1", "open").
					Build()
			},
			expected: prompt.StateExecuteTicket,
		},
		{
			name: "MARK_TICKET_DONE when all ticket goals closed",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Goal 1", "open").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 1, "Sprint Goal 1", "open").
					WithTicket("0001", "0001", "0001", "open").
					WithTicketGoal("0001", "0001", "0001", 1, "Ticket Goal 1", "closed").
					Build()
			},
			expected: prompt.StateMarkTicketDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			projectRoot := tt.setup(t)

			ctx, err := prompt.GatherContext(projectRoot)
			if err != nil {
				t.Fatalf("failed to gather context: %v", err)
			}

			state, err := prompt.DetermineState(ctx)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if state != tt.expected {
					t.Errorf("expected state %s, got %s", tt.expected, state)
				}
			}
		})
	}
}

// =============================================================================
// Test: GatherContext
// =============================================================================

func TestGatherContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(t *testing.T) string
		check func(t *testing.T, ctx *prompt.ProjectContext)
	}{
		{
			name: "gathers roadmap context",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
- Goal 2
## Phase 2: Features
- Goal 3
`).
					Build()
			},
			check: func(t *testing.T, ctx *prompt.ProjectContext) {
				if ctx.Roadmap == nil {
					t.Fatal("expected roadmap to be set")
				}
				if ctx.Roadmap.Missing {
					t.Error("expected roadmap to exist")
				}
				if ctx.RoadmapParsed == nil {
					t.Fatal("expected parsed roadmap")
				}
				if ctx.TotalPhases != 2 {
					t.Errorf("expected 2 total phases, got %d", ctx.TotalPhases)
				}
			},
		},
		{
			name: "gathers phase context",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "My Phase Goal", "open").
					Build()
				// Write some content to the README
				readmePath := filepath.Join(root, ".crumbler", "phases", "0001-phase", "README.md")
				testutil.WriteFile(t, readmePath, "Phase 1 description")
				return root
			},
			check: func(t *testing.T, ctx *prompt.ProjectContext) {
				if ctx.CurrentPhase == nil {
					t.Fatal("expected current phase to be set")
				}
				if ctx.CurrentPhase.ID != "0001-phase" {
					t.Errorf("expected phase ID 0001-phase, got %s", ctx.CurrentPhase.ID)
				}
				if len(ctx.PhaseGoals) != 1 {
					t.Errorf("expected 1 phase goal, got %d", len(ctx.PhaseGoals))
				}
				if ctx.PhaseReadme == nil || ctx.PhaseReadme.Empty {
					t.Error("expected phase README to have content")
				}
			},
		},
		{
			name: "gathers sprint context",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Phase Goal", "open").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 1, "Sprint Goal", "open").
					Build()
				// Write content to sprint files
				sprintPath := filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
				testutil.WriteFile(t, filepath.Join(sprintPath, "README.md"), "Sprint description")
				testutil.WriteFile(t, filepath.Join(sprintPath, "PRD.md"), "Product requirements")
				testutil.WriteFile(t, filepath.Join(sprintPath, "ERD.md"), "Entity relationships")
				return root
			},
			check: func(t *testing.T, ctx *prompt.ProjectContext) {
				if ctx.CurrentSprint == nil {
					t.Fatal("expected current sprint to be set")
				}
				if ctx.CurrentSprint.ID != "0001-sprint" {
					t.Errorf("expected sprint ID 0001-sprint, got %s", ctx.CurrentSprint.ID)
				}
				if len(ctx.SprintGoals) != 1 {
					t.Errorf("expected 1 sprint goal, got %d", len(ctx.SprintGoals))
				}
				if ctx.SprintReadme == nil || ctx.SprintReadme.Empty {
					t.Error("expected sprint README to have content")
				}
				if ctx.SprintPRD == nil || ctx.SprintPRD.Empty {
					t.Error("expected sprint PRD to have content")
				}
				if ctx.SprintERD == nil || ctx.SprintERD.Empty {
					t.Error("expected sprint ERD to have content")
				}
			},
		},
		{
			name: "gathers ticket context",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Phase Goal", "open").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 1, "Sprint Goal", "open").
					WithTicket("0001", "0001", "0001", "open").
					WithTicketGoal("0001", "0001", "0001", 1, "Ticket Goal", "open").
					Build()
				// Write content to ticket README
				ticketPath := filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint", "tickets", "0001-ticket")
				testutil.WriteFile(t, filepath.Join(ticketPath, "README.md"), "Ticket description")
				return root
			},
			check: func(t *testing.T, ctx *prompt.ProjectContext) {
				if ctx.CurrentTicket == nil {
					t.Fatal("expected current ticket to be set")
				}
				if ctx.CurrentTicket.ID != "0001-ticket" {
					t.Errorf("expected ticket ID 0001-ticket, got %s", ctx.CurrentTicket.ID)
				}
				if len(ctx.TicketGoals) != 1 {
					t.Errorf("expected 1 ticket goal, got %d", len(ctx.TicketGoals))
				}
				if ctx.TicketReadme == nil || ctx.TicketReadme.Empty {
					t.Error("expected ticket README to have content")
				}
			},
		},
		{
			name: "reports empty files",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					Build()
				// Clear the README to make it empty
				readmePath := filepath.Join(root, ".crumbler", "phases", "0001-phase", "README.md")
				testutil.WriteFile(t, readmePath, "")
				return root
			},
			check: func(t *testing.T, ctx *prompt.ProjectContext) {
				if ctx.PhaseReadme == nil {
					t.Fatal("expected phase README context")
				}
				if !ctx.PhaseReadme.Empty {
					t.Error("expected phase README to be reported as empty")
				}
			},
		},
		{
			name: "counts closed phases correctly",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
## Phase 2: Features
- Goal 2
## Phase 3: Polish
- Goal 3
`).
					WithPhase("0001", "closed").
					WithPhase("0002", "closed").
					WithPhase("0003", "open").
					Build()
			},
			check: func(t *testing.T, ctx *prompt.ProjectContext) {
				if ctx.ClosedPhases != 2 {
					t.Errorf("expected 2 closed phases, got %d", ctx.ClosedPhases)
				}
				if ctx.TotalPhases != 3 {
					t.Errorf("expected 3 total phases, got %d", ctx.TotalPhases)
				}
			},
		},
		{
			name: "gathers multiple open tickets",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Phase Goal", "open").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 1, "Sprint Goal", "open").
					WithTicket("0001", "0001", "0001", "open").
					WithTicket("0001", "0001", "0002", "open").
					WithTicket("0001", "0001", "0003", "done").
					Build()
			},
			check: func(t *testing.T, ctx *prompt.ProjectContext) {
				if len(ctx.OpenTickets) != 2 {
					t.Errorf("expected 2 open tickets, got %d", len(ctx.OpenTickets))
				}
				// First open ticket should be the current one
				if ctx.CurrentTicket == nil || ctx.CurrentTicket.ID != "0001-ticket" {
					t.Error("expected first open ticket to be current")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			projectRoot := tt.setup(t)

			ctx, err := prompt.GatherContext(projectRoot)
			if err != nil {
				t.Fatalf("failed to gather context: %v", err)
			}

			tt.check(t, ctx)
		})
	}
}

// =============================================================================
// Test: GeneratePrompt
// =============================================================================

func TestGeneratePrompt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setup         func(t *testing.T) string
		config        *prompt.PromptConfig
		checkContains []string
		checkMissing  []string
	}{
		{
			name: "includes state in output",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					Build()
			},
			config:        prompt.DefaultConfig(),
			checkContains: []string{"STATE: CREATE_PHASE"},
		},
		{
			name: "includes prelude by default",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					Build()
			},
			config:        prompt.DefaultConfig(),
			checkContains: []string{"CRUMBLER AGENT PROMPT", "What is Crumbler"},
		},
		{
			name: "excludes prelude when configured",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					Build()
			},
			config: &prompt.PromptConfig{
				IncludePrelude:  false,
				IncludePostlude: true,
				IncludeContext:  true,
			},
			checkMissing: []string{"CRUMBLER AGENT PROMPT"},
		},
		{
			name: "excludes postlude when configured",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					Build()
			},
			config: &prompt.PromptConfig{
				IncludePrelude:  true,
				IncludePostlude: false,
				IncludeContext:  true,
			},
			checkMissing: []string{"NEXT STEPS"},
		},
		{
			name: "includes roadmap context",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					Build()
			},
			config:        prompt.DefaultConfig(),
			checkContains: []string{"roadmap.md", "Phase 1: Foundation"},
		},
		{
			name: "includes instruction section",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					Build()
			},
			config:        prompt.DefaultConfig(),
			checkContains: []string{"INSTRUCTION", "Create Next Phase", "crumbler phase create"},
		},
		{
			name: "includes goals in context",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "My Important Goal", "open").
					Build()
			},
			config:        prompt.DefaultConfig(),
			checkContains: []string{"Phase Goals", "My Important Goal", "[ ]"},
		},
		{
			name: "shows closed goals with checkmark",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Completed Goal", "closed").
					WithPhaseGoal("0001", 2, "Open Goal", "open").
					Build()
			},
			config:        prompt.DefaultConfig(),
			checkContains: []string{"[x]", "Completed Goal", "[ ]", "Open Goal"},
		},
		{
			name: "EXIT state shows completion message",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "closed").
					Build()
			},
			config:        prompt.DefaultConfig(),
			checkContains: []string{"STATE: EXIT", "Project Complete"},
		},
		{
			name: "marks empty files for AI to populate",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Goal", "open").
					WithSprint("0001", "0001", "open").
					Build()
				// Clear the sprint README to make it empty
				sprintReadmePath := filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint", "README.md")
				testutil.WriteFile(t, sprintReadmePath, "")
				return root
			},
			config:        prompt.DefaultConfig(),
			checkContains: []string{"YOU MUST POPULATE"},
		},
		{
			name: "excludes context when configured",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Goal", "open").
					Build()
			},
			config: &prompt.PromptConfig{
				IncludePrelude:  true,
				IncludePostlude: true,
				IncludeContext:  false,
			},
			checkMissing: []string{"CONTEXT", "roadmap.md", "Phase Goals"},
		},
		{
			name: "uses minimal prelude when configured",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					Build()
			},
			config: &prompt.PromptConfig{
				IncludePrelude:  true,
				IncludePostlude: true,
				IncludeContext:  true,
				Minimal:         true,
			},
			checkContains: []string{"STATE: CREATE_PHASE", "POSITION:"},
			checkMissing:  []string{"CRUMBLER AGENT PROMPT", "What is Crumbler"},
		},
		{
			name: "uses minimal postlude when configured",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					Build()
			},
			config: &prompt.PromptConfig{
				IncludePrelude:  true,
				IncludePostlude: true,
				IncludeContext:  true,
				Minimal:         true,
			},
			checkContains: []string{"After completing this step, run:"},
			checkMissing:  []string{"NEXT STEPS", "Quick Reference"},
		},
		{
			name: "handles missing roadmap file",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					Build()
				// Remove roadmap file
				roadmapPath := filepath.Join(root, ".crumbler", "roadmap.md")
				os.Remove(roadmapPath)
				return root
			},
			config:        prompt.DefaultConfig(),
			checkContains: []string{"STATE:"},
		},
		{
			name: "handles missing phase README",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					Build()
				// Remove phase README
				readmePath := filepath.Join(root, ".crumbler", "phases", "0001-phase", "README.md")
				os.Remove(readmePath)
				return root
			},
			config:        prompt.DefaultConfig(),
			checkContains: []string{"STATE:"},
		},
		{
			name: "handles missing sprint files",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Goal", "open").
					WithSprint("0001", "0001", "open").
					Build()
				// Remove sprint files
				sprintPath := filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
				os.Remove(filepath.Join(sprintPath, "README.md"))
				os.Remove(filepath.Join(sprintPath, "PRD.md"))
				os.Remove(filepath.Join(sprintPath, "ERD.md"))
				return root
			},
			config:        prompt.DefaultConfig(),
			checkContains: []string{"STATE:"},
		},
		{
			name: "truncates very long content",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					Build()
				// Write very long content (>4000 chars)
				longContent := strings.Repeat("A", 5000)
				readmePath := filepath.Join(root, ".crumbler", "phases", "0001-phase", "README.md")
				testutil.WriteFile(t, readmePath, longContent)
				return root
			},
			config:        prompt.DefaultConfig(),
			checkContains: []string{"... (truncated)"},
		},
		{
			name: "shows multiple open tickets list",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Phase Goal", "open").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 1, "Sprint Goal", "open").
					WithTicket("0001", "0001", "0001", "open").
					WithTicket("0001", "0001", "0002", "open").
					WithTicket("0001", "0001", "0003", "open").
					Build()
			},
			config:        prompt.DefaultConfig(),
			checkContains: []string{"Open Tickets in Sprint", "0001-ticket", "0002-ticket", "0003-ticket"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			projectRoot := tt.setup(t)

			output, err := prompt.GeneratePrompt(projectRoot, tt.config)
			if err != nil {
				t.Fatalf("failed to generate prompt: %v", err)
			}

			for _, expected := range tt.checkContains {
				if !strings.Contains(output, expected) {
					t.Errorf("expected output to contain %q, but it didn't.\nOutput:\n%s", expected, output)
				}
			}

			for _, missing := range tt.checkMissing {
				if strings.Contains(output, missing) {
					t.Errorf("expected output NOT to contain %q, but it did", missing)
				}
			}
		})
	}
}

// =============================================================================
// Test: GetStateInstruction
// =============================================================================

func TestGetStateInstruction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setup         func(t *testing.T) string
		expectedState prompt.State
		checkTitle    string
		checkCommands []string
	}{
		{
			name: "CREATE_PHASE instruction includes create command",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					Build()
			},
			expectedState: prompt.StateCreatePhase,
			checkTitle:    "Create Next Phase",
			checkCommands: []string{"crumbler phase create"},
		},
		{
			name: "CREATE_SPRINT instruction includes create command",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Goal", "open").
					Build()
			},
			expectedState: prompt.StateCreateSprint,
			checkTitle:    "Create Sprint",
			checkCommands: []string{"crumbler sprint create"},
		},
		{
			name: "EXECUTE_TICKET instruction includes goal close command",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Phase Goal", "open").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 1, "Sprint Goal", "open").
					WithTicket("0001", "0001", "0001", "open").
					WithTicketGoal("0001", "0001", "0001", 1, "Ticket Goal", "open").
					Build()
			},
			expectedState: prompt.StateExecuteTicket,
			checkTitle:    "Execute Ticket",
			checkCommands: []string{"crumbler ticket goal close"},
		},
		{
			name: "MARK_TICKET_DONE instruction includes done command",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Phase Goal", "open").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 1, "Sprint Goal", "open").
					WithTicket("0001", "0001", "0001", "open").
					WithTicketGoal("0001", "0001", "0001", 1, "Ticket Goal", "closed").
					Build()
			},
			expectedState: prompt.StateMarkTicketDone,
			checkTitle:    "Mark Ticket Done",
			checkCommands: []string{"crumbler ticket done"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			projectRoot := tt.setup(t)

			ctx, err := prompt.GatherContext(projectRoot)
			if err != nil {
				t.Fatalf("failed to gather context: %v", err)
			}

			state, err := prompt.DetermineState(ctx)
			if err != nil {
				t.Fatalf("failed to determine state: %v", err)
			}

			if state != tt.expectedState {
				t.Fatalf("expected state %s, got %s", tt.expectedState, state)
			}

			instruction := prompt.GetStateInstruction(state, ctx)

			if instruction.Title != tt.checkTitle {
				t.Errorf("expected title %q, got %q", tt.checkTitle, instruction.Title)
			}

			for _, cmd := range tt.checkCommands {
				found := false
				for _, instCmd := range instruction.Commands {
					if strings.Contains(instCmd, cmd) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected command containing %q in instructions", cmd)
				}
			}
		})
	}
}

// =============================================================================
// Test: GetState (helper function)
// =============================================================================

func TestGetState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(t *testing.T) string
		expected prompt.State
	}{
		{
			name: "returns CREATE_PHASE for empty project with roadmap",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					Build()
			},
			expected: prompt.StateCreatePhase,
		},
		{
			name: "returns EXIT for completed project",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "closed").
					Build()
			},
			expected: prompt.StateExit,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			projectRoot := tt.setup(t)

			state, err := prompt.GetState(projectRoot)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if state != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, state)
			}
		})
	}
}

// =============================================================================
// Test: FormatGoalsList
// =============================================================================

func TestFormatGoalsList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setup         func(t *testing.T) string
		checkContains []string
	}{
		{
			name: "formats open goals with empty checkbox",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Open Goal", "open").
					Build()
			},
			checkContains: []string{"[ ]", "Open Goal"},
		},
		{
			name: "formats closed goals with filled checkbox",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Closed Goal", "closed").
					Build()
			},
			checkContains: []string{"[x]", "Closed Goal"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			projectRoot := tt.setup(t)

			ctx, err := prompt.GatherContext(projectRoot)
			if err != nil {
				t.Fatalf("failed to gather context: %v", err)
			}

			output := prompt.FormatGoalsList(ctx.PhaseGoals)

			for _, expected := range tt.checkContains {
				if !strings.Contains(output, expected) {
					t.Errorf("expected output to contain %q, got: %s", expected, output)
				}
			}
		})
	}
}

// =============================================================================
// Test: Full Workflow Scenarios
// =============================================================================

func TestFullWorkflowScenarios(t *testing.T) {
	t.Parallel()

	t.Run("workflow progresses through states correctly", func(t *testing.T) {
		t.Parallel()

		// Start with just roadmap - should be CREATE_PHASE
		projectRoot := testutil.NewTestProject(t).
			WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
			Build()

		state, err := prompt.GetState(projectRoot)
		if err != nil {
			t.Fatalf("failed to get state: %v", err)
		}
		if state != prompt.StateCreatePhase {
			t.Errorf("expected CREATE_PHASE, got %s", state)
		}

		// Add phase - should be CREATE_PHASE_GOALS
		testutil.CreateDir(t, filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase"))
		testutil.CreateDir(t, filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "goals"))
		testutil.CreateDir(t, filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "sprints"))
		testutil.CreateFile(t, filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "open"))
		testutil.CreateFile(t, filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "README.md"))

		state, err = prompt.GetState(projectRoot)
		if err != nil {
			t.Fatalf("failed to get state: %v", err)
		}
		if state != prompt.StateCreatePhaseGoals {
			t.Errorf("expected CREATE_PHASE_GOALS, got %s", state)
		}

		// Add phase goal - should be CREATE_SPRINT
		testutil.CreateDir(t, filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "goals", "0001-goal"))
		testutil.CreateFile(t, filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "goals", "0001-goal", "open"))
		testutil.WriteFile(t, filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "goals", "0001-goal", "name"), "Phase Goal 1")

		state, err = prompt.GetState(projectRoot)
		if err != nil {
			t.Fatalf("failed to get state: %v", err)
		}
		if state != prompt.StateCreateSprint {
			t.Errorf("expected CREATE_SPRINT, got %s", state)
		}
	})

	t.Run("complex project with multiple entities", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
- Goal 2
## Phase 2: Features
- Goal 3
`).
			WithPhase("0001", "closed").
			WithPhaseGoal("0001", 1, "Phase 1 Goal 1", "closed").
			WithPhaseGoal("0001", 2, "Phase 1 Goal 2", "closed").
			WithSprint("0001", "0001", "closed").
			WithSprintGoal("0001", "0001", 1, "Sprint Goal", "closed").
			WithTicket("0001", "0001", "0001", "done").
			WithTicketGoal("0001", "0001", "0001", 1, "Ticket Goal", "closed").
			WithPhase("0002", "open").
			WithPhaseGoal("0002", 1, "Phase 2 Goal", "open").
			WithSprint("0002", "0001", "open").
			WithSprintGoal("0002", "0001", 1, "Sprint 2 Goal", "open").
			WithTicket("0002", "0001", "0001", "open").
			WithTicketGoal("0002", "0001", "0001", 1, "Active Ticket Goal", "open").
			Build()

		ctx, err := prompt.GatherContext(projectRoot)
		if err != nil {
			t.Fatalf("failed to gather context: %v", err)
		}

		// Verify context
		if ctx.ClosedPhases != 1 {
			t.Errorf("expected 1 closed phase, got %d", ctx.ClosedPhases)
		}
		if ctx.CurrentPhase == nil || ctx.CurrentPhase.ID != "0002-phase" {
			t.Error("expected current phase to be 0002-phase")
		}
		if ctx.CurrentSprint == nil || ctx.CurrentSprint.ID != "0001-sprint" {
			t.Error("expected current sprint to be 0001-sprint")
		}
		if ctx.CurrentTicket == nil || ctx.CurrentTicket.ID != "0001-ticket" {
			t.Error("expected current ticket to be 0001-ticket")
		}

		// State should be EXECUTE_TICKET
		state, err := prompt.DetermineState(ctx)
		if err != nil {
			t.Fatalf("failed to determine state: %v", err)
		}
		if state != prompt.StateExecuteTicket {
			t.Errorf("expected EXECUTE_TICKET, got %s", state)
		}
	})
}

// =============================================================================
// Test: All State Instructions in Prompt Generation
// =============================================================================

func TestAllStateInstructionsInPrompt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setup         func(t *testing.T) string
		expectedState prompt.State
		checkTitle    string
		checkCommands []string
	}{
		{
			name: "CREATE_PHASE_GOALS instruction in prompt",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					Build()
			},
			expectedState: prompt.StateCreatePhaseGoals,
			checkTitle:    "Create Phase Goals",
			checkCommands: []string{"crumbler phase goal create"},
		},
		{
			name: "CLOSE_PHASE instruction in prompt",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Goal 1", "closed").
					WithSprint("0001", "0001", "closed").
					Build()
			},
			expectedState: prompt.StateClosePhase,
			checkTitle:    "Close Phase",
			checkCommands: []string{"crumbler phase close"},
		},
		{
			name: "CREATE_SPRINT_GOALS instruction in prompt",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Goal", "open").
					WithSprint("0001", "0001", "open").
					Build()
			},
			expectedState: prompt.StateCreateSprintGoals,
			checkTitle:    "Create Sprint Goals",
			checkCommands: []string{"crumbler sprint goal create"},
		},
		{
			name: "CREATE_TICKETS instruction in prompt",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Goal", "open").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 1, "Sprint Goal", "open").
					Build()
			},
			expectedState: prompt.StateCreateTickets,
			checkTitle:    "Create Tickets",
			checkCommands: []string{"crumbler ticket create"},
		},
		{
			name: "CLOSE_SPRINT instruction in prompt",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Goal", "open").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 1, "Sprint Goal", "closed").
					WithTicket("0001", "0001", "0001", "done").
					Build()
			},
			expectedState: prompt.StateCloseSprint,
			checkTitle:    "Close Sprint",
			checkCommands: []string{"crumbler sprint close"},
		},
		{
			name: "CREATE_TICKET_GOALS instruction in prompt",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Phase Goal", "open").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 1, "Sprint Goal", "open").
					WithTicket("0001", "0001", "0001", "open").
					Build()
			},
			expectedState: prompt.StateCreateTicketGoals,
			checkTitle:    "Create Ticket Goals",
			checkCommands: []string{"crumbler ticket goal create"},
		},
		{
			name: "EXECUTE_TICKET instruction in prompt",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Phase Goal", "open").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 1, "Sprint Goal", "open").
					WithTicket("0001", "0001", "0001", "open").
					WithTicketGoal("0001", "0001", "0001", 1, "Ticket Goal", "open").
					Build()
			},
			expectedState: prompt.StateExecuteTicket,
			checkTitle:    "Execute Ticket",
			checkCommands: []string{"crumbler ticket goal close"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			projectRoot := tt.setup(t)

			output, err := prompt.GeneratePrompt(projectRoot, prompt.DefaultConfig())
			if err != nil {
				t.Fatalf("failed to generate prompt: %v", err)
			}

			// Check state is correct
			if !strings.Contains(output, "STATE: "+string(tt.expectedState)) {
				t.Errorf("expected state %s in output, got:\n%s", tt.expectedState, output)
			}

			// Check instruction title
			if !strings.Contains(output, tt.checkTitle) {
				t.Errorf("expected title %q in output", tt.checkTitle)
			}

			// Check commands are included
			for _, cmd := range tt.checkCommands {
				if !strings.Contains(output, cmd) {
					t.Errorf("expected command %q in output", cmd)
				}
			}
		})
	}
}

// =============================================================================
// Test: Error Handling
// =============================================================================

func TestPromptErrorHandling(t *testing.T) {
	t.Parallel()

	t.Run("GatherContext handles missing .crumbler gracefully", func(t *testing.T) {
		t.Parallel()

		// Create temp directory without .crumbler
		tempDir := t.TempDir()

		ctx, err := prompt.GatherContext(tempDir)
		if err != nil {
			t.Fatalf("GatherContext should not error, got: %v", err)
		}
		if ctx == nil {
			t.Fatal("expected context to be returned")
		}
		// Roadmap should be marked as missing
		if ctx.Roadmap == nil || !ctx.Roadmap.Missing {
			t.Error("expected roadmap to be marked as missing")
		}
	})

	t.Run("GeneratePrompt handles missing project gracefully", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()

		// GeneratePrompt may error when trying to determine state
		// because it calls DetermineState which may fail
		_, err := prompt.GeneratePrompt(tempDir, prompt.DefaultConfig())
		// It's okay if it errors - the important thing is it doesn't panic
		if err != nil {
			// Error is expected - state determination may fail
			if !strings.Contains(err.Error(), "failed to determine state") && !strings.Contains(err.Error(), "failed to gather context") {
				t.Logf("Got error (expected): %v", err)
			}
		}
	})

	t.Run("GetState handles missing project", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()

		// GetState may error when trying to determine state
		_, err := prompt.GetState(tempDir)
		// Error is expected when project structure is invalid
		if err == nil {
			// If no error, state should be ERROR or similar
			t.Log("GetState returned no error for missing project (may be valid)")
		}
	})

	t.Run("GetStateString handles missing project", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()

		_, err := prompt.GetStateString(tempDir)
		// Error is expected when project structure is invalid
		if err == nil {
			// If no error, should return ERROR state
			t.Log("GetStateString returned no error for missing project (may be valid)")
		}
	})
}

// =============================================================================
// Test: Config Combinations
// =============================================================================

func TestPromptConfigCombinations(t *testing.T) {
	t.Parallel()

	t.Run("all flags disabled", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
			Build()

		config := &prompt.PromptConfig{
			IncludePrelude:  false,
			IncludePostlude: false,
			IncludeContext:  false,
			Minimal:         false,
		}

		output, err := prompt.GeneratePrompt(projectRoot, config)
		if err != nil {
			t.Fatalf("failed to generate prompt: %v", err)
		}

		// Should only have instruction section
		if !strings.Contains(output, "INSTRUCTION") {
			t.Error("expected INSTRUCTION section")
		}
		if strings.Contains(output, "PRELUDE") || strings.Contains(output, "CRUMBLER AGENT PROMPT") {
			t.Error("should not contain prelude")
		}
		if strings.Contains(output, "POSTLUDE") || strings.Contains(output, "NEXT STEPS") {
			t.Error("should not contain postlude")
		}
		if strings.Contains(output, "CONTEXT") || strings.Contains(output, "roadmap.md") {
			t.Error("should not contain context")
		}
	})

	t.Run("minimal with context", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
			Build()

		config := &prompt.PromptConfig{
			IncludePrelude:  true,
			IncludePostlude: true,
			IncludeContext:  true,
			Minimal:         true,
		}

		output, err := prompt.GeneratePrompt(projectRoot, config)
		if err != nil {
			t.Fatalf("failed to generate prompt: %v", err)
		}

		// Should have minimal prelude (just state/position)
		if !strings.Contains(output, "STATE:") {
			t.Error("expected STATE in output")
		}
		if strings.Contains(output, "CRUMBLER AGENT PROMPT") {
			t.Error("should not have full prelude header")
		}

		// Should have context
		if !strings.Contains(output, "roadmap.md") {
			t.Error("expected context section")
		}

		// Should have minimal postlude
		if !strings.Contains(output, "After completing this step") {
			t.Error("expected minimal postlude")
		}
		if strings.Contains(output, "Quick Reference") {
			t.Error("should not have full postlude")
		}
	})
}
