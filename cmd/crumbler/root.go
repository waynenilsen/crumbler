// Package crumbler provides the CLI interface for the crumbler tool.
package crumbler

import (
	"fmt"
	"os"
)

// Execute is the main entry point for the CLI, called from main.go.
// It parses command line arguments and routes to the appropriate subcommand.
func Execute() error {
	args := os.Args[1:]

	// Handle empty args - show help
	if len(args) == 0 {
		printTopLevelHelp()
		return nil
	}

	// Handle global flags first
	if args[0] == "--help" || args[0] == "-h" {
		printTopLevelHelp()
		return nil
	}

	// Route to subcommand
	switch args[0] {
	case "help":
		return runHelp(args[1:])
	case "init":
		return runInit(args[1:])
	case "status":
		return runStatus(args[1:])
	case "phase":
		return runPhase(args[1:])
	case "sprint":
		return runSprint(args[1:])
	case "ticket":
		return runTicket(args[1:])
	case "roadmap":
		return runRoadmap(args[1:])
	case "clean":
		return runClean(args[1:])
	case "get-next-prompt":
		return runGetNextPrompt(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "error: unknown command '%s'\n\n", args[0])
		printTopLevelHelp()
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

// printTopLevelHelp prints the top-level help message.
func printTopLevelHelp() {
	fmt.Print(`crumbler - Agentic SDLC State Machine Manager

crumbler is a lightweight CLI tool for managing software development lifecycle
state. It manages state transitions, directory structure, and validates the
state machine integrity. It does NOT generate content - AI agents are
responsible for populating document content.

USAGE:
    crumbler <command> [subcommand] [options]

COMMANDS:
    init            Initialize a new crumbler project in the current directory
    status          Show current state of the project
    phase           Manage phases (list, create, close, goal)
    sprint          Manage sprints (list, create, close, goal)
    ticket          Manage tickets (list, create, done, goal)
    roadmap         Manage roadmap (load, show)
    get-next-prompt Generate AI agent prompt based on current state
    clean           Format Claude Code streaming JSON output
    help            Show help for a command

FLAGS:
    -h, --help  Show this help message

EXAMPLES:
    crumbler init                           Initialize project
    crumbler status                         Show project status
    crumbler phase list                     List all phases
    crumbler phase create                   Create next phase
    crumbler sprint create                  Create sprint in current phase
    crumbler ticket done 0001               Mark ticket as done
    crumbler clean                           Format Claude Code JSON output
    crumbler help phase                     Show phase command help

PROJECT STRUCTURE:
    .crumbler/                              State directory (created on init)
    .crumbler/README.md                     Project overview
    .crumbler/roadmap.md                    Current roadmap
    .crumbler/phases/                       All phases
    .crumbler/phases/XXXX-phase/            Phase directory
    .crumbler/phases/XXXX-phase/open        Phase is open (empty file)
    .crumbler/phases/XXXX-phase/closed      Phase is closed (empty file)
    .crumbler/phases/XXXX-phase/sprints/    Sprints in this phase
    .crumbler/phases/XXXX-phase/goals/      Goals for this phase
    .crumbler/roadmaps/                     Roadmap archives

STATE FILES:
    open        Entity is open (phase, sprint, ticket, goal)
    closed      Entity is closed (phase, sprint, goal)
    done        Ticket is complete

For more information on a specific command, use:
    crumbler help <command>
    crumbler <command> --help
`)
}

// findProjectRoot locates the .crumbler directory by walking up from pwd.
// Returns the path to the project root (directory containing .crumbler) or
// an error if not found.
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	for {
		crumblerDir := dir + "/.crumbler"
		if info, err := os.Stat(crumblerDir); err == nil && info.IsDir() {
			return dir, nil
		}

		parent := dir[:max(0, len(dir)-len("/"+lastPathComponent(dir)))]
		if parent == dir || parent == "" {
			return "", fmt.Errorf("not a crumbler project (no .crumbler directory found)")
		}
		dir = parent
	}
}

// lastPathComponent returns the last component of a path.
func lastPathComponent(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}

// requireProject ensures we are in a crumbler-managed project.
// Returns the project root or an error if not in a project.
func requireProject() (string, error) {
	root, err := findProjectRoot()
	if err != nil {
		return "", fmt.Errorf("error: %w\n\nRun 'crumbler init' to initialize a new project", err)
	}
	return root, nil
}

// crumblerDir returns the path to the .crumbler directory.
func crumblerDir(projectRoot string) string {
	return projectRoot + "/.crumbler"
}

// phasesDir returns the path to the phases directory.
func phasesDir(projectRoot string) string {
	return crumblerDir(projectRoot) + "/phases"
}

// relPath returns a path relative to the project root.
func relPath(projectRoot, fullPath string) string {
	if len(fullPath) > len(projectRoot)+1 {
		return fullPath[len(projectRoot)+1:]
	}
	return fullPath
}
