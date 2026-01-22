// Package phase provides tests for Phase management functions.
package phase

import (
	"path/filepath"
	"testing"

	"github.com/waynenilsen/crumbler/internal/models"
	"github.com/waynenilsen/crumbler/internal/testutil"
)

// TestCreatePhase tests creating a new phase with proper directory structure.
func TestCreatePhase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		phaseIndex int
		wantID     string
	}{
		{
			name:       "create first phase",
			phaseIndex: 1,
			wantID:     "0001-phase",
		},
		{
			name:       "create second phase",
			phaseIndex: 2,
			wantID:     "0002-phase",
		},
		{
			name:       "create phase with high index",
			phaseIndex: 99,
			wantID:     "0099-phase",
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Build a minimal test project
			projectRoot := testutil.NewTestProject(t).Build()

			// Create the phase
			phasePath, err := CreatePhase(projectRoot, tt.phaseIndex)
			testutil.AssertNoError(t, err)

			// Verify directory was created
			testutil.AssertDirExists(t, phasePath)

			// Verify README.md was created
			readmePath := filepath.Join(phasePath, "README.md")
			testutil.AssertFileExists(t, readmePath)

			// Verify goals/ subdirectory was created
			goalsDir := filepath.Join(phasePath, "goals")
			testutil.AssertDirExists(t, goalsDir)

			// Verify sprints/ subdirectory was created
			sprintsDir := filepath.Join(phasePath, "sprints")
			testutil.AssertDirExists(t, sprintsDir)

			// Verify open file was created (phase is open)
			testutil.AssertStatus(t, phasePath, "open")
		})
	}
}

// TestCreatePhaseAlreadyExists tests that creating a duplicate phase fails.
func TestCreatePhaseAlreadyExists(t *testing.T) {
	t.Parallel()

	projectRoot := testutil.NewTestProject(t).
		WithPhase("0001", "open").
		Build()

	// Try to create the same phase again
	_, err := CreatePhase(projectRoot, 1)
	testutil.AssertError(t, err, "already exists")
}

// TestCreatePhaseGoal tests creating phase goals.
func TestCreatePhaseGoal(t *testing.T) {
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
			goalName:  "Implement authentication",
			wantID:    "0001-goal",
		},
		{
			name:      "create second goal",
			goalIndex: 2,
			goalName:  "Design database schema",
			wantID:    "0002-goal",
		},
		{
			name:      "create goal with special characters",
			goalIndex: 3,
			goalName:  "Add user-friendly error handling",
			wantID:    "0003-goal",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			projectRoot := testutil.NewTestProject(t).
				WithPhase("0001", "open").
				Build()

			phasePath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")

			// Create the goal
			goalPath, err := CreatePhaseGoal(phasePath, tt.goalIndex, tt.goalName)
			testutil.AssertNoError(t, err)

			// Verify goal directory was created
			testutil.AssertDirExists(t, goalPath)

			// Verify name file was created with correct content
			testutil.AssertGoalName(t, goalPath, tt.goalName)

			// Verify goal is open
			testutil.AssertGoalStatus(t, goalPath, "open")
		})
	}
}

// TestClosePhaseGoal tests closing phase goals.
func TestClosePhaseGoal(t *testing.T) {
	t.Parallel()

	projectRoot := testutil.NewTestProject(t).
		WithPhase("0001", "open").
		WithPhaseGoal("0001", 1, "First goal", "open").
		WithPhaseGoal("0001", 2, "Second goal", "open").
		Build()

	phasePath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")

	// Close the first goal
	err := ClosePhaseGoal(phasePath, "0001-goal")
	testutil.AssertNoError(t, err)

	// Verify first goal is closed
	goal1Path := filepath.Join(phasePath, "goals", "0001-goal")
	testutil.AssertGoalStatus(t, goal1Path, "closed")

	// Verify second goal is still open
	goal2Path := filepath.Join(phasePath, "goals", "0002-goal")
	testutil.AssertGoalStatus(t, goal2Path, "open")
}

