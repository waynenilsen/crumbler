// Package state provides tests for state machine validation.
package state

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/waynenilsen/crumbler/internal/models"
	"github.com/waynenilsen/crumbler/internal/testutil"
)

// TestValidStateTransitions tests valid state transitions for phases, sprints, and tickets.
func TestValidStateTransitions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func(t *testing.T) (string, string) // returns (projectRoot, entityPath)
		transition  func(path string) error
		wantStatus  models.Status
		description string
	}{
		{
			name: "phase open to closed",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					Build()
				return root, filepath.Join(root, ".crumbler", "phases", "0001-phase")
			},
			transition: func(path string) error {
				return SetClosedValidated(path)
			},
			wantStatus:  models.StatusClosed,
			description: "Phase should transition from open to closed",
		},
		{
			name: "sprint open to closed",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					Build()
				return root, filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
			},
			transition: func(path string) error {
				return SetClosedValidated(path)
			},
			wantStatus:  models.StatusClosed,
			description: "Sprint should transition from open to closed",
		},
		{
			name: "ticket open to done",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "open").
					Build()
				return root, filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint", "tickets", "0001-ticket")
			},
			transition: func(path string) error {
				return SetDoneValidated(path)
			},
			wantStatus:  models.StatusDone,
			description: "Ticket should transition from open to done",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, entityPath := tc.setup(t)

			// Perform the transition
			err := tc.transition(entityPath)
			testutil.AssertNoError(t, err)

			// Verify the new status
			status, err := GetStatus(entityPath)
			testutil.AssertNoError(t, err)
			if status != tc.wantStatus {
				t.Errorf("expected status %v, got %v", tc.wantStatus, status)
			}
		})
	}
}

// TestInvalidStateTransitions tests that invalid state transitions are rejected.
func TestInvalidStateTransitions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setup          func(t *testing.T) (string, string) // returns (projectRoot, entityPath)
		transition     func(path string) error
		wantErrContain string
		description    string
	}{
		{
			name: "closed to open not allowed for phase",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "closed").
					Build()
				return root, filepath.Join(root, ".crumbler", "phases", "0001-phase")
			},
			transition: func(path string) error {
				// Try to set closed again - should fail since already closed
				return SetClosedValidated(path)
			},
			wantErrContain: "cannot transition from closed to closed",
			description:    "Should error when trying to close an already closed phase",
		},
		{
			name: "done to open not allowed for ticket",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "done").
					Build()
				return root, filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint", "tickets", "0001-ticket")
			},
			transition: func(path string) error {
				// Try to set done again - should fail since already done
				return SetDoneValidated(path)
			},
			wantErrContain: "cannot transition from done to done",
			description:    "Should error when trying to mark done an already done ticket",
		},
		{
			name: "closed phase cannot transition to closed again",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "closed").
					Build()
				return root, filepath.Join(root, ".crumbler", "phases", "0001-phase")
			},
			transition: func(path string) error {
				return SetClosedValidated(path)
			},
			wantErrContain: "cannot transition",
			description:    "Should error when trying to close an already closed phase",
		},
		{
			name: "closed sprint cannot transition to closed again",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "closed").
					Build()
				return root, filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
			},
			transition: func(path string) error {
				return SetClosedValidated(path)
			},
			wantErrContain: "cannot transition",
			description:    "Should error when trying to close an already closed sprint",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, entityPath := tc.setup(t)

			// Attempt the invalid transition
			err := tc.transition(entityPath)
			if err == nil {
				t.Errorf("expected error but got nil")
				return
			}
			testutil.AssertError(t, err, tc.wantErrContain)
		})
	}
}

