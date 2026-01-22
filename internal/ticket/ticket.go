// Package ticket provides Ticket management functions for crumbler.
package ticket

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

var (
	// ticketDirPattern matches ticket directory names like "0001-ticket"
	ticketDirPattern = regexp.MustCompile(`^(\d{4})-ticket$`)
	// goalDirPattern matches goal directory names like "0001-goal"
	goalDirPattern = regexp.MustCompile(`^(\d{4})-goal$`)
)

// GetOpenTickets scans the tickets/ subdirectory for directories with an open file (no done file).
// Returns an empty slice if no open tickets are found.
func GetOpenTickets(sprintPath string) ([]models.Ticket, error) {
	ticketsPath := filepath.Join(sprintPath, models.TicketsDir)

	dirs, err := state.ListDirs(ticketsPath)
	if err != nil {
		return nil, nil // Return empty list if directory doesn't exist
	}

	var openTickets []models.Ticket
	for _, dir := range dirs {
		if !ticketDirPattern.MatchString(dir) {
			continue
		}

		ticketPath := filepath.Join(ticketsPath, dir)

		// Validate state
		if err := ValidateTicketState(ticketPath); err != nil {
			return nil, err
		}

		// Check if ticket is open (has open file, no done file)
		isOpen, err := state.IsOpen(ticketPath)
		if err != nil {
			return nil, fmt.Errorf("failed to check open state at %s: %w", getRelPath(ticketPath), err)
		}
		isDone, err := state.IsDone(ticketPath)
		if err != nil {
			return nil, fmt.Errorf("failed to check done state at %s: %w", getRelPath(ticketPath), err)
		}

		if isOpen && !isDone {
			ticket, err := loadTicket(ticketPath, dir)
			if err != nil {
				return nil, err
			}
			openTickets = append(openTickets, *ticket)
		}
	}

	// Sort by index
	sort.Slice(openTickets, func(i, j int) bool {
		return openTickets[i].Index < openTickets[j].Index
	})

	return openTickets, nil
}

