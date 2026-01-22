// Package testutil provides test infrastructure for the crumbler CLI tool.
package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestProjectBuilder provides a fluent API for building test project structures.
type TestProjectBuilder struct {
	t           *testing.T
	projectRoot string
	phases      map[string]*phaseConfig
	roadmap     string
	cleaned     bool
}

// phaseConfig holds configuration for a phase in the test project.
type phaseConfig struct {
	status  string
	goals   map[int]*goalConfig
	sprints map[string]*sprintConfig
}

// sprintConfig holds configuration for a sprint in the test project.
type sprintConfig struct {
	status  string
	goals   map[int]*goalConfig
	tickets map[string]*ticketConfig
	prd     string
	erd     string
}

// ticketConfig holds configuration for a ticket in the test project.
type ticketConfig struct {
	status      string
	goals       map[int]*goalConfig
	description string
}

// goalConfig holds configuration for a goal.
type goalConfig struct {
	name   string
	status string
}

// NewTestProject creates a new TestProjectBuilder for building test project structures.
// It creates a unique test directory under .test/ with format: .test/test-{timestamp}-{random}/
func NewTestProject(t *testing.T) *TestProjectBuilder {
	t.Helper()

	// Generate unique test directory name
	timestamp := time.Now().UnixNano()
	random := GenerateRandomString(6)
	testDirName := fmt.Sprintf("test-%d-%s", timestamp, random)

	// Create test directory path relative to project root
	// Find project root by looking for go.mod
	projectRoot := findProjectRoot(t)
	testRoot := filepath.Join(projectRoot, ".test", testDirName)

	// Ensure .test directory exists
	testBaseDir := filepath.Join(projectRoot, ".test")
	if err := os.MkdirAll(testBaseDir, 0755); err != nil {
		t.Fatalf("failed to create .test directory: %v", err)
	}

	// Create the unique test directory
	if err := os.MkdirAll(testRoot, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	builder := &TestProjectBuilder{
		t:           t,
		projectRoot: testRoot,
		phases:      make(map[string]*phaseConfig),
	}

	// Register cleanup
	t.Cleanup(func() {
		if !builder.cleaned {
			builder.cleanup()
		}
	})

	return builder
}

// findProjectRoot finds the project root by looking for go.mod
func findProjectRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find project root (no go.mod found)")
		}
		dir = parent
	}
}

// cleanup removes the test directory and the parent .test/ directory if empty
func (b *TestProjectBuilder) cleanup() {
	if b.projectRoot != "" {
		os.RemoveAll(b.projectRoot)
		b.cleaned = true

		// Try to remove the parent .test directory if it's empty
		testBaseDir := filepath.Dir(b.projectRoot)
		if filepath.Base(testBaseDir) == ".test" {
			// Only remove if directory is empty (will fail silently if not empty)
			os.Remove(testBaseDir)
		}
	}
}

// WithPhase adds a phase to the test project with the specified status.
// phaseID should be in format "0001" (4-digit zero-padded).
// status should be "open" or "closed".
func (b *TestProjectBuilder) WithPhase(phaseID string, status string) *TestProjectBuilder {
	b.t.Helper()

	if _, exists := b.phases[phaseID]; !exists {
		b.phases[phaseID] = &phaseConfig{
			status:  status,
			goals:   make(map[int]*goalConfig),
			sprints: make(map[string]*sprintConfig),
		}
	} else {
		b.phases[phaseID].status = status
	}

	return b
}

// WithSprint adds a sprint to a phase with the specified status.
// phaseID and sprintID should be in format "0001" (4-digit zero-padded).
// status should be "open" or "closed".
func (b *TestProjectBuilder) WithSprint(phaseID, sprintID string, status string) *TestProjectBuilder {
	b.t.Helper()

	// Ensure phase exists
	if _, exists := b.phases[phaseID]; !exists {
		b.phases[phaseID] = &phaseConfig{
			status:  "open",
			goals:   make(map[int]*goalConfig),
			sprints: make(map[string]*sprintConfig),
		}
	}

	if _, exists := b.phases[phaseID].sprints[sprintID]; !exists {
		b.phases[phaseID].sprints[sprintID] = &sprintConfig{
			status:  status,
			goals:   make(map[int]*goalConfig),
			tickets: make(map[string]*ticketConfig),
		}
	} else {
		b.phases[phaseID].sprints[sprintID].status = status
	}

	return b
}