// TestMutuallyExclusiveStates tests that having both open and closed (or done) files is detected as an error.
func TestMutuallyExclusiveStates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setup          func(t *testing.T) (string, string) // returns (projectRoot, entityPath)
		wantErrContain string
		description    string
	}{
		{
			name: "phase with both open and closed",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					Build()
				phasePath := filepath.Join(root, ".crumbler", "phases", "0001-phase")
				// Add conflicting closed file
				testutil.CreateFile(t, filepath.Join(phasePath, "closed"))
				return root, phasePath
			},
			wantErrContain: "both open and closed",
			description:    "Should detect mutually exclusive open/closed state in phase",
		},
		{
			name: "sprint with both open and closed",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					Build()
				sprintPath := filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
				// Add conflicting closed file
				testutil.CreateFile(t, filepath.Join(sprintPath, "closed"))
				return root, sprintPath
			},
			wantErrContain: "both open and closed",
			description:    "Should detect mutually exclusive open/closed state in sprint",
		},
		{
			name: "ticket with both open and done",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "open").
					Build()
				ticketPath := filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint", "tickets", "0001-ticket")
				// Add conflicting done file
				testutil.CreateFile(t, filepath.Join(ticketPath, "done"))
				return root, ticketPath
			},
			wantErrContain: "both open and done",
			description:    "Should detect mutually exclusive open/done state in ticket",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, entityPath := tc.setup(t)

			// Validate should fail
			_, err := GetStatus(entityPath)
			if err == nil {
				t.Errorf("expected error but got nil")
				return
			}
			testutil.AssertError(t, err, tc.wantErrContain)
		})
	}
}

// TestGoalStateTransitions tests valid goal state transitions (open to closed).
func TestGoalStateTransitions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func(t *testing.T) (string, string) // returns (projectRoot, goalPath)
		description string
	}{
		{
			name: "phase goal open to closed",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Test goal", "open").
					Build()
				return root, filepath.Join(root, ".crumbler", "phases", "0001-phase", "goals", "0001-goal")
			},
			description: "Phase goal should transition from open to closed",
		},
		{
			name: "sprint goal open to closed",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 1, "Sprint goal", "open").
					Build()
				return root, filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint", "goals", "0001-goal")
			},
			description: "Sprint goal should transition from open to closed",
		},
		{
			name: "ticket goal open to closed",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "open").
					WithTicketGoal("0001", "0001", "0001", 1, "Ticket goal", "open").
					Build()
				return root, filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint", "tickets", "0001-ticket", "goals", "0001-goal")
			},
			description: "Ticket goal should transition from open to closed",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, goalPath := tc.setup(t)

			// Verify initial state is open
			isOpen, err := IsOpen(goalPath)
			testutil.AssertNoError(t, err)
			if !isOpen {
				t.Errorf("expected goal to be open initially")
				return
			}

			// Perform the transition
			err = SetClosedValidated(goalPath)
			testutil.AssertNoError(t, err)

			// Verify the new status is closed
			isClosed, err := IsClosed(goalPath)
			testutil.AssertNoError(t, err)
			if !isClosed {
				t.Errorf("expected goal to be closed after transition")
			}

			// Verify open file is gone
			isOpen, err = IsOpen(goalPath)
			testutil.AssertNoError(t, err)
			if isOpen {
				t.Errorf("expected open file to be removed after transition")
			}
		})
	}
}

// TestGoalMutuallyExclusiveStates tests that goals with both open and closed files are detected as errors.
func TestGoalMutuallyExclusiveStates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setup          func(t *testing.T) (string, string) // returns (projectRoot, goalPath)
		wantErrContain string
		description    string
	}{
		{
			name: "phase goal with both open and closed",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Test goal", "open").
					Build()
				goalPath := filepath.Join(root, ".crumbler", "phases", "0001-phase", "goals", "0001-goal")
				// Add conflicting closed file
				testutil.CreateFile(t, filepath.Join(goalPath, "closed"))
				return root, goalPath
			},
			wantErrContain: "'open' and 'closed'",
			description:    "Should detect mutually exclusive open/closed state in phase goal",
		},
		{
			name: "sprint goal with both open and closed",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 1, "Sprint goal", "open").
					Build()
				goalPath := filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint", "goals", "0001-goal")
				// Add conflicting closed file
				testutil.CreateFile(t, filepath.Join(goalPath, "closed"))
				return root, goalPath
			},
			wantErrContain: "'open' and 'closed'",
			description:    "Should detect mutually exclusive open/closed state in sprint goal",
		},
		{
			name: "ticket goal with both open and closed",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "open").
					WithTicketGoal("0001", "0001", "0001", 1, "Ticket goal", "open").
					Build()
				goalPath := filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint", "tickets", "0001-ticket", "goals", "0001-goal")
				// Add conflicting closed file
				testutil.CreateFile(t, filepath.Join(goalPath, "closed"))
				return root, goalPath
			},
			wantErrContain: "'open' and 'closed'",
			description:    "Should detect mutually exclusive open/closed state in ticket goal",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			projectRoot, goalPath := tc.setup(t)

			// Use StateValidator to validate goal state
			validator := NewStateValidator(projectRoot)
			err := validator.ValidateGoalState(goalPath)
			if err == nil {
				t.Errorf("expected error but got nil")
				return
			}
			testutil.AssertError(t, err, tc.wantErrContain)
		})
	}
}

