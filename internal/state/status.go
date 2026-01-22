package state

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/waynenilsen/crumbler/internal/models"
)

// State file names
const (
	OpenFile   = "open"
	ClosedFile = "closed"
	DoneFile   = "done"
	NameFile   = "name"

	// Aliases for compatibility with existing code
	OpenFileName = OpenFile
	NameFileName = NameFile
)

// Note: Directory name constants are defined in state.go to avoid redeclaration

// File names
const (
	ReadmeFile = "README.md"
	PRDFile    = "PRD.md"
	ERDFile    = "ERD.md"
)

// ErrInvalidState is returned when an invalid state is detected.
var ErrInvalidState = errors.New("invalid state")

// IsOpen checks if the open file exists at the given path.
func IsOpen(path string) (bool, error) {
	return fileExists(filepath.Join(path, OpenFile))
}

// IsClosed checks if the closed file exists at the given path.
func IsClosed(path string) (bool, error) {
	return fileExists(filepath.Join(path, ClosedFile))
}

// IsDone checks if the done file exists at the given path.
func IsDone(path string) (bool, error) {
	return fileExists(filepath.Join(path, DoneFile))
}

// SetOpen creates the open file and removes the closed file.
func SetOpen(path string) error {
	if err := touchFile(filepath.Join(path, OpenFile)); err != nil {
		return err
	}
	return removeIfExists(filepath.Join(path, ClosedFile))
}

// SetClosed removes the open file and creates the closed file.
func SetClosed(path string) error {
	if err := removeIfExists(filepath.Join(path, OpenFile)); err != nil {
		return err
	}
	return touchFile(filepath.Join(path, ClosedFile))
}

// SetDone removes the open file and creates the done file.
func SetDone(path string) error {
	if err := removeIfExists(filepath.Join(path, OpenFile)); err != nil {
		return err
	}
	return touchFile(filepath.Join(path, DoneFile))
}

// ReadName reads the content of the name file in the given directory.
func ReadName(path string) (string, error) {
	content, err := os.ReadFile(filepath.Join(path, NameFile))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(content)), nil
}

// WriteName writes the given name to the name file in the given directory.
func WriteName(path, name string) error {
	return os.WriteFile(filepath.Join(path, NameFile), []byte(name), 0644)
}

// FormatIndex formats an index as a 4-digit zero-padded string.
func FormatIndex(index int) string {
	return fmt.Sprintf("%04d", index)
}

// ValidateState checks for invalid state (both open and closed exist).
// Returns an error with file paths if invalid state is detected.
func ValidateState(path, relPath string) error {
	openExists, err := fileExists(filepath.Join(path, OpenFile))
	if err != nil {
		return fmt.Errorf("failed to check open file at %s: %w", filepath.Join(relPath, OpenFile), err)
	}

	closedExists, err := fileExists(filepath.Join(path, ClosedFile))
	if err != nil {
		return fmt.Errorf("failed to check closed file at %s: %w", filepath.Join(relPath, ClosedFile), err)
	}

	if openExists && closedExists {
		return fmt.Errorf("invalid state: both 'open' and 'closed' exist in %s", relPath)
	}

	return nil
}

// ValidateTicketState checks for invalid ticket state (both open and done exist).
// Returns an error with file paths if invalid state is detected.
func ValidateTicketState(path, relPath string) error {
	openExists, err := fileExists(filepath.Join(path, OpenFile))
	if err != nil {
		return fmt.Errorf("failed to check open file at %s: %w", filepath.Join(relPath, OpenFile), err)
	}

	doneExists, err := fileExists(filepath.Join(path, DoneFile))
	if err != nil {
		return fmt.Errorf("failed to check done file at %s: %w", filepath.Join(relPath, DoneFile), err)
	}

	if openExists && doneExists {
		return fmt.Errorf("invalid state: both 'open' and 'done' exist in %s", relPath)
	}

	return nil
}

