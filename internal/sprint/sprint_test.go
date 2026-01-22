package sprint

import (
	"path/filepath"
	"testing"

	"github.com/waynenilsen/crumbler/internal/models"
	"github.com/waynenilsen/crumbler/internal/testutil"
)

// TestCreateSprint tests creating a sprint in a phase.
func TestCreateSprint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		sprintIdx  int
		wantID     string
		wantStatus string
	}{
		{
			name:       "create first sprint",
			sprintIdx:  1,
			wantID:     "0001-sprint",
			wantStatus: "open",
		},
		{
			name:       "create sprint with high index",
			sprintIdx:  42,
			wantID:     "0042-sprint",
			wantStatus: "open",
		},
		{
			name:       "create sprint with max 4-digit index",
			sprintIdx:  9999,
			wantID:     "9999-sprint",
			wantStatus: "open",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup: Create a project with a phase
			projectRoot := testutil.NewTestProject(t).
				WithPhase("0001", "open").
				Build()

			phasePath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")

			// Act: Create a sprint
			sprintPath, err := CreateSprint(phasePath, tt.sprintIdx)
			testutil.AssertNoError(t, err)

			// Assert: Sprint directory exists
			testutil.AssertDirExists(t, sprintPath)

			// Assert: Sprint has correct structure
			testutil.AssertFileExists(t, filepath.Join(sprintPath, "README.md"))
			testutil.AssertFileExists(t, filepath.Join(sprintPath, "PRD.md"))
			testutil.AssertFileExists(t, filepath.Join(sprintPath, "ERD.md"))
			testutil.AssertDirExists(t, filepath.Join(sprintPath, "goals"))
			testutil.AssertDirExists(t, filepath.Join(sprintPath, "tickets"))

			// Assert: Sprint is open
			testutil.AssertStatus(t, sprintPath, tt.wantStatus)

			// Assert: Sprint path contains expected ID
			if filepath.Base(sprintPath) != tt.wantID {
				t.Errorf("expected sprint ID %s, got %s", tt.wantID, filepath.Base(sprintPath))
			}
		})
	}
}

// TestCreateSprintGoal tests creating sprint goals.
func TestCreateSprintGoal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		goalIndex int
		goalName  string
		wantID    string
	}{
		{
			name:      "create first goal",
			goalIndex: 1,
			goalName:  "Complete sprint planning",
			wantID:    "0001-goal",
		},
		{
			name:      "create goal with high index",
			goalIndex: 100,
			goalName:  "Implement feature X",
			wantID:    "0100-goal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup: Create project with phase and sprint
			projectRoot := testutil.NewTestProject(t).
				WithPhase("0001", "open").
				WithSprint("0001", "0001", "open").
				Build()

			sprintPath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")

			// Act: Create a sprint goal
			goalPath, err := CreateSprintGoal(sprintPath, tt.goalIndex, tt.goalName)
			testutil.AssertNoError(t, err)

			// Assert: Goal directory exists
			testutil.AssertDirExists(t, goalPath)

			// Assert: Goal has correct ID
			if filepath.Base(goalPath) != tt.wantID {
				t.Errorf("expected goal ID %s, got %s", tt.wantID, filepath.Base(goalPath))
			}

			// Assert: Goal is open
			testutil.AssertGoalStatus(t, goalPath, "open")

			// Assert: Goal has correct name
			testutil.AssertGoalName(t, goalPath, tt.goalName)
		})
	}
}

// TestCloseSprintSuccess tests closing a sprint when all conditions are met.
func TestCloseSprintSuccess(t *testing.T) {
	t.Parallel()

	// Setup: Create a project with a sprint that has closed goals and done tickets
	projectRoot := testutil.NewTestProject(t).
		WithPhase("0001", "open").
		WithSprint("0001", "0001", "open").
		WithSprintGoal("0001", "0001", 1, "Goal 1", "closed").
		WithSprintGoal("0001", "0001", 2, "Goal 2", "closed").
		WithTicket("0001", "0001", "0001", "done").
		WithTicket("0001", "0001", "0002", "done").
		Build()

	sprintPath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")

	// Act: Close the sprint
	err := CloseSprint(sprintPath)
	testutil.AssertNoError(t, err)

	// Assert: Sprint is now closed
	testutil.AssertStatus(t, sprintPath, "closed")
}

