// Package sprint provides Sprint management functions for crumbler.
package sprint

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/waynenilsen/crumbler/internal/models"
	"github.com/waynenilsen/crumbler/internal/state"
)

var (
	// sprintDirPattern matches sprint directory names like "0001-sprint"
	sprintDirPattern = regexp.MustCompile(`^(\d{4})-sprint$`)
	// ticketDirPattern matches ticket directory names like "0001-ticket"
	ticketDirPattern = regexp.MustCompile(`^(\d{4})-ticket$`)
)

// GetOpenSprint scans the sprints/ subdirectory for a directory with an open file (no closed file).
// Returns nil if no open sprint is found.
func GetOpenSprint(phasePath string) (*models.Sprint, error) {
	sprintsPath := filepath.Join(phasePath, state.SprintsDirName)

	dirs, err := state.ListDirs(sprintsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list sprints in %s: %w", getRelPath(sprintsPath), err)
	}

	for _, dir := range dirs {
		if !sprintDirPattern.MatchString(dir) {
			continue
		}

		sprintPath := filepath.Join(sprintsPath, dir)

		// Validate state
		if err := ValidateSprintState(sprintPath); err != nil {
			return nil, err
		}

		isOpen, err := state.IsOpen(sprintPath)
		if err != nil {
			return nil, fmt.Errorf("failed to check open state at %s: %w", getRelPath(sprintPath), err)
		}
		if isOpen {
			sprint, err := loadSprint(sprintPath, dir)
			if err != nil {
				return nil, err
			}
			return sprint, nil
		}
	}

	return nil, nil
}

// GetNextSprintIndex finds the next sprint number.
func GetNextSprintIndex(phasePath string) (int, error) {
	sprintsPath := filepath.Join(phasePath, state.SprintsDirName)

	dirs, err := state.ListDirs(sprintsPath)
	if err != nil {
		return 1, nil // If directory doesn't exist or is empty, start at 1
	}

	maxIndex := 0
	for _, dir := range dirs {
		matches := sprintDirPattern.FindStringSubmatch(dir)
		if matches == nil {
			continue
		}

		index, err := strconv.Atoi(matches[1])
		if err != nil {
			continue
		}

		if index > maxIndex {
			maxIndex = index
		}
	}

	return maxIndex + 1, nil
}

// CreateSprint creates a new sprint directory structure.
// Returns the path to the created sprint.
func CreateSprint(phasePath string, index int) (string, error) {
	sprintID := state.FormatSprintID(index)
	sprintPath := filepath.Join(phasePath, state.SprintsDirName, sprintID)

	// Create sprints directory if it doesn't exist
	sprintsPath := filepath.Join(phasePath, state.SprintsDirName)
	if err := state.CreateDir(sprintsPath); err != nil {
		return "", fmt.Errorf("failed to create sprints directory at %s: %w", getRelPath(sprintsPath), err)
	}

	// Create sprint directory
	if err := state.CreateDir(sprintPath); err != nil {
		return "", fmt.Errorf("failed to create sprint directory at %s: %w", getRelPath(sprintPath), err)
	}

	// Create empty README.md
	readmePath := filepath.Join(sprintPath, state.ReadmeFile)
	if err := state.TouchFile(readmePath); err != nil {
		return "", fmt.Errorf("failed to create README.md at %s: %w", getRelPath(readmePath), err)
	}

	// Create empty PRD.md
	prdPath := filepath.Join(sprintPath, state.PRDFile)
	if err := state.TouchFile(prdPath); err != nil {
		return "", fmt.Errorf("failed to create PRD.md at %s: %w", getRelPath(prdPath), err)
	}

	// Create empty ERD.md
	erdPath := filepath.Join(sprintPath, state.ERDFile)
	if err := state.TouchFile(erdPath); err != nil {
		return "", fmt.Errorf("failed to create ERD.md at %s: %w", getRelPath(erdPath), err)
	}

	// Create goals/ subdirectory
	goalsPath := filepath.Join(sprintPath, state.GoalsDirName)
	if err := state.CreateDir(goalsPath); err != nil {
		return "", fmt.Errorf("failed to create goals directory at %s: %w", getRelPath(goalsPath), err)
	}

	// Create tickets/ subdirectory
	ticketsPath := filepath.Join(sprintPath, state.TicketsDirName)
	if err := state.CreateDir(ticketsPath); err != nil {
		return "", fmt.Errorf("failed to create tickets directory at %s: %w", getRelPath(ticketsPath), err)
	}

	// Touch open file
	if err := state.SetOpen(sprintPath); err != nil {
		return "", fmt.Errorf("failed to create open file at %s: %w", getRelPath(sprintPath), err)
	}

	return sprintPath, nil
}