// GetNextTicketIndex finds the next ticket number.
// Returns 1 if no tickets exist, otherwise returns max existing index + 1.
func GetNextTicketIndex(sprintPath string) (int, error) {
	ticketsPath := filepath.Join(sprintPath, models.TicketsDir)

	dirs, err := state.ListDirs(ticketsPath)
	if err != nil {
		return 1, nil // If directory doesn't exist or is empty, start at 1
	}

	maxIndex := 0
	for _, dir := range dirs {
		matches := ticketDirPattern.FindStringSubmatch(dir)
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

// CreateTicket creates a new ticket directory structure.
// Returns the path to the created ticket.
func CreateTicket(sprintPath string, index int) (string, error) {
	ticketID := state.FormatTicketID(index)
	ticketPath := filepath.Join(sprintPath, models.TicketsDir, ticketID)

	// Create tickets directory if it doesn't exist
	ticketsPath := filepath.Join(sprintPath, models.TicketsDir)
	if err := os.MkdirAll(ticketsPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create tickets directory at %s: %w", getRelPath(ticketsPath), err)
	}

	// Create ticket directory
	if err := os.MkdirAll(ticketPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create ticket directory at %s: %w", getRelPath(ticketPath), err)
	}

	// Create empty README.md
	readmePath := filepath.Join(ticketPath, models.ReadmeFile)
	if err := touchFile(readmePath); err != nil {
		return "", fmt.Errorf("failed to create README.md at %s: %w", getRelPath(readmePath), err)
	}

	// Create goals/ subdirectory
	goalsPath := filepath.Join(ticketPath, models.GoalsDir)
	if err := os.MkdirAll(goalsPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create goals directory at %s: %w", getRelPath(goalsPath), err)
	}

	// Touch open file
	openPath := filepath.Join(ticketPath, models.StatusFileOpen)
	if err := touchFile(openPath); err != nil {
		return "", fmt.Errorf("failed to create open file at %s: %w", getRelPath(openPath), err)
	}

	return ticketPath, nil
}

// GetTicket loads a ticket by ID.
func GetTicket(sprintPath, ticketID string) (*models.Ticket, error) {
	ticketPath := filepath.Join(sprintPath, models.TicketsDir, ticketID)

	info, err := os.Stat(ticketPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("ticket not found at %s", getRelPath(ticketPath))
		}
		return nil, fmt.Errorf("failed to check ticket directory at %s: %w", getRelPath(ticketPath), err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("ticket path is not a directory: %s", getRelPath(ticketPath))
	}

	if err := ValidateTicketState(ticketPath); err != nil {
		return nil, err
	}

	return loadTicket(ticketPath, ticketID)
}

// ListTickets lists all tickets sorted by ID.
func ListTickets(sprintPath string) ([]models.Ticket, error) {
	ticketsPath := filepath.Join(sprintPath, models.TicketsDir)

	dirs, err := state.ListDirs(ticketsPath)
	if err != nil {
		return nil, nil // Return empty list if directory doesn't exist
	}

	var ticketDirs []string
	for _, dir := range dirs {
		if ticketDirPattern.MatchString(dir) {
			ticketDirs = append(ticketDirs, dir)
		}
	}

	// Sort by ID
	sort.Strings(ticketDirs)

	var tickets []models.Ticket
	for _, dir := range ticketDirs {
		ticketPath := filepath.Join(ticketsPath, dir)

		if err := ValidateTicketState(ticketPath); err != nil {
			return nil, err
		}

		ticket, err := loadTicket(ticketPath, dir)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, *ticket)
	}

	return tickets, nil
}

// GetTicketGoals scans the goals/ directory and returns a list of goals.
func GetTicketGoals(ticketPath string) ([]models.Goal, error) {
	goalsPath := filepath.Join(ticketPath, models.GoalsDir)

	dirs, err := state.ListDirs(goalsPath)
	if err != nil {
		return nil, nil // Return empty list if directory doesn't exist
	}

	var goalDirs []string
	for _, dir := range dirs {
		if goalDirPattern.MatchString(dir) {
			goalDirs = append(goalDirs, dir)
		}
	}

	// Sort by ID
	sort.Strings(goalDirs)

	var goals []models.Goal
	for _, dir := range goalDirs {
		goalPath := filepath.Join(goalsPath, dir)

		goal, err := loadGoal(goalPath, dir)
		if err != nil {
			return nil, err
		}
		goals = append(goals, *goal)
	}

	return goals, nil
}

// CreateTicketGoal creates a goal in the ticket's goals/ directory.
// Returns the path to the created goal.
func CreateTicketGoal(ticketPath string, index int, goalName string) (string, error) {
	goalID := state.FormatGoalID(index)
	goalPath := filepath.Join(ticketPath, models.GoalsDir, goalID)

	// Create goals directory if it doesn't exist
	goalsPath := filepath.Join(ticketPath, models.GoalsDir)
	if err := os.MkdirAll(goalsPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create goals directory at %s: %w", getRelPath(goalsPath), err)
	}

	// Create goal directory
	if err := os.MkdirAll(goalPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create goal directory at %s: %w", getRelPath(goalPath), err)
	}

	// Write name file
	namePath := filepath.Join(goalPath, models.GoalNameFile)
	if err := os.WriteFile(namePath, []byte(goalName), 0644); err != nil {
		return "", fmt.Errorf("failed to write name file at %s: %w", getRelPath(namePath), err)
	}

	// Touch open file
	openPath := filepath.Join(goalPath, models.StatusFileOpen)
	if err := touchFile(openPath); err != nil {
		return "", fmt.Errorf("failed to create open file at %s: %w", getRelPath(openPath), err)
	}

	return goalPath, nil
}

// CloseTicketGoal closes a goal by deleting the open file and touching the closed file.
func CloseTicketGoal(ticketPath string, goalID string) error {
	goalPath := filepath.Join(ticketPath, models.GoalsDir, goalID)

	// Check goal exists
	info, err := os.Stat(goalPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("goal not found at %s", getRelPath(goalPath))
		}
		return fmt.Errorf("failed to check goal directory at %s: %w", getRelPath(goalPath), err)
	}
	if !info.IsDir() {
		return fmt.Errorf("goal path is not a directory: %s", getRelPath(goalPath))
	}

	// Validate current state (check for conflicting files)
	if err := validateGoalState(goalPath); err != nil {
		return err
	}

	// Check that goal is currently open
	isOpen, err := state.IsOpen(goalPath)
	if err != nil {
		return fmt.Errorf("failed to check open state at %s: %w", getRelPath(goalPath), err)
	}
	if !isOpen {
		isClosed, err := state.IsClosed(goalPath)
		if err != nil {
			return fmt.Errorf("failed to check closed state at %s: %w", getRelPath(goalPath), err)
		}
		if isClosed {
			return fmt.Errorf("goal already closed at %s", getRelPath(goalPath))
		}
		return fmt.Errorf("goal has no state at %s", getRelPath(goalPath))
	}

	// Delete open file
	openPath := filepath.Join(goalPath, models.StatusFileOpen)
	if err := os.Remove(openPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove open file at %s: %w", getRelPath(openPath), err)
	}

	// Touch closed file
	closedPath := filepath.Join(goalPath, models.StatusFileClosed)
	if err := touchFile(closedPath); err != nil {
		return fmt.Errorf("failed to create closed file at %s: %w", getRelPath(closedPath), err)
	}

	return nil
}