// TestCloseSprintWithOpenTickets tests error when closing sprint with open tickets.
func TestCloseSprintWithOpenTickets(t *testing.T) {
	t.Parallel()

	// Setup: Create a project with a sprint that has open tickets
	projectRoot := testutil.NewTestProject(t).
		WithPhase("0001", "open").
		WithSprint("0001", "0001", "open").
		WithSprintGoal("0001", "0001", 1, "Goal 1", "closed").
		WithTicket("0001", "0001", "0001", "open"). // Open ticket
		WithTicket("0001", "0001", "0002", "done").
		Build()

	sprintPath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")

	// Act: Try to close the sprint
	err := CloseSprint(sprintPath)

	// Assert: Error should be returned
	if err == nil {
		t.Fatal("expected error when closing sprint with open tickets, got nil")
	}
	testutil.AssertError(t, err, "tickets still open")

	// Assert: Sprint is still open
	testutil.AssertStatus(t, sprintPath, "open")
}

// TestCloseSprintWithOpenGoals tests error when closing sprint with open goals.
func TestCloseSprintWithOpenGoals(t *testing.T) {
	t.Parallel()

	// Setup: Create a project with a sprint that has open goals
	projectRoot := testutil.NewTestProject(t).
		WithPhase("0001", "open").
		WithSprint("0001", "0001", "open").
		WithSprintGoal("0001", "0001", 1, "Goal 1", "open"). // Open goal
		WithSprintGoal("0001", "0001", 2, "Goal 2", "closed").
		WithTicket("0001", "0001", "0001", "done").
		Build()

	sprintPath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")

	// Act: Try to close the sprint
	err := CloseSprint(sprintPath)

	// Assert: Error should be returned
	if err == nil {
		t.Fatal("expected error when closing sprint with open goals, got nil")
	}
	testutil.AssertError(t, err, "goals still open")

	// Assert: Sprint is still open
	testutil.AssertStatus(t, sprintPath, "open")
}

// TestGetOpenSprint tests retrieving the open sprint.
func TestGetOpenSprint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setup      func(t *testing.T) (string, string) // returns (projectRoot, phasePath)
		wantSprint bool
		wantID     string
	}{
		{
			name: "finds open sprint",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					Build()
				return root, filepath.Join(root, ".crumbler", "phases", "0001-phase")
			},
			wantSprint: true,
			wantID:     "0001-sprint",
		},
		{
			name: "no open sprint when all closed",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "closed").
					Build()
				return root, filepath.Join(root, ".crumbler", "phases", "0001-phase")
			},
			wantSprint: false,
		},
		{
			name: "finds open sprint among multiple",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "closed").
					WithSprint("0001", "0002", "open").
					WithSprint("0001", "0003", "closed").
					Build()
				return root, filepath.Join(root, ".crumbler", "phases", "0001-phase")
			},
			wantSprint: true,
			wantID:     "0002-sprint",
		},
		{
			name: "no sprints in phase",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					Build()
				return root, filepath.Join(root, ".crumbler", "phases", "0001-phase")
			},
			wantSprint: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, phasePath := tt.setup(t)

			// Act: Get open sprint
			sprint, err := GetOpenSprint(phasePath)
			testutil.AssertNoError(t, err)

			// Assert
			if tt.wantSprint {
				if sprint == nil {
					t.Fatal("expected to find open sprint, got nil")
				}
				if sprint.ID != tt.wantID {
					t.Errorf("expected sprint ID %s, got %s", tt.wantID, sprint.ID)
				}
				if sprint.Status != models.StatusOpen {
					t.Errorf("expected sprint status open, got %s", sprint.Status)
				}
			} else {
				if sprint != nil {
					t.Errorf("expected no open sprint, got %+v", sprint)
				}
			}
		})
	}
}

