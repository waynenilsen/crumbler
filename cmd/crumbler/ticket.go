package crumbler

import (
	"fmt"
	"os"
	"strings"
)

// runTicket handles the 'crumbler ticket' command and its subcommands.
func runTicket(args []string) error {
	if len(args) == 0 {
		printTicketHelp()
		return nil
	}

	// Handle help flag
	if args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		printTicketHelp()
		return nil
	}

	switch args[0] {
	case "list":
		return runTicketList(args[1:])
	case "create":
		return runTicketCreate(args[1:])
	case "done":
		return runTicketDone(args[1:])
	case "goal":
		return runTicketGoal(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "error: unknown ticket subcommand '%s'\n\n", args[0])
		printTicketHelp()
		return fmt.Errorf("unknown ticket subcommand: %s", args[0])
	}
}

// runTicketList handles 'crumbler ticket list [sprint-id]'.
func runTicketList(args []string) error {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help") {
		printTicketListHelp()
		return nil
	}

	projectRoot, err := requireProject()
	if err != nil {
		return err
	}

	// Determine which sprint to list tickets for
	var sprintID string
	if len(args) > 0 {
		sprintID = args[0]
	} else {
		// Find current open sprint
		status, err := getProjectStatus(projectRoot)
		if err != nil {
			return err
		}
		if status.CurrentSprint == nil {
			return fmt.Errorf("no open sprint found. Specify a sprint ID: crumbler ticket list <sprint-id>")
		}
		sprintID = status.CurrentSprint.ID
	}

	// Find sprint
	sprintPath, phaseID, err := findSprint(projectRoot, sprintID)
	if err != nil {
		return err
	}

	ticketsPath := sprintPath + "/tickets"
	tickets, err := listDirs(ticketsPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("No tickets found in sprint %s.\n", sprintID)
			return nil
		}
		return fmt.Errorf("failed to list tickets: %w", err)
	}

	fmt.Printf("Tickets in Sprint %s (Phase %s):\n", sprintID, phaseID)
	fmt.Println()

	count := 0
	for _, ticketName := range tickets {
		if !strings.HasSuffix(ticketName, "-ticket") {
			continue
		}

		count++
		ticketPath := ticketsPath + "/" + ticketName
		status := getEntityStatus(ticketPath, "open", "done")
		ticketID := strings.TrimSuffix(ticketName, "-ticket")

		marker := "[ ]"
		if status == "done" {
			marker = "[x]"
		}

		fmt.Printf("  %s %s  %s\n", marker, ticketID, relPath(projectRoot, ticketPath))

		// Show goals
		goals, _ := listGoals(ticketPath + "/goals")
		for _, goal := range goals {
			gMarker := "[ ]"
			if goal.Status == "closed" {
				gMarker = "[x]"
			}
			fmt.Printf("      %s Goal %s: %s\n", gMarker, goal.ID, goal.Name)
		}
	}

	if count == 0 {
		fmt.Println("  No tickets found.")
		fmt.Println()
		fmt.Printf("Create a ticket with: crumbler ticket create %s\n", sprintID)
	}

	return nil
}