// TestHierarchyConstraintPhaseWithOpenSprints tests that a phase cannot be closed with open sprints.
func TestHierarchyConstraintPhaseWithOpenSprints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setup          func(t *testing.T) (string, string) // returns (projectRoot, phasePath)
		wantErrContain string
		description    string
	}{
		{
			name: "cannot close phase with open sprint",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					Build()
				return root, filepath.Join(root, ".crumbler", "phases", "0001-phase")
			},
			wantErrContain: "open sprints",
			description:    "Should error when trying to close phase with open sprint",
		},
		{
			name: "cannot close phase with multiple open sprints",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithSprint("0001", "0002", "open").
					Build()
				return root, filepath.Join(root, ".crumbler", "phases", "0001-phase")
			},
			wantErrContain: "open sprints",
			description:    "Should error when trying to close phase with multiple open sprints",
		},
		{
			name: "cannot close phase with open phase goals",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Open goal", "open").
					Build()
				return root, filepath.Join(root, ".crumbler", "phases", "0001-phase")
			},
			wantErrContain: "open goals",
			description:    "Should error when trying to close phase with open phase goals",
		},
		{
			name: "can close phase with closed sprints and closed goals",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "closed").
					WithPhaseGoal("0001", 1, "Closed goal", "closed").
					Build()
				return root, filepath.Join(root, ".crumbler", "phases", "0001-phase")
			},
			wantErrContain: "", // No error expected
			description:    "Should allow closing phase with all sprints and goals closed",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			projectRoot, phasePath := tc.setup(t)
			validator := NewStateValidator(projectRoot)

			canClose, err := validator.CanClosePhase(phasePath)

			if tc.wantErrContain != "" {
				if err == nil {
					t.Errorf("expected error containing %q but got nil", tc.wantErrContain)
					return
				}
				testutil.AssertError(t, err, tc.wantErrContain)
				if canClose {
					t.Errorf("expected canClose to be false when error occurs")
				}
			} else {
				testutil.AssertNoError(t, err)
				if !canClose {
					t.Errorf("expected canClose to be true")
				}
			}
		})
	}
}

// TestHierarchyConstraintSprintWithOpenTickets tests that a sprint cannot be closed with open tickets.
func TestHierarchyConstraintSprintWithOpenTickets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setup          func(t *testing.T) (string, string) // returns (projectRoot, sprintPath)
		wantErrContain string
		description    string
	}{
		{
			name: "cannot close sprint with open ticket",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "open").
					Build()
				return root, filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
			},
			wantErrContain: "open tickets",
			description:    "Should error when trying to close sprint with open ticket",
		},
		{
			name: "cannot close sprint with multiple open tickets",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "open").
					WithTicket("0001", "0001", "0002", "open").
					Build()
				return root, filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
			},
			wantErrContain: "open tickets",
			description:    "Should error when trying to close sprint with multiple open tickets",
		},
		{
			name: "cannot close sprint with open sprint goals",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 1, "Open goal", "open").
					Build()
				return root, filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
			},
			wantErrContain: "open goals",
			description:    "Should error when trying to close sprint with open sprint goals",
		},
		{
			name: "can close sprint with done tickets and closed goals",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "done").
					WithSprintGoal("0001", "0001", 1, "Closed goal", "closed").
					Build()
				return root, filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
			},
			wantErrContain: "", // No error expected
			description:    "Should allow closing sprint with all tickets done and goals closed",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			projectRoot, sprintPath := tc.setup(t)
			validator := NewStateValidator(projectRoot)

			canClose, err := validator.CanCloseSprint(sprintPath)

			if tc.wantErrContain != "" {
				if err == nil {
					t.Errorf("expected error containing %q but got nil", tc.wantErrContain)
					return
				}
				testutil.AssertError(t, err, tc.wantErrContain)
				if canClose {
					t.Errorf("expected canClose to be false when error occurs")
				}
			} else {
				testutil.AssertNoError(t, err)
				if !canClose {
					t.Errorf("expected canClose to be true")
				}
			}
		})
	}
}

