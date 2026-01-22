// Package ticket provides Ticket management functions for crumbler.
package ticket

import (
	"path/filepath"
	"testing"

	"github.com/waynenilsen/crumbler/internal/models"
	"github.com/waynenilsen/crumbler/internal/testutil"
)

// TestCreateTicket tests creating a ticket in a sprint.
func TestCreateTicket(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		index       int
		wantID      string
		wantReadme  bool
		wantGoals   bool
		wantOpen    bool
	}{
		{
			name:       "create first ticket",
			index:      1,
			wantID:     "0001-ticket",
			wantReadme: true,
			wantGoals:  true,
			wantOpen:   true,
		},
		{
			name:       "create ticket with high index",
			index:      42,
			wantID:     "0042-ticket",
			wantReadme: true,
			wantGoals:  true,
			wantOpen:   true,
		},
		{
			name:       "create ticket with max 4-digit index",
			index:      9999,
			wantID:     "9999-ticket",
			wantReadme: true,
			wantGoals:  true,
			wantOpen:   true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Build test project with a sprint
			builder := testutil.NewTestProject(t).
				WithPhase("0001", "open").
				WithSprint("0001", "0001", "open")
			projectRoot := builder.Build()
			sprintPath := builder.SprintPath("0001", "0001")

			// Create ticket
			ticketPath, err := CreateTicket(sprintPath, tt.index)
			testutil.AssertNoError(t, err)

			// Verify ticket directory was created
			testutil.AssertDirExists(t, ticketPath)

			// Verify ticket ID
			expectedPath := filepath.Join(sprintPath, models.TicketsDir, tt.wantID)
			if ticketPath != expectedPath {
				t.Errorf("expected ticket path %s, got %s", expectedPath, ticketPath)
			}

			// Verify README.md was created
			if tt.wantReadme {
				readmePath := filepath.Join(ticketPath, models.ReadmeFile)
				testutil.AssertFileExists(t, readmePath)
			}

			// Verify goals/ directory was created
			if tt.wantGoals {
				goalsPath := filepath.Join(ticketPath, models.GoalsDir)
				testutil.AssertDirExists(t, goalsPath)
			}

			// Verify open file was created
			if tt.wantOpen {
				testutil.AssertStatus(t, ticketPath, "open")
			}

			_ = projectRoot // Use projectRoot to avoid unused variable warning
		})
	}
}

// TestCreateTicketGoal tests creating ticket goals.
func TestCreateTicketGoal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		goalIndex  int
		goalName   string
		wantID     string
		wantStatus string
	}{
		{
			name:       "create first goal",
			goalIndex:  1,
			goalName:   "Implement authentication",
			wantID:     "0001-goal",
			wantStatus: "open",
		},
		{
			name:       "create second goal",
			goalIndex:  2,
			goalName:   "Add unit tests",
			wantID:     "0002-goal",
			wantStatus: "open",
		},
		{
			name:       "create goal with long name",
			goalIndex:  3,
			goalName:   "Implement comprehensive error handling and logging system with detailed stack traces",
			wantID:     "0003-goal",
			wantStatus: "open",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Build test project with a ticket
			builder := testutil.NewTestProject(t).
				WithTicket("0001", "0001", "0001", "open")
			builder.Build()
			ticketPath := builder.TicketPath("0001", "0001", "0001")

			// Create ticket goal
			goalPath, err := CreateTicketGoal(ticketPath, tt.goalIndex, tt.goalName)
			testutil.AssertNoError(t, err)

			// Verify goal directory was created
			testutil.AssertDirExists(t, goalPath)

			// Verify goal ID
			expectedPath := filepath.Join(ticketPath, models.GoalsDir, tt.wantID)
			if goalPath != expectedPath {
				t.Errorf("expected goal path %s, got %s", expectedPath, goalPath)
			}

			// Verify name file was created with correct content
			testutil.AssertGoalName(t, goalPath, tt.goalName)

			// Verify goal status
			testutil.AssertGoalStatus(t, goalPath, tt.wantStatus)
		})
	}
}

