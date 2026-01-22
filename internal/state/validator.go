// Package state provides state validation functionality for crumbler.
package state

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/waynenilsen/crumbler/internal/models"
)

// StateValidator provides comprehensive state machine validation.
type StateValidator struct {
	// projectRoot is the absolute path to the project root directory.
	projectRoot string
}

// NewStateValidator creates a new StateValidator for the given project root.
func NewStateValidator(projectRoot string) *StateValidator {
	return &StateValidator{
		projectRoot: projectRoot,
	}
}

// ValidateStateMachine performs comprehensive validation on the entire state machine.
// This should be called on startup to ensure the state is valid before any operations.
func (v *StateValidator) ValidateStateMachine(projectRoot string) error {
	crumblerDir := filepath.Join(projectRoot, models.CrumblerDir)

	// Check if .crumbler directory exists
	exists, err := DirExists(crumblerDir)
	if err != nil {
		return fmt.Errorf("failed to check crumbler directory: %w", err)
	}
	if !exists {
		return &models.StateError{
			Type:         models.ErrorTypeOrphanedState,
			Message:      fmt.Sprintf("not a crumbler project: %s directory not found", models.CrumblerDir),
			Paths:        []string{models.CrumblerDir},
			SuggestedFix: "Run 'crumbler init' to initialize a new project",
		}
	}

	// Validate all phases
	phasesDir := filepath.Join(crumblerDir, models.PhasesDir)
	exists, err = DirExists(phasesDir)
	if err != nil {
		return fmt.Errorf("failed to check phases directory: %w", err)
	}
	if !exists {
		// No phases directory is valid (empty project)
		return nil
	}

	phaseDirs, err := ListDirs(phasesDir)
	if err != nil {
		return fmt.Errorf("failed to list phases: %w", err)
	}

	for _, phaseDir := range phaseDirs {
		if !strings.HasSuffix(phaseDir, "-phase") {
			continue // Skip non-phase directories
		}

		phasePath := filepath.Join(phasesDir, phaseDir)
		if err := v.ValidatePhaseState(phasePath); err != nil {
			return err
		}

		// Validate phase goals
		phaseGoalsDir := filepath.Join(phasePath, models.GoalsDir)
		if err := v.validateGoalsDir(phaseGoalsDir, projectRoot); err != nil {
			return err
		}

		// Validate sprints in this phase
		sprintsDir := filepath.Join(phasePath, models.SprintsDir)
		exists, err := DirExists(sprintsDir)
		if err != nil {
			return fmt.Errorf("failed to check sprints directory: %w", err)
		}
		if !exists {
			continue
		}

		sprintDirs, err := ListDirs(sprintsDir)
		if err != nil {
			return fmt.Errorf("failed to list sprints: %w", err)
		}

		for _, sprintDir := range sprintDirs {
			if !strings.HasSuffix(sprintDir, "-sprint") {
				continue // Skip non-sprint directories
			}

			sprintPath := filepath.Join(sprintsDir, sprintDir)
			if err := v.ValidateSprintState(sprintPath); err != nil {
				return err
			}

			// Validate sprint goals
			sprintGoalsDir := filepath.Join(sprintPath, models.GoalsDir)
			if err := v.validateGoalsDir(sprintGoalsDir, projectRoot); err != nil {
				return err
			}

			// Validate tickets in this sprint
			ticketsDir := filepath.Join(sprintPath, models.TicketsDir)
			exists, err := DirExists(ticketsDir)
			if err != nil {
				return fmt.Errorf("failed to check tickets directory: %w", err)
			}
			if !exists {
				continue
			}

			ticketDirs, err := ListDirs(ticketsDir)
			if err != nil {
				return fmt.Errorf("failed to list tickets: %w", err)
			}

			for _, ticketDir := range ticketDirs {
				if !strings.HasSuffix(ticketDir, "-ticket") {
					continue // Skip non-ticket directories
				}

				ticketPath := filepath.Join(ticketsDir, ticketDir)
				if err := v.ValidateTicketState(ticketPath); err != nil {
					return err
				}

				// Validate ticket goals
				ticketGoalsDir := filepath.Join(ticketPath, models.GoalsDir)
				if err := v.validateGoalsDir(ticketGoalsDir, projectRoot); err != nil {
					return err
				}
			}
		}
	}

	// Validate hierarchy constraints
	if err := v.ValidateHierarchy(projectRoot); err != nil {
		return err
	}

	return nil
}

// ValidatePhaseState checks for invalid state in a phase directory.
// Invalid state is when both 'open' and 'closed' files exist.
func (v *StateValidator) ValidatePhaseState(phasePath string) error {
	relPath, err := v.getRelPath(phasePath)
	if err != nil {
		relPath = phasePath // Fallback to absolute path
	}

	openPath := filepath.Join(phasePath, OpenFile)
	closedPath := filepath.Join(phasePath, ClosedFile)

	openExists, err := fileExists(openPath)
	if err != nil {
		return fmt.Errorf("failed to check open file at %s: %w", filepath.Join(relPath, OpenFile), err)
	}

	closedExists, err := fileExists(closedPath)
	if err != nil {
		return fmt.Errorf("failed to check closed file at %s: %w", filepath.Join(relPath, ClosedFile), err)
	}

	if openExists && closedExists {
		return &models.StateError{
			Type:    models.ErrorTypeMutuallyExclusiveState,
			Message: fmt.Sprintf("invalid state: both 'open' and 'closed' exist in %s", relPath),
			Paths: []string{
				filepath.Join(relPath, OpenFile),
				filepath.Join(relPath, ClosedFile),
			},
			SuggestedFix: fmt.Sprintf("Remove one of the state files: rm %s or rm %s", filepath.Join(relPath, OpenFile), filepath.Join(relPath, ClosedFile)),
		}
	}

	return nil
}