// TestAreSprintGoalsMet tests sprint goals met detection.
func TestAreSprintGoalsMet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T) string // returns sprintPath
		wantMet bool
	}{
		{
			name: "all goals closed and all tickets done",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 1, "Goal 1", "closed").
					WithSprintGoal("0001", "0001", 2, "Goal 2", "closed").
					WithTicket("0001", "0001", "0001", "done").
					WithTicket("0001", "0001", "0002", "done").
					Build()
				return filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
			},
			wantMet: true,
		},
		{
			name: "one goal still open",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 1, "Goal 1", "closed").
					WithSprintGoal("0001", "0001", 2, "Goal 2", "open"). // Open goal
					WithTicket("0001", "0001", "0001", "done").
					Build()
				return filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
			},
			wantMet: false,
		},
		{
			name: "one ticket still open",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 1, "Goal 1", "closed").
					WithTicket("0001", "0001", "0001", "done").
					WithTicket("0001", "0001", "0002", "open"). // Open ticket
					Build()
				return filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
			},
			wantMet: false,
		},
		{
			name: "no goals exist",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "done").
					Build()
				return filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
			},
			wantMet: false, // No goals means not met (allows CREATE_GOALS)
		},
		{
			name: "no tickets exist",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 1, "Goal 1", "closed").
					Build()
				return filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
			},
			wantMet: false, // No tickets means not met (allows CREATE_TICKETS)
		},
		{
			name: "no goals and no tickets",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					Build()
				return filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
			},
			wantMet: false, // Empty sprint is not met
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sprintPath := tt.setup(t)

			// Act: Check if sprint goals are met
			met, err := AreSprintGoalsMet(sprintPath)
			testutil.AssertNoError(t, err)

			// Assert
			if met != tt.wantMet {
				t.Errorf("expected goals met = %v, got %v", tt.wantMet, met)
			}
		})
	}
}

// TestPRDAndERDFileCreation tests that PRD.md and ERD.md are created.
func TestPRDAndERDFileCreation(t *testing.T) {
	t.Parallel()

	// Setup: Create a project with a phase
	projectRoot := testutil.NewTestProject(t).
		WithPhase("0001", "open").
		Build()

	phasePath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")

	// Act: Create a sprint
	sprintPath, err := CreateSprint(phasePath, 1)
	testutil.AssertNoError(t, err)

	// Assert: PRD.md exists
	prdPath := filepath.Join(sprintPath, "PRD.md")
	testutil.AssertFileExists(t, prdPath)

	// Assert: ERD.md exists
	erdPath := filepath.Join(sprintPath, "ERD.md")
	testutil.AssertFileExists(t, erdPath)
}

// TestGetNextSprintIndex tests GetNextSprintIndex returns correct index.
func TestGetNextSprintIndex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(t *testing.T) string // returns phasePath
		wantIndex int
	}{
		{
			name: "no sprints returns 1",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					Build()
				return filepath.Join(root, ".crumbler", "phases", "0001-phase")
			},
			wantIndex: 1,
		},
		{
			name: "one sprint returns 2",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					Build()
				return filepath.Join(root, ".crumbler", "phases", "0001-phase")
			},
			wantIndex: 2,
		},
		{
			name: "multiple sprints returns max+1",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "closed").
					WithSprint("0001", "0002", "closed").
					WithSprint("0001", "0005", "open"). // Gap in sequence
					Build()
				return filepath.Join(root, ".crumbler", "phases", "0001-phase")
			},
			wantIndex: 6, // max is 5, so next is 6
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			phasePath := tt.setup(t)

			// Act: Get next sprint index
			index, err := GetNextSprintIndex(phasePath)
			testutil.AssertNoError(t, err)

			// Assert
			if index != tt.wantIndex {
				t.Errorf("expected next index %d, got %d", tt.wantIndex, index)
			}
		})
	}
}

// TestGetSprintGoals tests getting sprint goals.
func TestGetSprintGoals(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(t *testing.T) string // returns sprintPath
		wantCount int
		wantNames []string
	}{
		{
			name: "no goals",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					Build()
				return filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
			},
			wantCount: 0,
		},
		{
			name: "multiple goals sorted by index",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 2, "Second Goal", "open").
					WithSprintGoal("0001", "0001", 1, "First Goal", "open").
					WithSprintGoal("0001", "0001", 3, "Third Goal", "closed").
					Build()
				return filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
			},
			wantCount: 3,
			wantNames: []string{"First Goal", "Second Goal", "Third Goal"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sprintPath := tt.setup(t)

			// Act: Get sprint goals
			goals, err := GetSprintGoals(sprintPath)
			testutil.AssertNoError(t, err)

			// Assert: Count
			if len(goals) != tt.wantCount {
				t.Errorf("expected %d goals, got %d", tt.wantCount, len(goals))
			}

			// Assert: Names in order (if specified)
			if tt.wantNames != nil {
				for i, wantName := range tt.wantNames {
					if i < len(goals) && goals[i].Name != wantName {
						t.Errorf("goal[%d] expected name %q, got %q", i, wantName, goals[i].Name)
					}
				}
			}
		})
	}
}