// TestHierarchyConstraintTicketWithOpenGoals tests that a ticket cannot be marked done with open goals.
func TestHierarchyConstraintTicketWithOpenGoals(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setup          func(t *testing.T) (string, string) // returns (projectRoot, ticketPath)
		wantErrContain string
		description    string
	}{
		{
			name: "cannot mark ticket done with open goal",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "open").
					WithTicketGoal("0001", "0001", "0001", 1, "Open goal", "open").
					Build()
				return root, filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint", "tickets", "0001-ticket")
			},
			wantErrContain: "open goals",
			description:    "Should error when trying to mark ticket done with open goal",
		},
		{
			name: "cannot mark ticket done with multiple open goals",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "open").
					WithTicketGoal("0001", "0001", "0001", 1, "Open goal 1", "open").
					WithTicketGoal("0001", "0001", "0001", 2, "Open goal 2", "open").
					Build()
				return root, filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint", "tickets", "0001-ticket")
			},
			wantErrContain: "open goals",
			description:    "Should error when trying to mark ticket done with multiple open goals",
		},
		{
			name: "can mark ticket done with all goals closed",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "open").
					WithTicketGoal("0001", "0001", "0001", 1, "Closed goal 1", "closed").
					WithTicketGoal("0001", "0001", "0001", 2, "Closed goal 2", "closed").
					Build()
				return root, filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint", "tickets", "0001-ticket")
			},
			wantErrContain: "", // No error expected
			description:    "Should allow marking ticket done with all goals closed",
		},
		{
			name: "can mark ticket done with no goals",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "open").
					Build()
				return root, filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint", "tickets", "0001-ticket")
			},
			wantErrContain: "", // No error expected - no goals means vacuously all closed
			description:    "Should allow marking ticket done with no goals (vacuously true)",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			projectRoot, ticketPath := tc.setup(t)
			validator := NewStateValidator(projectRoot)

			canMarkDone, err := validator.CanMarkTicketDone(ticketPath)

			if tc.wantErrContain != "" {
				if err == nil {
					t.Errorf("expected error containing %q but got nil", tc.wantErrContain)
					return
				}
				testutil.AssertError(t, err, tc.wantErrContain)
				if canMarkDone {
					t.Errorf("expected canMarkDone to be false when error occurs")
				}
			} else {
				testutil.AssertNoError(t, err)
				if !canMarkDone {
					t.Errorf("expected canMarkDone to be true")
				}
			}
		})
	}
}