// runTicketCreate handles 'crumbler ticket create [sprint-id]'.
func runTicketCreate(args []string) error {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help") {
		printTicketCreateHelp()
		return nil
	}

	projectRoot, err := requireProject()
	if err != nil {
		return err
	}

	// Determine which sprint to create ticket in
	var sprintID string
	if len(args) > 0 {
		sprintID = args[0]
	} else {
		// Find current open sprint
		status, err := getProjectStatus(projectRoot)
		if err != nil {
			return err
		}
		if status.CurrentSprint == nil {
			return fmt.Errorf("no open sprint found. Specify a sprint ID: crumbler ticket create <sprint-id>")
		}
		sprintID = status.CurrentSprint.ID
	}

	// Find sprint
	sprintPath, phaseID, err := findSprint(projectRoot, sprintID)
	if err != nil {
		return err
	}

	// Find next ticket number
	ticketsPath := sprintPath + "/tickets"
	nextIndex := 1

	tickets, err := listDirs(ticketsPath)
	if err == nil {
		for _, ticketName := range tickets {
			if strings.HasSuffix(ticketName, "-ticket") {
				nextIndex++
			}
		}
	}

	// Create ticket directory
	ticketID := fmt.Sprintf("%04d", nextIndex)
	ticketPath := fmt.Sprintf("%s/%s-ticket", ticketsPath, ticketID)

	if err := os.MkdirAll(ticketPath, 0755); err != nil {
		return fmt.Errorf("failed to create ticket directory: %w", err)
	}

	// Create subdirectories
	if err := os.MkdirAll(ticketPath+"/goals", 0755); err != nil {
		return fmt.Errorf("failed to create goals directory: %w", err)
	}

	// Create empty README.md
	if err := createEmptyFile(ticketPath + "/README.md"); err != nil {
		return fmt.Errorf("failed to create README.md: %w", err)
	}

	// Create open status file
	if err := createEmptyFile(ticketPath + "/open"); err != nil {
		return fmt.Errorf("failed to create open status file: %w", err)
	}

	relTicketPath := relPath(projectRoot, ticketPath)
	fmt.Printf("Created ticket %s in sprint %s (phase %s) at %s\n", ticketID, sprintID, phaseID, relTicketPath)
	fmt.Println()
	fmt.Println("Created:")
	fmt.Printf("  %s/\n", relTicketPath)
	fmt.Printf("  %s/open\n", relTicketPath)
	fmt.Printf("  %s/README.md\n", relTicketPath)
	fmt.Printf("  %s/goals/\n", relTicketPath)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  1. Edit %s/README.md to describe the ticket\n", relTicketPath)
	fmt.Printf("  2. Create ticket goals: crumbler ticket goal create %s <goal-name>\n", ticketID)
	fmt.Println("  3. Work on the ticket and close goals as you complete them")
	fmt.Printf("  4. Mark done when complete: crumbler ticket done %s\n", ticketID)

	return nil
}

// runTicketDone handles 'crumbler ticket done <ticket-id>'.
func runTicketDone(args []string) error {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		printTicketDoneHelp()
		return nil
	}

	projectRoot, err := requireProject()
	if err != nil {
		return err
	}

	ticketID := normalizeTicketID(args[0])

	// Find ticket by searching all phases and sprints
	ticketPath, sprintID, phaseID, err := findTicket(projectRoot, ticketID)
	if err != nil {
		return err
	}

	// Check ticket is open
	openPath := ticketPath + "/open"
	if _, err := os.Stat(openPath); os.IsNotExist(err) {
		donePath := ticketPath + "/done"
		if _, err := os.Stat(donePath); err == nil {
			return fmt.Errorf("ticket %s is already done", ticketID)
		}
		return fmt.Errorf("ticket %s has invalid state (no open or done file)", ticketID)
	}

	// Check all ticket goals are closed
	goalsPath := ticketPath + "/goals"
	goals, err := listGoals(goalsPath)
	if err == nil {
		for _, goal := range goals {
			if goal.Status == "open" {
				return fmt.Errorf("cannot mark ticket done: goal %s is still open at %s",
					goal.ID, relPath(projectRoot, goal.Path))
			}
		}
	}

	// Mark the ticket done: remove open, create done
	if err := os.Remove(openPath); err != nil {
		return fmt.Errorf("failed to remove open file: %w", err)
	}
	if err := createEmptyFile(ticketPath + "/done"); err != nil {
		return fmt.Errorf("failed to create done file: %w", err)
	}

	fmt.Printf("Marked ticket %s as done (sprint %s, phase %s)\n", ticketID, sprintID, phaseID)
	return nil
}

// normalizeTicketID strips the -ticket suffix if present, making it easier for agents.
func normalizeTicketID(ticketID string) string {
	return strings.TrimSuffix(ticketID, "-ticket")
}

// normalizeGoalID strips the -goal suffix if present, making it easier for agents.
func normalizeGoalID(goalID string) string {
	return strings.TrimSuffix(goalID, "-goal")
}

