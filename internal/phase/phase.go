// Package phase provides Phase management functions for the crumbler CLI tool.
package phase

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/waynenilsen/crumbler/internal/models"
	"github.com/waynenilsen/crumbler/internal/state"
)

// phasePattern matches phase directory names like "0001-phase"
var phasePattern = regexp.MustCompile(`^(\d{4})-phase$`)

// sprintPattern matches sprint directory names like "0001-sprint"
var sprintPattern = regexp.MustCompile(`^(\d{4})-sprint$`)

// GetOpenPhase scans .crumbler/phases/ for a directory with an open file (no closed file).
// Returns the open phase or nil if no open phase exists.
func GetOpenPhase(projectRoot string) (*models.Phase, error) {
	phasesPath := state.PhasesDir(projectRoot)

	dirs, err := state.ListDirs(phasesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list phases in %s: %w",
			relPhasesPath(), err)
	}

	for _, dir := range dirs {
		if !phasePattern.MatchString(dir) {
			continue
		}

		phasePath := filepath.Join(phasesPath, dir)
		relPath := relPhasePath(dir)

		// Validate state
		if err := state.ValidateStatus(phasePath); err != nil {
			return nil, wrapStateError(err, relPath)
		}

		isOpen, err := state.IsOpen(phasePath)
		if err != nil {
			return nil, err
		}
		if isOpen {
			phase, err := loadPhase(phasePath, dir)
			if err != nil {
				return nil, err
			}
			return phase, nil
		}
	}

	return nil, nil
}