// TestMarkTicketDone tests marking a ticket as done.
func TestMarkTicketDone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(*testutil.TestProjectBuilder)
		wantErr   bool
		errSubstr string
	}{
		{
			name: "mark ticket done with no goals",
			setup: func(b *testutil.TestProjectBuilder) {
				b.WithTicket("0001", "0001", "0001", "open")
			},
			wantErr: false,
		},
		{
			name: "mark ticket done with all goals closed",
			setup: func(b *testutil.TestProjectBuilder) {
				b.WithTicket("0001", "0001", "0001", "open").
					WithTicketGoal("0001", "0001", "0001", 1, "Goal 1", "closed").
					WithTicketGoal("0001", "0001", "0001", 2, "Goal 2", "closed")
			},
			wantErr: false,
		},
		{
			name: "error when marking ticket done with open goals",
			setup: func(b *testutil.TestProjectBuilder) {
				b.WithTicket("0001", "0001", "0001", "open").
					WithTicketGoal("0001", "0001", "0001", 1, "Goal 1", "open").
					WithTicketGoal("0001", "0001", "0001", 2, "Goal 2", "closed")
			},
			wantErr:   true,
			errSubstr: "goals still open",
		},
		{
			name: "error when marking already done ticket",
			setup: func(b *testutil.TestProjectBuilder) {
				b.WithTicket("0001", "0001", "0001", "done")
			},
			wantErr:   true,
			errSubstr: "already done",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			builder := testutil.NewTestProject(t)
			tt.setup(builder)
			builder.Build()
			ticketPath := builder.TicketPath("0001", "0001", "0001")

			err := MarkTicketDone(ticketPath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errSubstr)
				} else {
					testutil.AssertError(t, err, tt.errSubstr)
				}
			} else {
				testutil.AssertNoError(t, err)
				testutil.AssertStatus(t, ticketPath, "done")
			}
		})
	}
}

// TestGetOpenTickets tests getting open tickets from a sprint.
func TestGetOpenTickets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(*testutil.TestProjectBuilder)
		wantCount int
		wantIDs   []string
	}{
		{
			name: "no tickets",
			setup: func(b *testutil.TestProjectBuilder) {
				b.WithSprint("0001", "0001", "open")
			},
			wantCount: 0,
			wantIDs:   nil,
		},
		{
			name: "all tickets open",
			setup: func(b *testutil.TestProjectBuilder) {
				b.WithTicket("0001", "0001", "0001", "open").
					WithTicket("0001", "0001", "0002", "open").
					WithTicket("0001", "0001", "0003", "open")
			},
			wantCount: 3,
			wantIDs:   []string{"0001-ticket", "0002-ticket", "0003-ticket"},
		},
		{
			name: "mixed open and done tickets",
			setup: func(b *testutil.TestProjectBuilder) {
				b.WithTicket("0001", "0001", "0001", "done").
					WithTicket("0001", "0001", "0002", "open").
					WithTicket("0001", "0001", "0003", "done").
					WithTicket("0001", "0001", "0004", "open")
			},
			wantCount: 2,
			wantIDs:   []string{"0002-ticket", "0004-ticket"},
		},
		{
			name: "all tickets done",
			setup: func(b *testutil.TestProjectBuilder) {
				b.WithTicket("0001", "0001", "0001", "done").
					WithTicket("0001", "0001", "0002", "done")
			},
			wantCount: 0,
			wantIDs:   nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			builder := testutil.NewTestProject(t)
			tt.setup(builder)
			builder.Build()
			sprintPath := builder.SprintPath("0001", "0001")

			tickets, err := GetOpenTickets(sprintPath)
			testutil.AssertNoError(t, err)

			if len(tickets) != tt.wantCount {
				t.Errorf("expected %d open tickets, got %d", tt.wantCount, len(tickets))
			}

			if tt.wantIDs != nil {
				for i, wantID := range tt.wantIDs {
					if i >= len(tickets) {
						t.Errorf("missing ticket at index %d, expected %s", i, wantID)
						continue
					}
					if tickets[i].ID != wantID {
						t.Errorf("ticket[%d].ID = %s, want %s", i, tickets[i].ID, wantID)
					}
				}
			}
		})
	}
}

