package crumbler

import (
	"fmt"
	"os"
	"strings"
)

// runPhase handles the 'crumbler phase' command and its subcommands.
func runPhase(args []string) error {
	if len(args) == 0 {
		printPhaseHelp()
		return nil
	}

	// Handle help flag
	if args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		printPhaseHelp()
		return nil
	}

	switch args[0] {
	case "list":
		return runPhaseList(args[1:])
	case "create":
		return runPhaseCreate(args[1:])
	case "close":
		return runPhaseClose(args[1:])
	case "goal":
		return runPhaseGoal(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "error: unknown phase subcommand '%s'\n\n", args[0])
		printPhaseHelp()
		return fmt.Errorf("unknown phase subcommand: %s", args[0])
	}
}

// runPhaseList handles 'crumbler phase list'.
func runPhaseList(args []string) error {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help") {
		printPhaseListHelp()
		return nil
	}

	projectRoot, err := requireProject()
	if err != nil {
		return err
	}

	phasesPath := phasesDir(projectRoot)
	phases, err := listDirs(phasesPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No phases found.")
			return nil
		}
		return fmt.Errorf("failed to list phases: %w", err)
	}

	fmt.Println("Phases:")
	fmt.Println()

	count := 0
	for _, phaseName := range phases {
		if !strings.HasSuffix(phaseName, "-phase") {
			continue
		}

		count++
		phasePath := phasesPath + "/" + phaseName
		status := getEntityStatus(phasePath, "open", "closed")
		phaseID := strings.TrimSuffix(phaseName, "-phase")

		marker := "[ ]"
		if status == "closed" {
			marker = "[x]"
		}

		fmt.Printf("  %s %s  %s\n", marker, phaseID, relPath(projectRoot, phasePath))

		// Show goals
		goals, _ := listGoals(phasePath + "/goals")
		for _, goal := range goals {
			gMarker := "[ ]"
			if goal.Status == "closed" {
				gMarker = "[x]"
			}
			fmt.Printf("      %s Goal %s: %s\n", gMarker, goal.ID, goal.Name)
		}
	}

	if count == 0 {
		fmt.Println("  No phases found.")
		fmt.Println()
		fmt.Println("Create a phase with: crumbler phase create")
	}

	return nil
}

// runPhaseCreate handles 'crumbler phase create'.
func runPhaseCreate(args []string) error {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help") {
		printPhaseCreateHelp()
		return nil
	}

	projectRoot, err := requireProject()
	if err != nil {
		return err
	}

	// Find next phase number
	phasesPath := phasesDir(projectRoot)
	nextIndex := 1

	phases, err := listDirs(phasesPath)
	if err == nil {
		for _, phaseName := range phases {
			if strings.HasSuffix(phaseName, "-phase") {
				nextIndex++
			}
		}
	}

	// Create phase directory
	phaseID := fmt.Sprintf("%04d", nextIndex)
	phasePath := fmt.Sprintf("%s/%s-phase", phasesPath, phaseID)

	if err := os.MkdirAll(phasePath, 0755); err != nil {
		return fmt.Errorf("failed to create phase directory: %w", err)
	}

	// Create subdirectories
	if err := os.MkdirAll(phasePath+"/sprints", 0755); err != nil {
		return fmt.Errorf("failed to create sprints directory: %w", err)
	}
	if err := os.MkdirAll(phasePath+"/goals", 0755); err != nil {
		return fmt.Errorf("failed to create goals directory: %w", err)
	}

	// Create empty README.md
	if err := createEmptyFile(phasePath + "/README.md"); err != nil {
		return fmt.Errorf("failed to create README.md: %w", err)
	}

	// Create open status file
	if err := createEmptyFile(phasePath + "/open"); err != nil {
		return fmt.Errorf("failed to create open status file: %w", err)
	}

	relPhasePath := relPath(projectRoot, phasePath)
	fmt.Printf("Created phase %s at %s\n", phaseID, relPhasePath)
	fmt.Println()
	fmt.Println("Created:")
	fmt.Printf("  %s/\n", relPhasePath)
	fmt.Printf("  %s/open\n", relPhasePath)
	fmt.Printf("  %s/README.md\n", relPhasePath)
	fmt.Printf("  %s/sprints/\n", relPhasePath)
	fmt.Printf("  %s/goals/\n", relPhasePath)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  1. Edit %s/README.md to describe the phase\n", relPhasePath)
	fmt.Println("  2. Create phase goals: crumbler phase goal create", phaseID, "<goal-name>")
	fmt.Println("  3. Create a sprint: crumbler sprint create", phaseID)

	return nil
}

// normalizePhaseID strips the -phase suffix if present, making it easier for agents.
func normalizePhaseID(phaseID string) string {
	return strings.TrimSuffix(phaseID, "-phase")
}

