package crumbler

import (
	"fmt"
	"os"
	"strings"
)

// runSprint handles the 'crumbler sprint' command and its subcommands.
func runSprint(args []string) error {
	if len(args) == 0 {
		printSprintHelp()
		return nil
	}

	// Handle help flag
	if args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		printSprintHelp()
		return nil
	}

	switch args[0] {
	case "list":
		return runSprintList(args[1:])
	case "create":
		return runSprintCreate(args[1:])
	case "close":
		return runSprintClose(args[1:])
	case "goal":
		return runSprintGoal(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "error: unknown sprint subcommand '%s'\n\n", args[0])
		printSprintHelp()
		return fmt.Errorf("unknown sprint subcommand: %s", args[0])
	}
}

// runSprintList handles 'crumbler sprint list [phase-id]'.
func runSprintList(args []string) error {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help") {
		printSprintListHelp()
		return nil
	}

	projectRoot, err := requireProject()
	if err != nil {
		return err
	}

	// Determine which phase to list sprints for
	var phaseID string
	if len(args) > 0 {
		phaseID = args[0]
	} else {
		// Find current open phase
		status, err := getProjectStatus(projectRoot)
		if err != nil {
			return err
		}
		if status.CurrentPhase == nil {
			return fmt.Errorf("no open phase found. Specify a phase ID: crumbler sprint list <phase-id>")
		}
		phaseID = status.CurrentPhase.ID
	}

	phasePath := fmt.Sprintf("%s/%s-phase", phasesDir(projectRoot), phaseID)
	if _, err := os.Stat(phasePath); os.IsNotExist(err) {
		return fmt.Errorf("phase %s not found at %s", phaseID, relPath(projectRoot, phasePath))
	}

	sprintsPath := phasePath + "/sprints"
	sprints, err := listDirs(sprintsPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("No sprints found in phase %s.\n", phaseID)
			return nil
		}
		return fmt.Errorf("failed to list sprints: %w", err)
	}

	fmt.Printf("Sprints in Phase %s:\n", phaseID)
	fmt.Println()

	count := 0
	for _, sprintName := range sprints {
		if !strings.HasSuffix(sprintName, "-sprint") {
			continue
		}

		count++
		sprintPath := sprintsPath + "/" + sprintName
		status := getEntityStatus(sprintPath, "open", "closed")
		sprintID := strings.TrimSuffix(sprintName, "-sprint")

		marker := "[ ]"
		if status == "closed" {
			marker = "[x]"
		}

		fmt.Printf("  %s %s  %s\n", marker, sprintID, relPath(projectRoot, sprintPath))

		// Show goals
		goals, _ := listGoals(sprintPath + "/goals")
		for _, goal := range goals {
			gMarker := "[ ]"
			if goal.Status == "closed" {
				gMarker = "[x]"
			}
			fmt.Printf("      %s Goal %s: %s\n", gMarker, goal.ID, goal.Name)
		}
	}

	if count == 0 {
		fmt.Println("  No sprints found.")
		fmt.Println()
		fmt.Printf("Create a sprint with: crumbler sprint create %s\n", phaseID)
	}

	return nil
}

