// Package roadmap provides functionality for managing project roadmaps.
// Roadmaps define phases and their goals in a markdown format, which are
// used to drive the crumbler state machine.
package roadmap

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/waynenilsen/crumbler/internal/models"
	"github.com/waynenilsen/crumbler/internal/state"
)

// Roadmap markdown format regex patterns
var (
	// phaseHeaderRegex matches H2 headers for phases: ## Phase N: Name
	// Captures optional phase number and phase name
	phaseHeaderRegex = regexp.MustCompile(`^##\s+(?:Phase\s+\d+:\s*)?(.+)$`)

	// goalItemRegex matches bullet list items for goals: - Goal N: Description
	// Captures the goal description (with optional "Goal N:" prefix)
	goalItemRegex = regexp.MustCompile(`^[-*]\s+(?:Goal\s+\d+:\s*)?(.+)$`)

	// metadataRegex matches YAML-style metadata: key: value
	metadataRegex = regexp.MustCompile(`^(\w+):\s*(.+)$`)
)

// LoadRoadmap reads a roadmap from a markdown file at the given path.
// It returns the parsed Roadmap structure or an error if the file cannot be read.
func LoadRoadmap(path string) (*models.Roadmap, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read roadmap file %s: %w", path, err)
	}

	roadmap, err := ParseRoadmap(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse roadmap from %s: %w", path, err)
	}

	// Set the path to the roadmap file
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}
	roadmap.Path = absPath

	return roadmap, nil
}

// ParseRoadmap parses a markdown string into a Roadmap structure.
// The markdown format expects:
//   - Optional metadata at the top (key: value format)
//   - H2 headers (##) for phase names
//   - Bullet lists (- or *) for goals under each phase
//
// Example format:
//
//	# Project Roadmap
//
//	## Phase 1: Foundation
//	- Goal 1: Set up project structure
//	- Goal 2: Implement core models
//
//	## Phase 2: Core Features
//	- Goal 1: Implement state management
func ParseRoadmap(markdown string) (*models.Roadmap, error) {
	roadmap := &models.Roadmap{
		Phases:   []models.PhaseDefinition{},
		Metadata: make(map[string]string),
	}

	scanner := bufio.NewScanner(strings.NewReader(markdown))
	var currentPhase *models.PhaseDefinition
	var description strings.Builder
	inMetadata := true
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Skip empty lines
		if trimmedLine == "" {
			if currentPhase != nil && description.Len() > 0 {
				// Empty line ends description block
				if currentPhase.Description != "" {
					currentPhase.Description += "\n"
				}
				currentPhase.Description += strings.TrimSpace(description.String())
				description.Reset()
			}
			continue
		}

		// Handle H1 headers (title) - skip them
		if strings.HasPrefix(trimmedLine, "# ") && !strings.HasPrefix(trimmedLine, "##") {
			inMetadata = false
			continue
		}

		// Handle metadata at the top of the file
		if inMetadata && !strings.HasPrefix(trimmedLine, "#") && !strings.HasPrefix(trimmedLine, "-") && !strings.HasPrefix(trimmedLine, "*") {
			if matches := metadataRegex.FindStringSubmatch(trimmedLine); matches != nil {
				roadmap.Metadata[matches[1]] = matches[2]
				continue
			}
		}

		// Handle phase headers (## Phase N: Name)
		if strings.HasPrefix(trimmedLine, "## ") {
			inMetadata = false

			// Save description for previous phase
			if currentPhase != nil && description.Len() > 0 {
				if currentPhase.Description != "" {
					currentPhase.Description += "\n"
				}
				currentPhase.Description += strings.TrimSpace(description.String())
				description.Reset()
			}

			// Save previous phase if exists
			if currentPhase != nil {
				roadmap.Phases = append(roadmap.Phases, *currentPhase)
			}

			// Parse phase name from header
			matches := phaseHeaderRegex.FindStringSubmatch(trimmedLine)
			if matches == nil || len(matches) < 2 {
				return nil, fmt.Errorf("invalid phase header at line %d: %s", lineNum, trimmedLine)
			}

			currentPhase = &models.PhaseDefinition{
				Name:  strings.TrimSpace(matches[1]),
				Goals: []string{},
			}
			continue
		}

		// Handle goal items (- Goal N: Description or - Description)
		if strings.HasPrefix(trimmedLine, "- ") || strings.HasPrefix(trimmedLine, "* ") {
			inMetadata = false

			if currentPhase == nil {
				// Goals outside of a phase - skip or treat as top-level content
				continue
			}

			// Save any pending description
			if description.Len() > 0 {
				if currentPhase.Description != "" {
					currentPhase.Description += "\n"
				}
				currentPhase.Description += strings.TrimSpace(description.String())
				description.Reset()
			}

			matches := goalItemRegex.FindStringSubmatch(trimmedLine)
			if matches != nil && len(matches) >= 2 {
				goalText := strings.TrimSpace(matches[1])
				if goalText != "" {
					currentPhase.Goals = append(currentPhase.Goals, goalText)
				}
			}
			continue
		}

		// Handle description text (non-header, non-bullet text under a phase)
		if currentPhase != nil && !strings.HasPrefix(trimmedLine, "#") {
			inMetadata = false
			if description.Len() > 0 {
				description.WriteString("\n")
			}
			description.WriteString(trimmedLine)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading roadmap content: %w", err)
	}

	// Save final phase and its description
	if currentPhase != nil {
		if description.Len() > 0 {
			if currentPhase.Description != "" {
				currentPhase.Description += "\n"
			}
			currentPhase.Description += strings.TrimSpace(description.String())
		}
		roadmap.Phases = append(roadmap.Phases, *currentPhase)
	}

	return roadmap, nil
}

