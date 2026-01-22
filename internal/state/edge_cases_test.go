package state_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/waynenilsen/crumbler/internal/models"
	"github.com/waynenilsen/crumbler/internal/query"
	"github.com/waynenilsen/crumbler/internal/state"
	"github.com/waynenilsen/crumbler/internal/testutil"
)

// =============================================================================
// Edge Case Scenarios
// =============================================================================

func TestEmptyProject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		test func(t *testing.T, projectRoot string)
	}{
		{
			name: "no phases directory",
			test: func(t *testing.T, projectRoot string) {
				t.Parallel()

				// Remove the phases directory
				phasesDir := filepath.Join(projectRoot, ".crumbler", "phases")
				if err := os.RemoveAll(phasesDir); err != nil {
					t.Fatalf("failed to remove phases directory: %v", err)
				}

				// Validate state machine should pass with no phases
				validator := state.NewStateValidator(projectRoot)
				err := validator.ValidateStateMachine(projectRoot)
				testutil.AssertNoError(t, err)

				// OpenPhaseExists should return false
				exists, err := query.OpenPhaseExists(projectRoot)
				testutil.AssertNoError(t, err)
				if exists {
					t.Errorf("expected no open phase, but one exists")
				}

				// RoadmapComplete should return false (no phases means not complete)
				complete, err := query.RoadmapComplete(projectRoot)
				testutil.AssertNoError(t, err)
				if complete {
					t.Errorf("expected roadmap not complete with no phases")
				}
			},
		},
		{
			name: "empty phases directory",
			test: func(t *testing.T, projectRoot string) {
				t.Parallel()

				// Phases directory exists but is empty (default from builder)
				exists, err := query.OpenPhaseExists(projectRoot)
				testutil.AssertNoError(t, err)
				if exists {
					t.Errorf("expected no open phase in empty project")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectRoot := testutil.NewTestProject(t).Build()
			tt.test(t, projectRoot)
		})
	}
}

func TestPhaseWithNoSprints(t *testing.T) {
	t.Parallel()

	projectRoot := testutil.NewTestProject(t).
		WithPhase("0001", "open").
		Build()

	// Phase exists and is open
	exists, err := query.OpenPhaseExists(projectRoot)
	testutil.AssertNoError(t, err)
	if !exists {
		t.Errorf("expected open phase to exist")
	}

	// Phase goals should not be met because no sprints exist
	phasePath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")
	goalsMet, err := query.PhaseGoalsMet(phasePath)
	testutil.AssertNoError(t, err)
	if goalsMet {
		t.Errorf("expected phase goals not met with no sprints")
	}

	// OpenSprintExists should return false (no sprints)
	sprintExists, err := query.OpenSprintExists(phasePath)
	testutil.AssertNoError(t, err)
	if sprintExists {
		t.Errorf("expected no open sprints in phase with no sprints")
	}

	// SprintsExist should return false
	sprintsExist, err := query.SprintsExist(phasePath)
	testutil.AssertNoError(t, err)
	if sprintsExist {
		t.Errorf("expected no sprints to exist")
	}
}

func TestPhaseWithNoGoals(t *testing.T) {
	t.Parallel()

	projectRoot := testutil.NewTestProject(t).
		WithPhase("0001", "open").
		WithSprint("0001", "0001", "closed").
		Build()

	phasePath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")

	// PhaseGoalsExist should return false
	goalsExist, err := query.PhaseGoalsExist(phasePath)
	testutil.AssertNoError(t, err)
	if goalsExist {
		t.Errorf("expected no phase goals to exist")
	}

	// PhaseGoalsMet should still return false because no goals exist
	// (goals must exist AND be closed for goals to be "met")
	goalsMet, err := query.PhaseGoalsMet(phasePath)
	testutil.AssertNoError(t, err)
	if goalsMet {
		t.Errorf("expected phase goals not met when no goals exist")
	}
}

func TestSprintWithNoTickets(t *testing.T) {
	t.Parallel()

	projectRoot := testutil.NewTestProject(t).
		WithPhase("0001", "open").
		WithSprint("0001", "0001", "open").
		Build()

	sprintPath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")

	// OpenTicketsExist should return false
	ticketsExist, err := query.OpenTicketsExist(sprintPath)
	testutil.AssertNoError(t, err)
	if ticketsExist {
		t.Errorf("expected no open tickets in sprint with no tickets")
	}

	// TicketsExist should return false
	exists, err := query.TicketsExist(sprintPath)
	testutil.AssertNoError(t, err)
	if exists {
		t.Errorf("expected no tickets to exist")
	}

	// SprintGoalsMet should return false because no tickets exist
	goalsMet, err := query.SprintGoalsMet(sprintPath)
	testutil.AssertNoError(t, err)
	if goalsMet {
		t.Errorf("expected sprint goals not met when no tickets exist")
	}
}