// runSprintCreate handles 'crumbler sprint create [phase-id]'.
func runSprintCreate(args []string) error {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help") {
		printSprintCreateHelp()
		return nil
	}

	projectRoot, err := requireProject()
	if err != nil {
		return err
	}

	// Determine which phase to create sprint in
	var phaseID string
	if len(args) > 0 {
		phaseID = args[0]
	} else {
		// Find current open phase
		status, err := getProjectStatus(projectRoot)
		if err != nil {
			return err
		}
		if status.CurrentPhase == nil {
			return fmt.Errorf("no open phase found. Specify a phase ID: crumbler sprint create <phase-id>")
		}
		phaseID = status.CurrentPhase.ID
	}

	phasePath := fmt.Sprintf("%s/%s-phase", phasesDir(projectRoot), phaseID)
	if _, err := os.Stat(phasePath); os.IsNotExist(err) {
		return fmt.Errorf("phase %s not found at %s", phaseID, relPath(projectRoot, phasePath))
	}

	// Find next sprint number
	sprintsPath := phasePath + "/sprints"
	nextIndex := 1

	sprints, err := listDirs(sprintsPath)
	if err == nil {
		for _, sprintName := range sprints {
			if strings.HasSuffix(sprintName, "-sprint") {
				nextIndex++
			}
		}
	}

	// Create sprint directory
	sprintID := fmt.Sprintf("%04d", nextIndex)
	sprintPath := fmt.Sprintf("%s/%s-sprint", sprintsPath, sprintID)

	if err := os.MkdirAll(sprintPath, 0755); err != nil {
		return fmt.Errorf("failed to create sprint directory: %w", err)
	}

	// Create subdirectories
	if err := os.MkdirAll(sprintPath+"/tickets", 0755); err != nil {
		return fmt.Errorf("failed to create tickets directory: %w", err)
	}
	if err := os.MkdirAll(sprintPath+"/goals", 0755); err != nil {
		return fmt.Errorf("failed to create goals directory: %w", err)
	}

	// Create empty files
	if err := createEmptyFile(sprintPath + "/README.md"); err != nil {
		return fmt.Errorf("failed to create README.md: %w", err)
	}
	if err := createEmptyFile(sprintPath + "/PRD.md"); err != nil {
		return fmt.Errorf("failed to create PRD.md: %w", err)
	}
	if err := createEmptyFile(sprintPath + "/ERD.md"); err != nil {
		return fmt.Errorf("failed to create ERD.md: %w", err)
	}

	// Create open status file
	if err := createEmptyFile(sprintPath + "/open"); err != nil {
		return fmt.Errorf("failed to create open status file: %w", err)
	}

	relSprintPath := relPath(projectRoot, sprintPath)
	fmt.Printf("Created sprint %s in phase %s at %s\n", sprintID, phaseID, relSprintPath)
	fmt.Println()
	fmt.Println("Created:")
	fmt.Printf("  %s/\n", relSprintPath)
	fmt.Printf("  %s/open\n", relSprintPath)
	fmt.Printf("  %s/README.md\n", relSprintPath)
	fmt.Printf("  %s/PRD.md\n", relSprintPath)
	fmt.Printf("  %s/ERD.md\n", relSprintPath)
	fmt.Printf("  %s/tickets/\n", relSprintPath)
	fmt.Printf("  %s/goals/\n", relSprintPath)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  1. Edit %s/README.md to describe the sprint\n", relSprintPath)
	fmt.Printf("  2. Edit %s/PRD.md with Product Requirements Document\n", relSprintPath)
	fmt.Printf("  3. Edit %s/ERD.md with Entity Relationship Diagram\n", relSprintPath)
	fmt.Printf("  4. Create sprint goals: crumbler sprint goal create %s <goal-name>\n", sprintID)
	fmt.Printf("  5. Create tickets: crumbler ticket create %s\n", sprintID)

	return nil
}

// runSprintClose handles 'crumbler sprint close <sprint-id>'.
func runSprintClose(args []string) error {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		printSprintCloseHelp()
		return nil
	}

	projectRoot, err := requireProject()
	if err != nil {
		return err
	}

	sprintID := normalizeSprintID(args[0])

	// Find sprint by searching all phases
	sprintPath, phaseID, err := findSprint(projectRoot, sprintID)
	if err != nil {
		return err
	}

	// Check sprint is open
	openPath := sprintPath + "/open"
	if _, err := os.Stat(openPath); os.IsNotExist(err) {
		closedPath := sprintPath + "/closed"
		if _, err := os.Stat(closedPath); err == nil {
			return fmt.Errorf("sprint %s is already closed", sprintID)
		}
		return fmt.Errorf("sprint %s has invalid state (no open or closed file)", sprintID)
	}

	// Check all tickets are done
	ticketsPath := sprintPath + "/tickets"
	tickets, err := listDirs(ticketsPath)
	if err == nil {
		for _, ticketName := range tickets {
			if !strings.HasSuffix(ticketName, "-ticket") {
				continue
			}
			ticketPath := ticketsPath + "/" + ticketName
			if _, err := os.Stat(ticketPath + "/open"); err == nil {
				return fmt.Errorf("cannot close sprint: ticket %s is still open at %s",
					strings.TrimSuffix(ticketName, "-ticket"),
					relPath(projectRoot, ticketPath))
			}
		}
	}

	// Check all sprint goals are closed
	goalsPath := sprintPath + "/goals"
	goals, err := listGoals(goalsPath)
	if err == nil {
		for _, goal := range goals {
			if goal.Status == "open" {
				return fmt.Errorf("cannot close sprint: goal %s is still open at %s",
					goal.ID, relPath(projectRoot, goal.Path))
			}
		}
	}

	// Close the sprint: remove open, create closed
	if err := os.Remove(openPath); err != nil {
		return fmt.Errorf("failed to remove open file: %w", err)
	}
	if err := createEmptyFile(sprintPath + "/closed"); err != nil {
		return fmt.Errorf("failed to create closed file: %w", err)
	}

	fmt.Printf("Closed sprint %s in phase %s\n", sprintID, phaseID)
	return nil
}

