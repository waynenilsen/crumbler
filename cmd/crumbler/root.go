// Package crumbler provides the CLI interface for the crumbler tool.
package crumbler

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/waynenilsen/crumbler/internal/crumb"
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
		if len(args) > 1 {
			return runHelpFor(args[1])
		}
		printTopLevelHelp()
		return nil
	case "status":
		return runStatus(args[1:])
	case "create":
		return runCreate(args[1:])
	case "delete":
		return runDelete(args[1:])
	case "prompt":
		return runPrompt(args[1:])
	case "clean":
		return runClean(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "error: unknown command '%s'\n\n", args[0])
		printTopLevelHelp()
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

// runHelpFor prints help for a specific command.
func runHelpFor(cmd string) error {
	switch cmd {
	case "status":
		return runStatus([]string{"--help"})
	case "create":
		return runCreate([]string{"--help"})
	case "delete":
		return runDelete([]string{"--help"})
	case "prompt":
		return runPrompt([]string{"--help"})
	case "clean":
		return runClean([]string{"--help"})
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

// printTopLevelHelp prints the top-level help message.
func printTopLevelHelp() {
	fmt.Print(`crumbler - Simple Task Decomposition for AI Agents

crumbler organizes work into "crumbs" - directories with README.md files.
Work depth-first: complete children before parents. Delete crumbs when done.
The filesystem IS the state - existence means work to do, deletion means done.

USAGE:
    crumbler <command> [options]

COMMANDS:
    status    Show crumb tree and current state
    create    Create a new sub-crumb (auto-initializes if needed)
    delete    Delete the current crumb (mark work as done)
    prompt    Generate AI agent prompt for current state
    clean     Format Claude Code streaming JSON output
    help      Show help for a command

FLAGS:
    -h, --help  Show this help message

WORKFLOW:
    1. crumbler create "Task"     # Create first crumb (auto-inits)
    2. crumbler prompt            # Get AI instructions
    3. [Do the work]              # Follow instructions
    4. crumbler delete            # Mark crumb as done
    5. crumbler prompt            # Get next instructions
    6. Repeat until done

EXAMPLES:
    crumbler create "Setup Database"     # Create crumb (auto-inits)
    crumbler prompt                      # Get instructions
    crumbler status                      # View crumb tree
    crumbler delete                      # Mark done
    crumbler help create                 # Get command help

STRUCTURE:
    .crumbler/                           # Project root (auto-created)
    ├── README.md                        # Root crumb
    ├── 01-phase-one/                    # First child crumb
    │   ├── README.md                    # Task instructions
    │   └── 01-subtask/                  # Nested crumb
    │       └── README.md
    └── 02-phase-two/
        └── README.md

For more information: crumbler help <command>
`)
}

// findProjectRoot locates the .crumbler directory by walking up from pwd.
// If .crumbler exists, returns path to directory containing it.
// If .crumbler doesn't exist, returns current working directory.
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Walk up to find existing .crumbler
	searchDir := dir
	for {
		crumblerDir := filepath.Join(searchDir, crumb.CrumblerDir)
		if info, err := os.Stat(crumblerDir); err == nil && info.IsDir() {
			return searchDir, nil
		}

		parent := filepath.Dir(searchDir)
		if parent == searchDir {
			// No .crumbler found - return cwd (create will auto-init)
			return dir, nil
		}
		searchDir = parent
	}
}

// getProjectRoot returns the project root (cwd or directory with .crumbler).
func getProjectRoot() (string, error) {
	return findProjectRoot()
}

// crumblerDir returns the path to the .crumbler directory.
func crumblerDir(projectRoot string) string {
	return filepath.Join(projectRoot, crumb.CrumblerDir)
}

// relPath returns a path relative to the project root.
func relPath(projectRoot, fullPath string) string {
	rel, err := filepath.Rel(projectRoot, fullPath)
	if err != nil {
		return fullPath
	}
	return rel
}
