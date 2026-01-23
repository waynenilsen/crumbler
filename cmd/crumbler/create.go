package crumbler

import (
	"fmt"
	"strings"

	"github.com/waynenilsen/crumbler/internal/crumb"
)

// runCreate handles the 'crumbler create' command.
// It creates a new sub-crumb under the current crumb.
func runCreate(args []string) error {
	// Handle help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help") {
		printCreateHelp()
		return nil
	}

	// Require a name argument
	if len(args) == 0 {
		return fmt.Errorf("error: missing crumb name\n\nUsage: crumbler create \"Name\"\n\nRun 'crumbler create --help' for more information")
	}

	// Join all args as the name (in case it wasn't quoted)
	name := strings.Join(args, " ")

	projectRoot, err := getProjectRoot()
	if err != nil {
		return err
	}

	// Create the crumb
	path, err := crumb.Create(projectRoot, name)
	if err != nil {
		return fmt.Errorf("failed to create crumb: %w", err)
	}

	relPath := relPath(projectRoot, path)
	fmt.Printf("Created crumb: %s\n", relPath)
	fmt.Printf("README: %s/README.md\n", relPath)

	return nil
}

// printCreateHelp prints help for the create command.
func printCreateHelp() {
	fmt.Print(`crumbler create - Create a new sub-crumb

USAGE:
    crumbler create "Name"

DESCRIPTION:
    Creates a new sub-crumb under the current crumb. The name is automatically
    converted to kebab-case and assigned the next available ID (01-10).

    Crumbs are created with an empty README.md file that you should populate
    with task instructions.

ARGUMENTS:
    Name    Human-readable name for the crumb (required)
            Will be converted to kebab-case (e.g., "Add Auth" -> "add-auth")

CREATES:
    XX-name/           Crumb directory (XX is auto-assigned ID)
    XX-name/README.md  Empty README file

CONSTRAINTS:
    - Maximum 10 children per crumb (IDs 01-10)
    - Names are converted to kebab-case
    - Created under the current crumb (deepest leaf)

EXAMPLES:
    crumbler create "Setup Database"
    # Creates: 01-setup-database/README.md

    crumbler create "Add User Authentication"
    # Creates: 02-add-user-authentication/README.md

    crumbler create "Fix Bug"
    # Creates under current crumb

ERRORS:
    - "directory is full" - Parent crumb already has 10 children
    - "not a crumbler project" - No .crumbler directory found
`)
}