// normalizeSprintID strips the -sprint suffix if present, making it easier for agents.
func normalizeSprintID(sprintID string) string {
	return strings.TrimSuffix(sprintID, "-sprint")
}

// findSprint searches for a sprint by ID, checking the current sprint first.
// It prioritizes the current open sprint if one exists.
func findSprint(projectRoot, sprintID string) (string, string, error) {
	// Normalize sprint ID (strip -sprint suffix if present)
	sprintID = normalizeSprintID(sprintID)

	// First, check if there's a current sprint with matching ID
	status, err := getProjectStatus(projectRoot)
	if err == nil && status.CurrentSprint != nil && status.CurrentSprint.ID == sprintID {
		return status.CurrentSprint.Path, status.CurrentSprint.PhaseID, nil
	}

	// Fall back to searching all phases, prioritizing the current open phase
	phasesPath := phasesDir(projectRoot)
	phases, err := listDirs(phasesPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to list phases: %w", err)
	}

	// First, check the current open phase if it exists
	if status != nil && status.CurrentPhase != nil {
		currentPhaseName := fmt.Sprintf("%s-phase", status.CurrentPhase.ID)
		sprintPath := fmt.Sprintf("%s/%s/sprints/%s-sprint", phasesPath, currentPhaseName, sprintID)
		if _, err := os.Stat(sprintPath); err == nil {
			return sprintPath, status.CurrentPhase.ID, nil
		}
	}

	// Then search through all phases
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

		sprintPath := fmt.Sprintf("%s/%s/sprints/%s-sprint", phasesPath, phaseName, sprintID)
		if _, err := os.Stat(sprintPath); err == nil {
			phaseID := strings.TrimSuffix(phaseName, "-phase")
			return sprintPath, phaseID, nil
		}
	}

	return "", "", fmt.Errorf("sprint %s not found in any phase", sprintID)
}

// runSprintGoal handles 'crumbler sprint goal' subcommands.
func runSprintGoal(args []string) error {
	if len(args) == 0 {
		printSprintGoalHelp()
		return nil
	}

	if args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		printSprintGoalHelp()
		return nil
	}

	switch args[0] {
	case "create":
		return runSprintGoalCreate(args[1:])
	case "close":
		return runSprintGoalClose(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "error: unknown sprint goal subcommand '%s'\n\n", args[0])
		printSprintGoalHelp()
		return fmt.Errorf("unknown sprint goal subcommand: %s", args[0])
	}
}

// runSprintGoalCreate handles 'crumbler sprint goal create <sprint-id> <goal-name>'.
func runSprintGoalCreate(args []string) error {
	if len(args) < 2 || args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		printSprintGoalCreateHelp()
		return nil
	}

	projectRoot, err := requireProject()
	if err != nil {
		return err
	}

	sprintID := normalizeSprintID(args[0])
	goalName := strings.Join(args[1:], " ")

	// Find sprint
	sprintPath, phaseID, err := findSprint(projectRoot, sprintID)
	if err != nil {
		return err
	}

	goalsPath := sprintPath + "/goals"

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

	fmt.Printf("Created sprint goal %s in sprint %s (phase %s): %s\n", goalID, sprintID, phaseID, goalName)
	fmt.Printf("  Path: %s\n", relPath(projectRoot, goalPath))
	return nil
}