// TestValidateStateMachine tests the full state machine validation.
func TestValidateStateMachine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setup          func(t *testing.T) string // returns projectRoot
		wantErr        bool
		wantErrContain string
		description    string
	}{
		{
			name: "valid state machine with open phase",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithPhase("0001", "open").
					Build()
			},
			wantErr:     false,
			description: "Should pass validation with valid open phase",
		},
		{
			name: "valid state machine with closed phase and closed sprints",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithPhase("0001", "closed").
					WithSprint("0001", "0001", "closed").
					Build()
			},
			wantErr:     false,
			description: "Should pass validation with closed phase and closed sprints",
		},
		{
			name: "invalid state machine - closed phase with open sprint",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					Build()
				// Manually set phase to closed to create invalid state
				phasePath := filepath.Join(root, ".crumbler", "phases", "0001-phase")
				os.Remove(filepath.Join(phasePath, "open"))
				testutil.CreateFile(t, filepath.Join(phasePath, "closed"))
				return root
			},
			wantErr:        true,
			wantErrContain: "open sprints",
			description:    "Should fail validation when closed phase has open sprints",
		},
		{
			name: "invalid state machine - closed sprint with open tickets",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "open").
					Build()
				// Manually set sprint to closed to create invalid state
				sprintPath := filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
				os.Remove(filepath.Join(sprintPath, "open"))
				testutil.CreateFile(t, filepath.Join(sprintPath, "closed"))
				return root
			},
			wantErr:        true,
			wantErrContain: "open tickets",
			description:    "Should fail validation when closed sprint has open tickets",
		},
		{
			name: "invalid state machine - done ticket with open goals",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "open").
					WithTicketGoal("0001", "0001", "0001", 1, "Open goal", "open").
					Build()
				// Manually set ticket to done to create invalid state
				ticketPath := filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint", "tickets", "0001-ticket")
				os.Remove(filepath.Join(ticketPath, "open"))
				testutil.CreateFile(t, filepath.Join(ticketPath, "done"))
				return root
			},
			wantErr:        true,
			wantErrContain: "open goals",
			description:    "Should fail validation when done ticket has open goals",
		},
		{
			name: "invalid state machine - mutually exclusive state in phase",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					Build()
				// Add conflicting closed file
				phasePath := filepath.Join(root, ".crumbler", "phases", "0001-phase")
				testutil.CreateFile(t, filepath.Join(phasePath, "closed"))
				return root
			},
			wantErr:        true,
			wantErrContain: "both 'open' and 'closed'",
			description:    "Should fail validation when phase has both open and closed files",
		},
		{
			name: "invalid state machine - closed phase with open phase goals",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Open goal", "open").
					Build()
				// Manually set phase to closed to create invalid state
				phasePath := filepath.Join(root, ".crumbler", "phases", "0001-phase")
				os.Remove(filepath.Join(phasePath, "open"))
				testutil.CreateFile(t, filepath.Join(phasePath, "closed"))
				return root
			},
			wantErr:        true,
			wantErrContain: "open goals",
			description:    "Should fail validation when closed phase has open phase goals",
		},
		{
			name: "invalid state machine - closed sprint with open sprint goals",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 1, "Open goal", "open").
					Build()
				// Manually set sprint to closed to create invalid state
				sprintPath := filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
				os.Remove(filepath.Join(sprintPath, "open"))
				testutil.CreateFile(t, filepath.Join(sprintPath, "closed"))
				return root
			},
			wantErr:        true,
			wantErrContain: "open goals",
			description:    "Should fail validation when closed sprint has open sprint goals",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			projectRoot := tc.setup(t)
			validator := NewStateValidator(projectRoot)

			err := validator.ValidateStateMachine(projectRoot)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
					return
				}
				if tc.wantErrContain != "" {
					testutil.AssertError(t, err, tc.wantErrContain)
				}
			} else {
				testutil.AssertNoError(t, err)
			}
		})
	}
}