// ValidateSprintState checks for invalid state in a sprint directory.
// Invalid state is when both 'open' and 'closed' files exist.
func (v *StateValidator) ValidateSprintState(sprintPath string) error {
	relPath, err := v.getRelPath(sprintPath)
	if err != nil {
		relPath = sprintPath // Fallback to absolute path
	}

	openPath := filepath.Join(sprintPath, OpenFile)
	closedPath := filepath.Join(sprintPath, ClosedFile)

	openExists, err := fileExists(openPath)
	if err != nil {
		return fmt.Errorf("failed to check open file at %s: %w", filepath.Join(relPath, OpenFile), err)
	}

	closedExists, err := fileExists(closedPath)
	if err != nil {
		return fmt.Errorf("failed to check closed file at %s: %w", filepath.Join(relPath, ClosedFile), err)
	}

	if openExists && closedExists {
		return &models.StateError{
			Type:    models.ErrorTypeMutuallyExclusiveState,
			Message: fmt.Sprintf("invalid state: both 'open' and 'closed' exist in %s", relPath),
			Paths: []string{
				filepath.Join(relPath, OpenFile),
				filepath.Join(relPath, ClosedFile),
			},
			SuggestedFix: fmt.Sprintf("Remove one of the state files: rm %s or rm %s", filepath.Join(relPath, OpenFile), filepath.Join(relPath, ClosedFile)),
		}
	}

	return nil
}

// ValidateTicketState checks for invalid state in a ticket directory.
// Invalid state is when both 'open' and 'done' files exist.
func (v *StateValidator) ValidateTicketState(ticketPath string) error {
	relPath, err := v.getRelPath(ticketPath)
	if err != nil {
		relPath = ticketPath // Fallback to absolute path
	}

	openPath := filepath.Join(ticketPath, OpenFile)
	donePath := filepath.Join(ticketPath, DoneFile)

	openExists, err := fileExists(openPath)
	if err != nil {
		return fmt.Errorf("failed to check open file at %s: %w", filepath.Join(relPath, OpenFile), err)
	}

	doneExists, err := fileExists(donePath)
	if err != nil {
		return fmt.Errorf("failed to check done file at %s: %w", filepath.Join(relPath, DoneFile), err)
	}

	if openExists && doneExists {
		return &models.StateError{
			Type:    models.ErrorTypeMutuallyExclusiveState,
			Message: fmt.Sprintf("invalid state: both 'open' and 'done' exist in %s", relPath),
			Paths: []string{
				filepath.Join(relPath, OpenFile),
				filepath.Join(relPath, DoneFile),
			},
			SuggestedFix: fmt.Sprintf("Remove one of the state files: rm %s or rm %s", filepath.Join(relPath, OpenFile), filepath.Join(relPath, DoneFile)),
		}
	}

	return nil
}

// ValidateGoalState checks for invalid state in a goal directory.
// Invalid state is when both 'open' and 'closed' files exist.
func (v *StateValidator) ValidateGoalState(goalPath string) error {
	relPath, err := v.getRelPath(goalPath)
	if err != nil {
		relPath = goalPath // Fallback to absolute path
	}

	openPath := filepath.Join(goalPath, OpenFile)
	closedPath := filepath.Join(goalPath, ClosedFile)

	openExists, err := fileExists(openPath)
	if err != nil {
		return fmt.Errorf("failed to check open file at %s: %w", filepath.Join(relPath, OpenFile), err)
	}

	closedExists, err := fileExists(closedPath)
	if err != nil {
		return fmt.Errorf("failed to check closed file at %s: %w", filepath.Join(relPath, ClosedFile), err)
	}

	if openExists && closedExists {
		return &models.StateError{
			Type:    models.ErrorTypeMutuallyExclusiveState,
			Message: fmt.Sprintf("invalid state: both 'open' and 'closed' exist in %s", relPath),
			Paths: []string{
				filepath.Join(relPath, OpenFile),
				filepath.Join(relPath, ClosedFile),
			},
			SuggestedFix: fmt.Sprintf("Remove one of the state files: rm %s or rm %s", filepath.Join(relPath, OpenFile), filepath.Join(relPath, ClosedFile)),
		}
	}

	// Check for missing name file
	namePath := filepath.Join(goalPath, NameFile)
	nameExists, err := fileExists(namePath)
	if err != nil {
		return fmt.Errorf("failed to check name file at %s: %w", filepath.Join(relPath, NameFile), err)
	}

	if !nameExists {
		return &models.StateError{
			Type:         models.ErrorTypeMissingGoalName,
			Message:      fmt.Sprintf("goal missing name file: %s", relPath),
			Paths:        []string{filepath.Join(relPath, NameFile)},
			SuggestedFix: fmt.Sprintf("Create the name file: echo 'Goal Name' > %s", filepath.Join(relPath, NameFile)),
		}
	}

	return nil
}