// findTicket searches for a ticket by ID across all phases and sprints.
// It prioritizes the current open sprint if one exists.
func findTicket(projectRoot, ticketID string) (string, string, string, error) {
	// Normalize ticket ID (strip -ticket suffix if present)
	ticketID = normalizeTicketID(ticketID)

	// First, try to find the ticket in the current open sprint
	status, err := getProjectStatus(projectRoot)
	if err == nil && status.CurrentSprint != nil {
		ticketPath := fmt.Sprintf("%s/tickets/%s-ticket", status.CurrentSprint.Path, ticketID)
		if _, err := os.Stat(ticketPath); err == nil {
			return ticketPath, status.CurrentSprint.ID, status.CurrentSprint.PhaseID, nil
		}
	}

	// Fall back to searching all phases and sprints, prioritizing the current open phase
	phasesPath := phasesDir(projectRoot)
	phases, err := listDirs(phasesPath)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to list phases: %w", err)
	}

	// First, check the current open phase's sprints if it exists
	if status != nil && status.CurrentPhase != nil {
		currentPhaseName := fmt.Sprintf("%s-phase", status.CurrentPhase.ID)
		sprintsPath := fmt.Sprintf("%s/%s/sprints", phasesPath, currentPhaseName)
		sprints, err := listDirs(sprintsPath)
		if err == nil {
			for _, sprintName := range sprints {
				if !strings.HasSuffix(sprintName, "-sprint") {
					continue
				}

				ticketPath := fmt.Sprintf("%s/%s/tickets/%s-ticket", sprintsPath, sprintName, ticketID)
				if _, err := os.Stat(ticketPath); err == nil {
					sprintID := strings.TrimSuffix(sprintName, "-sprint")
					return ticketPath, sprintID, status.CurrentPhase.ID, nil
				}
			}
		}
	}

	// Then search through all other phases
	for _, phaseName := range phases {
		if !strings.HasSuffix(phaseName, "-phase") {
			continue
		}

		// Skip if we already checked the current phase
		if status != nil && status.CurrentPhase != nil {
			if phaseName == fmt.Sprintf("%s-phase", status.CurrentPhase.ID) {
				continue
			}
		}

		phaseID := strings.TrimSuffix(phaseName, "-phase")
		sprintsPath := fmt.Sprintf("%s/%s/sprints", phasesPath, phaseName)
		sprints, err := listDirs(sprintsPath)
		if err != nil {
			continue
		}

		for _, sprintName := range sprints {
			if !strings.HasSuffix(sprintName, "-sprint") {
				continue
			}

			ticketPath := fmt.Sprintf("%s/%s/tickets/%s-ticket", sprintsPath, sprintName, ticketID)
			if _, err := os.Stat(ticketPath); err == nil {
				sprintID := strings.TrimSuffix(sprintName, "-sprint")
				return ticketPath, sprintID, phaseID, nil
			}
		}
	}

	return "", "", "", fmt.Errorf("ticket %s not found in any sprint", ticketID)
}

// runTicketGoal handles 'crumbler ticket goal' subcommands.
func runTicketGoal(args []string) error {
	if len(args) == 0 {
		printTicketGoalHelp()
		return nil
	}

	if args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		printTicketGoalHelp()
		return nil
	}

	switch args[0] {
	case "create":
		return runTicketGoalCreate(args[1:])
	case "close":
		return runTicketGoalClose(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "error: unknown ticket goal subcommand '%s'\n\n", args[0])
		printTicketGoalHelp()
		return fmt.Errorf("unknown ticket goal subcommand: %s", args[0])
	}
}

// runTicketGoalCreate handles 'crumbler ticket goal create <ticket-id> <goal-name>'.
func runTicketGoalCreate(args []string) error {
	if len(args) < 2 || args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		printTicketGoalCreateHelp()
		return nil
	}

	projectRoot, err := requireProject()
	if err != nil {
		return err
	}

	ticketID := normalizeTicketID(args[0])
	goalName := strings.Join(args[1:], " ")

	// Find ticket
	ticketPath, sprintID, phaseID, err := findTicket(projectRoot, ticketID)
	if err != nil {
		return err
	}

	goalsPath := ticketPath + "/goals"

	// Find next goal number
	nextIndex := 1
	goals, err := listDirs(goalsPath)
	if err == nil {
		for _, goalDir := range goals {
			if strings.HasSuffix(goalDir, "-goal") {
				nextIndex++
			}
		}
	}

	// Create goal directory
	goalID := fmt.Sprintf("%04d", nextIndex)
	goalPath := fmt.Sprintf("%s/%s-goal", goalsPath, goalID)

	if err := os.MkdirAll(goalPath, 0755); err != nil {
		return fmt.Errorf("failed to create goal directory: %w", err)
	}

	// Create name file
	if err := os.WriteFile(goalPath+"/name", []byte(goalName), 0644); err != nil {
		return fmt.Errorf("failed to create name file: %w", err)
	}

	// Create open status file
	if err := createEmptyFile(goalPath + "/open"); err != nil {
		return fmt.Errorf("failed to create open status file: %w", err)
	}

	fmt.Printf("Created ticket goal %s in ticket %s (sprint %s, phase %s): %s\n",
		goalID, ticketID, sprintID, phaseID, goalName)
	fmt.Printf("  Path: %s\n", relPath(projectRoot, goalPath))
	return nil
}