// runSprintGoalClose handles 'crumbler sprint goal close <sprint-id> <goal-id>'.
func runSprintGoalClose(args []string) error {
	if len(args) < 2 || args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		printSprintGoalCloseHelp()
		return nil
	}

	projectRoot, err := requireProject()
	if err != nil {
		return err
	}

	sprintID := normalizeSprintID(args[0])
	goalID := normalizeGoalID(args[1])

	// Find sprint
	sprintPath, phaseID, err := findSprint(projectRoot, sprintID)
	if err != nil {
		return err
	}

	goalPath := fmt.Sprintf("%s/goals/%s-goal", sprintPath, goalID)

	// Check goal exists
	if _, err := os.Stat(goalPath); os.IsNotExist(err) {
		return fmt.Errorf("goal %s not found in sprint %s at %s", goalID, sprintID, relPath(projectRoot, goalPath))
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

	fmt.Printf("Closed goal %s in sprint %s (phase %s)\n", goalID, sprintID, phaseID)
	return nil
}

// printSprintHelp prints help for the sprint command.
func printSprintHelp() {
	fmt.Print(`crumbler sprint - Manage sprints

USAGE:
    crumbler sprint <subcommand> [arguments]

SUBCOMMANDS:
    list [phase-id]                   List sprints (in current or specified phase)
    create [phase-id]                 Create a sprint (in current or specified phase)
    close <sprint-id>                 Close a sprint
    goal create <sprint-id> <name>    Create a goal in a sprint (uses current sprint's phase if ID matches)
    goal close <sprint-id> <goal-id>  Close a goal in a sprint (uses current sprint's phase if ID matches)

DESCRIPTION:
    Sprints are time-boxed iterations within a phase. Each sprint contains
    tickets (work items) and has goals that must be met before closing.

SPRINT DIRECTORY STRUCTURE:
    .crumbler/phases/XXXX-phase/sprints/YYYY-sprint/
    .crumbler/phases/XXXX-phase/sprints/YYYY-sprint/open       Sprint is open
    .crumbler/phases/XXXX-phase/sprints/YYYY-sprint/closed     Sprint is closed
    .crumbler/phases/XXXX-phase/sprints/YYYY-sprint/README.md  Sprint description
    .crumbler/phases/XXXX-phase/sprints/YYYY-sprint/PRD.md     Product Requirements
    .crumbler/phases/XXXX-phase/sprints/YYYY-sprint/ERD.md     Entity Relationships
    .crumbler/phases/XXXX-phase/sprints/YYYY-sprint/tickets/   Tickets directory
    .crumbler/phases/XXXX-phase/sprints/YYYY-sprint/goals/     Goals directory

STATE TRANSITIONS:
    - Sprint starts as 'open' (open file exists)
    - Sprint can be closed when all tickets are done AND all goals are closed
    - Cannot close sprint with open tickets (error with ticket paths)
    - Cannot close sprint with open goals (error with goal paths)

EXAMPLES:
    crumbler sprint list                         List sprints in current phase
    crumbler sprint list 0001                    List sprints in phase 0001
    crumbler sprint create                       Create sprint in current phase
    crumbler sprint create 0001                  Create sprint in phase 0001
    crumbler sprint close 0001                   Close sprint 0001
    crumbler sprint goal create 0001 "Implement API"  Create goal in sprint
    crumbler sprint goal close 0001 0001         Close goal in sprint

For help on a specific subcommand:
    crumbler sprint <subcommand> --help
`)
}

// printSprintListHelp prints help for 'crumbler sprint list'.
func printSprintListHelp() {
	fmt.Print(`crumbler sprint list - List sprints

USAGE:
    crumbler sprint list [phase-id]

ARGUMENTS:
    phase-id    Optional. The 4-digit phase ID. If not specified, uses
                the current open phase.

DESCRIPTION:
    Lists all sprints in the specified (or current) phase with their
    status and goals.

OUTPUT FORMAT:
    [ ] 0001  .crumbler/phases/0001-phase/sprints/0001-sprint    (open)
        [ ] Goal 0001: Implement feature X
        [x] Goal 0002: Write unit tests
    [x] 0002  .crumbler/phases/0001-phase/sprints/0002-sprint    (closed)

EXAMPLES:
    crumbler sprint list              List sprints in current phase
    crumbler sprint list 0001         List sprints in phase 0001
`)
}