// TestClosePhaseWhenAllSprintsAndGoalsClosed tests closing phase when all conditions are met.
func TestClosePhaseWhenAllSprintsAndGoalsClosed(t *testing.T) {
	t.Parallel()

	projectRoot := testutil.NewTestProject(t).
		WithPhase("0001", "open").
		WithPhaseGoal("0001", 1, "Phase goal 1", "closed").
		WithPhaseGoal("0001", 2, "Phase goal 2", "closed").
		WithSprint("0001", "0001", "closed").
		WithSprint("0001", "0002", "closed").
		Build()

	phasePath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")

	// Close the phase
	err := ClosePhase(phasePath)
	testutil.AssertNoError(t, err)

	// Verify phase is closed
	testutil.AssertStatus(t, phasePath, "closed")
}

// TestClosePhaseErrorWithOpenSprints tests that closing phase fails when sprints are open.
func TestClosePhaseErrorWithOpenSprints(t *testing.T) {
	t.Parallel()

	projectRoot := testutil.NewTestProject(t).
		WithPhase("0001", "open").
		WithPhaseGoal("0001", 1, "Phase goal", "closed").
		WithSprint("0001", "0001", "closed").
		WithSprint("0001", "0002", "open"). // One open sprint
		Build()

	phasePath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")

	// Try to close the phase
	err := ClosePhase(phasePath)

	// Should fail with hierarchy constraint error
	if err == nil {
		t.Fatal("expected error when closing phase with open sprints")
	}
	testutil.AssertError(t, err, "open sprints")
	if !models.IsHierarchyConstraint(err) {
		t.Errorf("expected hierarchy constraint error, got: %v", err)
	}
}

// TestClosePhaseErrorWithOpenGoals tests that closing phase fails when goals are open.
func TestClosePhaseErrorWithOpenGoals(t *testing.T) {
	t.Parallel()

	projectRoot := testutil.NewTestProject(t).
		WithPhase("0001", "open").
		WithPhaseGoal("0001", 1, "Closed goal", "closed").
		WithPhaseGoal("0001", 2, "Open goal", "open"). // One open goal
		WithSprint("0001", "0001", "closed").
		Build()

	phasePath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")

	// Try to close the phase
	err := ClosePhase(phasePath)

	// Should fail with hierarchy constraint error
	if err == nil {
		t.Fatal("expected error when closing phase with open goals")
	}
	testutil.AssertError(t, err, "open goals")
	if !models.IsHierarchyConstraint(err) {
		t.Errorf("expected hierarchy constraint error, got: %v", err)
	}
}

// TestGetOpenPhase tests getting the current open phase.
func TestGetOpenPhase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func(*testing.T) string
		expectPhase string
		expectNil   bool
	}{
		{
			name: "single open phase",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithPhase("0001", "open").
					Build()
			},
			expectPhase: "0001-phase",
			expectNil:   false,
		},
		{
			name: "multiple phases one open",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithPhase("0001", "closed").
					WithPhase("0002", "open").
					Build()
			},
			expectPhase: "0002-phase",
			expectNil:   false,
		},
		{
			name: "no open phases",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithPhase("0001", "closed").
					WithPhase("0002", "closed").
					Build()
			},
			expectPhase: "",
			expectNil:   true,
		},
		{
			name: "no phases at all",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).Build()
			},
			expectPhase: "",
			expectNil:   true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			projectRoot := tt.setup(t)

			phase, err := GetOpenPhase(projectRoot)
			testutil.AssertNoError(t, err)

			if tt.expectNil {
				if phase != nil {
					t.Errorf("expected nil phase, got: %v", phase)
				}
			} else {
				if phase == nil {
					t.Fatal("expected non-nil phase")
				}
				if phase.ID != tt.expectPhase {
					t.Errorf("expected phase ID %s, got %s", tt.expectPhase, phase.ID)
				}
			}
		})
	}
}