// AreTicketGoalsMet checks if all ticket goals have closed file.
// Returns true if all goals are closed, false if any goal is open or if no goals exist.
func AreTicketGoalsMet(ticketPath string) (bool, error) {
	goals, err := GetTicketGoals(ticketPath)
	if err != nil {
		return false, err
	}

	// If no goals exist, return true (vacuously true - no goals to close)
	if len(goals) == 0 {
		return true, nil
	}

	// Check all goals are closed
	for _, goal := range goals {
		if goal.Status != models.StatusClosed {
			return false, nil
		}
	}

	return true, nil
}

// IsTicketComplete checks if done file exists AND all ticket goals have closed file.
func IsTicketComplete(ticketPath string) (bool, error) {
	// Check if done file exists
	isDone, err := state.IsDone(ticketPath)
	if err != nil {
		return false, fmt.Errorf("failed to check done state at %s: %w", getRelPath(ticketPath), err)
	}
	if !isDone {
		return false, nil
	}

	// Check if all goals are closed
	return AreTicketGoalsMet(ticketPath)
}

// MarkTicketDone marks a ticket as done by deleting the open file and touching the done file.
// Returns an error if ticket goals are still open.
func MarkTicketDone(ticketPath string) error {
	// Validate current state
	if err := ValidateTicketState(ticketPath); err != nil {
		return err
	}

	// Check that ticket is currently open
	isOpen, err := state.IsOpen(ticketPath)
	if err != nil {
		return fmt.Errorf("failed to check open state at %s: %w", getRelPath(ticketPath), err)
	}
	if !isOpen {
		isDone, err := state.IsDone(ticketPath)
		if err != nil {
			return fmt.Errorf("failed to check done state at %s: %w", getRelPath(ticketPath), err)
		}
		if isDone {
			return fmt.Errorf("ticket already done at %s", getRelPath(ticketPath))
		}
		return fmt.Errorf("ticket has no state at %s", getRelPath(ticketPath))
	}

	// Check for open goals
	openGoalPaths, err := getOpenGoalPaths(ticketPath)
	if err != nil {
		return err
	}
	if len(openGoalPaths) > 0 {
		return fmt.Errorf("cannot mark ticket done: goals still open: %s", strings.Join(openGoalPaths, ", "))
	}

	// Delete open file
	openPath := filepath.Join(ticketPath, models.StatusFileOpen)
	if err := os.Remove(openPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove open file at %s: %w", getRelPath(openPath), err)
	}

	// Touch done file
	donePath := filepath.Join(ticketPath, models.StatusFileDone)
	if err := touchFile(donePath); err != nil {
		return fmt.Errorf("failed to create done file at %s: %w", getRelPath(donePath), err)
	}

	return nil
}

// ValidateTicketState checks for invalid ticket state (both open and done exist).
func ValidateTicketState(ticketPath string) error {
	isOpen, err := state.IsOpen(ticketPath)
	if err != nil {
		return fmt.Errorf("failed to check open state at %s: %w", getRelPath(ticketPath), err)
	}
	isDone, err := state.IsDone(ticketPath)
	if err != nil {
		return fmt.Errorf("failed to check done state at %s: %w", getRelPath(ticketPath), err)
	}

	if isOpen && isDone {
		return fmt.Errorf("invalid state: both 'open' and 'done' exist in %s", getRelPath(ticketPath))
	}

	return nil
}

