// Package query provides state query functions for the AI agent loop.
// These functions are read-only and validate state, returning errors with file paths on invalid states.
package query

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	crumblerDir = ".crumbler"
	phasesDir   = "phases"
	sprintsDir  = "sprints"
	ticketsDir  = "tickets"
	goalsDir    = "goals"
	openFile    = "open"
	closedFile  = "closed"
	doneFile    = "done"
)

// OpenPhaseExists returns true if any phase has an `open` file (no `closed` file).
// It validates that no phase has both `open` and `closed` files (invalid state).
func OpenPhaseExists(projectRoot string) (bool, error) {
	phasesPath := filepath.Join(projectRoot, crumblerDir, phasesDir)

	entries, err := os.ReadDir(phasesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to read phases directory %s: %w", phasesPath, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		phasePath := filepath.Join(phasesPath, entry.Name())
		openPath := filepath.Join(phasePath, openFile)
		closedPath := filepath.Join(phasePath, closedFile)

		hasOpen := fileExists(openPath)
		hasClosed := fileExists(closedPath)

		// Validate state: cannot have both open and closed
		if hasOpen && hasClosed {
			return false, fmt.Errorf("invalid state: both 'open' and 'closed' exist in %s",
				relPath(projectRoot, phasePath))
		}

		if hasOpen && !hasClosed {
			return true, nil
		}
	}

	return false, nil
}

// RoadmapComplete returns true if all phases in the roadmap have a `closed` file.
// It validates state and returns errors with file paths on invalid states.
func RoadmapComplete(projectRoot string) (bool, error) {
	phasesPath := filepath.Join(projectRoot, crumblerDir, phasesDir)

	entries, err := os.ReadDir(phasesPath)
	if err != nil {
		if os.IsNotExist(err) {
			// No phases directory means no phases created yet, roadmap not complete
			return false, nil
		}
		return false, fmt.Errorf("failed to read phases directory %s: %w", phasesPath, err)
	}

	phaseCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		phaseCount++
		phasePath := filepath.Join(phasesPath, entry.Name())
		openPath := filepath.Join(phasePath, openFile)
		closedPath := filepath.Join(phasePath, closedFile)

		hasOpen := fileExists(openPath)
		hasClosed := fileExists(closedPath)

		// Validate state: cannot have both open and closed
		if hasOpen && hasClosed {
			return false, fmt.Errorf("invalid state: both 'open' and 'closed' exist in %s",
				relPath(projectRoot, phasePath))
		}

		// If any phase is not closed, roadmap is not complete
		if !hasClosed {
			return false, nil
		}
	}

	// If no phases exist, roadmap is not complete
	if phaseCount == 0 {
		return false, nil
	}

	return true, nil
}

// OpenSprintExists returns true if any sprint in the phase has an `open` file (no `closed` file).
// It validates that no sprint has both `open` and `closed` files (invalid state).
func OpenSprintExists(phasePath string) (bool, error) {
	sprintsPath := filepath.Join(phasePath, sprintsDir)

	entries, err := os.ReadDir(sprintsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to read sprints directory %s: %w", sprintsPath, err)
	}

	projectRoot := getProjectRoot(phasePath)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		sprintPath := filepath.Join(sprintsPath, entry.Name())
		openPath := filepath.Join(sprintPath, openFile)
		closedPath := filepath.Join(sprintPath, closedFile)

		hasOpen := fileExists(openPath)
		hasClosed := fileExists(closedPath)

		// Validate state: cannot have both open and closed
		if hasOpen && hasClosed {
			return false, fmt.Errorf("invalid state: both 'open' and 'closed' exist in %s",
				relPath(projectRoot, sprintPath))
		}

		if hasOpen && !hasClosed {
			return true, nil
		}
	}

	return false, nil
}

// PhaseGoalsMet returns true if all phase goals have a `closed` file AND all sprints have a `closed` file.
// Returns false if no goals or sprints exist yet.
func PhaseGoalsMet(phasePath string) (bool, error) {
	projectRoot := getProjectRoot(phasePath)

	// Check if goals exist and all are closed
	goalsPath := filepath.Join(phasePath, goalsDir)
	goalsExist, allGoalsClosed, err := checkGoalsStatus(goalsPath, projectRoot)
	if err != nil {
		return false, err
	}

	// Check if sprints exist and all are closed
	sprintsPath := filepath.Join(phasePath, sprintsDir)
	sprintsExist, allSprintsClosed, err := checkSprintsStatus(sprintsPath, projectRoot)
	if err != nil {
		return false, err
	}

	// Return false if no goals or sprints exist yet
	if !goalsExist || !sprintsExist {
		return false, nil
	}

	return allGoalsClosed && allSprintsClosed, nil
}

