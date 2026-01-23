package crumbler

import (
	"fmt"

	"github.com/waynenilsen/crumbler/internal/crumb"
)

// runCreate handles the 'crumbler create' command.
// It creates new sub-crumbs under the current crumb.
// Multiple names can be provided to create sibling crumbs.
func runCreate(args []string) error {
	// Handle help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help") {
		printCreateHelp()
		return nil
	}

	// Require at least one name argument
	if len(args) == 0 {
		return fmt.Errorf("error: missing crumb name\n\nUsage: crumbler create \"Name\" [\"Name2\" ...]\n\nRun 'crumbler create --help' for more information")
	}

	projectRoot, err := getProjectRoot()
	if err != nil {
		return err
	}

	// Create all crumbs as siblings
	paths, err := crumb.CreateMultiple(projectRoot, args)
	if err != nil {
		return fmt.Errorf("failed to create crumb(s): %w", err)
	}

	for _, path := range paths {
		rel := relPath(projectRoot, path)
		fmt.Printf("Created crumb: %s\n", rel)
	}

	return nil
}

// printCreateHelp prints help for the create command.
func printCreateHelp() {
	fmt.Print(`crumbler create - Create new sub-crumbs

USAGE:
    crumbler create "Name" ["Name2" ...]

DESCRIPTION:
    Creates new sub-crumbs under the current crumb. Names are automatically
    converted to kebab-case and assigned sequential IDs (01-10).

    Multiple names can be provided to create sibling crumbs in one command.
    This is useful during DECOMPOSE when planning multiple tasks at once.

    Crumbs are created with empty README.md files that you should populate
    with task instructions.

ARGUMENTS:
    Name    Human-readable name(s) for the crumb(s) (at least one required)
            Will be converted to kebab-case (e.g., "Add Auth" -> "add-auth")

CREATES:
    XX-name/           Crumb directory (XX is auto-assigned ID)
    XX-name/README.md  Empty README file

CONSTRAINTS:
    - Maximum 10 children per crumb (IDs 01-10)
    - Names are converted to kebab-case
    - All crumbs created as siblings under the current crumb

EXAMPLES:
    crumbler create "Setup Database"
    # Creates: 01-setup-database/README.md

    crumbler create "Add Auth" "Setup DB" "Write Tests"
    # Creates three sibling crumbs:
    #   01-add-auth/README.md
    #   02-setup-db/README.md
    #   03-write-tests/README.md

ERRORS:
    - "directory is full" - Parent crumb already has 10 children
    - "not a crumbler project" - No .crumbler directory found
`)
}