// WithTicket adds a ticket to a sprint with the specified status.
// phaseID, sprintID, and ticketID should be in format "0001" (4-digit zero-padded).
// status should be "open" or "done".
func (b *TestProjectBuilder) WithTicket(phaseID, sprintID, ticketID string, status string) *TestProjectBuilder {
	b.t.Helper()

	// Ensure phase and sprint exist
	b.WithSprint(phaseID, sprintID, "open")

	sprint := b.phases[phaseID].sprints[sprintID]
	if _, exists := sprint.tickets[ticketID]; !exists {
		sprint.tickets[ticketID] = &ticketConfig{
			status: status,
			goals:  make(map[int]*goalConfig),
		}
	} else {
		sprint.tickets[ticketID].status = status
	}

	return b
}

// WithRoadmap sets the roadmap content for the test project.
func (b *TestProjectBuilder) WithRoadmap(content string) *TestProjectBuilder {
	b.t.Helper()
	b.roadmap = content
	return b
}

// WithPhaseGoal adds a goal to a phase with the specified name and status.
// goalIndex is the 1-based index for the goal (will be zero-padded to 4 digits).
// status should be "open" or "closed".
func (b *TestProjectBuilder) WithPhaseGoal(phaseID string, goalIndex int, goalName string, status string) *TestProjectBuilder {
	b.t.Helper()

	// Ensure phase exists
	if _, exists := b.phases[phaseID]; !exists {
		b.phases[phaseID] = &phaseConfig{
			status:  "open",
			goals:   make(map[int]*goalConfig),
			sprints: make(map[string]*sprintConfig),
		}
	}

	b.phases[phaseID].goals[goalIndex] = &goalConfig{
		name:   goalName,
		status: status,
	}

	return b
}

// WithSprintGoal adds a goal to a sprint with the specified name and status.
// goalIndex is the 1-based index for the goal (will be zero-padded to 4 digits).
// status should be "open" or "closed".
func (b *TestProjectBuilder) WithSprintGoal(phaseID, sprintID string, goalIndex int, goalName string, status string) *TestProjectBuilder {
	b.t.Helper()

	// Ensure phase and sprint exist
	b.WithSprint(phaseID, sprintID, "open")

	b.phases[phaseID].sprints[sprintID].goals[goalIndex] = &goalConfig{
		name:   goalName,
		status: status,
	}

	return b
}

// WithTicketGoal adds a goal to a ticket with the specified name and status.
// goalIndex is the 1-based index for the goal (will be zero-padded to 4 digits).
// status should be "open" or "closed".
func (b *TestProjectBuilder) WithTicketGoal(phaseID, sprintID, ticketID string, goalIndex int, goalName string, status string) *TestProjectBuilder {
	b.t.Helper()

	// Ensure phase, sprint, and ticket exist
	b.WithTicket(phaseID, sprintID, ticketID, "open")

	b.phases[phaseID].sprints[sprintID].tickets[ticketID].goals[goalIndex] = &goalConfig{
		name:   goalName,
		status: status,
	}

	return b
}

// WithPRD sets the PRD.md content for a sprint.
func (b *TestProjectBuilder) WithPRD(phaseID, sprintID string, content string) *TestProjectBuilder {
	b.t.Helper()

	// Ensure phase and sprint exist
	b.WithSprint(phaseID, sprintID, "open")

	b.phases[phaseID].sprints[sprintID].prd = content

	return b
}

// WithERD sets the ERD.md content for a sprint.
func (b *TestProjectBuilder) WithERD(phaseID, sprintID string, content string) *TestProjectBuilder {
	b.t.Helper()

	// Ensure phase and sprint exist
	b.WithSprint(phaseID, sprintID, "open")

	b.phases[phaseID].sprints[sprintID].erd = content

	return b
}

// WithTicketDescription sets the README.md content for a ticket.
func (b *TestProjectBuilder) WithTicketDescription(phaseID, sprintID, ticketID string, content string) *TestProjectBuilder {
	b.t.Helper()

	// Ensure phase, sprint, and ticket exist
	b.WithTicket(phaseID, sprintID, ticketID, "open")

	b.phases[phaseID].sprints[sprintID].tickets[ticketID].description = content

	return b
}