// runTicketGoalClose handles 'crumbler ticket goal close <ticket-id> <goal-id>'.
func runTicketGoalClose(args []string) error {
	if len(args) < 2 || args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		printTicketGoalCloseHelp()
		return nil
	}

	projectRoot, err := requireProject()
	if err != nil {
		return err
	}

	ticketID := normalizeTicketID(args[0])
	goalID := normalizeGoalID(args[1])

	// Find ticket
	ticketPath, sprintID, phaseID, err := findTicket(projectRoot, ticketID)
	if err != nil {
		return err
	}

	goalPath := fmt.Sprintf("%s/goals/%s-goal", ticketPath, goalID)

	// Check goal exists
	if _, err := os.Stat(goalPath); os.IsNotExist(err) {
		return fmt.Errorf("goal %s not found in ticket %s at %s", goalID, ticketID, relPath(projectRoot, goalPath))
	}

	// Check goal is open
	openPath := goalPath + "/open"
	if _, err := os.Stat(openPath); os.IsNotExist(err) {
		closedPath := goalPath + "/closed"
		if _, err := os.Stat(closedPath); err == nil {
			return fmt.Errorf("goal %s is already closed", goalID)
		}
		return fmt.Errorf("goal %s has invalid state (no open or closed file)", goalID)
	}

	// Close the goal: remove open, create closed
	if err := os.Remove(openPath); err != nil {
		return fmt.Errorf("failed to remove open file: %w", err)
	}
	if err := createEmptyFile(goalPath + "/closed"); err != nil {
		return fmt.Errorf("failed to create closed file: %w", err)
	}

	fmt.Printf("Closed goal %s in ticket %s (sprint %s, phase %s)\n", goalID, ticketID, sprintID, phaseID)
	return nil
}

// printTicketHelp prints help for the ticket command.
func printTicketHelp() {
	fmt.Print(`crumbler ticket - Manage tickets

USAGE:
    crumbler ticket <subcommand> [arguments]

SUBCOMMANDS:
    list [sprint-id]                   List tickets (in current or specified sprint)
    create [sprint-id]                 Create a ticket (in current or specified sprint)
    done <ticket-id>                   Mark a ticket as done
    goal create <ticket-id> <name>     Create a goal in a ticket
    goal close <ticket-id> <goal-id>   Close a goal in a ticket

DESCRIPTION:
    Tickets are individual work items within a sprint. Each ticket has a
    description (README.md) and goals that must be completed before the
    ticket can be marked as done.

TICKET DIRECTORY STRUCTURE:
    .../sprints/YYYY-sprint/tickets/ZZZZ-ticket/
    .../sprints/YYYY-sprint/tickets/ZZZZ-ticket/open       Ticket is open
    .../sprints/YYYY-sprint/tickets/ZZZZ-ticket/done       Ticket is done
    .../sprints/YYYY-sprint/tickets/ZZZZ-ticket/README.md  Ticket description
    .../sprints/YYYY-sprint/tickets/ZZZZ-ticket/goals/     Goals directory

STATE TRANSITIONS:
    - Ticket starts as 'open' (open file exists)
    - Ticket can be marked 'done' when all goals are closed
    - Cannot mark ticket done with open goals (error with goal paths)

EXAMPLES:
    crumbler ticket list                          List tickets in current sprint
    crumbler ticket list 0001                     List tickets in sprint 0001
    crumbler ticket create                        Create ticket in current sprint
    crumbler ticket create 0001                   Create ticket in sprint 0001
    crumbler ticket done 0001                     Mark ticket 0001 as done
    crumbler ticket goal create 0001 "Write tests"  Create goal in ticket
    crumbler ticket goal close 0001 0001          Close goal in ticket

For help on a specific subcommand:
    crumbler ticket <subcommand> --help
`)
}