// ValidateHierarchy validates hierarchy constraints across the entire project.
// This checks that:
// - Closed phases don't have open sprints or open phase goals
// - Closed sprints don't have open tickets or open sprint goals
// - Done tickets don't have open ticket goals
func (v *StateValidator) ValidateHierarchy(projectRoot string) error {
	crumblerDir := filepath.Join(projectRoot, models.CrumblerDir)
	phasesDir := filepath.Join(crumblerDir, models.PhasesDir)

	exists, err := DirExists(phasesDir)
	if err != nil {
		return fmt.Errorf("failed to check phases directory: %w", err)
	}
	if !exists {
		return nil
	}

	phaseDirs, err := ListDirs(phasesDir)
	if err != nil {
		return fmt.Errorf("failed to list phases: %w", err)
	}

	for _, phaseDir := range phaseDirs {
		if !strings.HasSuffix(phaseDir, "-phase") {
			continue
		}

		phasePath := filepath.Join(phasesDir, phaseDir)
		phaseRelPath, _ := v.getRelPath(phasePath)

		phaseClosed, err := IsClosed(phasePath)
		if err != nil {
			return fmt.Errorf("failed to check phase closed state: %w", err)
		}

		if phaseClosed {
			// Check for open phase goals
			phaseGoalsDir := filepath.Join(phasePath, models.GoalsDir)
			hasOpenGoals, openGoalPaths, err := v.hasOpenGoals(phaseGoalsDir)
			if err != nil {
				return err
			}
			if hasOpenGoals {
				return &models.StateError{
					Type:         models.ErrorTypeHierarchyConstraint,
					Message:      fmt.Sprintf("invalid hierarchy: closed phase has open goals: %s", phaseRelPath),
					Paths:        openGoalPaths,
					SuggestedFix: "Close all phase goals before closing the phase, or reopen the phase",
				}
			}

			// Check for open sprints
			sprintsDir := filepath.Join(phasePath, models.SprintsDir)
			hasOpenSprints, openSprintPaths, err := v.hasOpenSprints(sprintsDir)
			if err != nil {
				return err
			}
			if hasOpenSprints {
				return &models.StateError{
					Type:         models.ErrorTypeHierarchyConstraint,
					Message:      fmt.Sprintf("invalid hierarchy: closed phase has open sprints: %s", phaseRelPath),
					Paths:        openSprintPaths,
					SuggestedFix: "Close all sprints before closing the phase, or reopen the phase",
				}
			}
		}

		// Check sprints within this phase
		sprintsDir := filepath.Join(phasePath, models.SprintsDir)
		exists, err := DirExists(sprintsDir)
		if err != nil {
			return fmt.Errorf("failed to check sprints directory: %w", err)
		}
		if !exists {
			continue
		}

		sprintDirs, err := ListDirs(sprintsDir)
		if err != nil {
			return fmt.Errorf("failed to list sprints: %w", err)
		}

		for _, sprintDir := range sprintDirs {
			if !strings.HasSuffix(sprintDir, "-sprint") {
				continue
			}

			sprintPath := filepath.Join(sprintsDir, sprintDir)
			sprintRelPath, _ := v.getRelPath(sprintPath)

			sprintClosed, err := IsClosed(sprintPath)
			if err != nil {
				return fmt.Errorf("failed to check sprint closed state: %w", err)
			}

			if sprintClosed {
				// Check for open sprint goals
				sprintGoalsDir := filepath.Join(sprintPath, models.GoalsDir)
				hasOpenGoals, openGoalPaths, err := v.hasOpenGoals(sprintGoalsDir)
				if err != nil {
					return err
				}
				if hasOpenGoals {
					return &models.StateError{
						Type:         models.ErrorTypeHierarchyConstraint,
						Message:      fmt.Sprintf("invalid hierarchy: closed sprint has open goals: %s", sprintRelPath),
						Paths:        openGoalPaths,
						SuggestedFix: "Close all sprint goals before closing the sprint, or reopen the sprint",
					}
				}

				// Check for open tickets
				ticketsDir := filepath.Join(sprintPath, models.TicketsDir)
				hasOpenTickets, openTicketPaths, err := v.hasOpenTickets(ticketsDir)
				if err != nil {
					return err
				}
				if hasOpenTickets {
					return &models.StateError{
						Type:         models.ErrorTypeHierarchyConstraint,
						Message:      fmt.Sprintf("invalid hierarchy: closed sprint has open tickets: %s", sprintRelPath),
						Paths:        openTicketPaths,
						SuggestedFix: "Mark all tickets as done before closing the sprint, or reopen the sprint",
					}
				}
			}

			// Check tickets within this sprint
			ticketsDir := filepath.Join(sprintPath, models.TicketsDir)
			exists, err := DirExists(ticketsDir)
			if err != nil {
				return fmt.Errorf("failed to check tickets directory: %w", err)
			}
			if !exists {
				continue
			}

			ticketDirs, err := ListDirs(ticketsDir)
			if err != nil {
				return fmt.Errorf("failed to list tickets: %w", err)
			}

			for _, ticketDir := range ticketDirs {
				if !strings.HasSuffix(ticketDir, "-ticket") {
					continue
				}

				ticketPath := filepath.Join(ticketsDir, ticketDir)
				ticketRelPath, _ := v.getRelPath(ticketPath)

				ticketDone, err := IsDone(ticketPath)
				if err != nil {
					return fmt.Errorf("failed to check ticket done state: %w", err)
				}

				if ticketDone {
					// Check for open ticket goals
					ticketGoalsDir := filepath.Join(ticketPath, models.GoalsDir)
					hasOpenGoals, openGoalPaths, err := v.hasOpenGoals(ticketGoalsDir)
					if err != nil {
						return err
					}
					if hasOpenGoals {
						return &models.StateError{
							Type:         models.ErrorTypeHierarchyConstraint,
							Message:      fmt.Sprintf("invalid hierarchy: done ticket has open goals: %s", ticketRelPath),
							Paths:        openGoalPaths,
							SuggestedFix: "Close all ticket goals before marking ticket as done, or reopen the ticket",
						}
					}
				}
			}
		}
	}

	return nil
}

