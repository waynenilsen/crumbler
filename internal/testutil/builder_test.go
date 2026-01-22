package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewTestProject(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.Build()

	// Verify the project root was created
	AssertDirExists(t, root)

	// Verify .crumbler directory structure
	crumblerDir := filepath.Join(root, ".crumbler")
	AssertDirExists(t, crumblerDir)
	AssertFileExists(t, filepath.Join(crumblerDir, "README.md"))
	AssertFileExists(t, filepath.Join(crumblerDir, "roadmap.md"))
	AssertDirExists(t, filepath.Join(crumblerDir, "phases"))
	AssertDirExists(t, filepath.Join(crumblerDir, "roadmaps"))
}

func TestWithPhase(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.
		WithPhase("0001", "open").
		Build()

	phasePath := filepath.Join(root, ".crumbler", "phases", "0001-phase")
	AssertDirExists(t, phasePath)
	AssertStatus(t, phasePath, "open")
	AssertFileExists(t, filepath.Join(phasePath, "README.md"))
	AssertDirExists(t, filepath.Join(phasePath, "goals"))
	AssertDirExists(t, filepath.Join(phasePath, "sprints"))
}

func TestWithPhaseGoal(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.
		WithPhase("0001", "open").
		WithPhaseGoal("0001", 1, "Implement authentication", "open").
		WithPhaseGoal("0001", 2, "Configure database", "closed").
		Build()

	phasePath := filepath.Join(root, ".crumbler", "phases", "0001-phase")
	goalPath1 := filepath.Join(phasePath, "goals", "0001-goal")
	goalPath2 := filepath.Join(phasePath, "goals", "0002-goal")

	AssertDirExists(t, goalPath1)
	AssertGoalStatus(t, goalPath1, "open")
	AssertGoalName(t, goalPath1, "Implement authentication")

	AssertDirExists(t, goalPath2)
	AssertGoalStatus(t, goalPath2, "closed")
	AssertGoalName(t, goalPath2, "Configure database")
}

func TestWithSprint(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.
		WithPhase("0001", "open").
		WithSprint("0001", "0001", "open").
		Build()

	sprintPath := filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
	AssertDirExists(t, sprintPath)
	AssertStatus(t, sprintPath, "open")
	AssertFileExists(t, filepath.Join(sprintPath, "README.md"))
	AssertFileExists(t, filepath.Join(sprintPath, "PRD.md"))
	AssertFileExists(t, filepath.Join(sprintPath, "ERD.md"))
	AssertDirExists(t, filepath.Join(sprintPath, "goals"))
	AssertDirExists(t, filepath.Join(sprintPath, "tickets"))
}

func TestWithSprintGoal(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.
		WithSprint("0001", "0001", "open").
		WithSprintGoal("0001", "0001", 1, "Complete API endpoints", "open").
		Build()

	sprintPath := filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
	goalPath := filepath.Join(sprintPath, "goals", "0001-goal")

	AssertDirExists(t, goalPath)
	AssertGoalStatus(t, goalPath, "open")
	AssertGoalName(t, goalPath, "Complete API endpoints")
}

func TestWithTicket(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.
		WithSprint("0001", "0001", "open").
		WithTicket("0001", "0001", "0001", "open").
		Build()

	ticketPath := filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint", "tickets", "0001-ticket")
	AssertDirExists(t, ticketPath)
	AssertStatus(t, ticketPath, "open")
	AssertFileExists(t, filepath.Join(ticketPath, "README.md"))
	AssertDirExists(t, filepath.Join(ticketPath, "goals"))
}

func TestWithTicketGoal(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.
		WithTicket("0001", "0001", "0001", "open").
		WithTicketGoal("0001", "0001", "0001", 1, "Write unit tests", "closed").
		Build()

	ticketPath := filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint", "tickets", "0001-ticket")
	goalPath := filepath.Join(ticketPath, "goals", "0001-goal")

	AssertDirExists(t, goalPath)
	AssertGoalStatus(t, goalPath, "closed")
	AssertGoalName(t, goalPath, "Write unit tests")
}

func TestWithRoadmap(t *testing.T) {
	t.Parallel()

	customRoadmap := "# Custom Roadmap\n\n## Phase 1\n- Task 1\n- Task 2\n"
	builder := NewTestProject(t)
	root := builder.
		WithRoadmap(customRoadmap).
		Build()

	roadmapPath := filepath.Join(root, ".crumbler", "roadmap.md")
	AssertFileContent(t, roadmapPath, customRoadmap)
}

func TestWithPRDAndERD(t *testing.T) {
	t.Parallel()

	customPRD := "# Custom PRD\n\nProduct requirements here.\n"
	customERD := "# Custom ERD\n\nEntity relationships here.\n"

	builder := NewTestProject(t)
	root := builder.
		WithSprint("0001", "0001", "open").
		WithPRD("0001", "0001", customPRD).
		WithERD("0001", "0001", customERD).
		Build()

	sprintPath := filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint")
	AssertFileContent(t, filepath.Join(sprintPath, "PRD.md"), customPRD)
	AssertFileContent(t, filepath.Join(sprintPath, "ERD.md"), customERD)
}