func TestSprintWithNoGoals(t *testing.T) {
	t.Parallel()

	projectRoot := testutil.NewTestProject(t).
		WithPhase("0001", "open").
		WithSprint("0001", "0001", "open").
		WithTicket("0001", "0001", "0001", "done").
		Build()

	sprintPath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")

	// SprintGoalsExist should return false
	goalsExist, err := query.SprintGoalsExist(sprintPath)
	testutil.AssertNoError(t, err)
	if goalsExist {
		t.Errorf("expected no sprint goals to exist")
	}

	// SprintGoalsMet should return false because no goals exist
	goalsMet, err := query.SprintGoalsMet(sprintPath)
	testutil.AssertNoError(t, err)
	if goalsMet {
		t.Errorf("expected sprint goals not met when no goals exist")
	}
}

func TestTicketWithNoGoals(t *testing.T) {
	t.Parallel()

	projectRoot := testutil.NewTestProject(t).
		WithPhase("0001", "open").
		WithSprint("0001", "0001", "open").
		WithTicket("0001", "0001", "0001", "open").
		Build()

	ticketPath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint", "tickets", "0001-ticket")

	// TicketGoalsExist should return false
	goalsExist, err := query.TicketGoalsExist(ticketPath)
	testutil.AssertNoError(t, err)
	if goalsExist {
		t.Errorf("expected no ticket goals to exist")
	}

	// Ticket should be able to be marked done (no goals = vacuously all goals closed)
	validator := state.NewStateValidator(projectRoot)
	canDone, err := validator.CanMarkTicketDone(ticketPath)
	testutil.AssertNoError(t, err)
	if !canDone {
		t.Errorf("expected ticket with no goals to be markable as done")
	}
}

func TestMultiplePhasesSprintsTicketsGoals(t *testing.T) {
	t.Parallel()

	// Build the project with all phases open first, then manually set closed status
	// after build to avoid builder order-dependency issues
	projectRoot := testutil.NewTestProject(t).
		// Phase 1: will be closed after build
		WithPhase("0001", "open").
		WithPhaseGoal("0001", 1, "Phase 1 Goal 1", "closed").
		WithPhaseGoal("0001", 2, "Phase 1 Goal 2", "closed").
		WithTicket("0001", "0001", "0001", "open"). // Will be set to done after build
		WithTicketGoal("0001", "0001", "0001", 1, "Ticket 1 Goal 1", "closed").
		WithTicket("0001", "0001", "0002", "open"). // Will be set to done after build
		WithSprintGoal("0001", "0001", 1, "Sprint 1 Goal 1", "closed").
		WithTicket("0001", "0002", "0001", "open"). // Will be set to done after build
		// Phase 2: open with sprints in progress
		WithPhase("0002", "open").
		WithPhaseGoal("0002", 1, "Phase 2 Goal 1", "open").
		WithTicket("0002", "0001", "0001", "open").
		WithTicketGoal("0002", "0001", "0001", 1, "Ticket Goal 1", "open").
		WithTicketGoal("0002", "0001", "0001", 2, "Ticket Goal 2", "closed").
		WithSprintGoal("0002", "0001", 1, "Sprint Goal", "open").
		// Phase 3: planned but not started
		WithPhase("0003", "open").
		Build()

	// Manually set closed status for phase 1, its sprints, and tickets
	phase1Path := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")
	sprint1Path := filepath.Join(phase1Path, "sprints", "0001-sprint")
	sprint2Path := filepath.Join(phase1Path, "sprints", "0002-sprint")

	// Mark all tickets in phase 1 as done first
	ticket1Path := filepath.Join(sprint1Path, "tickets", "0001-ticket")
	ticket2Path := filepath.Join(sprint1Path, "tickets", "0002-ticket")
	ticket3Path := filepath.Join(sprint2Path, "tickets", "0001-ticket")

	testutil.RemoveFile(t, filepath.Join(ticket1Path, "open"))
	testutil.CreateFile(t, filepath.Join(ticket1Path, "done"))
	testutil.RemoveFile(t, filepath.Join(ticket2Path, "open"))
	testutil.CreateFile(t, filepath.Join(ticket2Path, "done"))
	testutil.RemoveFile(t, filepath.Join(ticket3Path, "open"))
	testutil.CreateFile(t, filepath.Join(ticket3Path, "done"))

	// Close sprints (remove open, create closed)
	testutil.RemoveFile(t, filepath.Join(sprint1Path, "open"))
	testutil.CreateFile(t, filepath.Join(sprint1Path, "closed"))
	testutil.RemoveFile(t, filepath.Join(sprint2Path, "open"))
	testutil.CreateFile(t, filepath.Join(sprint2Path, "closed"))

	// Close phase 1
	testutil.RemoveFile(t, filepath.Join(phase1Path, "open"))
	testutil.CreateFile(t, filepath.Join(phase1Path, "closed"))

	// Verify state machine is valid
	validator := state.NewStateValidator(projectRoot)
	err := validator.ValidateStateMachine(projectRoot)
	testutil.AssertNoError(t, err)

	// Verify open phase exists (should find phase 2 or 3)
	exists, err := query.OpenPhaseExists(projectRoot)
	testutil.AssertNoError(t, err)
	if !exists {
		t.Errorf("expected open phases to exist")
	}

	// Verify roadmap is not complete (phases 2 and 3 still open)
	complete, err := query.RoadmapComplete(projectRoot)
	testutil.AssertNoError(t, err)
	if complete {
		t.Errorf("expected roadmap not complete")
	}

	// Verify phase 2 has open sprint
	phase2Path := filepath.Join(projectRoot, ".crumbler", "phases", "0002-phase")
	openSprint, err := query.OpenSprintExists(phase2Path)
	testutil.AssertNoError(t, err)
	if !openSprint {
		t.Errorf("expected open sprint in phase 2")
	}

	// Verify sprint 1 in phase 2 has open tickets
	phase2Sprint1Path := filepath.Join(phase2Path, "sprints", "0001-sprint")
	openTickets, err := query.OpenTicketsExist(phase2Sprint1Path)
	testutil.AssertNoError(t, err)
	if !openTickets {
		t.Errorf("expected open tickets in sprint")
	}

	// Verify phase 1's goals are met
	goalsMet, err := query.PhaseGoalsMet(phase1Path)
	testutil.AssertNoError(t, err)
	if !goalsMet {
		t.Errorf("expected phase 1 goals to be met")
	}
}