// printTicketListHelp prints help for 'crumbler ticket list'.
func printTicketListHelp() {
	fmt.Print(`crumbler ticket list - List tickets

USAGE:
    crumbler ticket list [sprint-id]

ARGUMENTS:
    sprint-id   Optional. The 4-digit sprint ID. If not specified, uses
                the current open sprint.

DESCRIPTION:
    Lists all tickets in the specified (or current) sprint with their
    status and goals.

OUTPUT FORMAT:
    [ ] 0001  .../tickets/0001-ticket    (open)
        [ ] Goal 0001: Write implementation
        [x] Goal 0002: Write tests
    [x] 0002  .../tickets/0002-ticket    (done)

EXAMPLES:
    crumbler ticket list              List tickets in current sprint
    crumbler ticket list 0001         List tickets in sprint 0001
`)
}

// printTicketCreateHelp prints help for 'crumbler ticket create'.
func printTicketCreateHelp() {
	fmt.Print(`crumbler ticket create - Create a ticket

USAGE:
    crumbler ticket create [sprint-id]

ARGUMENTS:
    sprint-id   Optional. The 4-digit sprint ID. If not specified, uses
                the current open sprint.

DESCRIPTION:
    Creates the next ticket in sequence within the specified sprint.
    Ticket IDs are 4-digit zero-padded numbers (0001, 0002, etc.).

CREATES:
    .../sprints/YYYY-sprint/tickets/ZZZZ-ticket/
    .../sprints/YYYY-sprint/tickets/ZZZZ-ticket/open        Open status
    .../sprints/YYYY-sprint/tickets/ZZZZ-ticket/README.md   Description
    .../sprints/YYYY-sprint/tickets/ZZZZ-ticket/goals/      Goals dir

EXAMPLES:
    crumbler ticket create            Create ticket in current sprint
    crumbler ticket create 0001       Create ticket in sprint 0001

FOR AI AGENTS:
    After creating a ticket, populate the README.md with proper structure:

    1. Write ticket description to README.md:
       cat > .crumbler/phases/0001-phase/sprints/0001-sprint/tickets/0001-ticket/README.md << 'EOF'
       # Ticket 0001: Implement User Registration Endpoint

       ## Description
       Create the POST /api/auth/register endpoint that allows users to
       register with email and password. The endpoint should validate input,
       hash passwords securely, and return a session token.

       ## Acceptance Criteria
       - [ ] Endpoint accepts email and password in request body
       - [ ] Email validation (format, uniqueness)
       - [ ] Password validation (min 8 chars, complexity requirements)
       - [ ] Password is hashed using bcrypt before storage
       - [ ] User record is created in database
       - [ ] Session token is generated and returned
       - [ ] Returns 201 Created on success
       - [ ] Returns 400 Bad Request for validation errors
       - [ ] Returns 409 Conflict for duplicate email

       ## Implementation Notes
       - Use bcrypt with cost factor 10
       - Generate JWT token with 24-hour expiration
       - Include user ID in token payload
       - Validate email format using regex
       - Check email uniqueness before creating user

       ## Testing
       - Unit tests for validation logic
       - Integration tests for endpoint
       - Test password hashing
       - Test duplicate email handling
       EOF

    2. Create ticket goals for tracking work:
       crumbler ticket goal create 0001 "Implement registration endpoint"
       crumbler ticket goal create 0001 "Add input validation"
       crumbler ticket goal create 0001 "Implement password hashing"
       crumbler ticket goal create 0001 "Write unit tests"
       crumbler ticket goal create 0001 "Write integration tests"

    3. Work on the ticket, closing goals as completed:
       crumbler ticket goal close 0001 0001
       crumbler ticket goal close 0001 0002

    4. When all goals are closed, mark ticket done:
       crumbler ticket done 0001

    DOCUMENTATION STRUCTURE:
    Ticket README.md should include:
    - H1 header with ticket number and title
    - Description section (what needs to be done)
    - Acceptance Criteria section (checklist of requirements)
    - Implementation Notes section (technical details, decisions)
    - Testing section (what tests are needed)
    - Any other relevant context (dependencies, references, etc.)
`)
}