// TestArePhaseGoalsMet tests the ArePhaseGoalsMet function.
func TestArePhaseGoalsMet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(*testing.T) string
		wantMet  bool
		wantErr  bool
	}{
		{
			name: "no goals or sprints - not met",
			setup: func(t *testing.T) string {
				projectRoot := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					Build()
				return filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")
			},
			wantMet: false,
			wantErr: false,
		},
		{
			name: "goals exist but not all closed - not met",
			setup: func(t *testing.T) string {
				projectRoot := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Goal 1", "closed").
					WithPhaseGoal("0001", 2, "Goal 2", "open").
					WithSprint("0001", "0001", "closed").
					Build()
				return filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")
			},
			wantMet: false,
			wantErr: false,
		},
		{
			name: "sprints exist but not all closed - not met",
			setup: func(t *testing.T) string {
				projectRoot := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Goal 1", "closed").
					WithSprint("0001", "0001", "closed").
					WithSprint("0001", "0002", "open").
					Build()
				return filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")
			},
			wantMet: false,
			wantErr: false,
		},
		{
			name: "all goals closed and all sprints closed - met",
			setup: func(t *testing.T) string {
				projectRoot := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Goal 1", "closed").
					WithPhaseGoal("0001", 2, "Goal 2", "closed").
					WithSprint("0001", "0001", "closed").
					WithSprint("0001", "0002", "closed").
					Build()
				return filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")
			},
			wantMet: true,
			wantErr: false,
		},
		{
			name: "goals exist but no sprints - not met",
			setup: func(t *testing.T) string {
				projectRoot := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Goal 1", "closed").
					Build()
				return filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")
			},
			wantMet: false,
			wantErr: false,
		},
		{
			name: "sprints exist but no goals - not met",
			setup: func(t *testing.T) string {
				projectRoot := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "closed").
					Build()
				return filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")
			},
			wantMet: false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			phasePath := tt.setup(t)

			met, err := ArePhaseGoalsMet(phasePath)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				testutil.AssertNoError(t, err)
			}

			if met != tt.wantMet {
				t.Errorf("expected goals met = %v, got %v", tt.wantMet, met)
			}
		})
	}
}

// TestGetNextPhaseIndex tests the GetNextPhaseIndex function.
func TestGetNextPhaseIndex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(*testing.T) string
		wantIndex int
	}{
		{
			name: "no phases - returns 1",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).Build()
			},
			wantIndex: 1,
		},
		{
			name: "one phase - returns 2",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithPhase("0001", "open").
					Build()
			},
			wantIndex: 2,
		},
		{
			name: "multiple phases - returns next index",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithPhase("0001", "closed").
					WithPhase("0002", "closed").
					WithPhase("0003", "open").
					Build()
			},
			wantIndex: 4,
		},
		{
			name: "non-sequential phases - returns max + 1",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithPhase("0001", "closed").
					WithPhase("0005", "open").
					Build()
			},
			wantIndex: 6,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			projectRoot := tt.setup(t)

			index, err := GetNextPhaseIndex(projectRoot)
			testutil.AssertNoError(t, err)

			if index != tt.wantIndex {
				t.Errorf("expected next index %d, got %d", tt.wantIndex, index)
			}
		})
	}
}

// TestGetPhase tests loading a specific phase by ID.
func TestGetPhase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setup      func(*testing.T) string
		phaseID    string
		wantStatus models.Status
		wantErr    bool
	}{
		{
			name: "existing open phase",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithPhase("0001", "open").
					Build()
			},
			phaseID:    "0001-phase",
			wantStatus: models.StatusOpen,
			wantErr:    false,
		},
		{
			name: "existing closed phase",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithPhase("0002", "closed").
					Build()
			},
			phaseID:    "0002-phase",
			wantStatus: models.StatusClosed,
			wantErr:    false,
		},
		{
			name: "non-existent phase",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithPhase("0001", "open").
					Build()
			},
			phaseID: "0099-phase",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			projectRoot := tt.setup(t)

			phase, err := GetPhase(projectRoot, tt.phaseID)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				testutil.AssertNoError(t, err)
				if phase == nil {
					t.Fatal("expected non-nil phase")
				}
				if phase.Status != tt.wantStatus {
					t.Errorf("expected status %v, got %v", tt.wantStatus, phase.Status)
				}
				if phase.ID != tt.phaseID {
					t.Errorf("expected phase ID %s, got %s", tt.phaseID, phase.ID)
				}
			}
		})
	}
}