// ValidateRoadmap checks the roadmap structure for validity.
// It returns an error if the roadmap is invalid:
//   - Roadmap is nil
//   - Roadmap has no phases
//   - Any phase has an empty name
func ValidateRoadmap(roadmap *models.Roadmap) error {
	if roadmap == nil {
		return fmt.Errorf("roadmap is nil")
	}

	if len(roadmap.Phases) == 0 {
		return fmt.Errorf("roadmap has no phases defined")
	}

	for i, phase := range roadmap.Phases {
		if strings.TrimSpace(phase.Name) == "" {
			return fmt.Errorf("phase %d has empty name", i+1)
		}
	}

	return nil
}

// IsRoadmapComplete checks if all phases defined in the roadmap have been
// created and closed in the project. It returns true if all phases have
// a corresponding closed phase directory.
func IsRoadmapComplete(roadmap *models.Roadmap, projectRoot string) (bool, error) {
	if roadmap == nil {
		return false, fmt.Errorf("roadmap is nil")
	}

	if len(roadmap.Phases) == 0 {
		return true, nil // Empty roadmap is complete
	}

	phasesDir := state.PhasesDirPath(projectRoot)

	// Check if phases directory exists
	exists, err := state.DirExists(phasesDir)
	if err != nil {
		return false, fmt.Errorf("failed to check phases directory: %w", err)
	}
	if !exists {
		return false, nil // No phases created yet
	}

	// List all phase directories
	phaseDirs, err := state.ListDirs(phasesDir)
	if err != nil {
		return false, fmt.Errorf("failed to list phases: %w", err)
	}

	// Check if we have enough phases
	if len(phaseDirs) < len(roadmap.Phases) {
		return false, nil // Not all phases created yet
	}

	// Check if all phases are closed
	for i := 1; i <= len(roadmap.Phases); i++ {
		phaseID := state.FormatPhaseID(i)
		phasePath := state.PhasePath(projectRoot, phaseID)

		// Check if phase directory exists
		exists, err := state.DirExists(phasePath)
		if err != nil {
			return false, fmt.Errorf("failed to check phase %s: %w", phaseID, err)
		}
		if !exists {
			return false, nil // Phase not created
		}

		// Check if phase is closed
		closed, err := state.IsClosed(phasePath)
		if err != nil {
			return false, fmt.Errorf("failed to check if phase %s is closed: %w", phaseID, err)
		}
		if !closed {
			return false, nil // Phase not closed
		}
	}

	return true, nil
}

// GetNextPhaseFromRoadmap returns the next phase definition to create based on
// the roadmap and current project state. It returns:
//   - The next PhaseDefinition to create
//   - The 1-based index for the new phase
//   - An error if all phases are already created or an error occurs
//
// Returns nil, 0, nil if the roadmap is complete (all phases created).
func GetNextPhaseFromRoadmap(roadmap *models.Roadmap, projectRoot string) (*models.PhaseDefinition, int, error) {
	if roadmap == nil {
		return nil, 0, fmt.Errorf("roadmap is nil")
	}

	if len(roadmap.Phases) == 0 {
		return nil, 0, nil // No phases in roadmap
	}

	phasesDir := state.PhasesDirPath(projectRoot)

	// Check if phases directory exists
	exists, err := state.DirExists(phasesDir)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to check phases directory: %w", err)
	}

	// If no phases directory, return first phase
	if !exists {
		return &roadmap.Phases[0], 1, nil
	}

	// Count existing phase directories
	phaseDirs, err := state.ListDirs(phasesDir)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list phases: %w", err)
	}

	// Filter to only count valid phase directories
	phaseCount := 0
	for _, dir := range phaseDirs {
		if strings.HasSuffix(dir, models.PhaseSuffix) {
			phaseCount++
		}
	}

	// Check if all phases from roadmap are created
	if phaseCount >= len(roadmap.Phases) {
		return nil, 0, nil // All phases created
	}

	// Return the next phase to create
	nextIndex := phaseCount + 1
	return &roadmap.Phases[phaseCount], nextIndex, nil
}