func TestRunningOutsideManagedProject(t *testing.T) {
	t.Parallel()

	// Create a temporary directory without .crumbler
	tempDir := t.TempDir()

	// Validator should return an error
	validator := state.NewStateValidator(tempDir)
	err := validator.ValidateStateMachine(tempDir)
	if err == nil {
		t.Errorf("expected error when running outside managed project")
	}

	// Error should indicate project not found
	testutil.AssertError(t, err, "not a crumbler project")
}

func TestMissingRequiredFiles(t *testing.T) {
	t.Parallel()

	t.Run("missing goal name file", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithPhase("0001", "open").
			WithPhaseGoal("0001", 1, "Test Goal", "open").
			Build()

		// Remove the name file from the goal
		goalPath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "goals", "0001-goal")
		nameFile := filepath.Join(goalPath, "name")
		if err := os.Remove(nameFile); err != nil {
			t.Fatalf("failed to remove name file: %v", err)
		}

		// Validation should fail
		validator := state.NewStateValidator(projectRoot)
		err := validator.ValidateStateMachine(projectRoot)
		if err == nil {
			t.Errorf("expected error when goal name file is missing")
		}

		// Error should indicate missing goal name
		testutil.AssertError(t, err, "missing name file")
	})

	t.Run("missing status file", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithPhase("0001", "open").
			Build()

		// Remove the open file from the phase
		phasePath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")
		openFile := filepath.Join(phasePath, "open")
		if err := os.Remove(openFile); err != nil {
			t.Fatalf("failed to remove open file: %v", err)
		}

		// GetStatus should return an error
		_, err := state.GetStatus(phasePath)
		if err == nil {
			t.Errorf("expected error when no status file exists")
		}
	})
}

func TestGoalNameFileReading(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		goalContent    string
		expectedResult string
	}{
		{
			name:           "simple goal name",
			goalContent:    "Implement authentication",
			expectedResult: "Implement authentication",
		},
		{
			name:           "goal name with whitespace",
			goalContent:    "  Set up CI/CD pipeline  \n",
			expectedResult: "Set up CI/CD pipeline",
		},
		{
			name:           "multiline goal name",
			goalContent:    "Create user interface\nwith dashboard",
			expectedResult: "Create user interface\nwith dashboard",
		},
		{
			name:           "empty goal name after clearing",
			goalContent:    "will be cleared",
			expectedResult: "",
		},
		{
			name:           "goal name with special characters",
			goalContent:    "Fix bug #123 - DB connection",
			expectedResult: "Fix bug #123 - DB connection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			projectRoot := testutil.NewTestProject(t).
				WithPhase("0001", "open").
				WithPhaseGoal("0001", 1, tt.goalContent, "open").
				Build()

			goalPath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "goals", "0001-goal")

			// For the "empty goal name after clearing" test, clear the file
			if tt.expectedResult == "" {
				testutil.WriteFile(t, filepath.Join(goalPath, "name"), "")
			}

			name, err := state.ReadGoalName(goalPath)
			testutil.AssertNoError(t, err)

			// Note: ReadGoalName trims whitespace
			expected := strings.TrimSpace(tt.expectedResult)
			if name != expected {
				t.Errorf("expected goal name %q, got %q", expected, name)
			}
		})
	}
}

// =============================================================================
// Error Handling Scenarios
// =============================================================================