// TestValidateHierarchy tests hierarchy validation specifically.
func TestValidateHierarchy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setup          func(t *testing.T) string // returns projectRoot
		wantErr        bool
		wantErrContain string
		description    string
	}{
		{
			name: "valid hierarchy - all open",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "open").
					Build()
			},
			wantErr:     false,
			description: "Should pass with all entities open",
		},
		{
			name: "valid hierarchy - properly closed",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithPhaseGoal("0001", 1, "Phase goal", "closed").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 1, "Sprint goal", "closed").
					WithTicket("0001", "0001", "0001", "open").
					WithTicketGoal("0001", "0001", "0001", 1, "Ticket goal", "closed").
					Build()
				// Manually close everything in the correct order (ticket -> sprint -> phase)
				ticketPath := filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint", "tickets", "0001-ticket")
				os.Remove(filepath.Join(ticketPath, "open"))
				testutil.CreateFile(t, filepath.Join(ticketPath, "done"))

				sprintPath := filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
				os.Remove(filepath.Join(sprintPath, "open"))
				testutil.CreateFile(t, filepath.Join(sprintPath, "closed"))

				phasePath := filepath.Join(root, ".crumbler", "phases", "0001-phase")
				os.Remove(filepath.Join(phasePath, "open"))
				testutil.CreateFile(t, filepath.Join(phasePath, "closed"))
				return root
			},
			wantErr:     false,
			description: "Should pass when all entities are properly closed with closed goals",
		},
		{
			name: "invalid hierarchy - closed phase with open sprint",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "closed").
					WithSprint("0001", "0001", "open").
					Build()
				return root
			},
			wantErr:        true,
			wantErrContain: "closed phase has open sprints",
			description:    "Should fail when closed phase has open sprint",
		},
		{
			name: "invalid hierarchy - closed phase with open phase goal",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithPhase("0001", "closed").
					WithPhaseGoal("0001", 1, "Open goal", "open").
					Build()
			},
			wantErr:        true,
			wantErrContain: "closed phase has open goals",
			description:    "Should fail when closed phase has open phase goal",
		},
		{
			name: "invalid hierarchy - closed sprint with open ticket",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "open").
					Build()
				// Manually close the sprint to create invalid hierarchy
				sprintPath := filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
				os.Remove(filepath.Join(sprintPath, "open"))
				testutil.CreateFile(t, filepath.Join(sprintPath, "closed"))
				return root
			},
			wantErr:        true,
			wantErrContain: "closed sprint has open tickets",
			description:    "Should fail when closed sprint has open ticket",
		},
		{
			name: "invalid hierarchy - closed sprint with open sprint goal",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithSprintGoal("0001", "0001", 1, "Open goal", "open").
					Build()
				// Manually close the sprint to create invalid hierarchy
				sprintPath := filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
				os.Remove(filepath.Join(sprintPath, "open"))
				testutil.CreateFile(t, filepath.Join(sprintPath, "closed"))
				return root
			},
			wantErr:        true,
			wantErrContain: "closed sprint has open goals",
			description:    "Should fail when closed sprint has open sprint goal",
		},
		{
			name: "invalid hierarchy - done ticket with open ticket goal",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "open").
					WithTicketGoal("0001", "0001", "0001", 1, "Open goal", "open").
					Build()
				// Manually mark ticket as done to create invalid hierarchy
				ticketPath := filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint", "tickets", "0001-ticket")
				os.Remove(filepath.Join(ticketPath, "open"))
				testutil.CreateFile(t, filepath.Join(ticketPath, "done"))
				return root
			},
			wantErr:        true,
			wantErrContain: "done ticket has open goals",
			description:    "Should fail when done ticket has open ticket goal",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			projectRoot := tc.setup(t)
			validator := NewStateValidator(projectRoot)

			err := validator.ValidateHierarchy(projectRoot)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
					return
				}
				if tc.wantErrContain != "" {
					testutil.AssertError(t, err, tc.wantErrContain)
				}
			} else {
				testutil.AssertNoError(t, err)
			}
		})
	}
}

// TestValidatePhaseState tests phase state validation.
func TestValidatePhaseState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setup          func(t *testing.T) (string, string) // returns (projectRoot, phasePath)
		wantErr        bool
		wantErrContain string
	}{
		{
			name: "valid open phase",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					Build()
				return root, filepath.Join(root, ".crumbler", "phases", "0001-phase")
			},
			wantErr: false,
		},
		{
			name: "valid closed phase",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "closed").
					Build()
				return root, filepath.Join(root, ".crumbler", "phases", "0001-phase")
			},
			wantErr: false,
		},
		{
			name: "invalid phase with both open and closed",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					Build()
				phasePath := filepath.Join(root, ".crumbler", "phases", "0001-phase")
				testutil.CreateFile(t, filepath.Join(phasePath, "closed"))
				return root, phasePath
			},
			wantErr:        true,
			wantErrContain: "both 'open' and 'closed'",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			projectRoot, phasePath := tc.setup(t)
			validator := NewStateValidator(projectRoot)

			err := validator.ValidatePhaseState(phasePath)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
					return
				}
				if tc.wantErrContain != "" {
					testutil.AssertError(t, err, tc.wantErrContain)
				}
			} else {
				testutil.AssertNoError(t, err)
			}
		})
	}
}

