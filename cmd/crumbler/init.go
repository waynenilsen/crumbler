package crumbler

import (
	"fmt"
	"os"
)

// runInit handles the 'crumbler init' command.
// It initializes a new crumbler project in the current directory.
func runInit(args []string) error {
	// Handle help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help") {
		printInitHelp()
		return nil
	}

	// Check if already initialized
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	crumblerPath := dir + "/.crumbler"
	if _, err := os.Stat(crumblerPath); err == nil {
		return fmt.Errorf("error: project already initialized at %s\n\nUse 'crumbler status' to see project state", crumblerPath)
	}

	// Create directory structure
	dirs := []string{
		crumblerPath,
		crumblerPath + "/phases",
		crumblerPath + "/roadmaps",
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", d, err)
		}
	}

	// Create empty files
	files := []string{
		crumblerPath + "/README.md",
		crumblerPath + "/roadmap.md",
	}

	for _, f := range files {
		file, err := os.Create(f)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", f, err)
		}
		file.Close()
	}

	fmt.Println("Initialized crumbler project in .crumbler/")
	fmt.Println()
	fmt.Println("Created:")
	fmt.Println("  .crumbler/")
	fmt.Println("  .crumbler/phases/")
	fmt.Println("  .crumbler/roadmaps/")
	fmt.Println("  .crumbler/README.md")
	fmt.Println("  .crumbler/roadmap.md")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Edit .crumbler/roadmap.md to define your project roadmap")
	fmt.Println("  2. Edit .crumbler/README.md to describe your project")
	fmt.Println("  3. Run 'crumbler phase create' to create your first phase")

	return nil
}

// printInitHelp prints help for the init command.
func printInitHelp() {
	fmt.Print(`crumbler init - Initialize a new crumbler project

USAGE:
    crumbler init

DESCRIPTION:
    Initializes a new crumbler project in the current directory by creating
    the .crumbler/ directory structure. This must be run before using any
    other crumbler commands.

    This command will fail if the current directory already contains a
    .crumbler/ directory (project is already initialized).

CREATES:
    .crumbler/              Main state directory
    .crumbler/phases/       Directory for phase subdirectories
    .crumbler/roadmaps/     Directory for roadmap archives
    .crumbler/README.md     Empty project overview (AI populates)
    .crumbler/roadmap.md    Empty roadmap file (AI populates)

ERRORS:
    - "project already initialized" - .crumbler/ already exists

EXAMPLES:
    # Initialize a new project
    crumbler init

    # Check if initialized
    crumbler status

NEXT STEPS AFTER INIT:
    1. Populate .crumbler/roadmap.md with your project roadmap (markdown format)
    2. Populate .crumbler/README.md with your project description
    3. Create first phase: crumbler phase create

FOR AI AGENTS:
    After running 'crumbler init', you should:
    1. Write content to .crumbler/README.md describing the project
    2. Write content to .crumbler/roadmap.md defining phases and goals
    3. Use 'crumbler phase create' to create the first phase

    Example commands to populate files:
        cat > .crumbler/README.md << 'EOF'
        # Project Name

        Project description here.
        EOF

        cat > .crumbler/roadmap.md << 'EOF'
        # Project Roadmap

        ## Phase 1: Foundation
        - Goal 1: Set up project structure
        - Goal 2: Implement core data models
        EOF
`)
}
