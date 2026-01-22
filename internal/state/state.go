// Package state provides utilities for managing file-based state.
package state

import (
	"os"

	"github.com/waynenilsen/crumbler/internal/models"
)

// Directory name constants with "DirName" suffix for clarity.
// These are used by sprint.go for consistency.
const (
	SprintsDirName  = models.SprintsDir
	GoalsDirName    = models.GoalsDir
	TicketsDirName  = models.TicketsDir
	PhasesDirName   = models.PhasesDir
	CrumblerDirName = models.CrumblerDir
)

// ValidateStatus checks that the status at the given path is valid
// (no conflicting status files exist). Alias for ValidateStatusFiles.
func ValidateStatus(path string) error {
	return ValidateStatusFiles(path)
}

// TouchFile creates an empty file at the given path.
// This is a public wrapper around the internal touchFile function.
func TouchFile(path string) error {
	return touchFile(path)
}

// CreateDir creates a directory with the given path.
func CreateDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// DirExists checks if a directory exists at the given path.
func DirExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}

// ListDirs returns a list of subdirectory names in the given directory.
func ListDirs(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}
	return dirs, nil
}