// OpenTicketsExist returns true if any ticket in the sprint has an `open` file (no `done` file).
// It validates state and returns errors with file paths on invalid states.
func OpenTicketsExist(sprintPath string) (bool, error) {
	ticketsPath := filepath.Join(sprintPath, ticketsDir)

	entries, err := os.ReadDir(ticketsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to read tickets directory %s: %w", ticketsPath, err)
	}

	projectRoot := getProjectRoot(sprintPath)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		ticketPath := filepath.Join(ticketsPath, entry.Name())
		openPath := filepath.Join(ticketPath, openFile)
		donePath := filepath.Join(ticketPath, doneFile)

		hasOpen := fileExists(openPath)
		hasDone := fileExists(donePath)

		// Validate state: cannot have both open and done
		if hasOpen && hasDone {
			return false, fmt.Errorf("invalid state: both 'open' and 'done' exist in %s",
				relPath(projectRoot, ticketPath))
		}

		if hasOpen && !hasDone {
			return true, nil
		}
	}

	return false, nil
}

// SprintGoalsMet returns true if all sprint goals have a `closed` file AND all tickets have a `done` file.
// Returns false if no goals or tickets exist yet.
func SprintGoalsMet(sprintPath string) (bool, error) {
	projectRoot := getProjectRoot(sprintPath)

	// Check if goals exist and all are closed
	goalsPath := filepath.Join(sprintPath, goalsDir)
	goalsExist, allGoalsClosed, err := checkGoalsStatus(goalsPath, projectRoot)
	if err != nil {
		return false, err
	}

	// Check if tickets exist and all are done
	ticketsPath := filepath.Join(sprintPath, ticketsDir)
	ticketsExist, allTicketsDone, err := checkTicketsStatus(ticketsPath, projectRoot)
	if err != nil {
		return false, err
	}

	// Return false if no goals or tickets exist yet
	if !goalsExist || !ticketsExist {
		return false, nil
	}

	return allGoalsClosed && allTicketsDone, nil
}

// TicketComplete returns true if the ticket has a `done` file AND all ticket goals have a `closed` file.
// It validates that the ticket does not have both `open` and `done` files (invalid state).
func TicketComplete(ticketPath string) (bool, error) {
	projectRoot := getProjectRoot(ticketPath)

	openPath := filepath.Join(ticketPath, openFile)
	donePath := filepath.Join(ticketPath, doneFile)

	hasOpen := fileExists(openPath)
	hasDone := fileExists(donePath)

	// Validate state: cannot have both open and done
	if hasOpen && hasDone {
		return false, fmt.Errorf("invalid state: both 'open' and 'done' exist in %s",
			relPath(projectRoot, ticketPath))
	}

	// If ticket doesn't have done file, it's not complete
	if !hasDone {
		return false, nil
	}

	// Check if all ticket goals are closed
	goalsPath := filepath.Join(ticketPath, goalsDir)
	goalsExist, allGoalsClosed, err := checkGoalsStatus(goalsPath, projectRoot)
	if err != nil {
		return false, err
	}

	// If goals exist, they must all be closed
	if goalsExist && !allGoalsClosed {
		return false, nil
	}

	return true, nil
}

// PhaseGoalsExist returns true if any goals exist in the phase's goals/ directory.
func PhaseGoalsExist(phasePath string) (bool, error) {
	return goalsExist(filepath.Join(phasePath, goalsDir))
}

// SprintGoalsExist returns true if any goals exist in the sprint's goals/ directory.
func SprintGoalsExist(sprintPath string) (bool, error) {
	return goalsExist(filepath.Join(sprintPath, goalsDir))
}

// TicketGoalsExist returns true if any goals exist in the ticket's goals/ directory.
func TicketGoalsExist(ticketPath string) (bool, error) {
	return goalsExist(filepath.Join(ticketPath, goalsDir))
}

// SprintsExist returns true if any sprints exist in the phase.
func SprintsExist(phasePath string) (bool, error) {
	sprintsPath := filepath.Join(phasePath, sprintsDir)

	entries, err := os.ReadDir(sprintsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to read sprints directory %s: %w", sprintsPath, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			return true, nil
		}
	}

	return false, nil
}