// CanClosePhase checks if a phase can be closed.
// A phase can be closed only if all its sprints are closed AND all phase goals are closed.
func (v *StateValidator) CanClosePhase(phasePath string) (bool, error) {
	relPath, err := v.getRelPath(phasePath)
	if err != nil {
		relPath = phasePath
	}

	// Check if phase is currently open
	isOpen, err := IsOpen(phasePath)
	if err != nil {
		return false, fmt.Errorf("failed to check phase open state: %w", err)
	}
	if !isOpen {
		return false, &models.StateError{
			Type:         models.ErrorTypeInvalidTransition,
			Message:      fmt.Sprintf("cannot close phase: phase is not open: %s", relPath),
			Paths:        []string{relPath},
			SuggestedFix: "Phase must be open before it can be closed",
		}
	}

	// Check phase goals
	phaseGoalsDir := filepath.Join(phasePath, models.GoalsDir)
	allGoalsClosed, openGoalPaths, err := v.areAllGoalsClosed(phaseGoalsDir)
	if err != nil {
		return false, err
	}
	if !allGoalsClosed {
		return false, &models.StateError{
			Type:         models.ErrorTypeInvalidTransition,
			Message:      fmt.Sprintf("invalid transition: cannot close phase with open goals: %s", relPath),
			Paths:        openGoalPaths,
			SuggestedFix: "Close all phase goals before closing the phase",
		}
	}

	// Check sprints
	sprintsDir := filepath.Join(phasePath, models.SprintsDir)
	allSprintsClosed, openSprintPaths, err := v.areAllSprintsClosed(sprintsDir)
	if err != nil {
		return false, err
	}
	if !allSprintsClosed {
		return false, &models.StateError{
			Type:         models.ErrorTypeInvalidTransition,
			Message:      fmt.Sprintf("invalid transition: cannot close phase with open sprints: %s", relPath),
			Paths:        openSprintPaths,
			SuggestedFix: "Close all sprints before closing the phase",
		}
	}

	return true, nil
}

// CanCloseSprint checks if a sprint can be closed.
// A sprint can be closed only if all its tickets are done AND all sprint goals are closed.
func (v *StateValidator) CanCloseSprint(sprintPath string) (bool, error) {
	relPath, err := v.getRelPath(sprintPath)
	if err != nil {
		relPath = sprintPath
	}

	// Check if sprint is currently open
	isOpen, err := IsOpen(sprintPath)
	if err != nil {
		return false, fmt.Errorf("failed to check sprint open state: %w", err)
	}
	if !isOpen {
		return false, &models.StateError{
			Type:         models.ErrorTypeInvalidTransition,
			Message:      fmt.Sprintf("cannot close sprint: sprint is not open: %s", relPath),
			Paths:        []string{relPath},
			SuggestedFix: "Sprint must be open before it can be closed",
		}
	}

	// Check sprint goals
	sprintGoalsDir := filepath.Join(sprintPath, models.GoalsDir)
	allGoalsClosed, openGoalPaths, err := v.areAllGoalsClosed(sprintGoalsDir)
	if err != nil {
		return false, err
	}
	if !allGoalsClosed {
		return false, &models.StateError{
			Type:         models.ErrorTypeInvalidTransition,
			Message:      fmt.Sprintf("invalid transition: cannot close sprint with open goals: %s", relPath),
			Paths:        openGoalPaths,
			SuggestedFix: "Close all sprint goals before closing the sprint",
		}
	}

	// Check tickets
	ticketsDir := filepath.Join(sprintPath, models.TicketsDir)
	allTicketsDone, openTicketPaths, err := v.areAllTicketsDone(ticketsDir)
	if err != nil {
		return false, err
	}
	if !allTicketsDone {
		return false, &models.StateError{
			Type:         models.ErrorTypeInvalidTransition,
			Message:      fmt.Sprintf("invalid transition: cannot close sprint with open tickets: %s", relPath),
			Paths:        openTicketPaths,
			SuggestedFix: "Mark all tickets as done before closing the sprint",
		}
	}

	return true, nil
}