// TestIsTicketComplete tests ticket completion detection.
func TestIsTicketComplete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		setup        func(*testutil.TestProjectBuilder)
		wantComplete bool
	}{
		{
			name: "open ticket is not complete",
			setup: func(b *testutil.TestProjectBuilder) {
				b.WithTicket("0001", "0001", "0001", "open")
			},
			wantComplete: false,
		},
		{
			name: "done ticket with no goals is complete",
			setup: func(b *testutil.TestProjectBuilder) {
				b.WithTicket("0001", "0001", "0001", "done")
			},
			wantComplete: true,
		},
		{
			name: "done ticket with all goals closed is complete",
			setup: func(b *testutil.TestProjectBuilder) {
				// Note: WithTicketGoal calls WithTicket internally with "open" status,
				// so we need to set goals first, then set the final status
				b.WithTicketGoal("0001", "0001", "0001", 1, "Goal 1", "closed").
					WithTicketGoal("0001", "0001", "0001", 2, "Goal 2", "closed").
					WithTicket("0001", "0001", "0001", "done")
			},
			wantComplete: true,
		},
		{
			name: "done ticket with open goals is not complete",
			setup: func(b *testutil.TestProjectBuilder) {
				// Note: WithTicketGoal calls WithTicket internally with "open" status,
				// so we need to set goals first, then set the final status
				b.WithTicketGoal("0001", "0001", "0001", 1, "Goal 1", "closed").
					WithTicketGoal("0001", "0001", "0001", 2, "Goal 2", "open").
					WithTicket("0001", "0001", "0001", "done")
			},
			wantComplete: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			builder := testutil.NewTestProject(t)
			tt.setup(builder)
			builder.Build()
			ticketPath := builder.TicketPath("0001", "0001", "0001")

			complete, err := IsTicketComplete(ticketPath)
			testutil.AssertNoError(t, err)

			if complete != tt.wantComplete {
				t.Errorf("IsTicketComplete() = %v, want %v", complete, tt.wantComplete)
			}
		})
	}
}

// TestGetNextTicketIndex tests getting the next ticket index.
func TestGetNextTicketIndex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(*testutil.TestProjectBuilder)
		wantIndex int
	}{
		{
			name: "no tickets returns 1",
			setup: func(b *testutil.TestProjectBuilder) {
				b.WithSprint("0001", "0001", "open")
			},
			wantIndex: 1,
		},
		{
			name: "one ticket returns 2",
			setup: func(b *testutil.TestProjectBuilder) {
				b.WithTicket("0001", "0001", "0001", "open")
			},
			wantIndex: 2,
		},
		{
			name: "multiple tickets returns max + 1",
			setup: func(b *testutil.TestProjectBuilder) {
				b.WithTicket("0001", "0001", "0001", "open").
					WithTicket("0001", "0001", "0002", "done").
					WithTicket("0001", "0001", "0003", "open")
			},
			wantIndex: 4,
		},
		{
			name: "gap in ticket numbers returns max + 1",
			setup: func(b *testutil.TestProjectBuilder) {
				b.WithTicket("0001", "0001", "0001", "open").
					WithTicket("0001", "0001", "0005", "open").
					WithTicket("0001", "0001", "0010", "done")
			},
			wantIndex: 11,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			builder := testutil.NewTestProject(t)
			tt.setup(builder)
			builder.Build()
			sprintPath := builder.SprintPath("0001", "0001")

			index, err := GetNextTicketIndex(sprintPath)
			testutil.AssertNoError(t, err)

			if index != tt.wantIndex {
				t.Errorf("GetNextTicketIndex() = %d, want %d", index, tt.wantIndex)
			}
		})
	}
}

// TestAreTicketGoalsMet tests checking if all ticket goals are met.
func TestAreTicketGoalsMet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(*testutil.TestProjectBuilder)
		wantMet bool
	}{
		{
			name: "no goals means goals are met (vacuously true)",
			setup: func(b *testutil.TestProjectBuilder) {
				b.WithTicket("0001", "0001", "0001", "open")
			},
			wantMet: true,
		},
		{
			name: "all goals closed means goals are met",
			setup: func(b *testutil.TestProjectBuilder) {
				b.WithTicket("0001", "0001", "0001", "open").
					WithTicketGoal("0001", "0001", "0001", 1, "Goal 1", "closed").
					WithTicketGoal("0001", "0001", "0001", 2, "Goal 2", "closed").
					WithTicketGoal("0001", "0001", "0001", 3, "Goal 3", "closed")
			},
			wantMet: true,
		},
		{
			name: "one open goal means goals are not met",
			setup: func(b *testutil.TestProjectBuilder) {
				b.WithTicket("0001", "0001", "0001", "open").
					WithTicketGoal("0001", "0001", "0001", 1, "Goal 1", "closed").
					WithTicketGoal("0001", "0001", "0001", 2, "Goal 2", "open").
					WithTicketGoal("0001", "0001", "0001", 3, "Goal 3", "closed")
			},
			wantMet: false,
		},
		{
			name: "all goals open means goals are not met",
			setup: func(b *testutil.TestProjectBuilder) {
				b.WithTicket("0001", "0001", "0001", "open").
					WithTicketGoal("0001", "0001", "0001", 1, "Goal 1", "open").
					WithTicketGoal("0001", "0001", "0001", 2, "Goal 2", "open")
			},
			wantMet: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			builder := testutil.NewTestProject(t)
			tt.setup(builder)
			builder.Build()
			ticketPath := builder.TicketPath("0001", "0001", "0001")

			met, err := AreTicketGoalsMet(ticketPath)
			testutil.AssertNoError(t, err)

			if met != tt.wantMet {
				t.Errorf("AreTicketGoalsMet() = %v, want %v", met, tt.wantMet)
			}
		})
	}
}