func TestErrorMessagesIncludeCorrectFilePaths(t *testing.T) {
	t.Parallel()

	t.Run("invalid phase state error includes path", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithPhase("0001", "open").
			Build()

		// Create conflicting state: add closed file while open exists
		phasePath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")
		closedFile := filepath.Join(phasePath, "closed")
		if err := os.WriteFile(closedFile, []byte{}, 0644); err != nil {
			t.Fatalf("failed to create closed file: %v", err)
		}

		validator := state.NewStateValidator(projectRoot)
		err := validator.ValidateStateMachine(projectRoot)
		if err == nil {
			t.Fatal("expected error for invalid state")
		}

		// Error should contain relative path
		errStr := err.Error()
		if !strings.Contains(errStr, ".crumbler/phases/0001-phase") {
			t.Errorf("error should contain relative path, got: %s", errStr)
		}
	})

	t.Run("sprint state error includes path", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithPhase("0001", "open").
			WithSprint("0001", "0001", "open").
			Build()

		// Create conflicting state
		sprintPath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
		closedFile := filepath.Join(sprintPath, "closed")
		if err := os.WriteFile(closedFile, []byte{}, 0644); err != nil {
			t.Fatalf("failed to create closed file: %v", err)
		}

		validator := state.NewStateValidator(projectRoot)
		err := validator.ValidateStateMachine(projectRoot)
		if err == nil {
			t.Fatal("expected error for invalid state")
		}

		// Error should contain relative path
		errStr := err.Error()
		if !strings.Contains(errStr, ".crumbler/phases/0001-phase/sprints/0001-sprint") {
			t.Errorf("error should contain relative path, got: %s", errStr)
		}
	})

	t.Run("ticket state error includes path", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithPhase("0001", "open").
			WithSprint("0001", "0001", "open").
			WithTicket("0001", "0001", "0001", "open").
			Build()

		// Create conflicting state
		ticketPath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint", "tickets", "0001-ticket")
		doneFile := filepath.Join(ticketPath, "done")
		if err := os.WriteFile(doneFile, []byte{}, 0644); err != nil {
			t.Fatalf("failed to create done file: %v", err)
		}

		validator := state.NewStateValidator(projectRoot)
		err := validator.ValidateStateMachine(projectRoot)
		if err == nil {
			t.Fatal("expected error for invalid state")
		}

		// Error should contain relative path
		errStr := err.Error()
		if !strings.Contains(errStr, "tickets/0001-ticket") {
			t.Errorf("error should contain relative path, got: %s", errStr)
		}
	})

	t.Run("goal state error includes path", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithPhase("0001", "open").
			WithPhaseGoal("0001", 1, "Test Goal", "open").
			Build()

		// Create conflicting state
		goalPath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "goals", "0001-goal")
		closedFile := filepath.Join(goalPath, "closed")
		if err := os.WriteFile(closedFile, []byte{}, 0644); err != nil {
			t.Fatalf("failed to create closed file: %v", err)
		}

		validator := state.NewStateValidator(projectRoot)
		err := validator.ValidateStateMachine(projectRoot)
		if err == nil {
			t.Fatal("expected error for invalid state")
		}

		// Error should contain relative path to goal
		errStr := err.Error()
		if !strings.Contains(errStr, "goals/0001-goal") {
			t.Errorf("error should contain goal path, got: %s", errStr)
		}
	})
}

func TestInvalidStateConflictDetection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setup          func(t *testing.T) string
		expectedError  string
		validatePaths  func(t *testing.T, err error)
	}{
		{
			name: "phase with both open and closed",
			setup: func(t *testing.T) string {
				projectRoot := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					Build()
				phasePath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")
				testutil.CreateFile(t, filepath.Join(phasePath, "closed"))
				return projectRoot
			},
			expectedError: "both 'open' and 'closed'",
			validatePaths: func(t *testing.T, err error) {
				errStr := err.Error()
				if !strings.Contains(errStr, "0001-phase") {
					t.Errorf("error should reference phase path")
				}
			},
		},
		{
			name: "sprint with both open and closed",
			setup: func(t *testing.T) string {
				projectRoot := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					Build()
				sprintPath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
				testutil.CreateFile(t, filepath.Join(sprintPath, "closed"))
				return projectRoot
			},
			expectedError: "both 'open' and 'closed'",
			validatePaths: func(t *testing.T, err error) {
				errStr := err.Error()
				if !strings.Contains(errStr, "0001-sprint") {
					t.Errorf("error should reference sprint path")
				}
			},
		},
		{
			name: "ticket with both open and done",
			setup: func(t *testing.T) string {
				projectRoot := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "open").
					Build()
				ticketPath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint", "tickets", "0001-ticket")
				testutil.CreateFile(t, filepath.Join(ticketPath, "done"))
				return projectRoot
			},
			expectedError: "both 'open' and 'done'",
			validatePaths: func(t *testing.T, err error) {
				errStr := err.Error()
				if !strings.Contains(errStr, "0001-ticket") {
					t.Errorf("error should reference ticket path")
				}
			},
		},
		{
			name: "goal with both open and closed",
			setup: func(t *testing.T) string {
				projectRoot := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Test Goal", "open").
					Build()
				goalPath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "goals", "0001-goal")
				testutil.CreateFile(t, filepath.Join(goalPath, "closed"))
				return projectRoot
			},
			expectedError: "both 'open' and 'closed'",
			validatePaths: func(t *testing.T, err error) {
				errStr := err.Error()
				if !strings.Contains(errStr, "0001-goal") {
					t.Errorf("error should reference goal path")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			projectRoot := tt.setup(t)
			validator := state.NewStateValidator(projectRoot)
			err := validator.ValidateStateMachine(projectRoot)

			if err == nil {
				t.Fatal("expected validation error")
			}

			testutil.AssertError(t, err, tt.expectedError)
			tt.validatePaths(t, err)
		})
	}
}