// loadTicket loads a ticket from the given path.
func loadTicket(ticketPath, id string) (*models.Ticket, error) {
	// Parse index from ID
	matches := ticketDirPattern.FindStringSubmatch(id)
	var index int
	if matches != nil {
		index, _ = strconv.Atoi(matches[1])
	}

	// Determine status
	var status models.Status
	isOpen, err := state.IsOpen(ticketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check open state at %s: %w", getRelPath(ticketPath), err)
	}
	isDone, err := state.IsDone(ticketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check done state at %s: %w", getRelPath(ticketPath), err)
	}

	if isOpen {
		status = models.StatusOpen
	} else if isDone {
		status = models.StatusDone
	} else {
		status = models.StatusUnknown
	}

	// Load goals
	goals, err := GetTicketGoals(ticketPath)
	if err != nil {
		return nil, err
	}

	return &models.Ticket{
		ID:              id,
		Path:            ticketPath,
		Goals:           goals,
		DescriptionPath: filepath.Join(ticketPath, models.ReadmeFile),
		Status:          status,
		Index:           index,
	}, nil
}

// loadGoal loads a goal from the given path.
func loadGoal(goalPath, id string) (*models.Goal, error) {
	// Validate state
	if err := validateGoalState(goalPath); err != nil {
		return nil, err
	}

	// Parse index from ID
	matches := goalDirPattern.FindStringSubmatch(id)
	var index int
	if matches != nil {
		index, _ = strconv.Atoi(matches[1])
	}

	// Determine status
	var status models.Status
	isOpen, err := state.IsOpen(goalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check open state at %s: %w", getRelPath(goalPath), err)
	}
	isClosed, err := state.IsClosed(goalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check closed state at %s: %w", getRelPath(goalPath), err)
	}

	if isOpen {
		status = models.StatusOpen
	} else if isClosed {
		status = models.StatusClosed
	} else {
		status = models.StatusUnknown
	}

	// Read name
	namePath := filepath.Join(goalPath, models.GoalNameFile)
	nameBytes, err := os.ReadFile(namePath)
	var name string
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("goal name file missing at %s", getRelPath(namePath))
		}
		return nil, fmt.Errorf("failed to read goal name at %s: %w", getRelPath(namePath), err)
	}
	name = strings.TrimSpace(string(nameBytes))

	return &models.Goal{
		ID:     id,
		Path:   goalPath,
		Name:   name,
		Status: status,
		Index:  index,
	}, nil
}

// validateGoalState checks for invalid goal state (both open and closed exist).
func validateGoalState(goalPath string) error {
	isOpen, err := state.IsOpen(goalPath)
	if err != nil {
		return fmt.Errorf("failed to check open state at %s: %w", getRelPath(goalPath), err)
	}
	isClosed, err := state.IsClosed(goalPath)
	if err != nil {
		return fmt.Errorf("failed to check closed state at %s: %w", getRelPath(goalPath), err)
	}

	if isOpen && isClosed {
		return fmt.Errorf("invalid state: both 'open' and 'closed' exist in %s", getRelPath(goalPath))
	}

	return nil
}

// getOpenGoalPaths returns the relative paths of all open goals.
func getOpenGoalPaths(ticketPath string) ([]string, error) {
	goalsPath := filepath.Join(ticketPath, models.GoalsDir)

	dirs, err := state.ListDirs(goalsPath)
	if err != nil {
		return nil, nil // If no goals directory, no open goals
	}

	var openPaths []string
	for _, dir := range dirs {
		if !goalDirPattern.MatchString(dir) {
			continue
		}

		goalPath := filepath.Join(goalsPath, dir)

		// Validate goal state
		if err := validateGoalState(goalPath); err != nil {
			return nil, err
		}

		isOpen, err := state.IsOpen(goalPath)
		if err != nil {
			return nil, fmt.Errorf("failed to check open state at %s: %w", getRelPath(goalPath), err)
		}

		if isOpen {
			openPaths = append(openPaths, getRelPath(goalPath))
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

// touchFile creates an empty file at the given path.
func touchFile(path string) error {
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	return file.Close()
}