// CanMarkTicketDone checks if a ticket can be marked as done.
// A ticket can be marked done only if all its goals are closed.
func (v *StateValidator) CanMarkTicketDone(ticketPath string) (bool, error) {
	relPath, err := v.getRelPath(ticketPath)
	if err != nil {
		relPath = ticketPath
	}

	// Check if ticket is currently open
	isOpen, err := IsOpen(ticketPath)
	if err != nil {
		return false, fmt.Errorf("failed to check ticket open state: %w", err)
	}
	if !isOpen {
		return false, &models.StateError{
			Type:         models.ErrorTypeInvalidTransition,
			Message:      fmt.Sprintf("cannot mark ticket done: ticket is not open: %s", relPath),
			Paths:        []string{relPath},
			SuggestedFix: "Ticket must be open before it can be marked done",
		}
	}

	// Check ticket goals
	ticketGoalsDir := filepath.Join(ticketPath, models.GoalsDir)
	allGoalsClosed, openGoalPaths, err := v.areAllGoalsClosed(ticketGoalsDir)
	if err != nil {
		return false, err
	}
	if !allGoalsClosed {
		return false, &models.StateError{
			Type:         models.ErrorTypeInvalidTransition,
			Message:      fmt.Sprintf("invalid transition: cannot mark ticket done with open goals: %s", relPath),
			Paths:        openGoalPaths,
			SuggestedFix: "Close all ticket goals before marking the ticket as done",
		}
	}

	return true, nil
}

// validateGoalsDir validates all goals in a goals directory.
func (v *StateValidator) validateGoalsDir(goalsDir, projectRoot string) error {
	exists, err := DirExists(goalsDir)
	if err != nil {
		return fmt.Errorf("failed to check goals directory: %w", err)
	}
	if !exists {
		return nil
	}

	goalDirs, err := ListDirs(goalsDir)
	if err != nil {
		return fmt.Errorf("failed to list goals: %w", err)
	}

	for _, goalDir := range goalDirs {
		if !strings.HasSuffix(goalDir, "-goal") {
			continue // Skip non-goal directories
		}

		goalPath := filepath.Join(goalsDir, goalDir)
		if err := v.ValidateGoalState(goalPath); err != nil {
			return err
		}
	}

	return nil
}

// hasOpenGoals checks if there are any open goals in the given goals directory.
// Returns (hasOpen, openGoalPaths, error).
func (v *StateValidator) hasOpenGoals(goalsDir string) (bool, []string, error) {
	exists, err := DirExists(goalsDir)
	if err != nil {
		return false, nil, fmt.Errorf("failed to check goals directory: %w", err)
	}
	if !exists {
		return false, nil, nil
	}

	goalDirs, err := ListDirs(goalsDir)
	if err != nil {
		return false, nil, fmt.Errorf("failed to list goals: %w", err)
	}

	var openGoalPaths []string
	for _, goalDir := range goalDirs {
		if !strings.HasSuffix(goalDir, "-goal") {
			continue
		}

		goalPath := filepath.Join(goalsDir, goalDir)
		isOpen, err := IsOpen(goalPath)
		if err != nil {
			return false, nil, fmt.Errorf("failed to check goal open state: %w", err)
		}
		if isOpen {
			relPath, _ := v.getRelPath(goalPath)
			openGoalPaths = append(openGoalPaths, relPath)
		}
	}

	return len(openGoalPaths) > 0, openGoalPaths, nil
}

// areAllGoalsClosed checks if all goals in the given goals directory are closed.
// Returns (allClosed, openGoalPaths, error).
func (v *StateValidator) areAllGoalsClosed(goalsDir string) (bool, []string, error) {
	exists, err := DirExists(goalsDir)
	if err != nil {
		return false, nil, fmt.Errorf("failed to check goals directory: %w", err)
	}
	if !exists {
		// No goals directory means all goals are "closed" (vacuously true)
		return true, nil, nil
	}

	goalDirs, err := ListDirs(goalsDir)
	if err != nil {
		return false, nil, fmt.Errorf("failed to list goals: %w", err)
	}

	if len(goalDirs) == 0 {
		// No goals means all goals are "closed" (vacuously true)
		return true, nil, nil
	}

	var openGoalPaths []string
	for _, goalDir := range goalDirs {
		if !strings.HasSuffix(goalDir, "-goal") {
			continue
		}

		goalPath := filepath.Join(goalsDir, goalDir)
		isClosed, err := IsClosed(goalPath)
		if err != nil {
			return false, nil, fmt.Errorf("failed to check goal closed state: %w", err)
		}
		if !isClosed {
			relPath, _ := v.getRelPath(goalPath)
			openGoalPaths = append(openGoalPaths, relPath)
		}
	}

	return len(openGoalPaths) == 0, openGoalPaths, nil
}

// hasOpenSprints checks if there are any open sprints in the given sprints directory.
// Returns (hasOpen, openSprintPaths, error).
func (v *StateValidator) hasOpenSprints(sprintsDir string) (bool, []string, error) {
	exists, err := DirExists(sprintsDir)
	if err != nil {
		return false, nil, fmt.Errorf("failed to check sprints directory: %w", err)
	}
	if !exists {
		return false, nil, nil
	}

	sprintDirs, err := ListDirs(sprintsDir)
	if err != nil {
		return false, nil, fmt.Errorf("failed to list sprints: %w", err)
	}

	var openSprintPaths []string
	for _, sprintDir := range sprintDirs {
		if !strings.HasSuffix(sprintDir, "-sprint") {
			continue
		}

		sprintPath := filepath.Join(sprintsDir, sprintDir)
		isOpen, err := IsOpen(sprintPath)
		if err != nil {
			return false, nil, fmt.Errorf("failed to check sprint open state: %w", err)
		}
		if isOpen {
			relPath, _ := v.getRelPath(sprintPath)
			openSprintPaths = append(openSprintPaths, relPath)
		}
	}

	return len(openSprintPaths) > 0, openSprintPaths, nil
}