func TestStateValidationErrorsContainProperPaths(t *testing.T) {
	t.Parallel()

	t.Run("hierarchy constraint error with phase paths", func(t *testing.T) {
		t.Parallel()

		// Create a phase that is closed but has open sprints (invalid hierarchy)
		projectRoot := testutil.NewTestProject(t).
			WithPhase("0001", "closed").
			WithSprint("0001", "0001", "open").
			Build()

		validator := state.NewStateValidator(projectRoot)
		err := validator.ValidateStateMachine(projectRoot)

		if err == nil {
			t.Fatal("expected hierarchy constraint error")
		}

		// Verify error contains phase and sprint paths
		errStr := err.Error()
		if !strings.Contains(errStr, "0001-phase") || !strings.Contains(errStr, "0001-sprint") {
			t.Errorf("error should reference both phase and sprint paths, got: %s", errStr)
		}
	})

	t.Run("hierarchy constraint error with sprint paths", func(t *testing.T) {
		t.Parallel()

		// Create a sprint that is closed but has open tickets (invalid hierarchy)
		// Note: WithTicket resets sprint status, so we need to manually set closed after build
		projectRoot := testutil.NewTestProject(t).
			WithPhase("0001", "open").
			WithTicket("0001", "0001", "0001", "open").
			Build()

		// Manually close the sprint (keeping ticket open = invalid hierarchy)
		sprintPath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
		testutil.RemoveFile(t, filepath.Join(sprintPath, "open"))
		testutil.CreateFile(t, filepath.Join(sprintPath, "closed"))

		validator := state.NewStateValidator(projectRoot)
		err := validator.ValidateStateMachine(projectRoot)

		if err == nil {
			t.Fatal("expected hierarchy constraint error")
		}

		// Verify error contains sprint and ticket paths
		errStr := err.Error()
		if !strings.Contains(errStr, "0001-sprint") || !strings.Contains(errStr, "0001-ticket") {
			t.Errorf("error should reference both sprint and ticket paths, got: %s", errStr)
		}
	})

	t.Run("hierarchy constraint error with ticket goal paths", func(t *testing.T) {
		t.Parallel()

		// Create a ticket that is done but has open goals (invalid hierarchy)
		// Note: WithTicketGoal uses WithTicket with open status, so we set done manually
		projectRoot := testutil.NewTestProject(t).
			WithPhase("0001", "open").
			WithTicket("0001", "0001", "0001", "open").
			WithTicketGoal("0001", "0001", "0001", 1, "Open Goal", "open").
			Build()

		// Manually mark ticket as done (keeping goal open = invalid hierarchy)
		ticketPath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint", "tickets", "0001-ticket")
		testutil.RemoveFile(t, filepath.Join(ticketPath, "open"))
		testutil.CreateFile(t, filepath.Join(ticketPath, "done"))

		validator := state.NewStateValidator(projectRoot)
		err := validator.ValidateStateMachine(projectRoot)

		if err == nil {
			t.Fatal("expected hierarchy constraint error")
		}

		// Verify error contains ticket and goal paths
		errStr := err.Error()
		if !strings.Contains(errStr, "0001-ticket") || !strings.Contains(errStr, "0001-goal") {
			t.Errorf("error should reference both ticket and goal paths, got: %s", errStr)
		}
	})

	t.Run("missing goal name error with proper path", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithPhase("0001", "open").
			WithPhaseGoal("0001", 1, "Goal Name", "open").
			Build()

		// Remove the name file
		goalPath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "goals", "0001-goal")
		testutil.RemoveFile(t, filepath.Join(goalPath, "name"))

		validator := state.NewStateValidator(projectRoot)
		err := validator.ValidateStateMachine(projectRoot)

		if err == nil {
			t.Fatal("expected missing goal name error")
		}

		// Verify error contains goal path
		errStr := err.Error()
		if !strings.Contains(errStr, "0001-goal") {
			t.Errorf("error should reference goal path, got: %s", errStr)
		}
	})
}