// printSprintCreateHelp prints help for 'crumbler sprint create'.
func printSprintCreateHelp() {
	fmt.Print(`crumbler sprint create - Create a sprint

USAGE:
    crumbler sprint create [phase-id]

ARGUMENTS:
    phase-id    Optional. The 4-digit phase ID. If not specified, uses
                the current open phase.

DESCRIPTION:
    Creates the next sprint in sequence within the specified phase.
    Sprint IDs are 4-digit zero-padded numbers (0001, 0002, etc.).

CREATES:
    .crumbler/phases/XXXX-phase/sprints/YYYY-sprint/
    .crumbler/phases/XXXX-phase/sprints/YYYY-sprint/open         Open status
    .crumbler/phases/XXXX-phase/sprints/YYYY-sprint/README.md    Description
    .crumbler/phases/XXXX-phase/sprints/YYYY-sprint/PRD.md       Requirements
    .crumbler/phases/XXXX-phase/sprints/YYYY-sprint/ERD.md       Entity diagram
    .crumbler/phases/XXXX-phase/sprints/YYYY-sprint/tickets/     Tickets dir
    .crumbler/phases/XXXX-phase/sprints/YYYY-sprint/goals/       Goals dir

EXAMPLES:
    crumbler sprint create            Create sprint in current phase
    crumbler sprint create 0001       Create sprint in phase 0001

FOR AI AGENTS:
    After creating a sprint, populate all three documentation files:

    1. Write sprint description to README.md:
       cat > .crumbler/phases/0001-phase/sprints/0001-sprint/README.md << 'EOF'
       # Sprint 1: User Authentication

       ## Sprint Goal
       Implement basic user authentication flow including registration,
       login, and session management.

       ## Scope
       - User registration endpoint
       - User login endpoint
       - Session token generation and validation
       - Password hashing and security

       ## Success Criteria
       - Users can register with email/password
       - Users can login and receive session tokens
       - Tokens expire after 24 hours
       - Passwords are securely hashed
       EOF

    2. Write Product Requirements Document to PRD.md:
       cat > .crumbler/phases/0001-phase/sprints/0001-sprint/PRD.md << 'EOF'
       # Product Requirements Document: User Authentication

       ## Overview
       This sprint implements the core user authentication system.

       ## Functional Requirements

       ### FR1: User Registration
       - User can register with email and password
       - Email must be unique and valid format
       - Password must meet security requirements (min 8 chars, mixed case, numbers)
       - Registration creates user account and returns session token

       ### FR2: User Login
       - User can login with email/password
       - Invalid credentials return error
       - Successful login returns session token

       ### FR3: Session Management
       - Session tokens expire after 24 hours
       - Tokens are JWT format
       - Tokens include user ID and expiration time

       ## Non-Functional Requirements
       - Passwords must be hashed using bcrypt
       - API responses must be under 200ms
       - System must handle 1000 concurrent logins
       EOF

    3. Write Entity Relationship Diagram to ERD.md:
       cat > .crumbler/phases/0001-phase/sprints/0001-sprint/ERD.md << 'EOF'
       # Entity Relationship Diagram: User Authentication

       ## Entities

       ### User
       - id (primary key)
       - email (unique, indexed)
       - password_hash
       - created_at
       - updated_at

       ### Session
       - id (primary key)
       - user_id (foreign key -> User.id)
       - token (unique, indexed)
       - expires_at
       - created_at

       ## Relationships
       - User has_many Sessions
       - Session belongs_to User
       EOF

    4. Create sprint goals:
       crumbler sprint goal create 0001 "Implement registration endpoint"
       crumbler sprint goal create 0001 "Implement login endpoint"
       crumbler sprint goal create 0001 "Add session token management"

    5. Decompose sprint into tickets:
       crumbler ticket create 0001
       crumbler ticket create 0001

    DOCUMENTATION STRUCTURE:
    - README.md: Sprint overview, goal, scope, success criteria
    - PRD.md: Detailed functional/non-functional requirements
    - ERD.md: Database schema, entities, relationships
`)
}

// printSprintCloseHelp prints help for 'crumbler sprint close'.
func printSprintCloseHelp() {
	fmt.Print(`crumbler sprint close - Close a sprint

USAGE:
    crumbler sprint close <sprint-id>

ARGUMENTS:
    sprint-id   The 4-digit sprint ID (e.g., 0001)

DESCRIPTION:
    Closes a sprint by removing the 'open' file and creating a 'closed' file.

    A sprint can only be closed when:
    - All tickets in the sprint are done
    - All sprint goals are closed

ERRORS:
    - "sprint not found" - sprint directory doesn't exist in any phase
    - "sprint is already closed" - sprint already has closed file
    - "cannot close sprint: ticket X is still open" - tickets must be done first
    - "cannot close sprint: goal X is still open" - goals must be closed first

EXAMPLES:
    crumbler sprint close 0001

STATE CHANGES:
    Before: .../sprints/0001-sprint/open exists
    After:  .../sprints/0001-sprint/closed exists (open removed)
`)
}