// TestValidateTicketState tests ticket state validation.
func TestValidateTicketState(t *testing.T) {
	t.Parallel()

	t.Run("valid open state", func(t *testing.T) {
		t.Parallel()

		builder := testutil.NewTestProject(t).
			WithTicket("0001", "0001", "0001", "open")
		builder.Build()
		ticketPath := builder.TicketPath("0001", "0001", "0001")

		err := ValidateTicketState(ticketPath)
		testutil.AssertNoError(t, err)
	})

	t.Run("valid done state", func(t *testing.T) {
		t.Parallel()

		builder := testutil.NewTestProject(t).
			WithTicket("0001", "0001", "0001", "done")
		builder.Build()
		ticketPath := builder.TicketPath("0001", "0001", "0001")

		err := ValidateTicketState(ticketPath)
		testutil.AssertNoError(t, err)
	})

	t.Run("invalid state with both open and done", func(t *testing.T) {
		t.Parallel()

		builder := testutil.NewTestProject(t).
			WithTicket("0001", "0001", "0001", "open")
		builder.Build()
		ticketPath := builder.TicketPath("0001", "0001", "0001")

		// Create conflicting state by adding 'done' file while 'open' already exists
		testutil.CreateFile(t, filepath.Join(ticketPath, "done"))

		err := ValidateTicketState(ticketPath)
		if err == nil {
			t.Error("expected error for invalid state, got nil")
		} else {
			testutil.AssertError(t, err, "both 'open' and 'done'")
		}
	})
}

// TestCloseTicketGoal tests closing a ticket goal.
func TestCloseTicketGoal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(*testutil.TestProjectBuilder)
		goalID    string
		wantErr   bool
		errSubstr string
	}{
		{
			name: "close open goal",
			setup: func(b *testutil.TestProjectBuilder) {
				b.WithTicket("0001", "0001", "0001", "open").
					WithTicketGoal("0001", "0001", "0001", 1, "Goal 1", "open")
			},
			goalID:  "0001-goal",
			wantErr: false,
		},
		{
			name: "error closing already closed goal",
			setup: func(b *testutil.TestProjectBuilder) {
				b.WithTicket("0001", "0001", "0001", "open").
					WithTicketGoal("0001", "0001", "0001", 1, "Goal 1", "closed")
			},
			goalID:    "0001-goal",
			wantErr:   true,
			errSubstr: "already closed",
		},
		{
			name: "error closing non-existent goal",
			setup: func(b *testutil.TestProjectBuilder) {
				b.WithTicket("0001", "0001", "0001", "open")
			},
			goalID:    "9999-goal",
			wantErr:   true,
			errSubstr: "not found",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			builder := testutil.NewTestProject(t)
			tt.setup(builder)
			builder.Build()
			ticketPath := builder.TicketPath("0001", "0001", "0001")

			err := CloseTicketGoal(ticketPath, tt.goalID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errSubstr)
				} else {
					testutil.AssertError(t, err, tt.errSubstr)
				}
			} else {
				testutil.AssertNoError(t, err)
				goalPath := filepath.Join(ticketPath, models.GoalsDir, tt.goalID)
				testutil.AssertGoalStatus(t, goalPath, "closed")
			}
		})
	}
}

