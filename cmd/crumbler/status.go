package crumbler

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// runStatus handles the 'crumbler status' command.
// It displays the current state of the project.
func runStatus(args []string) error {
	// Handle help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help") {
		printStatusHelp()
		return nil
	}

	projectRoot, err := requireProject()
	if err != nil {
		return err
	}

	status, err := getProjectStatus(projectRoot)
	if err != nil {
		return err
	}

	printProjectStatus(status)
	return nil
}

// ProjectStatus holds the aggregated project state.
type ProjectStatus struct {
	ProjectRoot    string
	TotalPhases    int
	OpenPhases     int
	ClosedPhases   int
	TotalSprints   int
	OpenSprints    int
	ClosedSprints  int
	TotalTickets   int
	OpenTickets    int
	DoneTickets    int
	CurrentPhase   *PhaseInfo
	CurrentSprint  *SprintInfo
	OpenTicketList []TicketInfo
}

// PhaseInfo holds information about a phase.
type PhaseInfo struct {
	ID     string
	Path   string
	Status string // "open" or "closed"
	Goals  []GoalInfo
}

// SprintInfo holds information about a sprint.
type SprintInfo struct {
	ID      string
	Path    string
	PhaseID string
	Status  string // "open" or "closed"
	Goals   []GoalInfo
}

// TicketInfo holds information about a ticket.
type TicketInfo struct {
	ID       string
	Path     string
	SprintID string
	PhaseID  string
	Status   string // "open" or "done"
	Goals    []GoalInfo
}

// GoalInfo holds information about a goal.
type GoalInfo struct {
	ID     string
	Name   string
	Path   string
	Status string // "open" or "closed"
}

// getProjectStatus gathers the project state.
func getProjectStatus(projectRoot string) (*ProjectStatus, error) {
	status := &ProjectStatus{
		ProjectRoot: projectRoot,
	}

	phasesPath := phasesDir(projectRoot)
	phases, err := listDirs(phasesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return status, nil
		}
		return nil, err
	}

	for _, phaseName := range phases {
		if !strings.HasSuffix(phaseName, "-phase") {
			continue
		}

		status.TotalPhases++
		phasePath := phasesPath + "/" + phaseName
		phaseStatus := getEntityStatus(phasePath, "open", "closed")

		if phaseStatus == "open" {
			status.OpenPhases++
			phaseID := strings.TrimSuffix(phaseName, "-phase")
			goals, _ := listGoals(phasePath + "/goals")
			status.CurrentPhase = &PhaseInfo{
				ID:     phaseID,
				Path:   phasePath,
				Status: phaseStatus,
				Goals:  goals,
			}
		} else if phaseStatus == "closed" {
			status.ClosedPhases++
		}

		// Count sprints in this phase
		sprintsPath := phasePath + "/sprints"
		sprints, err := listDirs(sprintsPath)
		if err != nil {
			continue
		}

		for _, sprintName := range sprints {
			if !strings.HasSuffix(sprintName, "-sprint") {
				continue
			}

			status.TotalSprints++
			sprintPath := sprintsPath + "/" + sprintName
			sprintStatus := getEntityStatus(sprintPath, "open", "closed")

			if sprintStatus == "open" {
				status.OpenSprints++
				phaseID := strings.TrimSuffix(phaseName, "-phase")
				sprintID := strings.TrimSuffix(sprintName, "-sprint")
				goals, _ := listGoals(sprintPath + "/goals")
				status.CurrentSprint = &SprintInfo{
					ID:      sprintID,
					Path:    sprintPath,
					PhaseID: phaseID,
					Status:  sprintStatus,
					Goals:   goals,
				}
			} else if sprintStatus == "closed" {
				status.ClosedSprints++
			}

			// Count tickets in this sprint
			ticketsPath := sprintPath + "/tickets"
			tickets, err := listDirs(ticketsPath)
			if err != nil {
				continue
			}

			for _, ticketName := range tickets {
				if !strings.HasSuffix(ticketName, "-ticket") {
					continue
				}

				status.TotalTickets++
				ticketPath := ticketsPath + "/" + ticketName
				ticketStatus := getEntityStatus(ticketPath, "open", "done")

				if ticketStatus == "open" {
					status.OpenTickets++
					phaseID := strings.TrimSuffix(phaseName, "-phase")
					sprintID := strings.TrimSuffix(sprintName, "-sprint")
					ticketID := strings.TrimSuffix(ticketName, "-ticket")
					goals, _ := listGoals(ticketPath + "/goals")
					status.OpenTicketList = append(status.OpenTicketList, TicketInfo{
						ID:       ticketID,
						Path:     ticketPath,
						SprintID: sprintID,
						PhaseID:  phaseID,
						Status:   ticketStatus,
						Goals:    goals,
					})
				} else if ticketStatus == "done" {
					status.DoneTickets++
				}
			}
		}
	}

	return status, nil
}