func TestWithTicketDescription(t *testing.T) {
	t.Parallel()

	customDesc := "# Custom Ticket\n\nThis is a custom ticket description.\n"
	builder := NewTestProject(t)
	root := builder.
		WithTicket("0001", "0001", "0001", "open").
		WithTicketDescription("0001", "0001", "0001", customDesc).
		Build()

	ticketPath := filepath.Join(root, ".crumbler", "phases", "0001-phase", "sprints", "0001-sprint", "tickets", "0001-ticket")
	AssertFileContent(t, filepath.Join(ticketPath, "README.md"), customDesc)
}

func TestBuilderPathHelpers(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.
		WithTicket("0001", "0001", "0001", "open").
		Build()

	// Test path helper methods
	expectedPhase := filepath.Join(root, ".crumbler", "phases", "0001-phase")
	if got := builder.PhasePath("0001"); got != expectedPhase {
		t.Errorf("PhasePath mismatch: expected %s, got %s", expectedPhase, got)
	}

	expectedSprint := filepath.Join(expectedPhase, "sprints", "0001-sprint")
	if got := builder.SprintPath("0001", "0001"); got != expectedSprint {
		t.Errorf("SprintPath mismatch: expected %s, got %s", expectedSprint, got)
	}

	expectedTicket := filepath.Join(expectedSprint, "tickets", "0001-ticket")
	if got := builder.TicketPath("0001", "0001", "0001"); got != expectedTicket {
		t.Errorf("TicketPath mismatch: expected %s, got %s", expectedTicket, got)
	}

	expectedGoal := filepath.Join(expectedTicket, "goals", "0001-goal")
	if got := builder.GoalPath(expectedTicket, 1); got != expectedGoal {
		t.Errorf("GoalPath mismatch: expected %s, got %s", expectedGoal, got)
	}
}

func TestComplexProjectStructure(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.
		// Phase 1 - closed
		WithPhase("0001", "closed").
		WithPhaseGoal("0001", 1, "Phase 1 Goal 1", "closed").
		WithPhaseGoal("0001", 2, "Phase 1 Goal 2", "closed").
		WithSprint("0001", "0001", "closed").
		WithSprintGoal("0001", "0001", 1, "Sprint 1 Goal", "closed").
		WithTicket("0001", "0001", "0001", "done").
		WithTicketGoal("0001", "0001", "0001", 1, "Ticket Goal", "closed").
		// Phase 2 - open with mixed sprint/ticket states
		WithPhase("0002", "open").
		WithPhaseGoal("0002", 1, "Phase 2 Goal", "open").
		WithSprint("0002", "0001", "closed").
		WithSprint("0002", "0002", "open").
		WithTicket("0002", "0002", "0001", "done").
		WithTicket("0002", "0002", "0002", "open").
		Build()

	// Verify Phase 1 structure
	phase1 := filepath.Join(root, ".crumbler", "phases", "0001-phase")
	AssertStatus(t, phase1, "closed")
	AssertGoalCount(t, phase1, 2)

	// Verify Phase 2 structure
	phase2 := filepath.Join(root, ".crumbler", "phases", "0002-phase")
	AssertStatus(t, phase2, "open")
	AssertGoalCount(t, phase2, 1)

	// Verify Sprint states
	sprint1 := filepath.Join(phase2, "sprints", "0001-sprint")
	sprint2 := filepath.Join(phase2, "sprints", "0002-sprint")
	AssertStatus(t, sprint1, "closed")
	AssertStatus(t, sprint2, "open")

	// Verify Ticket states
	ticket1 := filepath.Join(sprint2, "tickets", "0001-ticket")
	ticket2 := filepath.Join(sprint2, "tickets", "0002-ticket")
	AssertStatus(t, ticket1, "done")
	AssertStatus(t, ticket2, "open")
}

func TestParallelExecution(t *testing.T) {
	t.Parallel()

	// Run multiple subtests in parallel to verify isolation
	for i := 0; i < 5; i++ {
		i := i // capture loop variable
		t.Run("parallel", func(t *testing.T) {
			t.Parallel()

			builder := NewTestProject(t)
			root := builder.
				WithPhase("0001", "open").
				Build()

			// Verify each test has its own isolated directory
			AssertDirExists(t, root)
			phasePath := filepath.Join(root, ".crumbler", "phases", "0001-phase")
			AssertStatus(t, phasePath, "open")

			// Verify directory names are unique (contain timestamp and random)
			info, _ := os.Stat(root)
			t.Logf("Parallel test %d: created directory %s", i, info.Name())
		})
	}
}

func TestCleanup(t *testing.T) {
	// This test verifies that cleanup happens after test completion
	// We can't easily verify this within the test itself, but we can
	// verify the cleanup mechanism is in place by checking the builder
	builder := NewTestProject(t)
	root := builder.Build()

	// Directory should exist during the test
	AssertDirExists(t, root)

	// The cleanup function is registered via t.Cleanup(), which will
	// be called after this test function completes
}