// TestCloseSprintGoal tests closing a sprint goal.
func TestCloseSprintGoal(t *testing.T) {
	t.Parallel()

	// Setup: Create project with sprint and open goal
	projectRoot := testutil.NewTestProject(t).
		WithPhase("0001", "open").
		WithSprint("0001", "0001", "open").
		WithSprintGoal("0001", "0001", 1, "Test Goal", "open").
		Build()

	sprintPath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
	goalPath := filepath.Join(sprintPath, "goals", "0001-goal")

	// Pre-assert: Goal is open
	testutil.AssertGoalStatus(t, goalPath, "open")

	// Act: Close the goal
	err := CloseSprintGoal(sprintPath, "0001-goal")
	testutil.AssertNoError(t, err)

	// Assert: Goal is now closed
	testutil.AssertGoalStatus(t, goalPath, "closed")
}

// TestCloseSprintGoalNotFound tests error when closing non-existent goal.
func TestCloseSprintGoalNotFound(t *testing.T) {
	t.Parallel()

	// Setup: Create project with sprint but no goals
	projectRoot := testutil.NewTestProject(t).
		WithPhase("0001", "open").
		WithSprint("0001", "0001", "open").
		Build()

	sprintPath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")

	// Act: Try to close non-existent goal
	err := CloseSprintGoal(sprintPath, "0001-goal")

	// Assert: Error should be returned
	if err == nil {
		t.Fatal("expected error when closing non-existent goal, got nil")
	}
	testutil.AssertError(t, err, "goal not found")
}

// TestValidateSprintState tests sprint state validation.
func TestValidateSprintState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(t *testing.T) string // returns sprintPath
		wantError bool
	}{
		{
			name: "valid open state",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					Build()
				return filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
			},
			wantError: false,
		},
		{
			name: "valid closed state",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "closed").
					Build()
				return filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
			},
			wantError: false,
		},
		{
			name: "invalid state - both open and closed",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					Build()
				sprintPath := filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
				// Create both open and closed files (invalid state)
				testutil.CreateFile(t, filepath.Join(sprintPath, "closed"))
				return sprintPath
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sprintPath := tt.setup(t)

			// Act: Validate sprint state
			err := ValidateSprintState(sprintPath)

			// Assert
			if tt.wantError {
				if err == nil {
					t.Fatal("expected validation error, got nil")
				}
			} else {
				testutil.AssertNoError(t, err)
			}
		})
	}
}

// TestGetSprint tests loading a sprint by ID.
func TestGetSprint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		sprintID   string
		setup      func(t *testing.T) string // returns phasePath
		wantFound  bool
		wantStatus models.Status
	}{
		{
			name:     "found sprint",
			sprintID: "0001-sprint",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					Build()
				return filepath.Join(root, ".crumbler", "phases", "0001-phase")
			},
			wantFound:  true,
			wantStatus: models.StatusOpen,
		},
		{
			name:     "sprint not found",
			sprintID: "0099-sprint",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					Build()
				return filepath.Join(root, ".crumbler", "phases", "0001-phase")
			},
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			phasePath := tt.setup(t)

			// Act: Get sprint
			sprint, err := GetSprint(phasePath, tt.sprintID)

			// Assert
			if tt.wantFound {
				testutil.AssertNoError(t, err)
				if sprint == nil {
					t.Fatal("expected sprint, got nil")
				}
				if sprint.ID != tt.sprintID {
					t.Errorf("expected sprint ID %s, got %s", tt.sprintID, sprint.ID)
				}
				if sprint.Status != tt.wantStatus {
					t.Errorf("expected status %s, got %s", tt.wantStatus, sprint.Status)
				}
			} else {
				if err == nil {
					t.Fatal("expected error for non-existent sprint, got nil")
				}
				testutil.AssertError(t, err, "sprint not found")
			}
		})
	}
}

// TestListSprints tests listing all sprints.
func TestListSprints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(t *testing.T) string // returns phasePath
		wantCount int
		wantIDs   []string
	}{
		{
			name: "no sprints",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					Build()
				return filepath.Join(root, ".crumbler", "phases", "0001-phase")
			},
			wantCount: 0,
		},
		{
			name: "multiple sprints sorted",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0003", "open").
					WithSprint("0001", "0001", "closed").
					WithSprint("0001", "0002", "open").
					Build()
				return filepath.Join(root, ".crumbler", "phases", "0001-phase")
			},
			wantCount: 3,
			wantIDs:   []string{"0001-sprint", "0002-sprint", "0003-sprint"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			phasePath := tt.setup(t)

			// Act: List sprints
			sprints, err := ListSprints(phasePath)
			testutil.AssertNoError(t, err)

			// Assert: Count
			if len(sprints) != tt.wantCount {
				t.Errorf("expected %d sprints, got %d", tt.wantCount, len(sprints))
			}

			// Assert: Order
			if tt.wantIDs != nil {
				for i, wantID := range tt.wantIDs {
					if i < len(sprints) && sprints[i].ID != wantID {
						t.Errorf("sprint[%d] expected ID %s, got %s", i, wantID, sprints[i].ID)
					}
				}
			}
		})
	}
}