// listDirs returns sorted directory names in a path.
func listDirs(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}
	sort.Strings(dirs)
	return dirs, nil
}

// getEntityStatus returns the status of an entity based on status files.
func getEntityStatus(path, openFile, closedFile string) string {
	if _, err := os.Stat(filepath.Join(path, openFile)); err == nil {
		return "open"
	}
	if _, err := os.Stat(filepath.Join(path, closedFile)); err == nil {
		return closedFile // "closed" or "done"
	}
	return "unknown"
}

// listGoals returns goals in a goals directory.
func listGoals(goalsPath string) ([]GoalInfo, error) {
	dirs, err := listDirs(goalsPath)
	if err != nil {
		return nil, err
	}

	var goals []GoalInfo
	for _, goalName := range dirs {
		if !strings.HasSuffix(goalName, "-goal") {
			continue
		}

		goalPath := goalsPath + "/" + goalName
		goalID := strings.TrimSuffix(goalName, "-goal")

		// Read goal name from name file
		name := ""
		nameBytes, err := os.ReadFile(goalPath + "/name")
		if err == nil {
			name = strings.TrimSpace(string(nameBytes))
		}

		status := "open"
		if _, err := os.Stat(goalPath + "/closed"); err == nil {
			status = "closed"
		}

		goals = append(goals, GoalInfo{
			ID:     goalID,
			Name:   name,
			Path:   goalPath,
			Status: status,
		})
	}

	return goals, nil
}

// printProjectStatus displays the project status.
func printProjectStatus(status *ProjectStatus) {
	fmt.Println("Project Status")
	fmt.Println("==============")
	fmt.Println()

	// Overall progress
	fmt.Println("Progress:")
	fmt.Printf("  Phases:  %d/%d completed %s\n",
		status.ClosedPhases, status.TotalPhases,
		progressBar(status.ClosedPhases, status.TotalPhases))
	fmt.Printf("  Sprints: %d/%d completed %s\n",
		status.ClosedSprints, status.TotalSprints,
		progressBar(status.ClosedSprints, status.TotalSprints))
	fmt.Printf("  Tickets: %d/%d done %s\n",
		status.DoneTickets, status.TotalTickets,
		progressBar(status.DoneTickets, status.TotalTickets))
	fmt.Println()

	// Current phase
	if status.CurrentPhase != nil {
		fmt.Println("Current Phase:")
		fmt.Printf("  ID:     %s\n", status.CurrentPhase.ID)
		fmt.Printf("  Path:   %s\n", relPath(status.ProjectRoot, status.CurrentPhase.Path))
		fmt.Printf("  Status: %s\n", status.CurrentPhase.Status)
		if len(status.CurrentPhase.Goals) > 0 {
			fmt.Println("  Goals:")
			for _, goal := range status.CurrentPhase.Goals {
				marker := "[ ]"
				if goal.Status == "closed" {
					marker = "[x]"
				}
				fmt.Printf("    %s %s: %s\n", marker, goal.ID, goal.Name)
			}
		}
		fmt.Println()
	} else {
		fmt.Println("Current Phase: none")
		fmt.Println()
	}

	// Current sprint
	if status.CurrentSprint != nil {
		fmt.Println("Current Sprint:")
		fmt.Printf("  ID:     %s (in phase %s)\n", status.CurrentSprint.ID, status.CurrentSprint.PhaseID)
		fmt.Printf("  Path:   %s\n", relPath(status.ProjectRoot, status.CurrentSprint.Path))
		fmt.Printf("  Status: %s\n", status.CurrentSprint.Status)
		if len(status.CurrentSprint.Goals) > 0 {
			fmt.Println("  Goals:")
			for _, goal := range status.CurrentSprint.Goals {
				marker := "[ ]"
				if goal.Status == "closed" {
					marker = "[x]"
				}
				fmt.Printf("    %s %s: %s\n", marker, goal.ID, goal.Name)
			}
		}
		fmt.Println()
	} else {
		fmt.Println("Current Sprint: none")
		fmt.Println()
	}

	// Open tickets
	fmt.Println("Open Tickets:")
	if len(status.OpenTicketList) == 0 {
		fmt.Println("  none")
	} else {
		for _, ticket := range status.OpenTicketList {
			fmt.Printf("  [%s] %s (sprint %s, phase %s)\n",
				ticket.ID, relPath(status.ProjectRoot, ticket.Path),
				ticket.SprintID, ticket.PhaseID)
			if len(ticket.Goals) > 0 {
				for _, goal := range ticket.Goals {
					marker := "[ ]"
					if goal.Status == "closed" {
						marker = "[x]"
					}
					fmt.Printf("       %s %s: %s\n", marker, goal.ID, goal.Name)
				}
			}
		}
	}
}

