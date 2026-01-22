package query_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/waynenilsen/crumbler/internal/query"
	"github.com/waynenilsen/crumbler/internal/testutil"
)

// =============================================================================
// Test: OpenPhaseExists
// =============================================================================

func TestOpenPhaseExists(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(t *testing.T) string
		expected bool
		wantErr  bool
	}{
		{
			name: "returns true when open phase exists",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithPhase("0001", "open").
					Build()
			},
			expected: true,
			wantErr:  false,
		},
		{
			name: "returns false when no phases exist",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).Build()
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "returns false when all phases are closed",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithPhase("0001", "closed").
					WithPhase("0002", "closed").
					Build()
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "returns true when at least one phase is open among closed phases",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithPhase("0001", "closed").
					WithPhase("0002", "open").
					WithPhase("0003", "closed").
					Build()
			},
			expected: true,
			wantErr:  false,
		},
		{
			name: "returns error when phase has both open and closed files",
			setup: func(t *testing.T) string {
				projectRoot := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					Build()
				// Create invalid state: add closed file
				phasePath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")
				testutil.CreateFile(t, filepath.Join(phasePath, "closed"))
				return projectRoot
			},
			expected: false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			projectRoot := tt.setup(t)

			result, err := query.OpenPhaseExists(projectRoot)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				testutil.AssertNoError(t, err)
				if result != tt.expected {
					t.Errorf("expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

// =============================================================================
// Test: RoadmapComplete
// =============================================================================

func TestRoadmapComplete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(t *testing.T) string
		expected bool
		wantErr  bool
	}{
		{
			name: "returns true when all phases are closed",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithPhase("0001", "closed").
					WithPhase("0002", "closed").
					WithPhase("0003", "closed").
					Build()
			},
			expected: true,
			wantErr:  false,
		},
		{
			name: "returns false when no phases exist",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).Build()
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "returns false when at least one phase is open",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithPhase("0001", "closed").
					WithPhase("0002", "open").
					WithPhase("0003", "closed").
					Build()
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "returns false when any phase is not closed",
			setup: func(t *testing.T) string {
				projectRoot := testutil.NewTestProject(t).
					WithPhase("0001", "closed").
					Build()
				// Create phase without any status file
				phasePath := filepath.Join(projectRoot, ".crumbler", "phases", "0002-phase")
				testutil.CreateDir(t, phasePath)
				return projectRoot
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "returns error when phase has invalid state",
			setup: func(t *testing.T) string {
				projectRoot := testutil.NewTestProject(t).
					WithPhase("0001", "closed").
					Build()
				// Create invalid state
				phasePath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")
				testutil.CreateFile(t, filepath.Join(phasePath, "open"))
				return projectRoot
			},
			expected: false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			projectRoot := tt.setup(t)

			result, err := query.RoadmapComplete(projectRoot)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				testutil.AssertNoError(t, err)
				if result != tt.expected {
					t.Errorf("expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

// =============================================================================
// Test: PhaseGoalsMet
// =============================================================================

func TestPhaseGoalsMet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(t *testing.T) (projectRoot string, phasePath string)
		expected bool
		wantErr  bool
	}{
		{
			name: "returns true when all phase goals closed AND all sprints closed",
			setup: func(t *testing.T) (string, string) {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Goal 1", "closed").
					WithPhaseGoal("0001", 2, "Goal 2", "closed").
					WithSprint("0001", "0001", "closed").
					WithSprint("0001", "0002", "closed")
				projectRoot := builder.Build()
				return projectRoot, builder.PhasePath("0001")
			},
			expected: true,
			wantErr:  false,
		},
		{
			name: "returns false when no goals exist",
			setup: func(t *testing.T) (string, string) {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "closed")
				projectRoot := builder.Build()
				return projectRoot, builder.PhasePath("0001")
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "returns false when no sprints exist",
			setup: func(t *testing.T) (string, string) {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Goal 1", "closed")
				projectRoot := builder.Build()
				return projectRoot, builder.PhasePath("0001")
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "returns false when goals exist but not all closed",
			setup: func(t *testing.T) (string, string) {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Goal 1", "closed").
					WithPhaseGoal("0001", 2, "Goal 2", "open").
					WithSprint("0001", "0001", "closed")
				projectRoot := builder.Build()
				return projectRoot, builder.PhasePath("0001")
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "returns false when sprints exist but not all closed",
			setup: func(t *testing.T) (string, string) {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Goal 1", "closed").
					WithSprint("0001", "0001", "closed").
					WithSprint("0001", "0002", "open")
				projectRoot := builder.Build()
				return projectRoot, builder.PhasePath("0001")
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "returns false when neither goals nor sprints exist",
			setup: func(t *testing.T) (string, string) {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open")
				projectRoot := builder.Build()
				return projectRoot, builder.PhasePath("0001")
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "returns error when goal has invalid state",
			setup: func(t *testing.T) (string, string) {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Goal 1", "closed").
					WithSprint("0001", "0001", "closed")
				projectRoot := builder.Build()
				// Create invalid state: add open file to closed goal
				goalPath := builder.GoalPath(builder.PhasePath("0001"), 1)
				testutil.CreateFile(t, filepath.Join(goalPath, "open"))
				return projectRoot, builder.PhasePath("0001")
			},
			expected: false,
			wantErr:  true,
		},
		{
			name: "returns error when sprint has invalid state",
			setup: func(t *testing.T) (string, string) {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Goal 1", "closed").
					WithSprint("0001", "0001", "closed")
				projectRoot := builder.Build()
				// Create invalid state: add open file to closed sprint
				sprintPath := builder.SprintPath("0001", "0001")
				testutil.CreateFile(t, filepath.Join(sprintPath, "open"))
				return projectRoot, builder.PhasePath("0001")
			},
			expected: false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, phasePath := tt.setup(t)

			result, err := query.PhaseGoalsMet(phasePath)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				testutil.AssertNoError(t, err)
				if result != tt.expected {
					t.Errorf("expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

// =============================================================================
// Test: OpenSprintExists
// =============================================================================

func TestOpenSprintExists(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(t *testing.T) string
		expected bool
		wantErr  bool
	}{
		{
			name: "returns true when open sprint exists",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open")
				builder.Build()
				return builder.PhasePath("0001")
			},
			expected: true,
			wantErr:  false,
		},
		{
			name: "returns false when no sprints exist",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open")
				builder.Build()
				return builder.PhasePath("0001")
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "returns false when all sprints are closed",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "closed").
					WithSprint("0001", "0002", "closed")
				builder.Build()
				return builder.PhasePath("0001")
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "returns true when at least one sprint is open among closed sprints",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "closed").
					WithSprint("0001", "0002", "open").
					WithSprint("0001", "0003", "closed")
				builder.Build()
				return builder.PhasePath("0001")
			},
			expected: true,
			wantErr:  false,
		},
		{
			name: "returns error when sprint has both open and closed files",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open")
				builder.Build()
				// Create invalid state
				sprintPath := builder.SprintPath("0001", "0001")
				testutil.CreateFile(t, filepath.Join(sprintPath, "closed"))
				return builder.PhasePath("0001")
			},
			expected: false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			phasePath := tt.setup(t)

			result, err := query.OpenSprintExists(phasePath)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				testutil.AssertNoError(t, err)
				if result != tt.expected {
					t.Errorf("expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

// =============================================================================
// Test: SprintGoalsMet
// =============================================================================

func TestSprintGoalsMet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(t *testing.T) string
		expected bool
		wantErr  bool
	}{
		{
			name: "returns true when all sprint goals closed AND all tickets done",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 1, "Goal 1", "closed").
					WithSprintGoal("0001", "0001", 2, "Goal 2", "closed").
					WithTicket("0001", "0001", "0001", "done").
					WithTicket("0001", "0001", "0002", "done")
				builder.Build()
				return builder.SprintPath("0001", "0001")
			},
			expected: true,
			wantErr:  false,
		},
		{
			name: "returns false when no goals exist",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "done")
				builder.Build()
				return builder.SprintPath("0001", "0001")
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "returns false when no tickets exist",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 1, "Goal 1", "closed")
				builder.Build()
				return builder.SprintPath("0001", "0001")
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "returns false when goals exist but not all closed",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 1, "Goal 1", "closed").
					WithSprintGoal("0001", "0001", 2, "Goal 2", "open").
					WithTicket("0001", "0001", "0001", "done")
				builder.Build()
				return builder.SprintPath("0001", "0001")
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "returns false when tickets exist but not all done",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 1, "Goal 1", "closed").
					WithTicket("0001", "0001", "0001", "done").
					WithTicket("0001", "0001", "0002", "open")
				builder.Build()
				return builder.SprintPath("0001", "0001")
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "returns false when neither goals nor tickets exist",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open")
				builder.Build()
				return builder.SprintPath("0001", "0001")
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "returns error when ticket has invalid state",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 1, "Goal 1", "closed").
					WithTicket("0001", "0001", "0001", "done")
				builder.Build()
				// Create invalid state
				ticketPath := builder.TicketPath("0001", "0001", "0001")
				testutil.CreateFile(t, filepath.Join(ticketPath, "open"))
				return builder.SprintPath("0001", "0001")
			},
			expected: false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sprintPath := tt.setup(t)

			result, err := query.SprintGoalsMet(sprintPath)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				testutil.AssertNoError(t, err)
				if result != tt.expected {
					t.Errorf("expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

// =============================================================================
// Test: OpenTicketsExist
// =============================================================================

func TestOpenTicketsExist(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(t *testing.T) string
		expected bool
		wantErr  bool
	}{
		{
			name: "returns true when open tickets exist",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "open")
				builder.Build()
				return builder.SprintPath("0001", "0001")
			},
			expected: true,
			wantErr:  false,
		},
		{
			name: "returns false when no tickets exist",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open")
				builder.Build()
				return builder.SprintPath("0001", "0001")
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "returns false when all tickets are done",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "done").
					WithTicket("0001", "0001", "0002", "done")
				builder.Build()
				return builder.SprintPath("0001", "0001")
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "returns true when at least one ticket is open among done tickets",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "done").
					WithTicket("0001", "0001", "0002", "open").
					WithTicket("0001", "0001", "0003", "done")
				builder.Build()
				return builder.SprintPath("0001", "0001")
			},
			expected: true,
			wantErr:  false,
		},
		{
			name: "returns error when ticket has both open and done files",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "open")
				builder.Build()
				// Create invalid state
				ticketPath := builder.TicketPath("0001", "0001", "0001")
				testutil.CreateFile(t, filepath.Join(ticketPath, "done"))
				return builder.SprintPath("0001", "0001")
			},
			expected: false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sprintPath := tt.setup(t)

			result, err := query.OpenTicketsExist(sprintPath)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				testutil.AssertNoError(t, err)
				if result != tt.expected {
					t.Errorf("expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

// =============================================================================
// Test: TicketComplete
// =============================================================================

func TestTicketComplete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(t *testing.T) string
		expected bool
		wantErr  bool
	}{
		{
			name: "returns true when done file exists AND all goals closed",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicketGoal("0001", "0001", "0001", 1, "Goal 1", "closed").
					WithTicketGoal("0001", "0001", "0001", 2, "Goal 2", "closed").
					WithTicket("0001", "0001", "0001", "done") // Set done status after goals
				builder.Build()
				return builder.TicketPath("0001", "0001", "0001")
			},
			expected: true,
			wantErr:  false,
		},
		{
			name: "returns true when done file exists and no goals exist",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "done")
				builder.Build()
				return builder.TicketPath("0001", "0001", "0001")
			},
			expected: true,
			wantErr:  false,
		},
		{
			name: "returns false when done file does not exist",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "open").
					WithTicketGoal("0001", "0001", "0001", 1, "Goal 1", "closed")
				builder.Build()
				return builder.TicketPath("0001", "0001", "0001")
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "returns false when done file exists but goals not all closed",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicketGoal("0001", "0001", "0001", 1, "Goal 1", "closed").
					WithTicketGoal("0001", "0001", "0001", 2, "Goal 2", "open").
					WithTicket("0001", "0001", "0001", "done") // Set done status after goals
				builder.Build()
				return builder.TicketPath("0001", "0001", "0001")
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "returns error when ticket has both open and done files",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "done")
				builder.Build()
				// Create invalid state
				ticketPath := builder.TicketPath("0001", "0001", "0001")
				testutil.CreateFile(t, filepath.Join(ticketPath, "open"))
				return builder.TicketPath("0001", "0001", "0001")
			},
			expected: false,
			wantErr:  true,
		},
		{
			name: "returns error when goal has invalid state",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicketGoal("0001", "0001", "0001", 1, "Goal 1", "closed").
					WithTicket("0001", "0001", "0001", "done") // Set done status after goals
				builder.Build()
				// Create invalid state on goal
				goalPath := builder.GoalPath(builder.TicketPath("0001", "0001", "0001"), 1)
				testutil.CreateFile(t, filepath.Join(goalPath, "open"))
				return builder.TicketPath("0001", "0001", "0001")
			},
			expected: false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ticketPath := tt.setup(t)

			result, err := query.TicketComplete(ticketPath)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				testutil.AssertNoError(t, err)
				if result != tt.expected {
					t.Errorf("expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

// =============================================================================
// Test: PhaseGoalsExist
// =============================================================================

func TestPhaseGoalsExist(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(t *testing.T) string
		expected bool
		wantErr  bool
	}{
		{
			name: "returns true when phase goals exist",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Goal 1", "open")
				builder.Build()
				return builder.PhasePath("0001")
			},
			expected: true,
			wantErr:  false,
		},
		{
			name: "returns false when no phase goals exist",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open")
				builder.Build()
				return builder.PhasePath("0001")
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "returns true with multiple phase goals",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Goal 1", "open").
					WithPhaseGoal("0001", 2, "Goal 2", "closed").
					WithPhaseGoal("0001", 3, "Goal 3", "open")
				builder.Build()
				return builder.PhasePath("0001")
			},
			expected: true,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			phasePath := tt.setup(t)

			result, err := query.PhaseGoalsExist(phasePath)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				testutil.AssertNoError(t, err)
				if result != tt.expected {
					t.Errorf("expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

// =============================================================================
// Test: SprintGoalsExist
// =============================================================================

func TestSprintGoalsExist(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(t *testing.T) string
		expected bool
		wantErr  bool
	}{
		{
			name: "returns true when sprint goals exist",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 1, "Goal 1", "open")
				builder.Build()
				return builder.SprintPath("0001", "0001")
			},
			expected: true,
			wantErr:  false,
		},
		{
			name: "returns false when no sprint goals exist",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open")
				builder.Build()
				return builder.SprintPath("0001", "0001")
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "returns true with multiple sprint goals",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 1, "Goal 1", "open").
					WithSprintGoal("0001", "0001", 2, "Goal 2", "closed")
				builder.Build()
				return builder.SprintPath("0001", "0001")
			},
			expected: true,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sprintPath := tt.setup(t)

			result, err := query.SprintGoalsExist(sprintPath)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				testutil.AssertNoError(t, err)
				if result != tt.expected {
					t.Errorf("expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

// =============================================================================
// Test: TicketGoalsExist
// =============================================================================

func TestTicketGoalsExist(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(t *testing.T) string
		expected bool
		wantErr  bool
	}{
		{
			name: "returns true when ticket goals exist",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "open").
					WithTicketGoal("0001", "0001", "0001", 1, "Goal 1", "open")
				builder.Build()
				return builder.TicketPath("0001", "0001", "0001")
			},
			expected: true,
			wantErr:  false,
		},
		{
			name: "returns false when no ticket goals exist",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "open")
				builder.Build()
				return builder.TicketPath("0001", "0001", "0001")
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "returns true with multiple ticket goals",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "open").
					WithTicketGoal("0001", "0001", "0001", 1, "Goal 1", "open").
					WithTicketGoal("0001", "0001", "0001", 2, "Goal 2", "closed")
				builder.Build()
				return builder.TicketPath("0001", "0001", "0001")
			},
			expected: true,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ticketPath := tt.setup(t)

			result, err := query.TicketGoalsExist(ticketPath)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				testutil.AssertNoError(t, err)
				if result != tt.expected {
					t.Errorf("expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

// =============================================================================
// Test: SprintsExist
// =============================================================================

func TestSprintsExist(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(t *testing.T) string
		expected bool
		wantErr  bool
	}{
		{
			name: "returns true when sprints exist",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open")
				builder.Build()
				return builder.PhasePath("0001")
			},
			expected: true,
			wantErr:  false,
		},
		{
			name: "returns false when no sprints exist",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open")
				builder.Build()
				return builder.PhasePath("0001")
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "returns true with multiple sprints",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithSprint("0001", "0002", "closed").
					WithSprint("0001", "0003", "open")
				builder.Build()
				return builder.PhasePath("0001")
			},
			expected: true,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			phasePath := tt.setup(t)

			result, err := query.SprintsExist(phasePath)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				testutil.AssertNoError(t, err)
				if result != tt.expected {
					t.Errorf("expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

// =============================================================================
// Test: TicketsExist
// =============================================================================

func TestTicketsExist(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(t *testing.T) string
		expected bool
		wantErr  bool
	}{
		{
			name: "returns true when tickets exist",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "open")
				builder.Build()
				return builder.SprintPath("0001", "0001")
			},
			expected: true,
			wantErr:  false,
		},
		{
			name: "returns false when no tickets exist",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open")
				builder.Build()
				return builder.SprintPath("0001", "0001")
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "returns true with multiple tickets",
			setup: func(t *testing.T) string {
				builder := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "open").
					WithTicket("0001", "0001", "0002", "done").
					WithTicket("0001", "0001", "0003", "open")
				builder.Build()
				return builder.SprintPath("0001", "0001")
			},
			expected: true,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sprintPath := tt.setup(t)

			result, err := query.TicketsExist(sprintPath)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				testutil.AssertNoError(t, err)
				if result != tt.expected {
					t.Errorf("expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

// =============================================================================
// Test: Complete Workflow
// =============================================================================

func TestCompleteWorkflow(t *testing.T) {
	t.Parallel()

	t.Run("roadmap to phases to sprints to tickets to done", func(t *testing.T) {
		t.Parallel()

		// Start with empty project - simulate fresh start
		builder := testutil.NewTestProject(t)
		projectRoot := builder.Build()

		// Step 1: Check initial state - no phases exist
		hasOpenPhase, err := query.OpenPhaseExists(projectRoot)
		testutil.AssertNoError(t, err)
		if hasOpenPhase {
			t.Error("expected no open phase initially")
		}

		roadmapComplete, err := query.RoadmapComplete(projectRoot)
		testutil.AssertNoError(t, err)
		if roadmapComplete {
			t.Error("expected roadmap not complete initially")
		}

		// Step 2: Create phase 1 (simulate agent creating phase)
		phasePath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")
		testutil.CreateDir(t, phasePath)
		testutil.CreateDir(t, filepath.Join(phasePath, "goals"))
		testutil.CreateDir(t, filepath.Join(phasePath, "sprints"))
		testutil.CreateFile(t, filepath.Join(phasePath, "open"))

		// Verify open phase exists
		hasOpenPhase, err = query.OpenPhaseExists(projectRoot)
		testutil.AssertNoError(t, err)
		if !hasOpenPhase {
			t.Error("expected open phase after creation")
		}

		// Step 3: Check phase goals met (should be false - no goals/sprints yet)
		phaseGoalsMet, err := query.PhaseGoalsMet(phasePath)
		testutil.AssertNoError(t, err)
		if phaseGoalsMet {
			t.Error("expected phase goals not met without goals and sprints")
		}

		// Step 4: Create phase goal
		goalPath := filepath.Join(phasePath, "goals", "0001-goal")
		testutil.CreateDir(t, goalPath)
		testutil.WriteFile(t, filepath.Join(goalPath, "name"), "Complete phase setup")
		testutil.CreateFile(t, filepath.Join(goalPath, "open"))

		// Verify phase goals exist
		phaseGoalsExist, err := query.PhaseGoalsExist(phasePath)
		testutil.AssertNoError(t, err)
		if !phaseGoalsExist {
			t.Error("expected phase goals to exist")
		}

		// Step 5: Create sprint
		sprintPath := filepath.Join(phasePath, "sprints", "0001-sprint")
		testutil.CreateDir(t, sprintPath)
		testutil.CreateDir(t, filepath.Join(sprintPath, "goals"))
		testutil.CreateDir(t, filepath.Join(sprintPath, "tickets"))
		testutil.CreateFile(t, filepath.Join(sprintPath, "open"))

		// Verify open sprint exists
		hasOpenSprint, err := query.OpenSprintExists(phasePath)
		testutil.AssertNoError(t, err)
		if !hasOpenSprint {
			t.Error("expected open sprint after creation")
		}

		// Step 6: Check sprint goals met (should be false - no goals/tickets yet)
		sprintGoalsMet, err := query.SprintGoalsMet(sprintPath)
		testutil.AssertNoError(t, err)
		if sprintGoalsMet {
			t.Error("expected sprint goals not met without goals and tickets")
		}

		// Step 7: Create sprint goal
		sprintGoalPath := filepath.Join(sprintPath, "goals", "0001-goal")
		testutil.CreateDir(t, sprintGoalPath)
		testutil.WriteFile(t, filepath.Join(sprintGoalPath, "name"), "Complete sprint setup")
		testutil.CreateFile(t, filepath.Join(sprintGoalPath, "open"))

		// Step 8: Create ticket
		ticketPath := filepath.Join(sprintPath, "tickets", "0001-ticket")
		testutil.CreateDir(t, ticketPath)
		testutil.CreateDir(t, filepath.Join(ticketPath, "goals"))
		testutil.CreateFile(t, filepath.Join(ticketPath, "open"))

		// Verify open ticket exists
		hasOpenTickets, err := query.OpenTicketsExist(sprintPath)
		testutil.AssertNoError(t, err)
		if !hasOpenTickets {
			t.Error("expected open tickets after creation")
		}

		// Step 9: Create ticket goal
		ticketGoalPath := filepath.Join(ticketPath, "goals", "0001-goal")
		testutil.CreateDir(t, ticketGoalPath)
		testutil.WriteFile(t, filepath.Join(ticketGoalPath, "name"), "Implement feature")
		testutil.CreateFile(t, filepath.Join(ticketGoalPath, "open"))

		// Step 10: Close ticket goal and mark ticket done
		os.Remove(filepath.Join(ticketGoalPath, "open"))
		testutil.CreateFile(t, filepath.Join(ticketGoalPath, "closed"))

		os.Remove(filepath.Join(ticketPath, "open"))
		testutil.CreateFile(t, filepath.Join(ticketPath, "done"))

		// Verify ticket is complete
		ticketComplete, err := query.TicketComplete(ticketPath)
		testutil.AssertNoError(t, err)
		if !ticketComplete {
			t.Error("expected ticket to be complete")
		}

		// Verify no open tickets
		hasOpenTickets, err = query.OpenTicketsExist(sprintPath)
		testutil.AssertNoError(t, err)
		if hasOpenTickets {
			t.Error("expected no open tickets after completion")
		}

		// Step 11: Close sprint goal and sprint
		os.Remove(filepath.Join(sprintGoalPath, "open"))
		testutil.CreateFile(t, filepath.Join(sprintGoalPath, "closed"))

		// Verify sprint goals met
		sprintGoalsMet, err = query.SprintGoalsMet(sprintPath)
		testutil.AssertNoError(t, err)
		if !sprintGoalsMet {
			t.Error("expected sprint goals met")
		}

		os.Remove(filepath.Join(sprintPath, "open"))
		testutil.CreateFile(t, filepath.Join(sprintPath, "closed"))

		// Step 12: Close phase goal and phase
		os.Remove(filepath.Join(goalPath, "open"))
		testutil.CreateFile(t, filepath.Join(goalPath, "closed"))

		// Verify phase goals met
		phaseGoalsMet, err = query.PhaseGoalsMet(phasePath)
		testutil.AssertNoError(t, err)
		if !phaseGoalsMet {
			t.Error("expected phase goals met")
		}

		os.Remove(filepath.Join(phasePath, "open"))
		testutil.CreateFile(t, filepath.Join(phasePath, "closed"))

		// Step 13: Verify roadmap complete
		roadmapComplete, err = query.RoadmapComplete(projectRoot)
		testutil.AssertNoError(t, err)
		if !roadmapComplete {
			t.Error("expected roadmap complete")
		}
	})
}

// =============================================================================
// Test: Decision Point Logic
// =============================================================================

func TestDecisionPointLogic(t *testing.T) {
	t.Parallel()

	t.Run("CHECK_PHASE decision flow", func(t *testing.T) {
		t.Parallel()

		// Test: No open phase + roadmap not complete = CREATE_PHASE
		t.Run("no open phase and roadmap not complete leads to CREATE_PHASE", func(t *testing.T) {
			t.Parallel()

			projectRoot := testutil.NewTestProject(t).Build()

			openPhase, err := query.OpenPhaseExists(projectRoot)
			testutil.AssertNoError(t, err)
			if openPhase {
				t.Error("expected no open phase")
			}

			roadmapComplete, err := query.RoadmapComplete(projectRoot)
			testutil.AssertNoError(t, err)
			if roadmapComplete {
				t.Error("expected roadmap not complete")
			}

			// Decision: CREATE_PHASE
		})

		// Test: No open phase + roadmap complete = EXIT
		t.Run("no open phase and roadmap complete leads to EXIT", func(t *testing.T) {
			t.Parallel()

			projectRoot := testutil.NewTestProject(t).
				WithPhase("0001", "closed").
				WithPhase("0002", "closed").
				Build()

			openPhase, err := query.OpenPhaseExists(projectRoot)
			testutil.AssertNoError(t, err)
			if openPhase {
				t.Error("expected no open phase")
			}

			roadmapComplete, err := query.RoadmapComplete(projectRoot)
			testutil.AssertNoError(t, err)
			if !roadmapComplete {
				t.Error("expected roadmap complete")
			}

			// Decision: EXIT
		})

		// Test: Open phase exists = CHECK_SPRINT
		t.Run("open phase exists leads to CHECK_SPRINT", func(t *testing.T) {
			t.Parallel()

			builder := testutil.NewTestProject(t).
				WithPhase("0001", "open")
			projectRoot := builder.Build()

			openPhase, err := query.OpenPhaseExists(projectRoot)
			testutil.AssertNoError(t, err)
			if !openPhase {
				t.Error("expected open phase")
			}

			// Decision: CHECK_SPRINT (continue to sprint-level checks)
		})
	})

	t.Run("CHECK_SPRINT decision flow", func(t *testing.T) {
		t.Parallel()

		// Test: No open sprint + phase goals met = CLOSE_PHASE
		t.Run("no open sprint and phase goals met leads to CLOSE_PHASE", func(t *testing.T) {
			t.Parallel()

			builder := testutil.NewTestProject(t).
				WithPhase("0001", "open").
				WithPhaseGoal("0001", 1, "Goal 1", "closed").
				WithSprint("0001", "0001", "closed")
			builder.Build()
			phasePath := builder.PhasePath("0001")

			openSprint, err := query.OpenSprintExists(phasePath)
			testutil.AssertNoError(t, err)
			if openSprint {
				t.Error("expected no open sprint")
			}

			phaseGoalsMet, err := query.PhaseGoalsMet(phasePath)
			testutil.AssertNoError(t, err)
			if !phaseGoalsMet {
				t.Error("expected phase goals met")
			}

			// Decision: CLOSE_PHASE
		})

		// Test: No open sprint + no phase goals = CREATE_GOALS
		t.Run("no open sprint and no phase goals leads to CREATE_GOALS", func(t *testing.T) {
			t.Parallel()

			builder := testutil.NewTestProject(t).
				WithPhase("0001", "open")
			builder.Build()
			phasePath := builder.PhasePath("0001")

			openSprint, err := query.OpenSprintExists(phasePath)
			testutil.AssertNoError(t, err)
			if openSprint {
				t.Error("expected no open sprint")
			}

			phaseGoalsExist, err := query.PhaseGoalsExist(phasePath)
			testutil.AssertNoError(t, err)
			if phaseGoalsExist {
				t.Error("expected no phase goals")
			}

			// Decision: CREATE_GOALS
		})

		// Test: No open sprint + phase goals exist but not met + no sprints = CREATE_SPRINT
		t.Run("no open sprint and goals exist but no sprints leads to CREATE_SPRINT", func(t *testing.T) {
			t.Parallel()

			builder := testutil.NewTestProject(t).
				WithPhase("0001", "open").
				WithPhaseGoal("0001", 1, "Goal 1", "open")
			builder.Build()
			phasePath := builder.PhasePath("0001")

			openSprint, err := query.OpenSprintExists(phasePath)
			testutil.AssertNoError(t, err)
			if openSprint {
				t.Error("expected no open sprint")
			}

			phaseGoalsMet, err := query.PhaseGoalsMet(phasePath)
			testutil.AssertNoError(t, err)
			if phaseGoalsMet {
				t.Error("expected phase goals not met")
			}

			sprintsExist, err := query.SprintsExist(phasePath)
			testutil.AssertNoError(t, err)
			if sprintsExist {
				t.Error("expected no sprints")
			}

			// Decision: CREATE_SPRINT
		})

		// Test: Open sprint exists = CHECK_TICKETS
		t.Run("open sprint exists leads to CHECK_TICKETS", func(t *testing.T) {
			t.Parallel()

			builder := testutil.NewTestProject(t).
				WithPhase("0001", "open").
				WithSprint("0001", "0001", "open")
			builder.Build()
			phasePath := builder.PhasePath("0001")

			openSprint, err := query.OpenSprintExists(phasePath)
			testutil.AssertNoError(t, err)
			if !openSprint {
				t.Error("expected open sprint")
			}

			// Decision: CHECK_TICKETS (continue to ticket-level checks)
		})
	})

	t.Run("CHECK_TICKETS decision flow", func(t *testing.T) {
		t.Parallel()

		// Test: No open tickets + sprint goals met = CLOSE_SPRINT
		t.Run("no open tickets and sprint goals met leads to CLOSE_SPRINT", func(t *testing.T) {
			t.Parallel()

			builder := testutil.NewTestProject(t).
				WithPhase("0001", "open").
				WithSprint("0001", "0001", "open").
				WithSprintGoal("0001", "0001", 1, "Goal 1", "closed").
				WithTicket("0001", "0001", "0001", "done")
			builder.Build()
			sprintPath := builder.SprintPath("0001", "0001")

			openTickets, err := query.OpenTicketsExist(sprintPath)
			testutil.AssertNoError(t, err)
			if openTickets {
				t.Error("expected no open tickets")
			}

			sprintGoalsMet, err := query.SprintGoalsMet(sprintPath)
			testutil.AssertNoError(t, err)
			if !sprintGoalsMet {
				t.Error("expected sprint goals met")
			}

			// Decision: CLOSE_SPRINT
		})

		// Test: No open tickets + no sprint goals = CREATE_GOALS
		t.Run("no open tickets and no sprint goals leads to CREATE_GOALS", func(t *testing.T) {
			t.Parallel()

			builder := testutil.NewTestProject(t).
				WithPhase("0001", "open").
				WithSprint("0001", "0001", "open")
			builder.Build()
			sprintPath := builder.SprintPath("0001", "0001")

			openTickets, err := query.OpenTicketsExist(sprintPath)
			testutil.AssertNoError(t, err)
			if openTickets {
				t.Error("expected no open tickets")
			}

			sprintGoalsExist, err := query.SprintGoalsExist(sprintPath)
			testutil.AssertNoError(t, err)
			if sprintGoalsExist {
				t.Error("expected no sprint goals")
			}

			// Decision: CREATE_GOALS
		})

		// Test: No open tickets + sprint goals exist but not met + no tickets = CREATE_TICKETS
		t.Run("no open tickets and goals exist but no tickets leads to CREATE_TICKETS", func(t *testing.T) {
			t.Parallel()

			builder := testutil.NewTestProject(t).
				WithPhase("0001", "open").
				WithSprint("0001", "0001", "open").
				WithSprintGoal("0001", "0001", 1, "Goal 1", "open")
			builder.Build()
			sprintPath := builder.SprintPath("0001", "0001")

			openTickets, err := query.OpenTicketsExist(sprintPath)
			testutil.AssertNoError(t, err)
			if openTickets {
				t.Error("expected no open tickets")
			}

			sprintGoalsMet, err := query.SprintGoalsMet(sprintPath)
			testutil.AssertNoError(t, err)
			if sprintGoalsMet {
				t.Error("expected sprint goals not met")
			}

			ticketsExist, err := query.TicketsExist(sprintPath)
			testutil.AssertNoError(t, err)
			if ticketsExist {
				t.Error("expected no tickets")
			}

			// Decision: CREATE_TICKETS
		})

		// Test: Open tickets exist = EXECUTE
		t.Run("open tickets exist leads to EXECUTE", func(t *testing.T) {
			t.Parallel()

			builder := testutil.NewTestProject(t).
				WithPhase("0001", "open").
				WithSprint("0001", "0001", "open").
				WithTicket("0001", "0001", "0001", "open")
			builder.Build()
			sprintPath := builder.SprintPath("0001", "0001")

			openTickets, err := query.OpenTicketsExist(sprintPath)
			testutil.AssertNoError(t, err)
			if !openTickets {
				t.Error("expected open tickets")
			}

			// Decision: EXECUTE (execute the ticket)
		})
	})

	t.Run("TICKET_DONE decision", func(t *testing.T) {
		t.Parallel()

		// Test: Ticket not complete = continue EXECUTE
		t.Run("ticket not complete continues EXECUTE", func(t *testing.T) {
			t.Parallel()

			builder := testutil.NewTestProject(t).
				WithPhase("0001", "open").
				WithSprint("0001", "0001", "open").
				WithTicket("0001", "0001", "0001", "open").
				WithTicketGoal("0001", "0001", "0001", 1, "Goal 1", "open")
			builder.Build()
			ticketPath := builder.TicketPath("0001", "0001", "0001")

			ticketComplete, err := query.TicketComplete(ticketPath)
			testutil.AssertNoError(t, err)
			if ticketComplete {
				t.Error("expected ticket not complete")
			}

			// Decision: Continue EXECUTE
		})

		// Test: Ticket complete = MARK_DONE
		t.Run("ticket complete leads to MARK_DONE", func(t *testing.T) {
			t.Parallel()

			builder := testutil.NewTestProject(t).
				WithPhase("0001", "open").
				WithSprint("0001", "0001", "open").
				WithTicketGoal("0001", "0001", "0001", 1, "Goal 1", "closed").
				WithTicket("0001", "0001", "0001", "done") // Set done status after goals
			builder.Build()
			ticketPath := builder.TicketPath("0001", "0001", "0001")

			ticketComplete, err := query.TicketComplete(ticketPath)
			testutil.AssertNoError(t, err)
			if !ticketComplete {
				t.Error("expected ticket complete")
			}

			// Decision: MARK_DONE (ticket already marked done in this test setup)
		})
	})
}

// =============================================================================
// Test: Edge Cases
// =============================================================================

func TestEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("non-existent project root", func(t *testing.T) {
		t.Parallel()

		nonExistentPath := "/non/existent/path"

		result, err := query.OpenPhaseExists(nonExistentPath)
		testutil.AssertNoError(t, err)
		if result {
			t.Error("expected false for non-existent path")
		}

		result, err = query.RoadmapComplete(nonExistentPath)
		testutil.AssertNoError(t, err)
		if result {
			t.Error("expected false for non-existent path")
		}
	})

	t.Run("non-existent phase path", func(t *testing.T) {
		t.Parallel()

		nonExistentPath := "/non/existent/phase/path"

		result, err := query.OpenSprintExists(nonExistentPath)
		testutil.AssertNoError(t, err)
		if result {
			t.Error("expected false for non-existent path")
		}

		result, err = query.PhaseGoalsMet(nonExistentPath)
		testutil.AssertNoError(t, err)
		if result {
			t.Error("expected false for non-existent path")
		}

		result, err = query.PhaseGoalsExist(nonExistentPath)
		testutil.AssertNoError(t, err)
		if result {
			t.Error("expected false for non-existent path")
		}

		result, err = query.SprintsExist(nonExistentPath)
		testutil.AssertNoError(t, err)
		if result {
			t.Error("expected false for non-existent path")
		}
	})

	t.Run("non-existent sprint path", func(t *testing.T) {
		t.Parallel()

		nonExistentPath := "/non/existent/sprint/path"

		result, err := query.OpenTicketsExist(nonExistentPath)
		testutil.AssertNoError(t, err)
		if result {
			t.Error("expected false for non-existent path")
		}

		result, err = query.SprintGoalsMet(nonExistentPath)
		testutil.AssertNoError(t, err)
		if result {
			t.Error("expected false for non-existent path")
		}

		result, err = query.SprintGoalsExist(nonExistentPath)
		testutil.AssertNoError(t, err)
		if result {
			t.Error("expected false for non-existent path")
		}

		result, err = query.TicketsExist(nonExistentPath)
		testutil.AssertNoError(t, err)
		if result {
			t.Error("expected false for non-existent path")
		}
	})

	t.Run("non-existent ticket path", func(t *testing.T) {
		t.Parallel()

		nonExistentPath := "/non/existent/ticket/path"

		result, err := query.TicketComplete(nonExistentPath)
		testutil.AssertNoError(t, err)
		if result {
			t.Error("expected false for non-existent path")
		}

		result, err = query.TicketGoalsExist(nonExistentPath)
		testutil.AssertNoError(t, err)
		if result {
			t.Error("expected false for non-existent path")
		}
	})

	t.Run("empty goals directory", func(t *testing.T) {
		t.Parallel()

		builder := testutil.NewTestProject(t).
			WithPhase("0001", "open")
		projectRoot := builder.Build()
		phasePath := builder.PhasePath("0001")

		// Goals directory exists but is empty
		goalsDir := filepath.Join(phasePath, "goals")
		if _, err := os.Stat(goalsDir); os.IsNotExist(err) {
			testutil.CreateDir(t, goalsDir)
		}

		result, err := query.PhaseGoalsExist(phasePath)
		testutil.AssertNoError(t, err)
		if result {
			t.Error("expected no goals in empty directory")
		}

		// Verify the project was created correctly
		if projectRoot == "" {
			t.Error("expected project root to be set")
		}
	})

	t.Run("file in goals directory instead of directory", func(t *testing.T) {
		t.Parallel()

		builder := testutil.NewTestProject(t).
			WithPhase("0001", "open")
		builder.Build()
		phasePath := builder.PhasePath("0001")

		// Create a file in goals directory instead of a goal directory
		goalsDir := filepath.Join(phasePath, "goals")
		testutil.WriteFile(t, filepath.Join(goalsDir, "some-file.txt"), "not a goal")

		result, err := query.PhaseGoalsExist(phasePath)
		testutil.AssertNoError(t, err)
		if result {
			t.Error("expected files to not count as goals")
		}
	})

	t.Run("multiple phases with mixed states", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithPhase("0001", "closed").
			WithPhase("0002", "open").
			WithPhase("0003", "closed").
			WithPhase("0004", "open").
			WithPhase("0005", "closed").
			Build()

		// Should find open phase
		result, err := query.OpenPhaseExists(projectRoot)
		testutil.AssertNoError(t, err)
		if !result {
			t.Error("expected to find open phase among mixed states")
		}

		// Roadmap should not be complete
		complete, err := query.RoadmapComplete(projectRoot)
		testutil.AssertNoError(t, err)
		if complete {
			t.Error("expected roadmap not complete with open phases")
		}
	})

	t.Run("deep hierarchy with all levels", func(t *testing.T) {
		t.Parallel()

		builder := testutil.NewTestProject(t).
			WithPhase("0001", "open").
			WithPhaseGoal("0001", 1, "Phase Goal 1", "open").
			WithPhaseGoal("0001", 2, "Phase Goal 2", "closed").
			WithSprint("0001", "0001", "open").
			WithSprintGoal("0001", "0001", 1, "Sprint Goal 1", "open").
			WithTicket("0001", "0001", "0001", "open").
			WithTicketGoal("0001", "0001", "0001", 1, "Ticket Goal 1", "open").
			WithTicketGoal("0001", "0001", "0002", 1, "Ticket 2 Goal 1", "closed").
			WithTicket("0001", "0001", "0002", "done") // Set done status after goals
		builder.Build()

		phasePath := builder.PhasePath("0001")
		sprintPath := builder.SprintPath("0001", "0001")

		// Test all query functions work correctly in deep hierarchy
		phaseGoalsExist, err := query.PhaseGoalsExist(phasePath)
		testutil.AssertNoError(t, err)
		if !phaseGoalsExist {
			t.Error("expected phase goals to exist")
		}

		sprintGoalsExist, err := query.SprintGoalsExist(sprintPath)
		testutil.AssertNoError(t, err)
		if !sprintGoalsExist {
			t.Error("expected sprint goals to exist")
		}

		openTickets, err := query.OpenTicketsExist(sprintPath)
		testutil.AssertNoError(t, err)
		if !openTickets {
			t.Error("expected open tickets")
		}

		ticketsExist, err := query.TicketsExist(sprintPath)
		testutil.AssertNoError(t, err)
		if !ticketsExist {
			t.Error("expected tickets to exist")
		}

		// Ticket 2 should be complete
		ticket2Path := builder.TicketPath("0001", "0001", "0002")
		ticket2Complete, err := query.TicketComplete(ticket2Path)
		testutil.AssertNoError(t, err)
		if !ticket2Complete {
			t.Error("expected ticket 2 to be complete")
		}

		// Ticket 1 should not be complete
		ticket1Path := builder.TicketPath("0001", "0001", "0001")
		ticket1Complete, err := query.TicketComplete(ticket1Path)
		testutil.AssertNoError(t, err)
		if ticket1Complete {
			t.Error("expected ticket 1 to not be complete")
		}
	})
}