// printSprintGoalHelp prints help for 'crumbler sprint goal'.
func printSprintGoalHelp() {
	fmt.Print(`crumbler sprint goal - Manage sprint goals

USAGE:
    crumbler sprint goal <subcommand> [arguments]

SUBCOMMANDS:
    create <sprint-id> <goal-name>   Create a goal in a sprint
    close <sprint-id> <goal-id>      Close a goal in a sprint

DESCRIPTION:
    Sprint goals track objectives that must be completed within the sprint.
    Goals have a name and open/closed status.

    When specifying a sprint ID, if that sprint exists in the current open
    sprint's phase, that sprint will be used. Otherwise, the command searches
    all phases for a sprint with the matching ID.

GOAL DIRECTORY STRUCTURE:
    .../sprints/YYYY-sprint/goals/ZZZZ-goal/
    .../sprints/YYYY-sprint/goals/ZZZZ-goal/name     Goal name (text)
    .../sprints/YYYY-sprint/goals/ZZZZ-goal/open     Goal is open (empty)
    .../sprints/YYYY-sprint/goals/ZZZZ-goal/closed   Goal is closed (empty)

EXAMPLES:
    crumbler sprint goal create 0001 "Implement login API"
    crumbler sprint goal close 0001 0001
`)
}

// printSprintGoalCreateHelp prints help for 'crumbler sprint goal create'.
func printSprintGoalCreateHelp() {
	fmt.Print(`crumbler sprint goal create - Create a sprint goal

USAGE:
    crumbler sprint goal create <sprint-id> <goal-name>

ARGUMENTS:
    sprint-id   The 4-digit sprint ID (e.g., 0001)
    goal-name   The name of the goal (can contain spaces)

DESCRIPTION:
    Creates a new goal in the specified sprint. Goal IDs are 4-digit
    zero-padded numbers assigned automatically.

    If a sprint with the given ID exists in the current open sprint's phase,
    that sprint will be used. Otherwise, the command searches all phases
    for a sprint with the matching ID.

CREATES:
    .../sprints/YYYY-sprint/goals/ZZZZ-goal/
    .../sprints/YYYY-sprint/goals/ZZZZ-goal/name    Contains goal name
    .../sprints/YYYY-sprint/goals/ZZZZ-goal/open    Open status (empty)

EXAMPLES:
    crumbler sprint goal create 0001 "Implement login API"
    crumbler sprint goal create 0001 "Write unit tests for auth"

FOR AI AGENTS:
    Goal names should be:
    - Clear and actionable (verb + noun format)
    - Specific enough to track progress
    - Aligned with sprint objectives

    Good examples:
    - "Implement user registration endpoint"
    - "Add password validation logic"
    - "Create session token generation"
    - "Write integration tests for auth flow"

    Avoid vague names:
    - "Work on authentication" (too broad)
    - "Fix bugs" (not specific)
    - "Testing" (not actionable)
`)
}

// printSprintGoalCloseHelp prints help for 'crumbler sprint goal close'.
func printSprintGoalCloseHelp() {
	fmt.Print(`crumbler sprint goal close - Close a sprint goal

USAGE:
    crumbler sprint goal close <sprint-id> <goal-id>

ARGUMENTS:
    sprint-id   The 4-digit sprint ID (e.g., 0001)
    goal-id     The 4-digit goal ID (e.g., 0001)

DESCRIPTION:
    Closes a goal in the specified sprint.

    If a sprint with the given ID exists in the current open sprint's phase,
    that sprint will be used. Otherwise, the command searches all phases
    for a sprint with the matching ID.

ERRORS:
    - "goal not found" - goal directory doesn't exist
    - "goal is already closed" - goal already has closed file

EXAMPLES:
    crumbler sprint goal close 0001 0001
    crumbler sprint goal close 0001 0002

STATE CHANGES:
    Before: .../goals/0001-goal/open exists
    After:  .../goals/0001-goal/closed exists (open removed)
`)
}
