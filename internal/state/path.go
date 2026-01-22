// Package state provides state management functionality for crumbler.
package state

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/waynenilsen/crumbler/internal/models"
)

// FindProjectRoot locates the .crumbler/ directory by walking up from the
// current working directory. It returns the relative path from pwd to the
// project root (the directory containing .crumbler/).
// Returns an error if no .crumbler/ directory is found.
func FindProjectRoot() (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	current := pwd
	for {
		crumblerPath := filepath.Join(current, models.CrumblerDir)
		info, err := os.Stat(crumblerPath)
		if err == nil && info.IsDir() {
			// Found .crumbler/ directory
			rel, err := filepath.Rel(pwd, current)
			if err != nil {
				return "", fmt.Errorf("failed to compute relative path: %w", err)
			}
			return rel, nil
		}

		// Move to parent directory
		parent := filepath.Dir(current)
		if parent == current {
			// Reached filesystem root without finding .crumbler/
			return "", fmt.Errorf("not a crumbler project: %s directory not found (searched from %s)", models.CrumblerDir, pwd)
		}
		current = parent
	}
}

// CrumblerDirPath returns the path to the .crumbler directory.
func CrumblerDirPath(projectRoot string) string {
	return filepath.Join(projectRoot, models.CrumblerDir)
}

// PhasesDirPath returns the path to the .crumbler/phases directory.
func PhasesDirPath(projectRoot string) string {
	return filepath.Join(projectRoot, models.CrumblerDir, models.PhasesDir)
}

// PhasesDir returns the path to the .crumbler/phases directory.
// This is the primary function for getting the phases directory path.
func PhasesDir(projectRoot string) string {
	return PhasesDirPath(projectRoot)
}

// PhasePath returns the path to a specific phase directory.
// phaseID should be in the format "XXXX-phase" (e.g., "0001-phase").
func PhasePath(projectRoot, phaseID string) string {
	return filepath.Join(projectRoot, models.CrumblerDir, models.PhasesDir, phaseID)
}

// PhaseGoalsDir returns the path to the goals directory within a phase.
func PhaseGoalsDir(projectRoot, phaseID string) string {
	return filepath.Join(projectRoot, models.CrumblerDir, models.PhasesDir, phaseID, models.GoalsDir)
}

// PhaseGoalPath returns the path to a specific goal within a phase.
// goalID should be in the format "XXXX-goal" (e.g., "0001-goal").
func PhaseGoalPath(projectRoot, phaseID, goalID string) string {
	return filepath.Join(projectRoot, models.CrumblerDir, models.PhasesDir, phaseID, models.GoalsDir, goalID)
}

// SprintsDirPath returns the path to the sprints directory within a phase.
func SprintsDirPath(projectRoot, phaseID string) string {
	return filepath.Join(projectRoot, models.CrumblerDir, models.PhasesDir, phaseID, models.SprintsDir)
}

// SprintPath returns the path to a specific sprint directory within a phase.
// sprintID should be in the format "XXXX-sprint" (e.g., "0001-sprint").
func SprintPath(projectRoot, phaseID, sprintID string) string {
	return filepath.Join(projectRoot, models.CrumblerDir, models.PhasesDir, phaseID, models.SprintsDir, sprintID)
}

// SprintGoalsDir returns the path to the goals directory within a sprint.
func SprintGoalsDir(projectRoot, phaseID, sprintID string) string {
	return filepath.Join(projectRoot, models.CrumblerDir, models.PhasesDir, phaseID, models.SprintsDir, sprintID, models.GoalsDir)
}

// SprintGoalPath returns the path to a specific goal within a sprint.
func SprintGoalPath(projectRoot, phaseID, sprintID, goalID string) string {
	return filepath.Join(projectRoot, models.CrumblerDir, models.PhasesDir, phaseID, models.SprintsDir, sprintID, models.GoalsDir, goalID)
}

// TicketsDirPath returns the path to the tickets directory within a sprint.
func TicketsDirPath(projectRoot, phaseID, sprintID string) string {
	return filepath.Join(projectRoot, models.CrumblerDir, models.PhasesDir, phaseID, models.SprintsDir, sprintID, models.TicketsDir)
}

// TicketPath returns the path to a specific ticket directory within a sprint.
// ticketID should be in the format "XXXX-ticket" (e.g., "0001-ticket").
func TicketPath(projectRoot, phaseID, sprintID, ticketID string) string {
	return filepath.Join(projectRoot, models.CrumblerDir, models.PhasesDir, phaseID, models.SprintsDir, sprintID, models.TicketsDir, ticketID)
}

// TicketGoalsDir returns the path to the goals directory within a ticket.
func TicketGoalsDir(projectRoot, phaseID, sprintID, ticketID string) string {
	return filepath.Join(projectRoot, models.CrumblerDir, models.PhasesDir, phaseID, models.SprintsDir, sprintID, models.TicketsDir, ticketID, models.GoalsDir)
}

// TicketGoalPath returns the path to a specific goal within a ticket.
func TicketGoalPath(projectRoot, phaseID, sprintID, ticketID, goalID string) string {
	return filepath.Join(projectRoot, models.CrumblerDir, models.PhasesDir, phaseID, models.SprintsDir, sprintID, models.TicketsDir, ticketID, models.GoalsDir, goalID)
}

// RoadmapPath returns the path to the roadmap.md file.
func RoadmapPath(projectRoot string) string {
	return filepath.Join(projectRoot, models.CrumblerDir, models.RoadmapFile)
}

// FormatPhaseID returns a phase ID in the format "XXXX-phase".
func FormatPhaseID(index int) string {
	return FormatIndex(index) + models.PhaseSuffix
}

// FormatSprintID returns a sprint ID in the format "XXXX-sprint".
func FormatSprintID(index int) string {
	return FormatIndex(index) + models.SprintSuffix
}

// FormatTicketID returns a ticket ID in the format "XXXX-ticket".
func FormatTicketID(index int) string {
	return FormatIndex(index) + models.TicketSuffix
}

// FormatGoalID returns a goal ID in the format "XXXX-goal".
func FormatGoalID(index int) string {
	return FormatIndex(index) + models.GoalSuffix
}
