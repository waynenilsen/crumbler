package testutil

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestAssertFileExists(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.Build()

	// Test with existing file
	existingFile := filepath.Join(root, ".crumbler", "README.md")
	AssertFileExists(t, existingFile) // Should not fail
}

func TestAssertFileNotExists(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.Build()

	// Test with non-existing file
	nonExistingFile := filepath.Join(root, "non-existing.txt")
	AssertFileNotExists(t, nonExistingFile) // Should not fail
}

func TestAssertDirExists(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.Build()

	// Test with existing directory
	existingDir := filepath.Join(root, ".crumbler", "phases")
	AssertDirExists(t, existingDir) // Should not fail
}

func TestAssertDirNotExists(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.Build()

	// Test with non-existing directory
	nonExistingDir := filepath.Join(root, "non-existing-dir")
	AssertDirNotExists(t, nonExistingDir) // Should not fail
}

func TestAssertFileContent(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.WithRoadmap("# Test Roadmap\n").Build()

	roadmapFile := filepath.Join(root, ".crumbler", "roadmap.md")
	AssertFileContent(t, roadmapFile, "# Test Roadmap\n") // Should not fail
}

func TestAssertFileContains(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.WithRoadmap("# Test Roadmap\nThis is a test.\n").Build()

	roadmapFile := filepath.Join(root, ".crumbler", "roadmap.md")
	AssertFileContains(t, roadmapFile, "Test Roadmap") // Should not fail
	AssertFileContains(t, roadmapFile, "test")         // Should not fail
}

func TestAssertStatus(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.
		WithPhase("0001", "open").
		WithPhase("0002", "closed").
		WithTicket("0003", "0001", "0001", "done").
		Build()

	// Test open status
	phase1 := filepath.Join(root, ".crumbler", "phases", "0001-phase")
	AssertStatus(t, phase1, "open")

	// Test closed status
	phase2 := filepath.Join(root, ".crumbler", "phases", "0002-phase")
	AssertStatus(t, phase2, "closed")

	// Test done status
	ticket := filepath.Join(root, ".crumbler", "phases", "0003-phase", "sprints", "0001-sprint", "tickets", "0001-ticket")
	AssertStatus(t, ticket, "done")
}

func TestAssertGoalStatus(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.
		WithPhaseGoal("0001", 1, "Open Goal", "open").
		WithPhaseGoal("0001", 2, "Closed Goal", "closed").
		Build()

	phasePath := filepath.Join(root, ".crumbler", "phases", "0001-phase")
	goal1 := filepath.Join(phasePath, "goals", "0001-goal")
	goal2 := filepath.Join(phasePath, "goals", "0002-goal")

	AssertGoalStatus(t, goal1, "open")
	AssertGoalStatus(t, goal2, "closed")
}

func TestAssertGoalName(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.
		WithPhaseGoal("0001", 1, "My Custom Goal", "open").
		Build()

	goalPath := filepath.Join(root, ".crumbler", "phases", "0001-phase", "goals", "0001-goal")
	AssertGoalName(t, goalPath, "My Custom Goal")
}

func TestAssertError(t *testing.T) {
	t.Parallel()

	err := errors.New("this is a test error with specific text")
	AssertError(t, err, "test error")
	AssertError(t, err, "specific text")
}

func TestAssertErrorIs(t *testing.T) {
	t.Parallel()

	target := os.ErrNotExist
	err := os.ErrNotExist
	AssertErrorIs(t, err, target)
}

func TestAssertNoError(t *testing.T) {
	t.Parallel()

	AssertNoError(t, nil)
}

func TestReadFile(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.WithRoadmap("# Custom Content\n").Build()

	roadmapFile := filepath.Join(root, ".crumbler", "roadmap.md")
	content := ReadFile(t, roadmapFile)

	if content != "# Custom Content\n" {
		t.Errorf("ReadFile returned unexpected content: %s", content)
	}
}

func TestListDir(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.
		WithPhase("0001", "open").
		WithPhase("0002", "closed").
		Build()

	phasesDir := filepath.Join(root, ".crumbler", "phases")
	entries := ListDir(t, phasesDir)

	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d: %v", len(entries), entries)
	}
}

func TestListDirRecursive(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.
		WithPhase("0001", "open").
		WithPhaseGoal("0001", 1, "Goal", "open").
		Build()

	phasePath := filepath.Join(root, ".crumbler", "phases", "0001-phase")
	entries := ListDirRecursive(t, phasePath)

	// Should include: README.md, open, goals/, goals/0001-goal/, goals/0001-goal/name, goals/0001-goal/open, sprints/
	if len(entries) < 5 {
		t.Errorf("expected at least 5 entries, got %d: %v", len(entries), entries)
	}
}