// Build creates the test project file structure and returns the project root path.
func (b *TestProjectBuilder) Build() string {
	b.t.Helper()

	// Create .crumbler directory structure
	crumblerDir := filepath.Join(b.projectRoot, ".crumbler")
	if err := os.MkdirAll(crumblerDir, 0755); err != nil {
		b.t.Fatalf("failed to create .crumbler directory: %v", err)
	}

	// Create README.md with default content
	readmePath := filepath.Join(crumblerDir, "README.md")
	if err := os.WriteFile(readmePath, []byte(GenerateRealisticMarkdown("README")), 0644); err != nil {
		b.t.Fatalf("failed to create README.md: %v", err)
	}

	// Create roadmap.md
	roadmapPath := filepath.Join(crumblerDir, "roadmap.md")
	roadmapContent := b.roadmap
	if roadmapContent == "" {
		roadmapContent = GenerateRealisticMarkdown("roadmap")
	}
	if err := os.WriteFile(roadmapPath, []byte(roadmapContent), 0644); err != nil {
		b.t.Fatalf("failed to create roadmap.md: %v", err)
	}

	// Create phases directory
	phasesDir := filepath.Join(crumblerDir, "phases")
	if err := os.MkdirAll(phasesDir, 0755); err != nil {
		b.t.Fatalf("failed to create phases directory: %v", err)
	}

	// Create roadmaps directory
	roadmapsDir := filepath.Join(crumblerDir, "roadmaps")
	if err := os.MkdirAll(roadmapsDir, 0755); err != nil {
		b.t.Fatalf("failed to create roadmaps directory: %v", err)
	}

	// Build each phase
	for phaseID, phaseCfg := range b.phases {
		b.buildPhase(phasesDir, phaseID, phaseCfg)
	}

	return b.projectRoot
}

// buildPhase creates a phase directory structure.
func (b *TestProjectBuilder) buildPhase(phasesDir, phaseID string, cfg *phaseConfig) {
	b.t.Helper()

	phaseDir := filepath.Join(phasesDir, fmt.Sprintf("%s-phase", phaseID))
	if err := os.MkdirAll(phaseDir, 0755); err != nil {
		b.t.Fatalf("failed to create phase directory: %v", err)
	}

	// Create status file
	statusFile := filepath.Join(phaseDir, cfg.status)
	if err := os.WriteFile(statusFile, []byte{}, 0644); err != nil {
		b.t.Fatalf("failed to create status file: %v", err)
	}

	// Create README.md
	readmePath := filepath.Join(phaseDir, "README.md")
	if err := os.WriteFile(readmePath, []byte(GenerateRealisticMarkdown("phase")), 0644); err != nil {
		b.t.Fatalf("failed to create phase README.md: %v", err)
	}

	// Create goals directory
	goalsDir := filepath.Join(phaseDir, "goals")
	if err := os.MkdirAll(goalsDir, 0755); err != nil {
		b.t.Fatalf("failed to create goals directory: %v", err)
	}

	// Create phase goals
	for goalIndex, goalCfg := range cfg.goals {
		b.buildGoal(goalsDir, goalIndex, goalCfg)
	}

	// Create sprints directory
	sprintsDir := filepath.Join(phaseDir, "sprints")
	if err := os.MkdirAll(sprintsDir, 0755); err != nil {
		b.t.Fatalf("failed to create sprints directory: %v", err)
	}

	// Build each sprint
	for sprintID, sprintCfg := range cfg.sprints {
		b.buildSprint(sprintsDir, sprintID, sprintCfg)
	}
}

// buildSprint creates a sprint directory structure.
func (b *TestProjectBuilder) buildSprint(sprintsDir, sprintID string, cfg *sprintConfig) {
	b.t.Helper()

	sprintDir := filepath.Join(sprintsDir, fmt.Sprintf("%s-sprint", sprintID))
	if err := os.MkdirAll(sprintDir, 0755); err != nil {
		b.t.Fatalf("failed to create sprint directory: %v", err)
	}

	// Create status file
	statusFile := filepath.Join(sprintDir, cfg.status)
	if err := os.WriteFile(statusFile, []byte{}, 0644); err != nil {
		b.t.Fatalf("failed to create status file: %v", err)
	}

	// Create README.md
	readmePath := filepath.Join(sprintDir, "README.md")
	if err := os.WriteFile(readmePath, []byte(GenerateRealisticMarkdown("sprint")), 0644); err != nil {
		b.t.Fatalf("failed to create sprint README.md: %v", err)
	}

	// Create PRD.md
	prdPath := filepath.Join(sprintDir, "PRD.md")
	prdContent := cfg.prd
	if prdContent == "" {
		prdContent = GenerateRealisticMarkdown("PRD")
	}
	if err := os.WriteFile(prdPath, []byte(prdContent), 0644); err != nil {
		b.t.Fatalf("failed to create PRD.md: %v", err)
	}

	// Create ERD.md
	erdPath := filepath.Join(sprintDir, "ERD.md")
	erdContent := cfg.erd
	if erdContent == "" {
		erdContent = GenerateRealisticMarkdown("ERD")
	}
	if err := os.WriteFile(erdPath, []byte(erdContent), 0644); err != nil {
		b.t.Fatalf("failed to create ERD.md: %v", err)
	}

	// Create goals directory
	goalsDir := filepath.Join(sprintDir, "goals")
	if err := os.MkdirAll(goalsDir, 0755); err != nil {
		b.t.Fatalf("failed to create goals directory: %v", err)
	}

	// Create sprint goals
	for goalIndex, goalCfg := range cfg.goals {
		b.buildGoal(goalsDir, goalIndex, goalCfg)
	}

	// Create tickets directory
	ticketsDir := filepath.Join(sprintDir, "tickets")
	if err := os.MkdirAll(ticketsDir, 0755); err != nil {
		b.t.Fatalf("failed to create tickets directory: %v", err)
	}

	// Build each ticket
	for ticketID, ticketCfg := range cfg.tickets {
		b.buildTicket(ticketsDir, ticketID, ticketCfg)
	}
}

