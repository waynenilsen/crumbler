package crumbler

import (
	"fmt"

	"github.com/waynenilsen/crumbler/internal/crumb"
)

// runDelete handles the 'crumbler delete' command.
// It deletes the current crumb (marks work as done).
func runDelete(args []string) error {
	// Handle help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help") {
		printDeleteHelp()
		return nil
	}

	projectRoot, err := getProjectRoot()
	if err != nil {
		return err
	}

	// Get current crumb for display
	current, err := crumb.GetCurrent(projectRoot)
	if err != nil {
		return err
	}

	if current == nil {
		fmt.Println("No crumbs to delete. Project is done!")
		return nil
	}

	relPath := relPath(projectRoot, current.Path)

	// Delete the crumb
	if err := crumb.Delete(projectRoot); err != nil {
		return fmt.Errorf("failed to delete crumb: %w", err)
	}

	fmt.Printf("Deleted crumb: %s\n", relPath)

	// Check if project is now done
	done, err := crumb.IsDone(projectRoot)
	if err != nil {
		return err
	}

	if done {
		fmt.Println("\nAll crumbs completed. Project is done!")
	} else {
		fmt.Println("\nRun 'crumbler prompt' for next task.")
	}

	return nil
}

// printDeleteHelp prints help for the delete command.
func printDeleteHelp() {
	fmt.Print(`crumbler delete - Delete the current crumb

USAGE:
    crumbler delete

DESCRIPTION:
    Deletes the current crumb, marking its work as complete. The current
    crumb is always the deepest leaf (found via depth-first traversal).

    After deletion, the parent crumb becomes the new current crumb (if it
    has no other children), or the next sibling becomes current.

    You cannot delete a crumb that has children. Complete child crumbs first.

WORKFLOW:
    1. Execute the work described in the crumb's README
    2. Run 'crumbler delete' to mark the work as done
    3. Run 'crumbler prompt' to get the next task

CONSTRAINTS:
    - Cannot delete a crumb with children
    - Cannot delete the root .crumbler directory

EXAMPLES:
    # Delete current crumb after completing its work
    crumbler delete

    # Check what's next
    crumbler prompt

ERRORS:
    - "no crumb to delete" - No crumbs exist
    - "cannot delete crumb with children" - Must delete children first
    - "not a crumbler project" - No .crumbler directory found
`)
}
