package crumbler

import (
	"fmt"

	"github.com/waynenilsen/crumbler/internal/crumb"
	"github.com/waynenilsen/crumbler/internal/prompt"
)

// runStatus handles the 'crumbler status' command.
// It displays the current state of the project as a tree.
func runStatus(args []string) error {
	// Handle help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help") {
		printStatusHelp()
		return nil
	}

	projectRoot, err := getProjectRoot()
	if err != nil {
		return err
	}

	// Get crumb count
	count, err := crumb.Count(projectRoot)
	if err != nil {
		return err
	}

	// Get the tree
	tree, err := crumb.List(projectRoot)
	if err != nil {
		return err
	}

	// Print header
	if count == 0 {
		fmt.Println("Project Status: DONE (no crumbs remaining)")
		return nil
	}

	fmt.Printf("Project Status: %d crumb(s) remaining\n\n", count)

	// Get current crumb
	current, err := crumb.GetCurrent(projectRoot)
	if err != nil {
		return err
	}

	// Print tree with current crumb marked
	currentPath := ""
	if current != nil {
		currentPath = current.Path
	}
	fmt.Println(prompt.FormatTreeWithCurrent(tree, "", false, currentPath))

	if current != nil {
		fmt.Printf("Current: %s\n", current.RelPath)
	}

	return nil
}

// printStatusHelp prints help for the status command.
func printStatusHelp() {
	fmt.Print(`crumbler status - Show project status

USAGE:
    crumbler status

DESCRIPTION:
    Displays the current state of the crumbler project including:
    - Total number of remaining crumbs
    - Tree view of all crumbs
    - Current crumb (marked with arrow)

OUTPUT:
    Tree view shows the crumb hierarchy with the current crumb marked.
    Current crumb is always the deepest, first-by-ID leaf node.

EXAMPLES:
    crumbler status

OUTPUT EXAMPLE:
    Project Status: 3 crumb(s) remaining

    .crumbler/
    ├── 01-setup/
    │   └── 01-database/ ← current
    └── 02-features/

    Current: .crumbler/01-setup/01-database
`)
}