// areAllSprintsClosed checks if all sprints in the given sprints directory are closed.
// Returns (allClosed, openSprintPaths, error).
func (v *StateValidator) areAllSprintsClosed(sprintsDir string) (bool, []string, error) {
	exists, err := DirExists(sprintsDir)
	if err != nil {
		return false, nil, fmt.Errorf("failed to check sprints directory: %w", err)
	}
	if !exists {
		// No sprints directory means all sprints are "closed" (vacuously true)
		return true, nil, nil
	}

	sprintDirs, err := ListDirs(sprintsDir)
	if err != nil {
		return false, nil, fmt.Errorf("failed to list sprints: %w", err)
	}

	if len(sprintDirs) == 0 {
		// No sprints means all sprints are "closed" (vacuously true)
		return true, nil, nil
	}

	var openSprintPaths []string
	for _, sprintDir := range sprintDirs {
		if !strings.HasSuffix(sprintDir, "-sprint") {
			continue
		}

		sprintPath := filepath.Join(sprintsDir, sprintDir)
		isClosed, err := IsClosed(sprintPath)
		if err != nil {
			return false, nil, fmt.Errorf("failed to check sprint closed state: %w", err)
		}
		if !isClosed {
			relPath, _ := v.getRelPath(sprintPath)
			openSprintPaths = append(openSprintPaths, relPath)
		}
	}

	return len(openSprintPaths) == 0, openSprintPaths, nil
}

// hasOpenTickets checks if there are any open tickets in the given tickets directory.
// Returns (hasOpen, openTicketPaths, error).
func (v *StateValidator) hasOpenTickets(ticketsDir string) (bool, []string, error) {
	exists, err := DirExists(ticketsDir)
	if err != nil {
		return false, nil, fmt.Errorf("failed to check tickets directory: %w", err)
	}
	if !exists {
		return false, nil, nil
	}

	ticketDirs, err := ListDirs(ticketsDir)
	if err != nil {
		return false, nil, fmt.Errorf("failed to list tickets: %w", err)
	}

	var openTicketPaths []string
	for _, ticketDir := range ticketDirs {
		if !strings.HasSuffix(ticketDir, "-ticket") {
			continue
		}

		ticketPath := filepath.Join(ticketsDir, ticketDir)
		isOpen, err := IsOpen(ticketPath)
		if err != nil {
			return false, nil, fmt.Errorf("failed to check ticket open state: %w", err)
		}
		if isOpen {
			relPath, _ := v.getRelPath(ticketPath)
			openTicketPaths = append(openTicketPaths, relPath)
		}
	}

	return len(openTicketPaths) > 0, openTicketPaths, nil
}

// areAllTicketsDone checks if all tickets in the given tickets directory are done.
// Returns (allDone, openTicketPaths, error).
func (v *StateValidator) areAllTicketsDone(ticketsDir string) (bool, []string, error) {
	exists, err := DirExists(ticketsDir)
	if err != nil {
		return false, nil, fmt.Errorf("failed to check tickets directory: %w", err)
	}
	if !exists {
		// No tickets directory means all tickets are "done" (vacuously true)
		return true, nil, nil
	}

	ticketDirs, err := ListDirs(ticketsDir)
	if err != nil {
		return false, nil, fmt.Errorf("failed to list tickets: %w", err)
	}

	if len(ticketDirs) == 0 {
		// No tickets means all tickets are "done" (vacuously true)
		return true, nil, nil
	}

	var openTicketPaths []string
	for _, ticketDir := range ticketDirs {
		if !strings.HasSuffix(ticketDir, "-ticket") {
			continue
		}

		ticketPath := filepath.Join(ticketsDir, ticketDir)
		isDone, err := IsDone(ticketPath)
		if err != nil {
			return false, nil, fmt.Errorf("failed to check ticket done state: %w", err)
		}
		if !isDone {
			relPath, _ := v.getRelPath(ticketPath)
			openTicketPaths = append(openTicketPaths, relPath)
		}
	}

	return len(openTicketPaths) == 0, openTicketPaths, nil
}

// getRelPath returns the relative path from the project root to the given path.
func (v *StateValidator) getRelPath(path string) (string, error) {
	if v.projectRoot == "" {
		// Try to determine project root from the path
		// by finding the .crumbler directory
		absPath, err := filepath.Abs(path)
		if err != nil {
			return path, err
		}

		// Walk up to find .crumbler
		current := absPath
		for {
			parent := filepath.Dir(current)
			if parent == current {
				// Reached filesystem root
				return path, fmt.Errorf("could not find project root")
			}

			crumblerPath := filepath.Join(parent, models.CrumblerDir)
			if exists, _ := DirExists(crumblerPath); exists {
				// Found .crumbler, parent is project root
				relPath, err := filepath.Rel(parent, absPath)
				if err != nil {
					return path, err
				}
				return relPath, nil
			}

			// Check if current directory contains .crumbler
			crumblerPath = filepath.Join(current, models.CrumblerDir)
			if exists, _ := DirExists(crumblerPath); exists {
				// current is project root
				relPath, err := filepath.Rel(current, absPath)
				if err != nil {
					return path, err
				}
				return relPath, nil
			}

			current = parent
		}
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return path, err
	}

	absRoot, err := filepath.Abs(v.projectRoot)
	if err != nil {
		return path, err
	}

	relPath, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		return path, err
	}

	return relPath, nil
}