// SaveRoadmap writes the roadmap to .crumbler/roadmap.md in the project.
// It creates the file in markdown format with phases as H2 headers and
// goals as bullet lists.
func SaveRoadmap(roadmap *models.Roadmap, projectRoot string) error {
	if roadmap == nil {
		return fmt.Errorf("roadmap is nil")
	}

	roadmapPath := state.RoadmapPath(projectRoot)

	// Ensure .crumbler directory exists
	crumblerDir := state.CrumblerDirPath(projectRoot)
	if err := state.CreateDir(crumblerDir); err != nil {
		return fmt.Errorf("failed to create .crumbler directory: %w", err)
	}

	// Generate markdown content
	content := formatRoadmapMarkdown(roadmap)

	// Write to file
	if err := os.WriteFile(roadmapPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write roadmap to %s: %w", roadmapPath, err)
	}

	// Update the roadmap's path
	absPath, err := filepath.Abs(roadmapPath)
	if err != nil {
		absPath = roadmapPath
	}
	roadmap.Path = absPath

	return nil
}

// LoadProjectRoadmap loads the roadmap from .crumbler/roadmap.md in the project.
// Returns an error if the roadmap file doesn't exist or cannot be parsed.
func LoadProjectRoadmap(projectRoot string) (*models.Roadmap, error) {
	roadmapPath := state.RoadmapPath(projectRoot)

	// Check if roadmap file exists
	if _, err := os.Stat(roadmapPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("roadmap not found at %s", roadmapPath)
	}

	return LoadRoadmap(roadmapPath)
}

// GetExampleRoadmap returns an example roadmap in markdown format.
// This can be used as a template for creating new roadmaps.
func GetExampleRoadmap() string {
	return `# Project Roadmap

## Phase 1: Foundation
- Set up project structure and directory layout
- Initialize version control with .gitignore
- Create basic configuration files
- Set up development environment

## Phase 2: Core Implementation
- Implement core data models
- Create state management system
- Build file-based persistence layer
- Implement validation logic

## Phase 3: Feature Development
- Implement primary user workflows
- Add error handling and recovery
- Create logging and monitoring
- Build API or CLI interface

## Phase 4: Testing and Documentation
- Write unit tests for core functionality
- Create integration tests
- Write user documentation
- Create API documentation

## Phase 5: Polish and Release
- Perform security review
- Optimize performance
- Create release packaging
- Prepare deployment scripts
`
}

// formatRoadmapMarkdown formats a Roadmap structure into markdown content.
func formatRoadmapMarkdown(roadmap *models.Roadmap) string {
	var sb strings.Builder

	sb.WriteString("# Project Roadmap\n\n")

	// Write metadata if present
	if len(roadmap.Metadata) > 0 {
		for key, value := range roadmap.Metadata {
			sb.WriteString(fmt.Sprintf("%s: %s\n", key, value))
		}
		sb.WriteString("\n")
	}

	// Write phases
	for i, phase := range roadmap.Phases {
		sb.WriteString(fmt.Sprintf("## Phase %d: %s\n", i+1, phase.Name))

		// Write description if present
		if phase.Description != "" {
			sb.WriteString(phase.Description)
			sb.WriteString("\n")
		}

		// Write goals
		for _, goal := range phase.Goals {
			sb.WriteString(fmt.Sprintf("- %s\n", goal))
		}

		// Add blank line between phases
		if i < len(roadmap.Phases)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// CreateRoadmapFromTemplate creates a new Roadmap from a template string.
// This is a convenience function that combines parsing and validation.
func CreateRoadmapFromTemplate(template string) (*models.Roadmap, error) {
	roadmap, err := ParseRoadmap(template)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	if err := ValidateRoadmap(roadmap); err != nil {
		return nil, fmt.Errorf("invalid roadmap: %w", err)
	}

	return roadmap, nil
}

// GetPhaseDefinition returns the phase definition at the given 1-based index.
// Returns nil if the index is out of range.
func GetPhaseDefinition(roadmap *models.Roadmap, index int) *models.PhaseDefinition {
	if roadmap == nil || index < 1 || index > len(roadmap.Phases) {
		return nil
	}
	return &roadmap.Phases[index-1]
}

// CountPhases returns the number of phases in the roadmap.
func CountPhases(roadmap *models.Roadmap) int {
	if roadmap == nil {
		return 0
	}
	return len(roadmap.Phases)
}

// HasGoals returns true if the phase at the given index has goals defined.
func HasGoals(roadmap *models.Roadmap, phaseIndex int) bool {
	phase := GetPhaseDefinition(roadmap, phaseIndex)
	if phase == nil {
		return false
	}
	return len(phase.Goals) > 0
}