// TestGetTicketGoals tests getting ticket goals.
func TestGetTicketGoals(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(*testutil.TestProjectBuilder)
		wantCount int
		wantNames []string
	}{
		{
			name: "no goals",
			setup: func(b *testutil.TestProjectBuilder) {
				b.WithTicket("0001", "0001", "0001", "open")
			},
			wantCount: 0,
			wantNames: nil,
		},
		{
			name: "single goal",
			setup: func(b *testutil.TestProjectBuilder) {
				b.WithTicket("0001", "0001", "0001", "open").
					WithTicketGoal("0001", "0001", "0001", 1, "First goal", "open")
			},
			wantCount: 1,
			wantNames: []string{"First goal"},
		},
		{
			name: "multiple goals sorted by index",
			setup: func(b *testutil.TestProjectBuilder) {
				b.WithTicket("0001", "0001", "0001", "open").
					WithTicketGoal("0001", "0001", "0001", 3, "Third goal", "open").
					WithTicketGoal("0001", "0001", "0001", 1, "First goal", "closed").
					WithTicketGoal("0001", "0001", "0001", 2, "Second goal", "open")
			},
			wantCount: 3,
			wantNames: []string{"First goal", "Second goal", "Third goal"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			builder := testutil.NewTestProject(t)
			tt.setup(builder)
			builder.Build()
			ticketPath := builder.TicketPath("0001", "0001", "0001")

			goals, err := GetTicketGoals(ticketPath)
			testutil.AssertNoError(t, err)

			if len(goals) != tt.wantCount {
				t.Errorf("expected %d goals, got %d", tt.wantCount, len(goals))
			}

			if tt.wantNames != nil {
				for i, wantName := range tt.wantNames {
					if i >= len(goals) {
						t.Errorf("missing goal at index %d, expected name %s", i, wantName)
						continue
					}
					if goals[i].Name != wantName {
						t.Errorf("goal[%d].Name = %s, want %s", i, goals[i].Name, wantName)
					}
				}
			}
		})
	}
}

// TestGetTicket tests loading a ticket by ID.
func TestGetTicket(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setup      func(*testutil.TestProjectBuilder)
		ticketID   string
		wantStatus models.Status
		wantErr    bool
		errSubstr  string
	}{
		{
			name: "get open ticket",
			setup: func(b *testutil.TestProjectBuilder) {
				b.WithTicket("0001", "0001", "0001", "open")
			},
			ticketID:   "0001-ticket",
			wantStatus: models.StatusOpen,
			wantErr:    false,
		},
		{
			name: "get done ticket",
			setup: func(b *testutil.TestProjectBuilder) {
				b.WithTicket("0001", "0001", "0001", "done")
			},
			ticketID:   "0001-ticket",
			wantStatus: models.StatusDone,
			wantErr:    false,
		},
		{
			name: "error for non-existent ticket",
			setup: func(b *testutil.TestProjectBuilder) {
				b.WithSprint("0001", "0001", "open")
			},
			ticketID:  "9999-ticket",
			wantErr:   true,
			errSubstr: "not found",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			builder := testutil.NewTestProject(t)
			tt.setup(builder)
			builder.Build()
			sprintPath := builder.SprintPath("0001", "0001")

			ticket, err := GetTicket(sprintPath, tt.ticketID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errSubstr)
				} else {
					testutil.AssertError(t, err, tt.errSubstr)
				}
			} else {
				testutil.AssertNoError(t, err)
				if ticket == nil {
					t.Fatal("expected ticket, got nil")
				}
				if ticket.Status != tt.wantStatus {
					t.Errorf("ticket.Status = %v, want %v", ticket.Status, tt.wantStatus)
				}
				if ticket.ID != tt.ticketID {
					t.Errorf("ticket.ID = %s, want %s", ticket.ID, tt.ticketID)
				}
			}
		})
	}
}

