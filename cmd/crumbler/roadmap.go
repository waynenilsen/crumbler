package crumbler

import (
	"fmt"
	"io"
	"os"
)

// runRoadmap handles the 'crumbler roadmap' command and its subcommands.
func runRoadmap(args []string) error {
	if len(args) == 0 {
		printRoadmapHelp()
		return nil
	}

	// Handle help flag
	if args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		printRoadmapHelp()
		return nil
	}

	switch args[0] {
	case "load":
		return runRoadmapLoad(args[1:])
	case "show":
		return runRoadmapShow(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "error: unknown roadmap subcommand '%s'\n\n", args[0])
		printRoadmapHelp()
		return fmt.Errorf("unknown roadmap subcommand: %s", args[0])
	}
}

// runRoadmapLoad handles 'crumbler roadmap load <file>'.
func runRoadmapLoad(args []string) error {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		printRoadmapLoadHelp()
		return nil
	}

	projectRoot, err := requireProject()
	if err != nil {
		return err
	}

	sourceFile := args[0]

	// Check source file exists
	if _, err := os.Stat(sourceFile); os.IsNotExist(err) {
		return fmt.Errorf("source file not found: %s", sourceFile)
	}

	// Read source file
	content, err := os.ReadFile(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to read source file %s: %w", sourceFile, err)
	}

	// Write to .crumbler/roadmap.md
	destPath := crumblerDir(projectRoot) + "/roadmap.md"
	if err := os.WriteFile(destPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write roadmap to %s: %w", relPath(projectRoot, destPath), err)
	}

	// Also archive to roadmaps directory
	archivePath := crumblerDir(projectRoot) + "/roadmaps/" + lastPathComponent(sourceFile)
	if err := os.WriteFile(archivePath, content, 0644); err != nil {
		// Non-fatal - just warn
		fmt.Fprintf(os.Stderr, "warning: failed to archive roadmap to %s: %v\n",
			relPath(projectRoot, archivePath), err)
	}

	fmt.Printf("Loaded roadmap from %s to %s\n", sourceFile, relPath(projectRoot, destPath))
	if _, err := os.Stat(archivePath); err == nil {
		fmt.Printf("Archived copy at %s\n", relPath(projectRoot, archivePath))
	}

	return nil
}

// runRoadmapShow handles 'crumbler roadmap show'.
func runRoadmapShow(args []string) error {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help") {
		printRoadmapShowHelp()
		return nil
	}

	projectRoot, err := requireProject()
	if err != nil {
		return err
	}

	roadmapPath := crumblerDir(projectRoot) + "/roadmap.md"

	// Check roadmap exists
	if _, err := os.Stat(roadmapPath); os.IsNotExist(err) {
		fmt.Println("No roadmap found.")
		fmt.Println()
		fmt.Println("Create a roadmap by:")
		fmt.Println("  1. Edit .crumbler/roadmap.md directly, or")
		fmt.Println("  2. Load from file: crumbler roadmap load <file>")
		return nil
	}

	// Read and display roadmap
	file, err := os.Open(roadmapPath)
	if err != nil {
		return fmt.Errorf("failed to open roadmap: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read roadmap: %w", err)
	}

	if len(content) == 0 {
		fmt.Println("Roadmap is empty.")
		fmt.Println()
		fmt.Println("Populate the roadmap by:")
		fmt.Println("  1. Edit .crumbler/roadmap.md directly, or")
		fmt.Println("  2. Load from file: crumbler roadmap load <file>")
		return nil
	}

	fmt.Printf("Roadmap (%s):\n", relPath(projectRoot, roadmapPath))
	fmt.Println("================================================================================")
	fmt.Println(string(content))

	return nil
}

// printRoadmapHelp prints help for the roadmap command.
func printRoadmapHelp() {
	fmt.Print(`crumbler roadmap - Manage project roadmap

USAGE:
    crumbler roadmap <subcommand> [arguments]

SUBCOMMANDS:
    load <file>    Load a roadmap file into the project
    show           Display the current roadmap

DESCRIPTION:
    The roadmap defines the high-level plan for the project. It is stored
    in .crumbler/roadmap.md as a markdown file that AI agents can read
    and reference when creating phases and sprints.

ROADMAP FILE:
    .crumbler/roadmap.md    Current project roadmap (markdown)
    .crumbler/roadmaps/     Archive of loaded roadmaps

EXAMPLES:
    crumbler roadmap show                Show current roadmap
    crumbler roadmap load my-roadmap.md  Load roadmap from file

For help on a specific subcommand:
    crumbler roadmap <subcommand> --help
`)
}

// printRoadmapLoadHelp prints help for 'crumbler roadmap load'.
func printRoadmapLoadHelp() {
	fmt.Print(`crumbler roadmap load - Load a roadmap file

USAGE:
    crumbler roadmap load <file>

ARGUMENTS:
    file    Path to the roadmap file to load (markdown format)

DESCRIPTION:
    Loads a roadmap file into the project by copying its contents to
    .crumbler/roadmap.md. The original file is also archived to
    .crumbler/roadmaps/ for reference.

DESTINATION:
    .crumbler/roadmap.md                   Active roadmap
    .crumbler/roadmaps/<filename>          Archived copy

EXAMPLES:
    crumbler roadmap load roadmap.md
    crumbler roadmap load docs/project-plan.md

ROADMAP FORMAT:
    The roadmap is a markdown file with a structured format:

    # Project Roadmap

    ## Phase 1: Foundation
    Optional description text for phase 1.

    - Set up project structure
    - Implement core data models
    - Create basic configuration files

    ## Phase 2: Core Features
    Optional description text for phase 2.

    - Implement user authentication
    - Create API endpoints
    - Build state management system

    STRUCTURE:
    - H2 headers (##) define phases
    - Bullet lists (- or *) under each phase are goals
    - Text between header and bullets becomes phase description
    - Optional metadata (key: value) at the top

FOR AI AGENTS:
    The roadmap provides context for creating phases and sprints. When
    creating a new phase, reference the roadmap to determine:
    - Phase name and description
    - Phase goals to create
    - Sprint planning for the phase

    Example workflow:
    1. crumbler roadmap show           # Read the roadmap
    2. crumbler phase create           # Create next phase
    3. (edit phase README.md based on roadmap)
    4. crumbler phase goal create ...  # Create goals from roadmap
`)
}

// printRoadmapShowHelp prints help for 'crumbler roadmap show'.
func printRoadmapShowHelp() {
	fmt.Print(`crumbler roadmap show - Display the current roadmap

USAGE:
    crumbler roadmap show

DESCRIPTION:
    Displays the contents of .crumbler/roadmap.md.

    If the roadmap file doesn't exist or is empty, instructions for
    creating/loading a roadmap are shown.

EXAMPLES:
    crumbler roadmap show

OUTPUT:
    Roadmap (.crumbler/roadmap.md):
    ================================================================================
    # Project Roadmap

    ## Phase 1: Foundation
    ...

FOR AI AGENTS:
    Use 'crumbler roadmap show' to read the project plan before:
    - Creating new phases (to understand phase goals)
    - Creating sprints (to understand sprint scope)
    - Planning work (to understand project context)
`)
}