// GetNextPhaseIndex finds the next phase number (returns int to be formatted as 4-digit).
func GetNextPhaseIndex(projectRoot string) (int, error) {
	phasesPath := state.PhasesDir(projectRoot)

	dirs, err := state.ListDirs(phasesPath)
	if err != nil {
		return 1, nil // If no phases directory exists, start at 1
	}

	maxIndex := 0
	for _, dir := range dirs {
		matches := phasePattern.FindStringSubmatch(dir)
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

// CreatePhase creates a new phase directory structure.
// Returns the phase path (absolute).
func CreatePhase(projectRoot string, index int) (string, error) {
	phasesPath := state.PhasesDir(projectRoot)
	phaseID := state.FormatPhaseID(index)
	phasePath := filepath.Join(phasesPath, phaseID)
	relPath := relPhasePath(phaseID)

	// Check if phase already exists
	exists, err := state.DirExists(phasePath)
	if err != nil {
		return "", fmt.Errorf("failed to check if phase exists at %s: %w", relPath, err)
	}
	if exists {
		return "", fmt.Errorf("phase already exists at %s", relPath)
	}

	// Create phase directory
	if err := state.CreateDir(phasePath); err != nil {
		return "", fmt.Errorf("failed to create phase directory at %s: %w", relPath, err)
	}

	// Create empty README.md
	readmePath := filepath.Join(phasePath, state.ReadmeFile)
	if err := state.TouchFile(readmePath); err != nil {
		os.RemoveAll(phasePath)
		return "", fmt.Errorf("failed to create README.md at %s: %w",
			filepath.Join(relPath, state.ReadmeFile), err)
	}

	// Create goals/ subdirectory
	goalsPath := filepath.Join(phasePath, state.GoalsDirName)
	if err := state.CreateDir(goalsPath); err != nil {
		os.RemoveAll(phasePath)
		return "", fmt.Errorf("failed to create goals directory at %s: %w",
			filepath.Join(relPath, state.GoalsDirName), err)
	}

	// Create sprints/ subdirectory
	sprintsPath := filepath.Join(phasePath, state.SprintsDirName)
	if err := state.CreateDir(sprintsPath); err != nil {
		os.RemoveAll(phasePath)
		return "", fmt.Errorf("failed to create sprints directory at %s: %w",
			filepath.Join(relPath, state.SprintsDirName), err)
	}

	// Touch open file
	if err := state.SetOpen(phasePath); err != nil {
		os.RemoveAll(phasePath)
		return "", fmt.Errorf("failed to set phase as open at %s: %w",
			filepath.Join(relPath, state.OpenFileName), err)
	}

	return phasePath, nil
}

// GetPhase loads a phase by ID.
func GetPhase(projectRoot, phaseID string) (*models.Phase, error) {
	phasePath := state.PhasePath(projectRoot, phaseID)
	relPath := relPhasePath(phaseID)

	exists, err := state.DirExists(phasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to check if phase exists at %s: %w", relPath, err)
	}
	if !exists {
		return nil, fmt.Errorf("phase not found at %s", relPath)
	}

	return loadPhase(phasePath, phaseID)
}

// ListPhases lists all phases sorted by ID.
func ListPhases(projectRoot string) ([]models.Phase, error) {
	phasesPath := state.PhasesDir(projectRoot)

	dirs, err := state.ListDirs(phasesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list phases in %s: %w",
			relPhasesPath(), err)
	}

	var phases []models.Phase
	for _, dir := range dirs {
		if !phasePattern.MatchString(dir) {
			continue
		}

		phasePath := filepath.Join(phasesPath, dir)
		phase, err := loadPhase(phasePath, dir)
		if err != nil {
			return nil, err
		}
		phases = append(phases, *phase)
	}

	// Sort by index
	sort.Slice(phases, func(i, j int) bool {
		return phases[i].Index < phases[j].Index
	})

	return phases, nil
}

// GetPhaseGoals scans the goals/ directory and returns a list of goals.
func GetPhaseGoals(phasePath string) ([]models.Goal, error) {
	goalsPath := filepath.Join(phasePath, state.GoalsDirName)
	return state.ListGoals(goalsPath)
}

// CreatePhaseGoal creates a goal in the phase's goals/ directory.
// Returns the goal path (absolute).
func CreatePhaseGoal(phasePath string, index int, goalName string) (string, error) {
	goalsPath := filepath.Join(phasePath, state.GoalsDirName)
	goalID := state.FormatGoalID(index)
	goalPath := filepath.Join(goalsPath, goalID)
	relPath := filepath.Join(getRelPath(phasePath), state.GoalsDirName, goalID)

	// Check if goal already exists
	exists, err := state.DirExists(goalPath)
	if err != nil {
		return "", fmt.Errorf("failed to check if goal exists at %s: %w", relPath, err)
	}
	if exists {
		return "", fmt.Errorf("goal already exists at %s", relPath)
	}

	// Create goal directory
	if err := state.CreateDir(goalPath); err != nil {
		return "", fmt.Errorf("failed to create goal directory at %s: %w", relPath, err)
	}

	// Create name file with goalName
	if err := state.WriteGoalName(goalPath, goalName); err != nil {
		os.RemoveAll(goalPath)
		return "", fmt.Errorf("failed to create name file at %s: %w",
			filepath.Join(relPath, state.NameFileName), err)
	}

	// Touch open file
	if err := state.SetOpen(goalPath); err != nil {
		os.RemoveAll(goalPath)
		return "", fmt.Errorf("failed to set goal as open at %s: %w",
			filepath.Join(relPath, state.OpenFileName), err)
	}

	return goalPath, nil
}

// ClosePhaseGoal validates state and closes a phase goal.
func ClosePhaseGoal(phasePath string, goalID string) error {
	goalPath := filepath.Join(phasePath, state.GoalsDirName, goalID)
	relPath := filepath.Join(getRelPath(phasePath), state.GoalsDirName, goalID)

	// Check if goal exists
	exists, err := state.DirExists(goalPath)
	if err != nil {
		return fmt.Errorf("failed to check if goal exists at %s: %w", relPath, err)
	}
	if !exists {
		return fmt.Errorf("goal not found at %s", relPath)
	}

	// Validate state and close
	if err := state.CloseGoal(goalPath); err != nil {
		return wrapStateError(err, relPath)
	}

	return nil
}

// ArePhaseGoalsMet checks if all phase goals have closed file AND all sprints have closed file.
// Returns false if no goals or sprints exist yet.
func ArePhaseGoalsMet(phasePath string) (bool, error) {
	relPath := getRelPath(phasePath)

	// Check phase goals
	goalsPath := filepath.Join(phasePath, state.GoalsDirName)
	goalsExist, err := state.GoalsExist(goalsPath)
	if err != nil {
		return false, err
	}

	if !goalsExist {
		return false, nil // No goals exist yet
	}

	allGoalsClosed, err := state.AreAllGoalsClosed(goalsPath)
	if err != nil {
		return false, err
	}

	if !allGoalsClosed {
		return false, nil
	}

	// Check sprints
	sprintsPath := filepath.Join(phasePath, state.SprintsDirName)
	dirs, err := state.ListDirs(sprintsPath)
	if err != nil {
		return false, fmt.Errorf("failed to list sprints in %s: %w",
			filepath.Join(relPath, state.SprintsDirName), err)
	}

	sprintCount := 0
	for _, dir := range dirs {
		if !sprintPattern.MatchString(dir) {
			continue
		}
		sprintCount++

		sprintPath := filepath.Join(sprintsPath, dir)
		sprintRelPath := filepath.Join(relPath, state.SprintsDirName, dir)

		// Validate sprint state
		if err := state.ValidateStatus(sprintPath); err != nil {
			return false, wrapStateError(err, sprintRelPath)
		}

		isClosed, err := state.IsClosed(sprintPath)
		if err != nil {
			return false, err
		}
		if !isClosed {
			return false, nil
		}
	}

	if sprintCount == 0 {
		return false, nil // No sprints exist yet
	}

	return true, nil
}

// ClosePhase validates state and closes the phase.
// Returns error if sprints still open or phase goals still open.
func ClosePhase(phasePath string) error {
	relPath := getRelPath(phasePath)

	// Validate phase state
	if err := state.ValidateStatus(phasePath); err != nil {
		return wrapStateError(err, relPath)
	}

	// Check if phase is open
	isOpen, err := state.IsOpen(phasePath)
	if err != nil {
		return err
	}
	if !isOpen {
		return fmt.Errorf("cannot close phase that is not open at %s", relPath)
	}

	// Check for open sprints
	openSprints, err := getOpenSprintPaths(phasePath)
	if err != nil {
		return err
	}
	if len(openSprints) > 0 {
		return models.NewHierarchyConstraintError(
			fmt.Sprintf("cannot close phase with open sprints: %s", strings.Join(openSprints, ", ")),
			openSprints...,
		)
	}

	// Check for open phase goals
	openGoals, err := getOpenGoalPaths(phasePath)
	if err != nil {
		return err
	}
	if len(openGoals) > 0 {
		return models.NewHierarchyConstraintError(
			fmt.Sprintf("cannot close phase with open goals: %s", strings.Join(openGoals, ", ")),
			openGoals...,
		)
	}

	// Close the phase
	if err := state.SetClosed(phasePath); err != nil {
		return fmt.Errorf("failed to close phase at %s: %w", relPath, err)
	}

	return nil
}

// ValidatePhaseState checks for invalid state in a phase.
func ValidatePhaseState(phasePath string) error {
	relPath := getRelPath(phasePath)

	// Validate phase state
	if err := state.ValidateStatus(phasePath); err != nil {
		return wrapStateError(err, relPath)
	}

	// Validate all goals
	goals, err := GetPhaseGoals(phasePath)
	if err != nil {
		return err
	}

	for _, goal := range goals {
		goalRelPath := filepath.Join(relPath, state.GoalsDirName, goal.ID)
		if err := state.ValidateStatus(goal.Path); err != nil {
			return wrapStateError(err, goalRelPath)
		}
	}

	// Validate all sprints
	sprintsPath := filepath.Join(phasePath, state.SprintsDirName)
	dirs, err := state.ListDirs(sprintsPath)
	if err != nil {
		return fmt.Errorf("failed to list sprints in %s: %w",
			filepath.Join(relPath, state.SprintsDirName), err)
	}

	for _, dir := range dirs {
		if !sprintPattern.MatchString(dir) {
			continue
		}

		sprintPath := filepath.Join(sprintsPath, dir)
		sprintRelPath := filepath.Join(relPath, state.SprintsDirName, dir)

		if err := state.ValidateStatus(sprintPath); err != nil {
			return wrapStateError(err, sprintRelPath)
		}
	}

	return nil
}

// loadPhase loads a phase from its directory.
func loadPhase(phasePath, phaseID string) (*models.Phase, error) {
	relPath := getRelPath(phasePath)

	// Validate state
	if err := state.ValidateStatus(phasePath); err != nil {
		return nil, wrapStateError(err, relPath)
	}

	// Determine status
	phaseStatus, err := state.GetStatus(phasePath)
	if err != nil {
		return nil, wrapStateError(err, relPath)
	}

	// Extract index from ID
	matches := phasePattern.FindStringSubmatch(phaseID)
	if matches == nil {
		return nil, fmt.Errorf("invalid phase ID format: %s", phaseID)
	}
	index, _ := strconv.Atoi(matches[1])

	// Load goals
	goals, err := GetPhaseGoals(phasePath)
	if err != nil {
		return nil, err
	}

	return &models.Phase{
		ID:     phaseID,
		Path:   phasePath,
		Status: phaseStatus,
		Index:  index,
		Goals:  goals,
	}, nil
}

// getRelPath converts an absolute path to a relative path from .crumbler/.
func getRelPath(absPath string) string {
	// Find .crumbler in the path and return everything from there
	idx := strings.Index(absPath, state.CrumblerDirName)
	if idx == -1 {
		return absPath
	}
	return absPath[idx:]
}

// relPhasesPath returns the relative path to the phases directory.
func relPhasesPath() string {
	return filepath.Join(state.CrumblerDirName, state.PhasesDirName)
}

// relPhasePath returns the relative path to a specific phase.
func relPhasePath(phaseID string) string {
	return filepath.Join(state.CrumblerDirName, state.PhasesDirName, phaseID)
}

// wrapStateError wraps a state error with a relative path.
func wrapStateError(err error, relPath string) error {
	if statusErr, ok := err.(*state.StatusError); ok {
		return models.NewInvalidStateError(
			fmt.Sprintf("%s at %s", statusErr.Message, relPath),
			statusErr.ConflictingFiles...,
		)
	}
	return fmt.Errorf("%s at %s: %w", err.Error(), relPath, err)
}

// getOpenSprintPaths returns relative paths of all open sprints in a phase.
func getOpenSprintPaths(phasePath string) ([]string, error) {
	sprintsPath := filepath.Join(phasePath, state.SprintsDirName)
	relPath := getRelPath(phasePath)

	dirs, err := state.ListDirs(sprintsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list sprints in %s: %w",
			filepath.Join(relPath, state.SprintsDirName), err)
	}

	var openPaths []string
	for _, dir := range dirs {
		if !sprintPattern.MatchString(dir) {
			continue
		}

		sprintPath := filepath.Join(sprintsPath, dir)
		sprintRelPath := filepath.Join(relPath, state.SprintsDirName, dir)

		// Validate state
		if err := state.ValidateStatus(sprintPath); err != nil {
			return nil, wrapStateError(err, sprintRelPath)
		}

		isOpen, err := state.IsOpen(sprintPath)
		if err != nil {
			return nil, err
		}
		if isOpen {
			openPaths = append(openPaths, sprintRelPath)
		}
	}

	return openPaths, nil
}

// getOpenGoalPaths returns relative paths of all open goals in a phase.
func getOpenGoalPaths(phasePath string) ([]string, error) {
	goalsPath := filepath.Join(phasePath, state.GoalsDirName)
	relPath := getRelPath(phasePath)

	goals, err := state.ListGoals(goalsPath)
	if err != nil {
		return nil, err
	}

	var openPaths []string
	for _, goal := range goals {
		if goal.Status == models.StatusOpen {
			goalRelPath := filepath.Join(relPath, state.GoalsDirName, goal.ID)
			openPaths = append(openPaths, goalRelPath)
		}
	}

	return openPaths, nil
}