// runPhaseClose handles 'crumbler phase close <phase-id>'.
func runPhaseClose(args []string) error {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		printPhaseCloseHelp()
		return nil
	}

	projectRoot, err := requireProject()
	if err != nil {
		return err
	}

	phaseID := normalizePhaseID(args[0])
	phasePath := fmt.Sprintf("%s/%s-phase", phasesDir(projectRoot), phaseID)

	// Check phase exists
	if _, err := os.Stat(phasePath); os.IsNotExist(err) {
		return fmt.Errorf("phase %s not found at %s", phaseID, relPath(projectRoot, phasePath))
	}

	// Check phase is open
	openPath := phasePath + "/open"
	if _, err := os.Stat(openPath); os.IsNotExist(err) {
		closedPath := phasePath + "/closed"
		if _, err := os.Stat(closedPath); err == nil {
			return fmt.Errorf("phase %s is already closed", phaseID)
		}
		return fmt.Errorf("phase %s has invalid state (no open or closed file)", phaseID)
	}

	// Check all sprints are closed
	sprintsPath := phasePath + "/sprints"
	sprints, err := listDirs(sprintsPath)
	if err == nil {
		for _, sprintName := range sprints {
			if !strings.HasSuffix(sprintName, "-sprint") {
				continue
			}
			sprintPath := sprintsPath + "/" + sprintName
			if _, err := os.Stat(sprintPath + "/open"); err == nil {
				return fmt.Errorf("cannot close phase: sprint %s is still open at %s",
					strings.TrimSuffix(sprintName, "-sprint"),
					relPath(projectRoot, sprintPath))
			}
		}
	}

	// Check all phase goals are closed
	goalsPath := phasePath + "/goals"
	goals, err := listGoals(goalsPath)
	if err == nil {
		for _, goal := range goals {
			if goal.Status == "open" {
				return fmt.Errorf("cannot close phase: goal %s is still open at %s",
					goal.ID, relPath(projectRoot, goal.Path))
			}
		}
	}

	// Close the phase: remove open, create closed
	if err := os.Remove(openPath); err != nil {
		return fmt.Errorf("failed to remove open file: %w", err)
	}
	if err := createEmptyFile(phasePath + "/closed"); err != nil {
		return fmt.Errorf("failed to create closed file: %w", err)
	}

	fmt.Printf("Closed phase %s\n", phaseID)
	return nil
}

// runPhaseGoal handles 'crumbler phase goal' subcommands.
func runPhaseGoal(args []string) error {
	if len(args) == 0 {
		printPhaseGoalHelp()
		return nil
	}

	if args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		printPhaseGoalHelp()
		return nil
	}

	switch args[0] {
	case "create":
		return runPhaseGoalCreate(args[1:])
	case "close":
		return runPhaseGoalClose(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "error: unknown phase goal subcommand '%s'\n\n", args[0])
		printPhaseGoalHelp()
		return fmt.Errorf("unknown phase goal subcommand: %s", args[0])
	}
}

// runPhaseGoalCreate handles 'crumbler phase goal create <phase-id> <goal-name>'.
func runPhaseGoalCreate(args []string) error {
	if len(args) < 2 || args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		printPhaseGoalCreateHelp()
		return nil
	}

	projectRoot, err := requireProject()
	if err != nil {
		return err
	}

	phaseID := normalizePhaseID(args[0])
	goalName := strings.Join(args[1:], " ")

	phasePath := fmt.Sprintf("%s/%s-phase", phasesDir(projectRoot), phaseID)
	goalsPath := phasePath + "/goals"

	// Check phase exists
	if _, err := os.Stat(phasePath); os.IsNotExist(err) {
		return fmt.Errorf("phase %s not found at %s", phaseID, relPath(projectRoot, phasePath))
	}

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

	fmt.Printf("Created phase goal %s in phase %s: %s\n", goalID, phaseID, goalName)
	fmt.Printf("  Path: %s\n", relPath(projectRoot, goalPath))
	return nil
}

// runPhaseGoalClose handles 'crumbler phase goal close <phase-id> <goal-id>'.
func runPhaseGoalClose(args []string) error {
	if len(args) < 2 || args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		printPhaseGoalCloseHelp()
		return nil
	}

	projectRoot, err := requireProject()
	if err != nil {
		return err
	}

	phaseID := normalizePhaseID(args[0])
	goalID := normalizeGoalID(args[1])

	phasePath := fmt.Sprintf("%s/%s-phase", phasesDir(projectRoot), phaseID)
	goalPath := fmt.Sprintf("%s/goals/%s-goal", phasePath, goalID)

	// Check goal exists
	if _, err := os.Stat(goalPath); os.IsNotExist(err) {
		return fmt.Errorf("goal %s not found in phase %s at %s", goalID, phaseID, relPath(projectRoot, goalPath))
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

	fmt.Printf("Closed goal %s in phase %s\n", goalID, phaseID)
	return nil
}