// fileExists checks if a file exists at the given path.
func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// removeIfExists removes a file if it exists.
func removeIfExists(path string) error {
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// ErrInvalidTransition is returned when an invalid state transition is attempted.
var ErrInvalidTransition = errors.New("invalid state transition")

// StatusError provides detailed error information about state issues.
type StatusError struct {
	// Err is the underlying error type.
	Err error
	// Path is the relative path where the error occurred.
	Path string
	// Message provides additional context about the error.
	Message string
	// ConflictingFiles lists any conflicting status files.
	ConflictingFiles []string
}

// Error implements the error interface.
func (e *StatusError) Error() string {
	if len(e.ConflictingFiles) > 0 {
		return fmt.Sprintf("%s: %s (conflicting files: %v)", e.Err, e.Message, e.ConflictingFiles)
	}
	return fmt.Sprintf("%s: %s at %s", e.Err, e.Message, e.Path)
}

// Unwrap returns the underlying error.
func (e *StatusError) Unwrap() error {
	return e.Err
}

// GetStatus returns the current status of the item at the given path.
// It validates that no conflicting status files exist.
func GetStatus(path string) (models.Status, error) {
	isOpen, err := IsOpen(path)
	if err != nil {
		return models.StatusUnknown, fmt.Errorf("failed to check open status: %w", err)
	}

	isClosed, err := IsClosed(path)
	if err != nil {
		return models.StatusUnknown, fmt.Errorf("failed to check closed status: %w", err)
	}

	isDone, err := IsDone(path)
	if err != nil {
		return models.StatusUnknown, fmt.Errorf("failed to check done status: %w", err)
	}

	// Check for invalid states (multiple status files exist)
	if isOpen && isClosed {
		return models.StatusUnknown, &StatusError{
			Err:              ErrInvalidState,
			Path:             path,
			Message:          "both open and closed files exist",
			ConflictingFiles: []string{filepath.Join(path, OpenFile), filepath.Join(path, ClosedFile)},
		}
	}
	if isOpen && isDone {
		return models.StatusUnknown, &StatusError{
			Err:              ErrInvalidState,
			Path:             path,
			Message:          "both open and done files exist",
			ConflictingFiles: []string{filepath.Join(path, OpenFile), filepath.Join(path, DoneFile)},
		}
	}
	if isClosed && isDone {
		return models.StatusUnknown, &StatusError{
			Err:              ErrInvalidState,
			Path:             path,
			Message:          "both closed and done files exist",
			ConflictingFiles: []string{filepath.Join(path, ClosedFile), filepath.Join(path, DoneFile)},
		}
	}

	// Return the current status
	if isOpen {
		return models.StatusOpen, nil
	}
	if isClosed {
		return models.StatusClosed, nil
	}
	if isDone {
		return models.StatusDone, nil
	}

	// No status file exists
	return models.StatusUnknown, &StatusError{
		Err:     ErrInvalidState,
		Path:    path,
		Message: "no status file exists (open, closed, or done)",
	}
}

// ValidateStatusFiles checks that the status at the given path is valid
// (no conflicting status files exist).
func ValidateStatusFiles(path string) error {
	_, err := GetStatus(path)
	return err
}

// SetOpenValidated sets the status to open by removing closed/done files and creating the open file.
// It validates the current state before making changes.
func SetOpenValidated(path string) error {
	// Validate current state - allow if no status exists
	if err := ValidateStatusFiles(path); err != nil {
		var statusErr *StatusError
		if errors.As(err, &statusErr) && statusErr.Message == "no status file exists (open, closed, or done)" {
			// This is OK, we're initializing the status
		} else {
			return err
		}
	}

	// Remove closed file if it exists
	closedPath := filepath.Join(path, ClosedFile)
	if err := os.Remove(closedPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove closed file at %s: %w", closedPath, err)
	}

	// Remove done file if it exists
	donePath := filepath.Join(path, DoneFile)
	if err := os.Remove(donePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove done file at %s: %w", donePath, err)
	}

	// Create open file atomically
	openPath := filepath.Join(path, OpenFile)
	if err := touchFile(openPath); err != nil {
		return fmt.Errorf("failed to create open file at %s: %w", openPath, err)
	}

	return nil
}

// SetClosedValidated sets the status to closed by removing open file and creating the closed file.
// It validates the current state before making changes and ensures proper state transition.
func SetClosedValidated(path string) error {
	// Validate current state
	status, err := GetStatus(path)
	if err != nil {
		return err
	}

	// Can only transition from open to closed
	if status != models.StatusOpen {
		return &StatusError{
			Err:     ErrInvalidTransition,
			Path:    path,
			Message: fmt.Sprintf("cannot transition from %s to closed", status),
		}
	}

	// Remove open file
	openPath := filepath.Join(path, OpenFile)
	if err := os.Remove(openPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove open file at %s: %w", openPath, err)
	}

	// Create closed file atomically
	closedPath := filepath.Join(path, ClosedFile)
	if err := touchFile(closedPath); err != nil {
		return fmt.Errorf("failed to create closed file at %s: %w", closedPath, err)
	}

	return nil
}

// SetDoneValidated sets the status to done by removing open file and creating the done file.
// This is used for tickets. It validates the current state before making changes.
func SetDoneValidated(path string) error {
	// Validate current state
	status, err := GetStatus(path)
	if err != nil {
		return err
	}

	// Can only transition from open to done
	if status != models.StatusOpen {
		return &StatusError{
			Err:     ErrInvalidTransition,
			Path:    path,
			Message: fmt.Sprintf("cannot transition from %s to done", status),
		}
	}

	// Remove open file
	openPath := filepath.Join(path, OpenFile)
	if err := os.Remove(openPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove open file at %s: %w", openPath, err)
	}

	// Create done file atomically
	donePath := filepath.Join(path, DoneFile)
	if err := touchFile(donePath); err != nil {
		return fmt.Errorf("failed to create done file at %s: %w", donePath, err)
	}

	return nil
}

// touchFile creates an empty file at the given path atomically.
// It creates a temp file and renames it to ensure atomic creation.
// This is the internal implementation used by other functions.
func touchFile(path string) error {
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Create a temporary file in the same directory
	tempFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()

	// Close the temp file
	if err := tempFile.Close(); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Rename temp file to target path (atomic on most filesystems)
	if err := os.Rename(tempPath, path); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename temp file to %s: %w", path, err)
	}

	return nil
}

// Note: TouchFile, CreateDir, DirExists, and ListDirs are defined in state.go