// buildTicket creates a ticket directory structure.
func (b *TestProjectBuilder) buildTicket(ticketsDir, ticketID string, cfg *ticketConfig) {
	b.t.Helper()

	ticketDir := filepath.Join(ticketsDir, fmt.Sprintf("%s-ticket", ticketID))
	if err := os.MkdirAll(ticketDir, 0755); err != nil {
		b.t.Fatalf("failed to create ticket directory: %v", err)
	}

	// Create status file
	statusFile := filepath.Join(ticketDir, cfg.status)
	if err := os.WriteFile(statusFile, []byte{}, 0644); err != nil {
		b.t.Fatalf("failed to create status file: %v", err)
	}

	// Create README.md
	readmePath := filepath.Join(ticketDir, "README.md")
	descContent := cfg.description
	if descContent == "" {
		descContent = GenerateRealisticMarkdown("ticket")
	}
	if err := os.WriteFile(readmePath, []byte(descContent), 0644); err != nil {
		b.t.Fatalf("failed to create ticket README.md: %v", err)
	}

	// Create goals directory
	goalsDir := filepath.Join(ticketDir, "goals")
	if err := os.MkdirAll(goalsDir, 0755); err != nil {
		b.t.Fatalf("failed to create goals directory: %v", err)
	}

	// Create ticket goals
	for goalIndex, goalCfg := range cfg.goals {
		b.buildGoal(goalsDir, goalIndex, goalCfg)
	}
}

// buildGoal creates a goal directory structure.
func (b *TestProjectBuilder) buildGoal(goalsDir string, goalIndex int, cfg *goalConfig) {
	b.t.Helper()

	goalDir := filepath.Join(goalsDir, fmt.Sprintf("%04d-goal", goalIndex))
	if err := os.MkdirAll(goalDir, 0755); err != nil {
		b.t.Fatalf("failed to create goal directory: %v", err)
	}

	// Create name file
	namePath := filepath.Join(goalDir, "name")
	goalName := cfg.name
	if goalName == "" {
		goalName = GenerateGoalName()
	}
	if err := os.WriteFile(namePath, []byte(goalName), 0644); err != nil {
		b.t.Fatalf("failed to create goal name file: %v", err)
	}

	// Create status file
	statusFile := filepath.Join(goalDir, cfg.status)
	if err := os.WriteFile(statusFile, []byte{}, 0644); err != nil {
		b.t.Fatalf("failed to create goal status file: %v", err)
	}
}

// ProjectRoot returns the project root path without building.
// Useful for tests that need the path before building.
func (b *TestProjectBuilder) ProjectRoot() string {
	return b.projectRoot
}

// CrumblerDir returns the path to the .crumbler directory.
func (b *TestProjectBuilder) CrumblerDir() string {
	return filepath.Join(b.projectRoot, ".crumbler")
}

// PhasePath returns the path to a specific phase directory.
func (b *TestProjectBuilder) PhasePath(phaseID string) string {
	return filepath.Join(b.projectRoot, ".crumbler", "phases", fmt.Sprintf("%s-phase", phaseID))
}

// SprintPath returns the path to a specific sprint directory.
func (b *TestProjectBuilder) SprintPath(phaseID, sprintID string) string {
	return filepath.Join(b.PhasePath(phaseID), "sprints", fmt.Sprintf("%s-sprint", sprintID))
}

// TicketPath returns the path to a specific ticket directory.
func (b *TestProjectBuilder) TicketPath(phaseID, sprintID, ticketID string) string {
	return filepath.Join(b.SprintPath(phaseID, sprintID), "tickets", fmt.Sprintf("%s-ticket", ticketID))
}

// GoalPath returns the path to a specific goal directory within a parent.
func (b *TestProjectBuilder) GoalPath(parentPath string, goalIndex int) string {
	return filepath.Join(parentPath, "goals", fmt.Sprintf("%04d-goal", goalIndex))
}