// createEmptyFile creates an empty file.
func createEmptyFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	return file.Close()
}

// printPhaseHelp prints help for the phase command.
func printPhaseHelp() {
	fmt.Print(`crumbler phase - Manage phases

USAGE:
    crumbler phase <subcommand> [arguments]

SUBCOMMANDS:
    list                            List all phases
    create                          Create the next phase
    close <phase-id>                Close a phase
    goal create <phase-id> <name>   Create a goal in a phase
    goal close <phase-id> <goal-id> Close a goal in a phase

DESCRIPTION:
    Phases are the top-level organizational unit in crumbler. Each phase
    contains sprints, which contain tickets. Phases have goals that must
    be met before the phase can be closed.

PHASE DIRECTORY STRUCTURE:
    .crumbler/phases/XXXX-phase/
    .crumbler/phases/XXXX-phase/open           Phase is open (empty file)
    .crumbler/phases/XXXX-phase/closed         Phase is closed (empty file)
    .crumbler/phases/XXXX-phase/README.md      Phase description (AI populates)
    .crumbler/phases/XXXX-phase/goals/         Phase goals directory
    .crumbler/phases/XXXX-phase/sprints/       Sprints in this phase

STATE TRANSITIONS:
    - Phase starts as 'open' (open file exists)
    - Phase can be closed when all sprints are closed AND all goals are closed
    - Cannot close phase with open sprints (error with sprint paths)
    - Cannot close phase with open goals (error with goal paths)

EXAMPLES:
    crumbler phase list                         List all phases
    crumbler phase create                       Create next phase (0001, 0002, etc.)
    crumbler phase close 0001                   Close phase 0001
    crumbler phase goal create 0001 "Set up CI" Create goal in phase 0001
    crumbler phase goal close 0001 0001         Close goal 0001 in phase 0001

For help on a specific subcommand:
    crumbler phase <subcommand> --help
`)
}

// printPhaseListHelp prints help for 'crumbler phase list'.
func printPhaseListHelp() {
	fmt.Print(`crumbler phase list - List all phases

USAGE:
    crumbler phase list

DESCRIPTION:
    Lists all phases in the project with their status and goals.
    Phases are shown with checkmarks indicating completion status.

OUTPUT FORMAT:
    [ ] 0001  .crumbler/phases/0001-phase    (open phase)
        [ ] Goal 0001: Goal name here
        [x] Goal 0002: Another goal (closed)
    [x] 0002  .crumbler/phases/0002-phase    (closed phase)

EXAMPLES:
    crumbler phase list
`)
}

// printPhaseCreateHelp prints help for 'crumbler phase create'.
func printPhaseCreateHelp() {
	fmt.Print(`crumbler phase create - Create the next phase

USAGE:
    crumbler phase create

DESCRIPTION:
    Creates the next phase in sequence. Phase IDs are 4-digit zero-padded
    numbers (0001, 0002, 0003, etc.).

CREATES:
    .crumbler/phases/XXXX-phase/            Phase directory
    .crumbler/phases/XXXX-phase/open        Open status file (empty)
    .crumbler/phases/XXXX-phase/README.md   Phase description (empty, AI populates)
    .crumbler/phases/XXXX-phase/sprints/    Sprints directory
    .crumbler/phases/XXXX-phase/goals/      Goals directory

EXAMPLES:
    crumbler phase create

FOR AI AGENTS:
    After creating a phase, populate the README.md with proper structure:

    1. Write phase description to README.md:
       cat > .crumbler/phases/0001-phase/README.md << 'EOF'
       # Phase 1: Foundation

       ## Overview
       Set up the basic project infrastructure and establish development workflows.

       ## Objectives
       - Initialize project structure and directory layout
       - Set up version control and CI/CD pipeline
       - Establish coding standards and development environment

       ## Deliverables
       - Project directory structure
       - CI/CD configuration
       - Development environment setup
       - Initial documentation
       EOF

    2. Create phase goals (reference roadmap for goal names):
       crumbler phase goal create 0001 "Initialize project structure"
       crumbler phase goal create 0001 "Set up CI/CD pipeline"
       crumbler phase goal create 0001 "Configure development environment"

    3. Create sprints to work on goals:
       crumbler sprint create 0001

    DOCUMENTATION STRUCTURE:
    Phase README.md should include:
    - H1 header with phase name and number
    - Overview section describing the phase purpose
    - Objectives section (bullet list of high-level goals)
    - Deliverables section (what will be produced)
    - Any other relevant context for the phase
`)
}