func TestStateErrorTypes(t *testing.T) {
	t.Parallel()

	t.Run("invalid state error type", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithPhase("0001", "open").
			Build()

		// Create conflicting state
		phasePath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")
		testutil.CreateFile(t, filepath.Join(phasePath, "closed"))

		validator := state.NewStateValidator(projectRoot)
		err := validator.ValidateStateMachine(projectRoot)

		if err == nil {
			t.Fatal("expected error")
		}

		// Check if error is a StateError
		if !models.IsStateError(err) {
			t.Errorf("expected StateError, got %T", err)
		}

		// Check error type
		stateErr := models.AsStateError(err)
		if stateErr == nil {
			t.Fatal("expected non-nil StateError")
		}
		if stateErr.Type != models.ErrorTypeMutuallyExclusiveState {
			t.Errorf("expected error type %s, got %s", models.ErrorTypeMutuallyExclusiveState, stateErr.Type)
		}
	})

	t.Run("hierarchy constraint error type", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithPhase("0001", "closed").
			WithSprint("0001", "0001", "open").
			Build()

		validator := state.NewStateValidator(projectRoot)
		err := validator.ValidateStateMachine(projectRoot)

		if err == nil {
			t.Fatal("expected error")
		}

		stateErr := models.AsStateError(err)
		if stateErr == nil {
			t.Fatal("expected non-nil StateError")
		}
		if stateErr.Type != models.ErrorTypeHierarchyConstraint {
			t.Errorf("expected error type %s, got %s", models.ErrorTypeHierarchyConstraint, stateErr.Type)
		}
	})

	t.Run("missing goal name error type", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithPhase("0001", "open").
			WithPhaseGoal("0001", 1, "Test", "open").
			Build()

		goalPath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "goals", "0001-goal")
		testutil.RemoveFile(t, filepath.Join(goalPath, "name"))

		validator := state.NewStateValidator(projectRoot)
		err := validator.ValidateStateMachine(projectRoot)

		if err == nil {
			t.Fatal("expected error")
		}

		stateErr := models.AsStateError(err)
		if stateErr == nil {
			t.Fatal("expected non-nil StateError")
		}
		if stateErr.Type != models.ErrorTypeMissingGoalName {
			t.Errorf("expected error type %s, got %s", models.ErrorTypeMissingGoalName, stateErr.Type)
		}
	})

	t.Run("project not found error type", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		validator := state.NewStateValidator(tempDir)
		err := validator.ValidateStateMachine(tempDir)

		if err == nil {
			t.Fatal("expected error")
		}

		stateErr := models.AsStateError(err)
		if stateErr == nil {
			t.Fatal("expected non-nil StateError")
		}
		if stateErr.Type != models.ErrorTypeOrphanedState {
			t.Errorf("expected error type %s, got %s", models.ErrorTypeOrphanedState, stateErr.Type)
		}
	})
}

func TestCollectAllErrorsEdgeCases(t *testing.T) {
	t.Parallel()

	// Create a project with multiple errors
	projectRoot := testutil.NewTestProject(t).
		// Phase 1 with conflicting state
		WithPhase("0001", "open").
		WithPhaseGoal("0001", 1, "Goal 1", "open").
		// Phase 2 closed with open sprint (hierarchy error)
		WithPhase("0002", "closed").
		WithSprint("0002", "0001", "open").
		Build()

	// Add conflicting state to phase 1
	phase1Path := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")
	testutil.CreateFile(t, filepath.Join(phase1Path, "closed"))

	validator := state.NewStateValidator(projectRoot)
	errors := validator.CollectAllErrors(projectRoot)

	// Should have at least 2 errors
	if len(errors) < 2 {
		t.Errorf("expected at least 2 errors, got %d", len(errors))
	}

	// Verify errors include both types
	hasMutuallyExclusive := false
	hasHierarchy := false

	for _, err := range errors {
		if err.Type == models.ErrorTypeMutuallyExclusiveState {
			hasMutuallyExclusive = true
		}
		if err.Type == models.ErrorTypeHierarchyConstraint {
			hasHierarchy = true
		}
	}

	if !hasMutuallyExclusive {
		t.Error("expected mutually exclusive state error")
	}
	if !hasHierarchy {
		t.Error("expected hierarchy constraint error")
	}
}

// =============================================================================
// Query Function Error Handling
// =============================================================================

func TestQueryFunctionErrorsWithConflictingState(t *testing.T) {
	t.Parallel()

	t.Run("OpenPhaseExists with conflicting state", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithPhase("0001", "open").
			Build()

		phasePath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")
		testutil.CreateFile(t, filepath.Join(phasePath, "closed"))

		_, err := query.OpenPhaseExists(projectRoot)
		if err == nil {
			t.Error("expected error for conflicting state")
		}
		testutil.AssertError(t, err, "invalid state")
	})

	t.Run("OpenSprintExists with conflicting state", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithPhase("0001", "open").
			WithSprint("0001", "0001", "open").
			Build()

		sprintPath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
		testutil.CreateFile(t, filepath.Join(sprintPath, "closed"))

		phasePath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")
		_, err := query.OpenSprintExists(phasePath)
		if err == nil {
			t.Error("expected error for conflicting state")
		}
		testutil.AssertError(t, err, "invalid state")
	})

	t.Run("OpenTicketsExist with conflicting state", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithPhase("0001", "open").
			WithSprint("0001", "0001", "open").
			WithTicket("0001", "0001", "0001", "open").
			Build()

		ticketPath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint", "tickets", "0001-ticket")
		testutil.CreateFile(t, filepath.Join(ticketPath, "done"))

		sprintPath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
		_, err := query.OpenTicketsExist(sprintPath)
		if err == nil {
			t.Error("expected error for conflicting state")
		}
		testutil.AssertError(t, err, "invalid state")
	})

	t.Run("TicketComplete with conflicting state", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithPhase("0001", "open").
			WithSprint("0001", "0001", "open").
			WithTicket("0001", "0001", "0001", "open").
			Build()

		ticketPath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint", "tickets", "0001-ticket")
		testutil.CreateFile(t, filepath.Join(ticketPath, "done"))

		_, err := query.TicketComplete(ticketPath)
		if err == nil {
			t.Error("expected error for conflicting state")
		}
		testutil.AssertError(t, err, "invalid state")
	})

	t.Run("PhaseGoalsMet with conflicting goal state", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithPhase("0001", "open").
			WithPhaseGoal("0001", 1, "Test Goal", "open").
			WithSprint("0001", "0001", "closed").
			Build()

		goalPath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "goals", "0001-goal")
		testutil.CreateFile(t, filepath.Join(goalPath, "closed"))

		phasePath := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase")
		_, err := query.PhaseGoalsMet(phasePath)
		if err == nil {
			t.Error("expected error for conflicting goal state")
		}
		testutil.AssertError(t, err, "invalid state")
	})
}

