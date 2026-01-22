package state

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/waynenilsen/crumbler/internal/models"
)

// goalIDPattern matches goal directory names in the format "XXXX-goal".
var goalIDPattern = regexp.MustCompile(`^(\d{4})-goal$`)

// CreateGoal creates a new goal directory at the specified goals directory.
// It creates the goal directory structure: goals/XXXX-goal/ with an empty name file
// and touches the open status file.
// Returns the path to the created goal directory.
func CreateGoal(goalsDir string, index int) (string, error) {
	// Format the goal ID
	goalID := FormatGoalID(index)
	goalPath := filepath.Join(goalsDir, goalID)

	// Create the goal directory
	if err := os.MkdirAll(goalPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create goal directory at %s: %w", goalPath, err)
	}

	// Create empty name file
	namePath := filepath.Join(goalPath, NameFile)
	if err := TouchFile(namePath); err != nil {
		// Clean up on failure
		os.RemoveAll(goalPath)
		return "", fmt.Errorf("failed to create name file at %s: %w", namePath, err)
	}

	// Touch open file to mark goal as open
	if err := SetOpen(goalPath); err != nil {
		// Clean up on failure
		os.RemoveAll(goalPath)
		return "", fmt.Errorf("failed to set goal status to open at %s: %w", goalPath, err)
	}

	return goalPath, nil
}

// ReadGoalName reads the goal name from the name file at the given goal path.
func ReadGoalName(goalPath string) (string, error) {
	name, err := ReadName(goalPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("goal name file not found at %s", filepath.Join(goalPath, NameFile))
		}
		return "", fmt.Errorf("failed to read goal name from %s: %w", goalPath, err)
	}
	return strings.TrimSpace(name), nil
}

// WriteGoalName writes the goal name to the name file at the given goal path.
// It uses atomic write (temp file + rename) for safety.
func WriteGoalName(goalPath, name string) error {
	namePath := filepath.Join(goalPath, NameFile)
	dir := filepath.Dir(namePath)

	// Create a temporary file in the same directory
	tempFile, err := os.CreateTemp(dir, ".tmp-name-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file for goal name: %w", err)
	}
	tempPath := tempFile.Name()

	// Write the name to the temp file
	if _, err := tempFile.WriteString(name); err != nil {
		tempFile.Close()
		os.Remove(tempPath)
		return fmt.Errorf("failed to write goal name to temp file: %w", err)
	}

	// Close the temp file
	if err := tempFile.Close(); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Rename temp file to target path (atomic on most filesystems)
	if err := os.Rename(tempPath, namePath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename temp file to %s: %w", namePath, err)
	}

	return nil
}

// ListGoals scans the goals directory and returns a sorted list of goals.
// Goals are sorted by their index (ascending).
func ListGoals(goalsDir string) ([]models.Goal, error) {
	// Check if the goals directory exists
	exists, err := DirExists(goalsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to check goals directory %s: %w", goalsDir, err)
	}
	if !exists {
		// No goals directory means no goals
		return nil, nil
	}

	// Read the directory entries
	entries, err := os.ReadDir(goalsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read goals directory %s: %w", goalsDir, err)
	}

	var goals []models.Goal
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if this is a valid goal directory
		matches := goalIDPattern.FindStringSubmatch(entry.Name())
		if matches == nil {
			continue
		}

		// Parse the index
		index, err := strconv.Atoi(matches[1])
		if err != nil {
			continue // Invalid index, skip
		}

		goalPath := filepath.Join(goalsDir, entry.Name())

		// Read goal name
		name, err := ReadGoalName(goalPath)
		if err != nil {
			// Name file might be empty or missing, that's OK
			name = ""
		}

		// Get status
		status, err := GetStatus(goalPath)
		if err != nil {
			// If we can't determine status, mark as unknown
			status = models.StatusUnknown
		}

		goals = append(goals, models.Goal{
			ID:     entry.Name(),
			Path:   goalPath,
			Name:   name,
			Status: status,
			Index:  index,
		})
	}

	// Sort goals by index
	sort.Slice(goals, func(i, j int) bool {
		return goals[i].Index < goals[j].Index
	})

	return goals, nil
}

// GetNextGoalIndex finds the next available goal index by scanning the goals directory.
// Returns 1 if no goals exist, otherwise returns max existing index + 1.
func GetNextGoalIndex(goalsDir string) (int, error) {
	goals, err := ListGoals(goalsDir)
	if err != nil {
		return 0, err
	}

	if len(goals) == 0 {
		return 1, nil
	}

	// Find the maximum index
	maxIndex := 0
	for _, goal := range goals {
		if goal.Index > maxIndex {
			maxIndex = goal.Index
		}
	}

	return maxIndex + 1, nil
}

// AreAllGoalsClosed checks if all goals in the goals directory have the closed status.
// Returns true if there are no goals or if all goals are closed.
// Returns false if any goal is open or has an invalid state.
func AreAllGoalsClosed(goalsDir string) (bool, error) {
	goals, err := ListGoals(goalsDir)
	if err != nil {
		return false, err
	}

	// If no goals exist, consider all goals "closed" (vacuously true)
	if len(goals) == 0 {
		return true, nil
	}

	for _, goal := range goals {
		if goal.Status != models.StatusClosed {
			return false, nil
		}
	}

	return true, nil
}

// CloseGoal closes a goal by transitioning its status from open to closed.
// It validates the current state before making the transition.
func CloseGoal(goalPath string) error {
	// Validate that the goal exists
	exists, err := DirExists(goalPath)
	if err != nil {
		return fmt.Errorf("failed to check goal at %s: %w", goalPath, err)
	}
	if !exists {
		return fmt.Errorf("goal not found at %s", goalPath)
	}

	// Use SetClosedValidated which validates the state transition
	return SetClosedValidated(goalPath)
}

// GoalExists checks if a goal directory exists at the given path.
func GoalExists(goalPath string) bool {
	exists, _ := DirExists(goalPath)
	return exists
}

// GoalsExist checks if any goals exist in the goals directory.
func GoalsExist(goalsDir string) (bool, error) {
	goals, err := ListGoals(goalsDir)
	if err != nil {
		return false, err
	}
	return len(goals) > 0, nil
}

// ParseGoalIndex extracts the numeric index from a goal ID (e.g., "0001-goal" -> 1).
func ParseGoalIndex(goalID string) (int, error) {
	matches := goalIDPattern.FindStringSubmatch(goalID)
	if matches == nil {
		return 0, fmt.Errorf("invalid goal ID format: %s (expected XXXX-goal)", goalID)
	}
	return strconv.Atoi(matches[1])
}