// FileExists checks if a file exists at the given path.
// This is exported for use by other packages.
func FileExists(path string) (bool, error) {
	return fileExists(path)
}

// GetRelPathFromProjectRoot returns the relative path from the project root to the given path.
// This is a standalone function that doesn't require a StateValidator instance.
func GetRelPathFromProjectRoot(projectRoot, path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return path, err
	}

	absRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return path, err
	}

	relPath, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		return path, err
	}

	return relPath, nil
}

// ValidateGoalExists checks if a goal directory exists and has valid structure.
func (v *StateValidator) ValidateGoalExists(goalPath string) error {
	exists, err := DirExists(goalPath)
	if err != nil {
		return fmt.Errorf("failed to check goal directory: %w", err)
	}
	if !exists {
		relPath, _ := v.getRelPath(goalPath)
		return &models.StateError{
			Type:         models.ErrorTypeOrphanedState,
			Message:      fmt.Sprintf("goal directory does not exist: %s", relPath),
			Paths:        []string{relPath},
			SuggestedFix: "Create the goal directory or remove references to it",
		}
	}

	return v.ValidateGoalState(goalPath)
}

// CollectAllErrors performs validation and collects all errors instead of stopping at the first one.
// This is useful for providing comprehensive validation feedback.
func (v *StateValidator) CollectAllErrors(projectRoot string) []*models.StateError {
	var errors []*models.StateError

	crumblerDir := filepath.Join(projectRoot, models.CrumblerDir)
	exists, err := DirExists(crumblerDir)
	if err != nil || !exists {
		errors = append(errors, &models.StateError{
			Type:         models.ErrorTypeOrphanedState,
			Message:      fmt.Sprintf("not a crumbler project: %s directory not found", models.CrumblerDir),
			Paths:        []string{models.CrumblerDir},
			SuggestedFix: "Run 'crumbler init' to initialize a new project",
		})
		return errors
	}

	// Walk through all entities and collect errors
	phasesDir := filepath.Join(crumblerDir, models.PhasesDir)
	exists, _ = DirExists(phasesDir)
	if !exists {
		return errors
	}

	phaseDirs, err := ListDirs(phasesDir)
	if err != nil {
		return errors
	}

	for _, phaseDir := range phaseDirs {
		if !strings.HasSuffix(phaseDir, "-phase") {
			continue
		}

		phasePath := filepath.Join(phasesDir, phaseDir)

		// Validate phase state
		if err := v.ValidatePhaseState(phasePath); err != nil {
			if stateErr, ok := err.(*models.StateError); ok {
				errors = append(errors, stateErr)
			}
		}

		// Validate phase goals
		phaseGoalsDir := filepath.Join(phasePath, models.GoalsDir)
		errors = append(errors, v.collectGoalErrors(phaseGoalsDir)...)

		// Validate sprints
		sprintsDir := filepath.Join(phasePath, models.SprintsDir)
		sprintDirs, err := ListDirs(sprintsDir)
		if err != nil {
			continue
		}

		for _, sprintDir := range sprintDirs {
			if !strings.HasSuffix(sprintDir, "-sprint") {
				continue
			}

			sprintPath := filepath.Join(sprintsDir, sprintDir)

			// Validate sprint state
			if err := v.ValidateSprintState(sprintPath); err != nil {
				if stateErr, ok := err.(*models.StateError); ok {
					errors = append(errors, stateErr)
				}
			}

			// Validate sprint goals
			sprintGoalsDir := filepath.Join(sprintPath, models.GoalsDir)
			errors = append(errors, v.collectGoalErrors(sprintGoalsDir)...)

			// Validate tickets
			ticketsDir := filepath.Join(sprintPath, models.TicketsDir)
			ticketDirs, err := ListDirs(ticketsDir)
			if err != nil {
				continue
			}

			for _, ticketDir := range ticketDirs {
				if !strings.HasSuffix(ticketDir, "-ticket") {
					continue
				}

				ticketPath := filepath.Join(ticketsDir, ticketDir)

				// Validate ticket state
				if err := v.ValidateTicketState(ticketPath); err != nil {
					if stateErr, ok := err.(*models.StateError); ok {
						errors = append(errors, stateErr)
					}
				}

				// Validate ticket goals
				ticketGoalsDir := filepath.Join(ticketPath, models.GoalsDir)
				errors = append(errors, v.collectGoalErrors(ticketGoalsDir)...)
			}
		}
	}

	// Collect hierarchy errors
	hierarchyErrors := v.collectHierarchyErrors(projectRoot)
	errors = append(errors, hierarchyErrors...)

	return errors
}

// collectGoalErrors collects all goal validation errors in a goals directory.
func (v *StateValidator) collectGoalErrors(goalsDir string) []*models.StateError {
	var errors []*models.StateError

	exists, _ := DirExists(goalsDir)
	if !exists {
		return errors
	}

	goalDirs, err := ListDirs(goalsDir)
	if err != nil {
		return errors
	}

	for _, goalDir := range goalDirs {
		if !strings.HasSuffix(goalDir, "-goal") {
			continue
		}

		goalPath := filepath.Join(goalsDir, goalDir)
		if err := v.ValidateGoalState(goalPath); err != nil {
			if stateErr, ok := err.(*models.StateError); ok {
				errors = append(errors, stateErr)
			}
		}
	}

	return errors
}