// =============================================================================
// GetStatus Function Tests
// =============================================================================

func TestGetStatusFunction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setup          func(t *testing.T, path string)
		expectedStatus models.Status
		expectError    bool
	}{
		{
			name: "status open",
			setup: func(t *testing.T, path string) {
				testutil.CreateFile(t, filepath.Join(path, "open"))
			},
			expectedStatus: models.StatusOpen,
			expectError:    false,
		},
		{
			name: "status closed",
			setup: func(t *testing.T, path string) {
				testutil.CreateFile(t, filepath.Join(path, "closed"))
			},
			expectedStatus: models.StatusClosed,
			expectError:    false,
		},
		{
			name: "status done",
			setup: func(t *testing.T, path string) {
				testutil.CreateFile(t, filepath.Join(path, "done"))
			},
			expectedStatus: models.StatusDone,
			expectError:    false,
		},
		{
			name: "no status file",
			setup: func(t *testing.T, path string) {
				// Don't create any status file
			},
			expectedStatus: models.StatusUnknown,
			expectError:    true,
		},
		{
			name: "open and closed conflict",
			setup: func(t *testing.T, path string) {
				testutil.CreateFile(t, filepath.Join(path, "open"))
				testutil.CreateFile(t, filepath.Join(path, "closed"))
			},
			expectedStatus: models.StatusUnknown,
			expectError:    true,
		},
		{
			name: "open and done conflict",
			setup: func(t *testing.T, path string) {
				testutil.CreateFile(t, filepath.Join(path, "open"))
				testutil.CreateFile(t, filepath.Join(path, "done"))
			},
			expectedStatus: models.StatusUnknown,
			expectError:    true,
		},
		{
			name: "closed and done conflict",
			setup: func(t *testing.T, path string) {
				testutil.CreateFile(t, filepath.Join(path, "closed"))
				testutil.CreateFile(t, filepath.Join(path, "done"))
			},
			expectedStatus: models.StatusUnknown,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			testDir := filepath.Join(tempDir, "test-entity")
			testutil.CreateDir(t, testDir)

			tt.setup(t, testDir)

			status, err := state.GetStatus(testDir)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				testutil.AssertNoError(t, err)
			}

			if status != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, status)
			}
		})
	}
}

// =============================================================================
// Status Transition Validation Tests
// =============================================================================

func TestSetClosedValidated(t *testing.T) {
	t.Parallel()

	t.Run("valid transition from open to closed", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		entityPath := filepath.Join(tempDir, "entity")
		testutil.CreateDir(t, entityPath)
		testutil.CreateFile(t, filepath.Join(entityPath, "open"))

		err := state.SetClosedValidated(entityPath)
		testutil.AssertNoError(t, err)

		// Verify closed file exists and open file doesn't
		testutil.AssertFileExists(t, filepath.Join(entityPath, "closed"))
		testutil.AssertFileNotExists(t, filepath.Join(entityPath, "open"))
	})

	t.Run("invalid transition from closed", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		entityPath := filepath.Join(tempDir, "entity")
		testutil.CreateDir(t, entityPath)
		testutil.CreateFile(t, filepath.Join(entityPath, "closed"))

		err := state.SetClosedValidated(entityPath)
		if err == nil {
			t.Error("expected error when transitioning from closed")
		}
	})

	t.Run("invalid transition from done", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		entityPath := filepath.Join(tempDir, "entity")
		testutil.CreateDir(t, entityPath)
		testutil.CreateFile(t, filepath.Join(entityPath, "done"))

		err := state.SetClosedValidated(entityPath)
		if err == nil {
			t.Error("expected error when transitioning from done")
		}
	})
}