// TicketsExist returns true if any tickets exist in the sprint.
func TicketsExist(sprintPath string) (bool, error) {
	ticketsPath := filepath.Join(sprintPath, ticketsDir)

	entries, err := os.ReadDir(ticketsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to read tickets directory %s: %w", ticketsPath, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			return true, nil
		}
	}

	return false, nil
}

// goalsExist checks if any goal directories exist in the given goals path.
func goalsExist(goalsPath string) (bool, error) {
	entries, err := os.ReadDir(goalsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to read goals directory %s: %w", goalsPath, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			return true, nil
		}
	}

	return false, nil
}

// checkGoalsStatus checks if goals exist and if all are closed.
// Returns (goalsExist, allClosed, error).
func checkGoalsStatus(goalsPath string, projectRoot string) (bool, bool, error) {
	entries, err := os.ReadDir(goalsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, false, nil
		}
		return false, false, fmt.Errorf("failed to read goals directory %s: %w", goalsPath, err)
	}

	goalCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		goalCount++
		goalPath := filepath.Join(goalsPath, entry.Name())
		openPath := filepath.Join(goalPath, openFile)
		closedPath := filepath.Join(goalPath, closedFile)

		hasOpen := fileExists(openPath)
		hasClosed := fileExists(closedPath)

		// Validate state: cannot have both open and closed
		if hasOpen && hasClosed {
			return false, false, fmt.Errorf("invalid state: both 'open' and 'closed' exist in %s",
				relPath(projectRoot, goalPath))
		}

		// If any goal is not closed, not all are closed
		if !hasClosed {
			return true, false, nil
		}
	}

	if goalCount == 0 {
		return false, false, nil
	}

	return true, true, nil
}

// checkSprintsStatus checks if sprints exist and if all are closed.
// Returns (sprintsExist, allClosed, error).
func checkSprintsStatus(sprintsPath string, projectRoot string) (bool, bool, error) {
	entries, err := os.ReadDir(sprintsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, false, nil
		}
		return false, false, fmt.Errorf("failed to read sprints directory %s: %w", sprintsPath, err)
	}

	sprintCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		sprintCount++
		sprintPath := filepath.Join(sprintsPath, entry.Name())
		openPath := filepath.Join(sprintPath, openFile)
		closedPath := filepath.Join(sprintPath, closedFile)

		hasOpen := fileExists(openPath)
		hasClosed := fileExists(closedPath)

		// Validate state: cannot have both open and closed
		if hasOpen && hasClosed {
			return false, false, fmt.Errorf("invalid state: both 'open' and 'closed' exist in %s",
				relPath(projectRoot, sprintPath))
		}

		// If any sprint is not closed, not all are closed
		if !hasClosed {
			return true, false, nil
		}
	}

	if sprintCount == 0 {
		return false, false, nil
	}

	return true, true, nil
}

// checkTicketsStatus checks if tickets exist and if all are done.
// Returns (ticketsExist, allDone, error).
func checkTicketsStatus(ticketsPath string, projectRoot string) (bool, bool, error) {
	entries, err := os.ReadDir(ticketsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, false, nil
		}
		return false, false, fmt.Errorf("failed to read tickets directory %s: %w", ticketsPath, err)
	}

	ticketCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		ticketCount++
		ticketPath := filepath.Join(ticketsPath, entry.Name())
		openPath := filepath.Join(ticketPath, openFile)
		donePath := filepath.Join(ticketPath, doneFile)

		hasOpen := fileExists(openPath)
		hasDone := fileExists(donePath)

		// Validate state: cannot have both open and done
		if hasOpen && hasDone {
			return false, false, fmt.Errorf("invalid state: both 'open' and 'done' exist in %s",
				relPath(projectRoot, ticketPath))
		}

		// If any ticket is not done, not all are done
		if !hasDone {
			return true, false, nil
		}
	}

	if ticketCount == 0 {
		return false, false, nil
	}

	return true, true, nil
}

// fileExists checks if a file exists at the given path.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// relPath returns the relative path from projectRoot to path.
// If it fails, it returns the absolute path.
func relPath(projectRoot, path string) string {
	rel, err := filepath.Rel(projectRoot, path)
	if err != nil {
		return path
	}
	return rel
}

// getProjectRoot extracts the project root from a path within the .crumbler directory.
// It finds the parent of .crumbler in the path.
func getProjectRoot(path string) string {
	// Walk up the path to find the project root (parent of .crumbler)
	dir := path
	for {
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			return path
		}
		if filepath.Base(dir) == crumblerDir {
			return parent
		}
		dir = parent
	}
}