// TestValidateSprintState tests sprint state validation.
func TestValidateSprintState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setup          func(t *testing.T) (string, string) // returns (projectRoot, sprintPath)
		wantErr        bool
		wantErrContain string
	}{
		{
			name: "valid open sprint",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					Build()
				return root, filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
			},
			wantErr: false,
		},
		{
			name: "valid closed sprint",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "closed").
					Build()
				return root, filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
			},
			wantErr: false,
		},
		{
			name: "invalid sprint with both open and closed",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					Build()
				sprintPath := filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
				testutil.CreateFile(t, filepath.Join(sprintPath, "closed"))
				return root, sprintPath
			},
			wantErr:        true,
			wantErrContain: "both 'open' and 'closed'",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			projectRoot, sprintPath := tc.setup(t)
			validator := NewStateValidator(projectRoot)

			err := validator.ValidateSprintState(sprintPath)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
					return
				}
				if tc.wantErrContain != "" {
					testutil.AssertError(t, err, tc.wantErrContain)
				}
			} else {
				testutil.AssertNoError(t, err)
			}
		})
	}
}

// TestValidateTicketState tests ticket state validation.
func TestValidateTicketState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setup          func(t *testing.T) (string, string) // returns (projectRoot, ticketPath)
		wantErr        bool
		wantErrContain string
	}{
		{
			name: "valid open ticket",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "open").
					Build()
				return root, filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint", "tickets", "0001-ticket")
			},
			wantErr: false,
		},
		{
			name: "valid done ticket",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "done").
					Build()
				return root, filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint", "tickets", "0001-ticket")
			},
			wantErr: false,
		},
		{
			name: "invalid ticket with both open and done",
			setup: func(t *testing.T) (string, string) {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					WithTicket("0001", "0001", "0001", "open").
					Build()
				ticketPath := filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint", "tickets", "0001-ticket")
				testutil.CreateFile(t, filepath.Join(ticketPath, "done"))
				return root, ticketPath
			},
			wantErr:        true,
			wantErrContain: "both 'open' and 'done'",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			projectRoot, ticketPath := tc.setup(t)
			validator := NewStateValidator(projectRoot)

			err := validator.ValidateTicketState(ticketPath)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
					return
				}
				if tc.wantErrContain != "" {
					testutil.AssertError(t, err, tc.wantErrContain)
				}
			} else {
				testutil.AssertNoError(t, err)
			}
		})
	}
}

// TestCollectAllErrors tests that multiple errors can be collected.
func TestCollectAllErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setup          func(t *testing.T) string // returns projectRoot
		wantErrCount   int
		wantErrTypes   []string
		description    string
	}{
		{
			name: "no errors in valid project",
			setup: func(t *testing.T) string {
				return testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					Build()
			},
			wantErrCount: 0,
			description:  "Should have no errors in valid project",
		},
		{
			name: "multiple mutually exclusive errors",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "open").
					WithSprint("0001", "0001", "open").
					Build()
				// Create conflicting states
				phasePath := filepath.Join(root, ".crumbler", "phases", "0001-phase")
				testutil.CreateFile(t, filepath.Join(phasePath, "closed"))
				sprintPath := filepath.Join(phasePath, "sprints", "0001-sprint")
				testutil.CreateFile(t, filepath.Join(sprintPath, "closed"))
				return root
			},
			// 2 mutually exclusive errors + 1 hierarchy error (closed phase has open sprint due to the conflicting state)
			wantErrCount: 3,
			wantErrTypes: []string{models.ErrorTypeMutuallyExclusiveState, models.ErrorTypeMutuallyExclusiveState, models.ErrorTypeHierarchyConstraint},
			description:  "Should collect multiple mutually exclusive state errors plus hierarchy errors",
		},
		{
			name: "hierarchy and goal errors",
			setup: func(t *testing.T) string {
				root := testutil.NewTestProject(t).
					WithPhase("0001", "closed").
					WithSprint("0001", "0001", "open").
					WithPhaseGoal("0001", 1, "Open goal", "open").
					Build()
				return root
			},
			wantErrCount: 2,
			wantErrTypes: []string{models.ErrorTypeHierarchyConstraint, models.ErrorTypeHierarchyConstraint},
			description:  "Should collect hierarchy constraint errors for open sprint and open goal in closed phase",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			projectRoot := tc.setup(t)
			validator := NewStateValidator(projectRoot)

			errors := validator.CollectAllErrors(projectRoot)

			if len(errors) != tc.wantErrCount {
				t.Errorf("expected %d errors, got %d", tc.wantErrCount, len(errors))
				for i, err := range errors {
					t.Logf("error %d: %v", i, err)
				}
				return
			}

			// Verify error types if specified
			if len(tc.wantErrTypes) > 0 {
				for i, wantType := range tc.wantErrTypes {
					if i < len(errors) && errors[i].Type != wantType {
						t.Errorf("error %d: expected type %q, got %q", i, wantType, errors[i].Type)
					}
				}
			}
		})
	}
}