func TestSetDoneValidated(t *testing.T) {
	t.Parallel()

	t.Run("valid transition from open to done", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		entityPath := filepath.Join(tempDir, "entity")
		testutil.CreateDir(t, entityPath)
		testutil.CreateFile(t, filepath.Join(entityPath, "open"))

		err := state.SetDoneValidated(entityPath)
		testutil.AssertNoError(t, err)

		// Verify done file exists and open file doesn't
		testutil.AssertFileExists(t, filepath.Join(entityPath, "done"))
		testutil.AssertFileNotExists(t, filepath.Join(entityPath, "open"))
	})

	t.Run("invalid transition from done", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		entityPath := filepath.Join(tempDir, "entity")
		testutil.CreateDir(t, entityPath)
		testutil.CreateFile(t, filepath.Join(entityPath, "done"))

		err := state.SetDoneValidated(entityPath)
		if err == nil {
			t.Error("expected error when transitioning from done")
		}
	})

	t.Run("invalid transition from closed", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		entityPath := filepath.Join(tempDir, "entity")
		testutil.CreateDir(t, entityPath)
		testutil.CreateFile(t, filepath.Join(entityPath, "closed"))

		err := state.SetDoneValidated(entityPath)
		if err == nil {
			t.Error("expected error when transitioning from closed")
		}
	})
}

// =============================================================================
// Goal List and Management Tests
// =============================================================================

func TestListGoals(t *testing.T) {
	t.Parallel()

	t.Run("list goals sorted by index", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithPhase("0001", "open").
			WithPhaseGoal("0001", 3, "Goal Three", "open").
			WithPhaseGoal("0001", 1, "Goal One", "closed").
			WithPhaseGoal("0001", 2, "Goal Two", "open").
			Build()

		goalsDir := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "goals")
		goals, err := state.ListGoals(goalsDir)
		testutil.AssertNoError(t, err)

		if len(goals) != 3 {
			t.Fatalf("expected 3 goals, got %d", len(goals))
		}

		// Verify sorted order
		if goals[0].Index != 1 || goals[1].Index != 2 || goals[2].Index != 3 {
			t.Errorf("goals not sorted by index: %v, %v, %v", goals[0].Index, goals[1].Index, goals[2].Index)
		}

		// Verify names
		if goals[0].Name != "Goal One" {
			t.Errorf("expected 'Goal One', got %q", goals[0].Name)
		}
	})

	t.Run("list goals with empty directory", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithPhase("0001", "open").
			Build()

		goalsDir := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "goals")
		goals, err := state.ListGoals(goalsDir)
		testutil.AssertNoError(t, err)

		if len(goals) != 0 {
			t.Errorf("expected 0 goals, got %d", len(goals))
		}
	})

	t.Run("list goals with non-existent directory", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		goalsDir := filepath.Join(tempDir, "nonexistent", "goals")

		goals, err := state.ListGoals(goalsDir)
		testutil.AssertNoError(t, err)

		if goals != nil {
			t.Errorf("expected nil for non-existent directory, got %v", goals)
		}
	})
}

func TestGetNextGoalIndex(t *testing.T) {
	t.Parallel()

	t.Run("first goal", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithPhase("0001", "open").
			Build()

		goalsDir := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "goals")
		index, err := state.GetNextGoalIndex(goalsDir)
		testutil.AssertNoError(t, err)

		if index != 1 {
			t.Errorf("expected index 1 for first goal, got %d", index)
		}
	})

	t.Run("next goal after existing", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithPhase("0001", "open").
			WithPhaseGoal("0001", 1, "Goal 1", "open").
			WithPhaseGoal("0001", 2, "Goal 2", "open").
			Build()

		goalsDir := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "goals")
		index, err := state.GetNextGoalIndex(goalsDir)
		testutil.AssertNoError(t, err)

		if index != 3 {
			t.Errorf("expected index 3, got %d", index)
		}
	})

	t.Run("handles gaps in numbering", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithPhase("0001", "open").
			WithPhaseGoal("0001", 1, "Goal 1", "open").
			WithPhaseGoal("0001", 5, "Goal 5", "open").
			Build()

		goalsDir := filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "goals")
		index, err := state.GetNextGoalIndex(goalsDir)
		testutil.AssertNoError(t, err)

		// Should return max + 1, not fill in gaps
		if index != 6 {
			t.Errorf("expected index 6 (max+1), got %d", index)
		}
	})
}

func TestAreAllGoalsClosed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(t *testing.T) string
		expected bool
	}{
		{
			name: "no goals",
			setup: func(t *testing.T) string {
				projectRoot := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					Build()
				return filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "goals")
			},
			expected: true, // Vacuously true
		},
		{
			name: "all goals closed",
			setup: func(t *testing.T) string {
				projectRoot := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Goal 1", "closed").
					WithPhaseGoal("0001", 2, "Goal 2", "closed").
					Build()
				return filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "goals")
			},
			expected: true,
		},
		{
			name: "one goal open",
			setup: func(t *testing.T) string {
				projectRoot := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Goal 1", "closed").
					WithPhaseGoal("0001", 2, "Goal 2", "open").
					Build()
				return filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "goals")
			},
			expected: false,
		},
		{
			name: "all goals open",
			setup: func(t *testing.T) string {
				projectRoot := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Goal 1", "open").
					WithPhaseGoal("0001", 2, "Goal 2", "open").
					Build()
				return filepath.Join(projectRoot, ".crumbler", "phases", "0001-phase", "goals")
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			goalsDir := tt.setup(t)
			result, err := state.AreAllGoalsClosed(goalsDir)
			testutil.AssertNoError(t, err)

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