// printPhaseCloseHelp prints help for 'crumbler phase close'.
func printPhaseCloseHelp() {
	fmt.Print(`crumbler phase close - Close a phase

USAGE:
    crumbler phase close <phase-id>

ARGUMENTS:
    phase-id    The 4-digit phase ID (e.g., 0001)

DESCRIPTION:
    Closes a phase by removing the 'open' file and creating a 'closed' file.

    A phase can only be closed when:
    - All sprints in the phase are closed
    - All phase goals are closed

ERRORS:
    - "phase not found" - phase directory doesn't exist
    - "phase is already closed" - phase already has closed file
    - "cannot close phase: sprint X is still open" - sprints must be closed first
    - "cannot close phase: goal X is still open" - goals must be closed first

EXAMPLES:
    crumbler phase close 0001

STATE CHANGES:
    Before: .crumbler/phases/0001-phase/open exists
    After:  .crumbler/phases/0001-phase/closed exists (open removed)
`)
}

// printPhaseGoalHelp prints help for 'crumbler phase goal'.
func printPhaseGoalHelp() {
	fmt.Print(`crumbler phase goal - Manage phase goals

USAGE:
    crumbler phase goal <subcommand> [arguments]

SUBCOMMANDS:
    create <phase-id> <goal-name>   Create a goal in a phase
    close <phase-id> <goal-id>      Close a goal in a phase

DESCRIPTION:
    Phase goals track high-level objectives that must be completed before
    a phase can be closed. Goals have a name and open/closed status.

GOAL DIRECTORY STRUCTURE:
    .crumbler/phases/XXXX-phase/goals/XXXX-goal/
    .crumbler/phases/XXXX-phase/goals/XXXX-goal/name     Goal name (text)
    .crumbler/phases/XXXX-phase/goals/XXXX-goal/open     Goal is open (empty)
    .crumbler/phases/XXXX-phase/goals/XXXX-goal/closed   Goal is closed (empty)

EXAMPLES:
    crumbler phase goal create 0001 "Set up CI/CD"
    crumbler phase goal close 0001 0001
`)
}

// printPhaseGoalCreateHelp prints help for 'crumbler phase goal create'.
func printPhaseGoalCreateHelp() {
	fmt.Print(`crumbler phase goal create - Create a phase goal

USAGE:
    crumbler phase goal create <phase-id> <goal-name>

ARGUMENTS:
    phase-id    The 4-digit phase ID (e.g., 0001)
    goal-name   The name of the goal (can contain spaces)

DESCRIPTION:
    Creates a new goal in the specified phase. Goal IDs are 4-digit
    zero-padded numbers assigned automatically (0001, 0002, etc.).

CREATES:
    .crumbler/phases/XXXX-phase/goals/YYYY-goal/
    .crumbler/phases/XXXX-phase/goals/YYYY-goal/name    Contains goal name
    .crumbler/phases/XXXX-phase/goals/YYYY-goal/open    Open status (empty)

EXAMPLES:
    crumbler phase goal create 0001 "Set up CI/CD pipeline"
    crumbler phase goal create 0001 "Initialize project structure"

FOR AI AGENTS:
    Goal names should be:
    - High-level phase objectives (not implementation details)
    - Clear and measurable outcomes
    - Aligned with roadmap phase goals

    Create goals that represent high-level phase objectives:

    crumbler phase goal create 0001 "Implement user authentication"
    crumbler phase goal create 0001 "Set up database schema"
    crumbler phase goal create 0001 "Create API endpoints"

    Good phase goal examples:
    - "Establish project foundation and structure"
    - "Implement core authentication system"
    - "Set up CI/CD pipeline and deployment"
    - "Create comprehensive test coverage"

    Goal names are stored in the name file as plain text.
`)
}

// printPhaseGoalCloseHelp prints help for 'crumbler phase goal close'.
func printPhaseGoalCloseHelp() {
	fmt.Print(`crumbler phase goal close - Close a phase goal

USAGE:
    crumbler phase goal close <phase-id> <goal-id>

ARGUMENTS:
    phase-id    The 4-digit phase ID (e.g., 0001)
    goal-id     The 4-digit goal ID (e.g., 0001)

DESCRIPTION:
    Closes a goal in the specified phase by removing the 'open' file
    and creating a 'closed' file.

ERRORS:
    - "goal not found" - goal directory doesn't exist
    - "goal is already closed" - goal already has closed file
    - "goal has invalid state" - neither open nor closed file exists

EXAMPLES:
    crumbler phase goal close 0001 0001
    crumbler phase goal close 0001 0002

STATE CHANGES:
    Before: .crumbler/phases/0001-phase/goals/0001-goal/open exists
    After:  .crumbler/phases/0001-phase/goals/0001-goal/closed exists (open removed)
`)
}