// TestListPhases tests listing all phases.
func TestListPhases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setup      func(*testing.T) string
		wantCount  int
		wantPhases []string
	}{
		{
			name: "no phases",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).Build()
			},
			wantCount:  0,
			wantPhases: []string{},
		},
		{
			name: "single phase",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithPhase("0001", "open").
					Build()
			},
			wantCount:  1,
			wantPhases: []string{"0001-phase"},
		},
		{
			name: "multiple phases sorted by index",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithPhase("0003", "open").
					WithPhase("0001", "closed").
					WithPhase("0002", "closed").
					Build()
			},
			wantCount:  3,
			wantPhases: []string{"0001-phase", "0002-phase", "0003-phase"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			projectRoot := tt.setup(t)

			phases, err := ListPhases(projectRoot)
			testutil.AssertNoError(t, err)

			if len(phases) != tt.wantCount {
				t.Errorf("expected %d phases, got %d", tt.wantCount, len(phases))
			}

			for i, wantID := range tt.wantPhases {
				if i >= len(phases) {
					break
				}
				if phases[i].ID != wantID {
					t.Errorf("expected phase[%d].ID = %s, got %s", i, wantID, phases[i].ID)
				}
			}
		})
	}
}

// TestGetPhaseGoals tests retrieving phase goals.
func TestGetPhaseGoals(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		setup        func(*testing.T) string
		wantCount    int
		wantGoalIDs  []string
		wantStatuses []models.Status
	}{
		{
			name: "no goals",
			setup: func(t *testing.T) string {
				projectRoot := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					Build()
				return filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")
			},
			wantCount:    0,
			wantGoalIDs:  []string{},
			wantStatuses: []models.Status{},
		},
		{
			name: "single goal",
			setup: func(t *testing.T) string {
				projectRoot := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "First goal", "open").
					Build()
				return filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")
			},
			wantCount:    1,
			wantGoalIDs:  []string{"0001-goal"},
			wantStatuses: []models.Status{models.StatusOpen},
		},
		{
			name: "multiple goals sorted",
			setup: func(t *testing.T) string {
				projectRoot := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 3, "Third goal", "open").
					WithPhaseGoal("0001", 1, "First goal", "closed").
					WithPhaseGoal("0001", 2, "Second goal", "closed").
					Build()
				return filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")
			},
			wantCount:    3,
			wantGoalIDs:  []string{"0001-goal", "0002-goal", "0003-goal"},
			wantStatuses: []models.Status{models.StatusClosed, models.StatusClosed, models.StatusOpen},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			phasePath := tt.setup(t)

			goals, err := GetPhaseGoals(phasePath)
			testutil.AssertNoError(t, err)

			if len(goals) != tt.wantCount {
				t.Errorf("expected %d goals, got %d", tt.wantCount, len(goals))
			}

			for i, wantID := range tt.wantGoalIDs {
				if i >= len(goals) {
					break
				}
				if goals[i].ID != wantID {
					t.Errorf("expected goal[%d].ID = %s, got %s", i, wantID, goals[i].ID)
				}
				if goals[i].Status != tt.wantStatuses[i] {
					t.Errorf("expected goal[%d].Status = %v, got %v", i, tt.wantStatuses[i], goals[i].Status)
				}
			}
		})
	}
}

// TestValidatePhaseState tests phase state validation.
func TestValidatePhaseState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(*testing.T) string
		wantErr bool
	}{
		{
			name: "valid open phase",
			setup: func(t *testing.T) string {
				projectRoot := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					Build()
				return filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")
			},
			wantErr: false,
		},
		{
			name: "valid closed phase",
			setup: func(t *testing.T) string {
				projectRoot := testutil.NewTestProject(t).
					WithPhase("0001", "closed").
					Build()
				return filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")
			},
			wantErr: false,
		},
		{
			name: "invalid phase with both open and closed",
			setup: func(t *testing.T) string {
				projectRoot := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					Build()
				phasePath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")
				// Add a closed file as well to create invalid state
				testutil.CreateFile(t, filepath.Join(phasePath, "closed"))
				return phasePath
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			phasePath := tt.setup(t)

			err := ValidatePhaseState(phasePath)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				testutil.AssertNoError(t, err)
			}
		})
	}
}