// collectHierarchyErrors collects all hierarchy constraint errors.
func (v *StateValidator) collectHierarchyErrors(projectRoot string) []*models.StateError {
	var errors []*models.StateError

	crumblerDir := filepath.Join(projectRoot, models.CrumblerDir)
	phasesDir := filepath.Join(crumblerDir, models.PhasesDir)

	exists, _ := DirExists(phasesDir)
	if !exists {
		return errors
	}

	phaseDirs, err := ListDirs(phasesDir)
	if err != nil {
		return errors
	}

	for _, phaseDir := range phaseDirs {
		if !strings.HasSuffix(phaseDir, "-phase") {
			continue
		}

		phasePath := filepath.Join(phasesDir, phaseDir)
		phaseRelPath, _ := v.getRelPath(phasePath)

		phaseClosed, _ := IsClosed(phasePath)
		if phaseClosed {
			// Check for open phase goals
			phaseGoalsDir := filepath.Join(phasePath, models.GoalsDir)
			hasOpenGoals, openGoalPaths, _ := v.hasOpenGoals(phaseGoalsDir)
			if hasOpenGoals {
				errors = append(errors, &models.StateError{
					Type:         models.ErrorTypeHierarchyConstraint,
					Message:      fmt.Sprintf("invalid hierarchy: closed phase has open goals: %s", phaseRelPath),
					Paths:        openGoalPaths,
					SuggestedFix: "Close all phase goals before closing the phase, or reopen the phase",
				})
			}

			// Check for open sprints
			sprintsDir := filepath.Join(phasePath, models.SprintsDir)
			hasOpenSprints, openSprintPaths, _ := v.hasOpenSprints(sprintsDir)
			if hasOpenSprints {
				errors = append(errors, &models.StateError{
					Type:         models.ErrorTypeHierarchyConstraint,
					Message:      fmt.Sprintf("invalid hierarchy: closed phase has open sprints: %s", phaseRelPath),
					Paths:        openSprintPaths,
					SuggestedFix: "Close all sprints before closing the phase, or reopen the phase",
				})
			}
		}

		// Check sprints
		sprintsDir := filepath.Join(phasePath, models.SprintsDir)
		sprintDirs, _ := ListDirs(sprintsDir)

		for _, sprintDir := range sprintDirs {
			if !strings.HasSuffix(sprintDir, "-sprint") {
				continue
			}

			sprintPath := filepath.Join(sprintsDir, sprintDir)
			sprintRelPath, _ := v.getRelPath(sprintPath)

			sprintClosed, _ := IsClosed(sprintPath)
			if sprintClosed {
				// Check for open sprint goals
				sprintGoalsDir := filepath.Join(sprintPath, models.GoalsDir)
				hasOpenGoals, openGoalPaths, _ := v.hasOpenGoals(sprintGoalsDir)
				if hasOpenGoals {
					errors = append(errors, &models.StateError{
						Type:         models.ErrorTypeHierarchyConstraint,
						Message:      fmt.Sprintf("invalid hierarchy: closed sprint has open goals: %s", sprintRelPath),
						Paths:        openGoalPaths,
						SuggestedFix: "Close all sprint goals before closing the sprint, or reopen the sprint",
					})
				}

				// Check for open tickets
				ticketsDir := filepath.Join(sprintPath, models.TicketsDir)
				hasOpenTickets, openTicketPaths, _ := v.hasOpenTickets(ticketsDir)
				if hasOpenTickets {
					errors = append(errors, &models.StateError{
						Type:         models.ErrorTypeHierarchyConstraint,
						Message:      fmt.Sprintf("invalid hierarchy: closed sprint has open tickets: %s", sprintRelPath),
						Paths:        openTicketPaths,
						SuggestedFix: "Mark all tickets as done before closing the sprint, or reopen the sprint",
					})
				}
			}

			// Check tickets
			ticketsDir := filepath.Join(sprintPath, models.TicketsDir)
			ticketDirs, _ := ListDirs(ticketsDir)

			for _, ticketDir := range ticketDirs {
				if !strings.HasSuffix(ticketDir, "-ticket") {
					continue
				}

				ticketPath := filepath.Join(ticketsDir, ticketDir)
				ticketRelPath, _ := v.getRelPath(ticketPath)

				ticketDone, _ := IsDone(ticketPath)
				if ticketDone {
					// Check for open ticket goals
					ticketGoalsDir := filepath.Join(ticketPath, models.GoalsDir)
					hasOpenGoals, openGoalPaths, _ := v.hasOpenGoals(ticketGoalsDir)
					if hasOpenGoals {
						errors = append(errors, &models.StateError{
							Type:         models.ErrorTypeHierarchyConstraint,
							Message:      fmt.Sprintf("invalid hierarchy: done ticket has open goals: %s", ticketRelPath),
							Paths:        openGoalPaths,
							SuggestedFix: "Close all ticket goals before marking ticket as done, or reopen the ticket",
						})
					}
				}
			}
		}
	}

	return errors
}

// FileExistsAt checks if a file exists at the given path.
// This is an alias for FileExists for clarity.
func FileExistsAt(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return !info.IsDir(), nil
}