// progressBar returns a simple text progress bar.
func progressBar(completed, total int) string {
	if total == 0 {
		return "[----------]"
	}

	width := 10
	filled := (completed * width) / total
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("#", filled) + strings.Repeat("-", width-filled)
	return "[" + bar + "]"
}

// printStatusHelp prints help for the status command.
func printStatusHelp() {
	fmt.Print(`crumbler status - Show current project state

USAGE:
    crumbler status

DESCRIPTION:
    Displays the current state of the crumbler project including:
    - Overall progress (phases, sprints, tickets)
    - Current open phase and its goals
    - Current open sprint and its goals
    - List of open tickets with their goals

    Progress is shown with visual indicators and completion counts.

OUTPUT:
    Progress:         Completion counts with progress bars
    Current Phase:    ID, path, status, and goals of open phase
    Current Sprint:   ID, path, status, and goals of open sprint
    Open Tickets:     List of tickets that need work

EXAMPLES:
    crumbler status

OUTPUT EXAMPLE:
    Project Status
    ==============

    Progress:
      Phases:  1/3 completed [###-------]
      Sprints: 2/5 completed [####------]
      Tickets: 8/12 done [######----]

    Current Phase:
      ID:     0001
      Path:   .crumbler/phases/0001-phase
      Status: open
      Goals:
        [x] 0001: Implement core data models
        [ ] 0002: Set up API endpoints

    Current Sprint:
      ID:     0002 (in phase 0001)
      Path:   .crumbler/phases/0001-phase/sprints/0002-sprint
      Status: open
      Goals:
        [x] 0001: Complete user authentication
        [ ] 0002: Add input validation

    Open Tickets:
      [0003] .crumbler/phases/0001-phase/sprints/0002-sprint/tickets/0003-ticket
         [ ] 0001: Write unit tests

FOR AI AGENTS:
    Use 'crumbler status' to understand the current project state before
    deciding on next actions. The status output shows:

    1. Overall progress - how much work is complete
    2. Current phase - what high-level work is in progress
    3. Current sprint - what sprint iteration is active
    4. Open tickets - what specific tasks need to be done

    Based on status, you can determine:
    - If no open phase: check if roadmap is complete or create next phase
    - If no open sprint: check if phase goals are met or create next sprint
    - If no open tickets: check if sprint goals are met or create tickets
    - If open tickets exist: work on completing them
`)
}