// TestClosePhaseNotOpen tests that closing a non-open phase fails.
func TestClosePhaseNotOpen(t *testing.T) {
	t.Parallel()

	projectRoot := testutil.NewTestProject(t).
		WithPhase("0001", "closed").
		Build()

	phasePath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")

	err := ClosePhase(phasePath)
	if err == nil {
		t.Fatal("expected error when closing already closed phase")
	}
	testutil.AssertError(t, err, "not open")
}

// TestPhaseGoalsWithNames tests that goals are created with proper names.
func TestPhaseGoalsWithNames(t *testing.T) {
	t.Parallel()

	projectRoot := testutil.NewTestProject(t).
		WithPhase("0001", "open").
		WithPhaseGoal("0001", 1, "Implement user authentication", "open").
		WithPhaseGoal("0001", 2, "Set up CI/CD pipeline", "closed").
		Build()

	phasePath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")

	goals, err := GetPhaseGoals(phasePath)
	testutil.AssertNoError(t, err)

	if len(goals) != 2 {
		t.Fatalf("expected 2 goals, got %d", len(goals))
	}

	// Verify names
	if goals[0].Name != "Implement user authentication" {
		t.Errorf("expected goal[0].Name = 'Implement user authentication', got '%s'", goals[0].Name)
	}
	if goals[1].Name != "Set up CI/CD pipeline" {
		t.Errorf("expected goal[1].Name = 'Set up CI/CD pipeline', got '%s'", goals[1].Name)
	}
}

// TestClosePhaseGoalNotFound tests closing a non-existent goal.
func TestClosePhaseGoalNotFound(t *testing.T) {
	t.Parallel()

	projectRoot := testutil.NewTestProject(t).
		WithPhase("0001", "open").
		WithPhaseGoal("0001", 1, "Existing goal", "open").
		Build()

	phasePath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")

	err := ClosePhaseGoal(phasePath, "9999-goal")
	if err == nil {
		t.Fatal("expected error when closing non-existent goal")
	}
	testutil.AssertError(t, err, "not found")
}

// TestPhaseWithMultipleGoalsAndSprints tests a complex phase setup.
func TestPhaseWithMultipleGoalsAndSprints(t *testing.T) {
	t.Parallel()

	projectRoot := testutil.NewTestProject(t).
		WithPhase("0001", "open").
		WithPhaseGoal("0001", 1, "Goal 1", "closed").
		WithPhaseGoal("0001", 2, "Goal 2", "closed").
		WithPhaseGoal("0001", 3, "Goal 3", "closed").
		WithSprint("0001", "0001", "closed").
		WithSprint("0001", "0002", "closed").
		Build()

	phasePath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")

	// Verify goals met
	met, err := ArePhaseGoalsMet(phasePath)
	testutil.AssertNoError(t, err)
	if !met {
		t.Error("expected phase goals to be met")
	}

	// Close the phase
	err = ClosePhase(phasePath)
	testutil.AssertNoError(t, err)

	// Verify phase is closed
	testutil.AssertStatus(t, phasePath, "closed")

	// Get the phase to verify status
	phase, err := GetPhase(projectRoot, "0001-phase")
	testutil.AssertNoError(t, err)
	if phase.Status != models.StatusClosed {
		t.Errorf("expected phase status to be closed, got %v", phase.Status)
	}
}

// TestPhaseIndex tests that phases have correct indices.
func TestPhaseIndex(t *testing.T) {
	t.Parallel()

	projectRoot := testutil.NewTestProject(t).
		WithPhase("0001", "closed").
		WithPhase("0010", "closed").
		WithPhase("0100", "open").
		Build()

	phases, err := ListPhases(projectRoot)
	testutil.AssertNoError(t, err)

	if len(phases) != 3 {
		t.Fatalf("expected 3 phases, got %d", len(phases))
	}

	expectedIndices := []int{1, 10, 100}
	for i, phase := range phases {
		if phase.Index != expectedIndices[i] {
			t.Errorf("expected phase[%d].Index = %d, got %d", i, expectedIndices[i], phase.Index)
		}
	}
}