// printTicketDoneHelp prints help for 'crumbler ticket done'.
func printTicketDoneHelp() {
	fmt.Print(`crumbler ticket done - Mark a ticket as done

USAGE:
    crumbler ticket done <ticket-id>

ARGUMENTS:
    ticket-id   The 4-digit ticket ID (e.g., 0001)

DESCRIPTION:
    Marks a ticket as done by removing the 'open' file and creating a 'done' file.

    A ticket can only be marked done when:
    - All ticket goals are closed

ERRORS:
    - "ticket not found" - ticket doesn't exist in any sprint
    - "ticket is already done" - ticket already has done file
    - "cannot mark ticket done: goal X is still open" - goals must be closed first

EXAMPLES:
    crumbler ticket done 0001

STATE CHANGES:
    Before: .../tickets/0001-ticket/open exists
    After:  .../tickets/0001-ticket/done exists (open removed)
`)
}

// printTicketGoalHelp prints help for 'crumbler ticket goal'.
func printTicketGoalHelp() {
	fmt.Print(`crumbler ticket goal - Manage ticket goals

USAGE:
    crumbler ticket goal <subcommand> [arguments]

SUBCOMMANDS:
    create <ticket-id> <goal-name>   Create a goal in a ticket
    close <ticket-id> <goal-id>      Close a goal in a ticket

DESCRIPTION:
    Ticket goals track specific tasks that must be completed for the ticket.
    Goals have a name and open/closed status.

GOAL DIRECTORY STRUCTURE:
    .../tickets/ZZZZ-ticket/goals/WWWW-goal/
    .../tickets/ZZZZ-ticket/goals/WWWW-goal/name     Goal name (text)
    .../tickets/ZZZZ-ticket/goals/WWWW-goal/open     Goal is open (empty)
    .../tickets/ZZZZ-ticket/goals/WWWW-goal/closed   Goal is closed (empty)

EXAMPLES:
    crumbler ticket goal create 0001 "Write implementation"
    crumbler ticket goal close 0001 0001
`)
}

// printTicketGoalCreateHelp prints help for 'crumbler ticket goal create'.
func printTicketGoalCreateHelp() {
	fmt.Print(`crumbler ticket goal create - Create a ticket goal

USAGE:
    crumbler ticket goal create <ticket-id> <goal-name>

ARGUMENTS:
    ticket-id   The 4-digit ticket ID (e.g., 0001)
    goal-name   The name of the goal (can contain spaces)

DESCRIPTION:
    Creates a new goal in the specified ticket. Goal IDs are 4-digit
    zero-padded numbers assigned automatically.

CREATES:
    .../tickets/ZZZZ-ticket/goals/WWWW-goal/
    .../tickets/ZZZZ-ticket/goals/WWWW-goal/name    Contains goal name
    .../tickets/ZZZZ-ticket/goals/WWWW-goal/open    Open status (empty)

EXAMPLES:
    crumbler ticket goal create 0001 "Write implementation"
    crumbler ticket goal create 0001 "Write unit tests"
    crumbler ticket goal create 0001 "Update documentation"

FOR AI AGENTS:
    Ticket goal names should be:
    - Specific tasks that can be completed independently
    - Actionable (verb + noun format)
    - Small enough to track incremental progress

    Good examples:
    - "Implement registration endpoint handler"
    - "Add email validation logic"
    - "Write unit tests for registration"
    - "Add error handling for duplicate email"
    - "Update API documentation"

    Goal names are stored in the name file as plain text.
`)
}

// printTicketGoalCloseHelp prints help for 'crumbler ticket goal close'.
func printTicketGoalCloseHelp() {
	fmt.Print(`crumbler ticket goal close - Close a ticket goal

USAGE:
    crumbler ticket goal close <ticket-id> <goal-id>

ARGUMENTS:
    ticket-id   The 4-digit ticket ID (e.g., 0001)
    goal-id     The 4-digit goal ID (e.g., 0001)

DESCRIPTION:
    Closes a goal in the specified ticket.

ERRORS:
    - "goal not found" - goal directory doesn't exist
    - "goal is already closed" - goal already has closed file

EXAMPLES:
    crumbler ticket goal close 0001 0001
    crumbler ticket goal close 0001 0002

STATE CHANGES:
    Before: .../goals/0001-goal/open exists
    After:  .../goals/0001-goal/closed exists (open removed)
`)
}