// TestCloseSprintAlreadyClosed tests error when closing already closed sprint.
func TestCloseSprintAlreadyClosed(t *testing.T) {
	t.Parallel()

	// Setup: Create project with closed sprint
	projectRoot := testutil.NewTestProject(t).
		WithPhase("0001", "open").
		WithSprint("0001", "0001", "closed").
		Build()

	sprintPath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")

	// Act: Try to close already closed sprint
	err := CloseSprint(sprintPath)

	// Assert: Error should be returned
	if err == nil {
		t.Fatal("expected error when closing already closed sprint, got nil")
	}
	testutil.AssertError(t, err, "already closed")
}

// TestSprintWithGoalsAndTickets tests sprint with complex structure.
func TestSprintWithGoalsAndTickets(t *testing.T) {
	t.Parallel()

	// Setup: Create complex sprint structure
	// Note: WithTicketGoal sets ticket status to "open", so we need to add goals first
	// then set the final status by calling WithTicket with "done" last
	projectRoot := testutil.NewTestProject(t).
		WithPhase("0001", "open").
		WithSprint("0001", "0001", "open").
		WithSprintGoal("0001", "0001", 1, "Implement API", "closed").
		WithSprintGoal("0001", "0001", 2, "Write tests", "closed").
		WithTicketGoal("0001", "0001", "0001", 1, "Create endpoint", "closed").
		WithTicket("0001", "0001", "0001", "done"). // Set final status after goals
		WithTicketGoal("0001", "0001", "0002", 1, "Add validation", "closed").
		WithTicket("0001", "0001", "0002", "done"). // Set final status after goals
		Build()

	sprintPath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")

	// Assert: Sprint goals met
	met, err := AreSprintGoalsMet(sprintPath)
	testutil.AssertNoError(t, err)
	if !met {
		t.Error("expected sprint goals to be met")
	}

	// Assert: Sprint can be closed
	err = CloseSprint(sprintPath)
	testutil.AssertNoError(t, err)

	// Assert: Sprint is closed
	testutil.AssertStatus(t, sprintPath, "closed")
}

// TestMultipleSprintsInPhase tests managing multiple sprints.
func TestMultipleSprintsInPhase(t *testing.T) {
	t.Parallel()

	// Setup: Create phase with multiple sprints in different states
	projectRoot := testutil.NewTestProject(t).
		WithPhase("0001", "open").
		WithSprint("0001", "0001", "closed").
		WithSprint("0001", "0002", "closed").
		WithSprint("0001", "0003", "open").
		Build()

	phasePath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")

	// Act: List sprints
	sprints, err := ListSprints(phasePath)
	testutil.AssertNoError(t, err)

	// Assert: Correct count
	if len(sprints) != 3 {
		t.Fatalf("expected 3 sprints, got %d", len(sprints))
	}

	// Assert: Status of each sprint
	statusByID := make(map[string]models.Status)
	for _, s := range sprints {
		statusByID[s.ID] = s.Status
	}

	if statusByID["0001-sprint"] != models.StatusClosed {
		t.Errorf("sprint 0001 expected closed, got %s", statusByID["0001-sprint"])
	}
	if statusByID["0002-sprint"] != models.StatusClosed {
		t.Errorf("sprint 0002 expected closed, got %s", statusByID["0002-sprint"])
	}
	if statusByID["0003-sprint"] != models.StatusOpen {
		t.Errorf("sprint 0003 expected open, got %s", statusByID["0003-sprint"])
	}

	// Assert: GetOpenSprint finds the right one
	openSprint, err := GetOpenSprint(phasePath)
	testutil.AssertNoError(t, err)
	if openSprint == nil {
		t.Fatal("expected to find open sprint")
	}
	if openSprint.ID != "0003-sprint" {
		t.Errorf("expected open sprint 0003, got %s", openSprint.ID)
	}

	// Assert: Next index is correct
	nextIndex, err := GetNextSprintIndex(phasePath)
	testutil.AssertNoError(t, err)
	if nextIndex != 4 {
		t.Errorf("expected next index 4, got %d", nextIndex)
	}
}