// TestListTickets tests listing all tickets in a sprint.
func TestListTickets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(*testutil.TestProjectBuilder)
		wantCount int
		wantIDs   []string
	}{
		{
			name: "no tickets",
			setup: func(b *testutil.TestProjectBuilder) {
				b.WithSprint("0001", "0001", "open")
			},
			wantCount: 0,
			wantIDs:   nil,
		},
		{
			name: "single ticket",
			setup: func(b *testutil.TestProjectBuilder) {
				b.WithTicket("0001", "0001", "0001", "open")
			},
			wantCount: 1,
			wantIDs:   []string{"0001-ticket"},
		},
		{
			name: "multiple tickets sorted by ID",
			setup: func(b *testutil.TestProjectBuilder) {
				b.WithTicket("0001", "0001", "0003", "done").
					WithTicket("0001", "0001", "0001", "open").
					WithTicket("0001", "0001", "0002", "done")
			},
			wantCount: 3,
			wantIDs:   []string{"0001-ticket", "0002-ticket", "0003-ticket"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			builder := testutil.NewTestProject(t)
			tt.setup(builder)
			builder.Build()
			sprintPath := builder.SprintPath("0001", "0001")

			tickets, err := ListTickets(sprintPath)
			testutil.AssertNoError(t, err)

			if len(tickets) != tt.wantCount {
				t.Errorf("expected %d tickets, got %d", tt.wantCount, len(tickets))
			}

			if tt.wantIDs != nil {
				for i, wantID := range tt.wantIDs {
					if i >= len(tickets) {
						t.Errorf("missing ticket at index %d, expected %s", i, wantID)
						continue
					}
					if tickets[i].ID != wantID {
						t.Errorf("ticket[%d].ID = %s, want %s", i, tickets[i].ID, wantID)
					}
				}
			}
		})
	}
}

// TestTicketGoalIntegration tests the full workflow of ticket goals.
func TestTicketGoalIntegration(t *testing.T) {
	t.Parallel()

	t.Run("create goals, close them, then mark ticket done", func(t *testing.T) {
		t.Parallel()

		builder := testutil.NewTestProject(t).
			WithTicket("0001", "0001", "0001", "open")
		builder.Build()
		ticketPath := builder.TicketPath("0001", "0001", "0001")

		// Create goals
		goal1Path, err := CreateTicketGoal(ticketPath, 1, "Implement feature")
		testutil.AssertNoError(t, err)
		testutil.AssertGoalStatus(t, goal1Path, "open")

		goal2Path, err := CreateTicketGoal(ticketPath, 2, "Add tests")
		testutil.AssertNoError(t, err)
		testutil.AssertGoalStatus(t, goal2Path, "open")

		// Verify goals are not met
		met, err := AreTicketGoalsMet(ticketPath)
		testutil.AssertNoError(t, err)
		if met {
			t.Error("expected goals not to be met when goals are open")
		}

		// Try to mark ticket done (should fail)
		err = MarkTicketDone(ticketPath)
		if err == nil {
			t.Error("expected error when marking ticket done with open goals")
		}

		// Close goals
		err = CloseTicketGoal(ticketPath, "0001-goal")
		testutil.AssertNoError(t, err)
		testutil.AssertGoalStatus(t, goal1Path, "closed")

		err = CloseTicketGoal(ticketPath, "0002-goal")
		testutil.AssertNoError(t, err)
		testutil.AssertGoalStatus(t, goal2Path, "closed")

		// Verify goals are now met
		met, err = AreTicketGoalsMet(ticketPath)
		testutil.AssertNoError(t, err)
		if !met {
			t.Error("expected goals to be met when all goals are closed")
		}

		// Mark ticket done (should succeed)
		err = MarkTicketDone(ticketPath)
		testutil.AssertNoError(t, err)
		testutil.AssertStatus(t, ticketPath, "done")

		// Verify ticket is complete
		complete, err := IsTicketComplete(ticketPath)
		testutil.AssertNoError(t, err)
		if !complete {
			t.Error("expected ticket to be complete")
		}
	})
}

// TestTicketCreationIntegration tests creating multiple tickets in sequence.
func TestTicketCreationIntegration(t *testing.T) {
	t.Parallel()

	builder := testutil.NewTestProject(t).
		WithPhase("0001", "open").
		WithSprint("0001", "0001", "open")
	builder.Build()
	sprintPath := builder.SprintPath("0001", "0001")

	// Get initial next index
	idx, err := GetNextTicketIndex(sprintPath)
	testutil.AssertNoError(t, err)
	if idx != 1 {
		t.Errorf("expected initial index 1, got %d", idx)
	}

	// Create first ticket
	ticket1Path, err := CreateTicket(sprintPath, idx)
	testutil.AssertNoError(t, err)
	testutil.AssertStatus(t, ticket1Path, "open")

	// Get next index
	idx, err = GetNextTicketIndex(sprintPath)
	testutil.AssertNoError(t, err)
	if idx != 2 {
		t.Errorf("expected next index 2, got %d", idx)
	}

	// Create second ticket
	ticket2Path, err := CreateTicket(sprintPath, idx)
	testutil.AssertNoError(t, err)
	testutil.AssertStatus(t, ticket2Path, "open")

	// List tickets
	tickets, err := ListTickets(sprintPath)
	testutil.AssertNoError(t, err)
	if len(tickets) != 2 {
		t.Errorf("expected 2 tickets, got %d", len(tickets))
	}

	// Get open tickets (should be 2)
	openTickets, err := GetOpenTickets(sprintPath)
	testutil.AssertNoError(t, err)
	if len(openTickets) != 2 {
		t.Errorf("expected 2 open tickets, got %d", len(openTickets))
	}

	// Mark first ticket done
	err = MarkTicketDone(ticket1Path)
	testutil.AssertNoError(t, err)

	// Get open tickets (should be 1)
	openTickets, err = GetOpenTickets(sprintPath)
	testutil.AssertNoError(t, err)
	if len(openTickets) != 1 {
		t.Errorf("expected 1 open ticket, got %d", len(openTickets))
	}
	if openTickets[0].ID != "0002-ticket" {
		t.Errorf("expected open ticket 0002-ticket, got %s", openTickets[0].ID)
	}
}