// TestGoalMissingNameFile tests that goals without name files are detected.
func TestGoalMissingNameFile(t *testing.T) {
	t.Parallel()

	root := testutil.NewTestProject(t).
		WithPhase("0001", "open").
		WithPhaseGoal("0001", 1, "Test goal", "open").
		Build()

	goalPath := filepath.Join(root, ".crumbler", "phases", "0001-phase", "goals", "0001-goal")

	// Remove the name file
	os.Remove(filepath.Join(goalPath, "name"))

	validator := NewStateValidator(root)
	err := validator.ValidateGoalState(goalPath)

	if err == nil {
		t.Errorf("expected error for missing name file")
		return
	}

	testutil.AssertError(t, err, "name")
}

// TestEmptyProject tests validation of an empty project.
func TestEmptyProject(t *testing.T) {
	t.Parallel()

	root := testutil.NewTestProject(t).Build()

	validator := NewStateValidator(root)
	err := validator.ValidateStateMachine(root)

	testutil.AssertNoError(t, err)
}

// TestMultiplePhases tests validation with multiple phases.
func TestMultiplePhases(t *testing.T) {
	t.Parallel()

	root := testutil.NewTestProject(t).
		WithPhase("0001", "closed").
		WithSprint("0001", "0001", "closed").
		WithPhase("0002", "open").
		WithSprint("0002", "0001", "open").
		Build()

	validator := NewStateValidator(root)
	err := validator.ValidateStateMachine(root)

	testutil.AssertNoError(t, err)
}

// TestComplexHierarchy tests validation of a complex hierarchy.
func TestComplexHierarchy(t *testing.T) {
	t.Parallel()

	// Build a complex valid hierarchy where phase 1 is closed properly and phase 2 is open
	root := testutil.NewTestProject(t).
		// Phase 1 - will be closed manually after setup
		WithPhase("0001", "open").
		WithPhaseGoal("0001", 1, "Phase 1 Goal 1", "closed").
		WithPhaseGoal("0001", 2, "Phase 1 Goal 2", "closed").
		WithSprint("0001", "0001", "open").
		WithSprintGoal("0001", "0001", 1, "Sprint 1.1 Goal", "closed").
		WithTicket("0001", "0001", "0001", "open").
		WithTicketGoal("0001", "0001", "0001", 1, "Ticket Goal", "closed").
		WithSprint("0001", "0002", "open").
		// Phase 2 - open and in progress
		WithPhase("0002", "open").
		WithPhaseGoal("0002", 1, "Phase 2 Goal 1", "open").
		WithSprint("0002", "0001", "open").
		WithSprintGoal("0002", "0001", 1, "Sprint Goal", "open").
		WithTicket("0002", "0001", "0001", "open").
		WithTicketGoal("0002", "0001", "0001", 1, "Ticket Goal", "open").
		Build()

	// Properly close phase 1's hierarchy (ticket -> sprint -> phase)
	ticket1Path := filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint", "tickets", "0001-ticket")
	os.Remove(filepath.Join(ticket1Path, "open"))
	testutil.CreateFile(t, filepath.Join(ticket1Path, "done"))

	sprint1Path := filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
	os.Remove(filepath.Join(sprint1Path, "open"))
	testutil.CreateFile(t, filepath.Join(sprint1Path, "closed"))

	sprint2Path := filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0002-sprint")
	os.Remove(filepath.Join(sprint2Path, "open"))
	testutil.CreateFile(t, filepath.Join(sprint2Path, "closed"))

	phase1Path := filepath.Join(root, ".crumbler", "phases", "0001-phase")
	os.Remove(filepath.Join(phase1Path, "open"))
	testutil.CreateFile(t, filepath.Join(phase1Path, "closed"))

	validator := NewStateValidator(root)
	err := validator.ValidateStateMachine(root)

	testutil.AssertNoError(t, err)
}