func TestWriteFile(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.Build()

	testFile := filepath.Join(root, "test-dir", "test.txt")
	WriteFile(t, testFile, "test content")

	content := ReadFile(t, testFile)
	if content != "test content" {
		t.Errorf("WriteFile/ReadFile mismatch: got %s", content)
	}
}

func TestCreateFile(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.Build()

	testFile := filepath.Join(root, "empty.txt")
	CreateFile(t, testFile)

	AssertFileExists(t, testFile)
	content := ReadFile(t, testFile)
	if content != "" {
		t.Errorf("CreateFile should create empty file, got: %s", content)
	}
}

func TestRemoveFile(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.Build()

	testFile := filepath.Join(root, "to-remove.txt")
	WriteFile(t, testFile, "content")
	AssertFileExists(t, testFile)

	RemoveFile(t, testFile)
	AssertFileNotExists(t, testFile)
}

func TestCreateDir(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.Build()

	testDir := filepath.Join(root, "new-dir", "nested")
	CreateDir(t, testDir)

	AssertDirExists(t, testDir)
}

func TestTouchStatusFile(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.Build()

	testDir := filepath.Join(root, "status-test")
	CreateDir(t, testDir)
	TouchStatusFile(t, testDir, "open")

	AssertFileExists(t, filepath.Join(testDir, "open"))
}

func TestSetStatus(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.WithPhase("0001", "open").Build()

	phasePath := filepath.Join(root, ".crumbler", "phases", "0001-phase")

	// Initially open
	AssertStatus(t, phasePath, "open")

	// Change to closed
	SetStatus(t, phasePath, "closed")
	AssertStatus(t, phasePath, "closed")

	// Verify open file was removed
	AssertFileNotExists(t, filepath.Join(phasePath, "open"))
}

func TestCountGoals(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.
		WithPhase("0001", "open").
		WithPhaseGoal("0001", 1, "Goal 1", "open").
		WithPhaseGoal("0001", 2, "Goal 2", "closed").
		WithPhaseGoal("0001", 3, "Goal 3", "open").
		Build()

	phasePath := filepath.Join(root, ".crumbler", "phases", "0001-phase")
	count := CountGoals(t, phasePath)

	if count != 3 {
		t.Errorf("expected 3 goals, got %d", count)
	}
}

func TestCountGoalsEmpty(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.WithPhase("0001", "open").Build()

	phasePath := filepath.Join(root, ".crumbler", "phases", "0001-phase")
	count := CountGoals(t, phasePath)

	if count != 0 {
		t.Errorf("expected 0 goals, got %d", count)
	}
}

func TestGetOpenGoals(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.
		WithPhase("0001", "open").
		WithPhaseGoal("0001", 1, "Goal 1", "open").
		WithPhaseGoal("0001", 2, "Goal 2", "closed").
		WithPhaseGoal("0001", 3, "Goal 3", "open").
		Build()

	phasePath := filepath.Join(root, ".crumbler", "phases", "0001-phase")
	openGoals := GetOpenGoals(t, phasePath)

	if len(openGoals) != 2 {
		t.Errorf("expected 2 open goals, got %d: %v", len(openGoals), openGoals)
	}
}

func TestGetClosedGoals(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.
		WithPhase("0001", "open").
		WithPhaseGoal("0001", 1, "Goal 1", "open").
		WithPhaseGoal("0001", 2, "Goal 2", "closed").
		WithPhaseGoal("0001", 3, "Goal 3", "closed").
		Build()

	phasePath := filepath.Join(root, ".crumbler", "phases", "0001-phase")
	closedGoals := GetClosedGoals(t, phasePath)

	if len(closedGoals) != 2 {
		t.Errorf("expected 2 closed goals, got %d: %v", len(closedGoals), closedGoals)
	}
}

func TestAssertAllGoalsClosed(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.
		WithPhase("0001", "open").
		WithPhaseGoal("0001", 1, "Goal 1", "closed").
		WithPhaseGoal("0001", 2, "Goal 2", "closed").
		Build()

	phasePath := filepath.Join(root, ".crumbler", "phases", "0001-phase")
	AssertAllGoalsClosed(t, phasePath) // Should not fail
}

func TestAssertGoalCount(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.
		WithPhase("0001", "open").
		WithPhaseGoal("0001", 1, "Goal 1", "open").
		WithPhaseGoal("0001", 2, "Goal 2", "closed").
		Build()

	phasePath := filepath.Join(root, ".crumbler", "phases", "0001-phase")
	AssertGoalCount(t, phasePath, 2) // Should not fail
}