// TestTicketEdgeCases tests edge cases and error conditions.
func TestTicketEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("GetOpenTickets with non-existent sprint", func(t *testing.T) {
		t.Parallel()

		tickets, err := GetOpenTickets("/non/existent/path")
		testutil.AssertNoError(t, err)
		if tickets != nil && len(tickets) != 0 {
			t.Errorf("expected empty slice, got %v", tickets)
		}
	})

	t.Run("GetNextTicketIndex with non-existent sprint", func(t *testing.T) {
		t.Parallel()

		idx, err := GetNextTicketIndex("/non/existent/path")
		testutil.AssertNoError(t, err)
		if idx != 1 {
			t.Errorf("expected index 1 for non-existent path, got %d", idx)
		}
	})

	t.Run("GetTicketGoals with non-existent ticket", func(t *testing.T) {
		t.Parallel()

		goals, err := GetTicketGoals("/non/existent/path")
		testutil.AssertNoError(t, err)
		if goals != nil && len(goals) != 0 {
			t.Errorf("expected empty slice, got %v", goals)
		}
	})

	t.Run("ListTickets with non-existent sprint", func(t *testing.T) {
		t.Parallel()

		tickets, err := ListTickets("/non/existent/path")
		testutil.AssertNoError(t, err)
		if tickets != nil && len(tickets) != 0 {
			t.Errorf("expected empty slice, got %v", tickets)
		}
	})
}

// TestTicketWithGoalsWorkflow tests the complete workflow with ticket goals.
func TestTicketWithGoalsWorkflow(t *testing.T) {
	t.Parallel()

	builder := testutil.NewTestProject(t).
		WithTicket("0001", "0001", "0001", "open").
		WithTicketGoal("0001", "0001", "0001", 1, "Design API", "closed").
		WithTicketGoal("0001", "0001", "0001", 2, "Implement handlers", "closed").
		WithTicketGoal("0001", "0001", "0001", 3, "Write tests", "open")
	builder.Build()
	ticketPath := builder.TicketPath("0001", "0001", "0001")

	// Verify initial state
	goals, err := GetTicketGoals(ticketPath)
	testutil.AssertNoError(t, err)
	if len(goals) != 3 {
		t.Fatalf("expected 3 goals, got %d", len(goals))
	}

	// Verify goals are not met (one is still open)
	met, err := AreTicketGoalsMet(ticketPath)
	testutil.AssertNoError(t, err)
	if met {
		t.Error("expected goals not to be met")
	}

	// Try to mark ticket done (should fail)
	err = MarkTicketDone(ticketPath)
	if err == nil {
		t.Error("expected error when marking ticket done with open goals")
	}
	testutil.AssertError(t, err, "goals still open")

	// Close the remaining goal
	err = CloseTicketGoal(ticketPath, "0003-goal")
	testutil.AssertNoError(t, err)

	// Now goals should be met
	met, err = AreTicketGoalsMet(ticketPath)
	testutil.AssertNoError(t, err)
	if !met {
		t.Error("expected goals to be met after closing all goals")
	}

	// Mark ticket done (should succeed)
	err = MarkTicketDone(ticketPath)
	testutil.AssertNoError(t, err)

	// Verify ticket is complete
	complete, err := IsTicketComplete(ticketPath)
	testutil.AssertNoError(t, err)
	if !complete {
		t.Error("expected ticket to be complete")
	}
}