// GetSprint loads a sprint by ID.
func GetSprint(phasePath, sprintID string) (*models.Sprint, error) {
	sprintPath := filepath.Join(phasePath, state.SprintsDirName, sprintID)

	exists, err := state.DirExists(sprintPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check sprint directory at %s: %w", getRelPath(sprintPath), err)
	}
	if !exists {
		return nil, fmt.Errorf("sprint not found at %s", getRelPath(sprintPath))
	}

	if err := ValidateSprintState(sprintPath); err != nil {
		return nil, err
	}

	return loadSprint(sprintPath, sprintID)
}

// ListSprints lists all sprints sorted by ID.
func ListSprints(phasePath string) ([]models.Sprint, error) {
	sprintsPath := filepath.Join(phasePath, state.SprintsDirName)

	dirs, err := state.ListDirs(sprintsPath)
	if err != nil {
		return nil, nil // Return empty list if directory doesn't exist
	}

	var sprintDirs []string
	for _, dir := range dirs {
		if sprintDirPattern.MatchString(dir) {
			sprintDirs = append(sprintDirs, dir)
		}
	}

	// Sort by ID
	sort.Strings(sprintDirs)

	var sprints []models.Sprint
	for _, dir := range sprintDirs {
		sprintPath := filepath.Join(sprintsPath, dir)

		if err := ValidateSprintState(sprintPath); err != nil {
			return nil, err
		}

		sprint, err := loadSprint(sprintPath, dir)
		if err != nil {
			return nil, err
		}
		sprints = append(sprints, *sprint)
	}

	return sprints, nil
}

// GetSprintGoals scans the goals/ directory and returns a list of goals.
func GetSprintGoals(sprintPath string) ([]models.Goal, error) {
	goalsPath := filepath.Join(sprintPath, state.GoalsDirName)
	return state.ListGoals(goalsPath)
}

// CreateSprintGoal creates a goal in the sprint's goals/ directory.
// Returns the path to the created goal.
func CreateSprintGoal(sprintPath string, index int, goalName string) (string, error) {
	// Create goals directory if it doesn't exist
	goalsPath := filepath.Join(sprintPath, state.GoalsDirName)
	if err := state.CreateDir(goalsPath); err != nil {
		return "", fmt.Errorf("failed to create goals directory at %s: %w", getRelPath(goalsPath), err)
	}

	// Create goal directory using state package
	goalPath, err := state.CreateGoal(goalsPath, index)
	if err != nil {
		return "", fmt.Errorf("failed to create goal at %s: %w", getRelPath(goalsPath), err)
	}

	// Write the goal name
	if err := state.WriteGoalName(goalPath, goalName); err != nil {
		return "", fmt.Errorf("failed to write goal name at %s: %w", getRelPath(goalPath), err)
	}

	return goalPath, nil
}

// CloseSprintGoal closes a goal by deleting the open file and touching the closed file.
func CloseSprintGoal(sprintPath string, goalID string) error {
	goalPath := filepath.Join(sprintPath, state.GoalsDirName, goalID)

	// Check goal exists
	exists, err := state.DirExists(goalPath)
	if err != nil {
		return fmt.Errorf("failed to check goal directory at %s: %w", getRelPath(goalPath), err)
	}
	if !exists {
		return fmt.Errorf("goal not found at %s", getRelPath(goalPath))
	}

	// Use state.CloseGoal which handles validation and state transition
	if err := state.CloseGoal(goalPath); err != nil {
		return fmt.Errorf("failed to close goal at %s: %w", getRelPath(goalPath), err)
	}

	return nil
}

// AreSprintGoalsMet checks if all sprint goals have closed file AND all tickets have done file.
// Returns false if no goals or tickets exist yet.
func AreSprintGoalsMet(sprintPath string) (bool, error) {
	// Check goals
	goalsPath := filepath.Join(sprintPath, state.GoalsDirName)
	goals, err := state.ListGoals(goalsPath)
	if err != nil {
		return false, err
	}

	// If no goals exist, return false
	if len(goals) == 0 {
		return false, nil
	}

	// Check all goals are closed
	for _, goal := range goals {
		if goal.Status != models.StatusClosed {
			return false, nil
		}
	}

	// Check tickets
	ticketsPath := filepath.Join(sprintPath, state.TicketsDirName)
	ticketDirs, err := state.ListDirs(ticketsPath)
	if err != nil {
		return false, nil // If no tickets directory, return false
	}

	// Filter to valid ticket directories
	var validTicketDirs []string
	for _, dir := range ticketDirs {
		if ticketDirPattern.MatchString(dir) {
			validTicketDirs = append(validTicketDirs, dir)
		}
	}

	// If no tickets exist, return false
	if len(validTicketDirs) == 0 {
		return false, nil
	}

	// Check all tickets are done
	for _, dir := range validTicketDirs {
		ticketPath := filepath.Join(ticketsPath, dir)

		// Validate ticket state
		if err := state.ValidateStatus(ticketPath); err != nil {
			return false, err
		}

		isDone, err := state.IsDone(ticketPath)
		if err != nil {
			return false, fmt.Errorf("failed to check done state at %s: %w", getRelPath(ticketPath), err)
		}
		if !isDone {
			return false, nil
		}
	}

	return true, nil
}

// CloseSprint closes a sprint by deleting the open file and touching the closed file.
// Returns an error if tickets are still open or sprint goals are still open.
func CloseSprint(sprintPath string) error {
	// Validate current state
	if err := ValidateSprintState(sprintPath); err != nil {
		return err
	}

	// Check that sprint is currently open
	isOpen, err := state.IsOpen(sprintPath)
	if err != nil {
		return fmt.Errorf("failed to check open state at %s: %w", getRelPath(sprintPath), err)
	}
	if !isOpen {
		isClosed, err := state.IsClosed(sprintPath)
		if err != nil {
			return fmt.Errorf("failed to check closed state at %s: %w", getRelPath(sprintPath), err)
		}
		if isClosed {
			return fmt.Errorf("sprint already closed at %s", getRelPath(sprintPath))
		}
		return fmt.Errorf("sprint has no state at %s", getRelPath(sprintPath))
	}

	// Check for open tickets
	openTicketPaths, err := getOpenTicketPaths(sprintPath)
	if err != nil {
		return err
	}
	if len(openTicketPaths) > 0 {
		return fmt.Errorf("cannot close sprint: tickets still open: %s", strings.Join(openTicketPaths, ", "))
	}

	// Check for open goals
	openGoalPaths, err := getOpenGoalPaths(sprintPath)
	if err != nil {
		return err
	}
	if len(openGoalPaths) > 0 {
		return fmt.Errorf("cannot close sprint: goals still open: %s", strings.Join(openGoalPaths, ", "))
	}

	// Set closed state
	if err := state.SetClosed(sprintPath); err != nil {
		return fmt.Errorf("failed to close sprint at %s: %w", getRelPath(sprintPath), err)
	}

	return nil
}

// ValidateSprintState checks for invalid sprint state (both open and closed exist).
func ValidateSprintState(sprintPath string) error {
	if err := state.ValidateStatus(sprintPath); err != nil {
		return fmt.Errorf("invalid sprint state at %s: %w", getRelPath(sprintPath), err)
	}
	return nil
}

// loadSprint loads a sprint from the given path.
func loadSprint(sprintPath, id string) (*models.Sprint, error) {
	// Get status
	status, err := state.GetStatus(sprintPath)
	if err != nil {
		// If no status exists, treat as unknown but don't fail
		status = models.StatusUnknown
	}

	// Parse index from ID
	index := 0
	matches := sprintDirPattern.FindStringSubmatch(id)
	if matches != nil {
		index, _ = strconv.Atoi(matches[1])
	}

	// Get goals
	goalsPath := filepath.Join(sprintPath, state.GoalsDirName)
	goals, err := state.ListGoals(goalsPath)
	if err != nil {
		return nil, err
	}

	return &models.Sprint{
		ID:      id,
		Path:    sprintPath,
		Goals:   goals,
		PRDPath: filepath.Join(sprintPath, state.PRDFile),
		ERDPath: filepath.Join(sprintPath, state.ERDFile),
		Status:  status,
		Index:   index,
	}, nil
}

// getOpenTicketPaths returns the relative paths of all open tickets.
func getOpenTicketPaths(sprintPath string) ([]string, error) {
	ticketsPath := filepath.Join(sprintPath, state.TicketsDirName)

	dirs, err := state.ListDirs(ticketsPath)
	if err != nil {
		return nil, nil // If no tickets directory, no open tickets
	}

	var openPaths []string
	for _, dir := range dirs {
		if !ticketDirPattern.MatchString(dir) {
			continue
		}

		ticketPath := filepath.Join(ticketsPath, dir)

		// Validate ticket state
		if err := state.ValidateStatus(ticketPath); err != nil {
			return nil, err
		}

		isOpen, err := state.IsOpen(ticketPath)
		if err != nil {
			return nil, err
		}
		if isOpen {
			openPaths = append(openPaths, getRelPath(ticketPath))
		}
	}

	return openPaths, nil
}

// getOpenGoalPaths returns the relative paths of all open goals.
func getOpenGoalPaths(sprintPath string) ([]string, error) {
	goalsPath := filepath.Join(sprintPath, state.GoalsDirName)

	goals, err := state.ListGoals(goalsPath)
	if err != nil {
		return nil, nil // If no goals directory, no open goals
	}

	var openPaths []string
	for _, goal := range goals {
		if goal.Status == models.StatusOpen {
			openPaths = append(openPaths, getRelPath(goal.Path))
		}
	}

	return openPaths, nil
}

// getRelPath converts an absolute path to a relative path from the project root.
// This is a simplified implementation that looks for .crumbler in the path.
func getRelPath(absPath string) string {
	// Find .crumbler in the path and return from there
	idx := strings.Index(absPath, ".crumbler")
	if idx >= 0 {
		return absPath[idx:]
	}
	// If .crumbler not found, return the original path
	return absPath
}
