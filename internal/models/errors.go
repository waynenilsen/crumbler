// Package models provides error types for the crumbler state machine.
package models

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Error type constants for categorizing state machine errors.
const (
	// ErrInvalidState indicates the state is invalid (e.g., both open and closed exist).
	ErrInvalidState = "invalid_state"

	// ErrInvalidTransition indicates an invalid state transition was attempted.
	ErrInvalidTransition = "invalid_transition"

	// ErrHierarchyConstraint indicates a hierarchy constraint was violated.
	// For example, trying to close a phase with open sprints.
	ErrHierarchyConstraint = "hierarchy_constraint"

	// ErrMissingFile indicates a required file is missing.
	// For example, a goal missing its "name" file.
	ErrMissingFile = "missing_file"
)

// ErrorType constants for detailed error categorization (used by validator).
const (
	// ErrorTypeOrphanedState indicates orphaned state files were found.
	ErrorTypeOrphanedState = "orphaned_state"

	// ErrorTypeMutuallyExclusiveState indicates both open and closed/done files exist.
	ErrorTypeMutuallyExclusiveState = "mutually_exclusive_state"

	// ErrorTypeHierarchyConstraint indicates a hierarchy constraint violation.
	ErrorTypeHierarchyConstraint = "hierarchy_constraint"

	// ErrorTypeInvalidTransition indicates an invalid state transition.
	ErrorTypeInvalidTransition = "invalid_transition"

	// ErrorTypeMissingGoalName indicates a goal is missing its name file.
	ErrorTypeMissingGoalName = "missing_goal_name"

	// ErrorTypeNotFound indicates the requested entity was not found.
	ErrorTypeNotFound = "not_found"

	// ErrorTypeProjectNotFound indicates the .crumbler directory was not found.
	ErrorTypeProjectNotFound = "project_not_found"
)

// StateError represents an error in the crumbler state machine.
// It includes the error type, affected file paths, and a descriptive message.
// All file paths are stored as relative paths from the project root for
// clear error reporting.
type StateError struct {
	// Type categorizes the error (e.g., ErrInvalidState, ErrInvalidTransition).
	Type string

	// FilePaths contains the file paths involved in the error (legacy field).
	// These are stored as relative paths from the project root.
	FilePaths []string

	// Paths is an alias for FilePaths for compatibility with validator.
	// These are stored as relative paths from the project root.
	Paths []string

	// Message is a human-readable description of the error.
	Message string

	// SuggestedFix provides a suggested action to fix the error.
	SuggestedFix string
}

// Error implements the error interface for StateError.
// It formats the error message to include the type, message, and affected files.
func (e *StateError) Error() string {
	// Prefer Paths over FilePaths (for validator compatibility)
	paths := e.Paths
	if len(paths) == 0 {
		paths = e.FilePaths
	}

	if len(paths) == 0 {
		if e.SuggestedFix != "" {
			return fmt.Sprintf("%s: %s\n  suggested fix: %s", e.Type, e.Message, e.SuggestedFix)
		}
		return fmt.Sprintf("%s: %s", e.Type, e.Message)
	}

	if e.SuggestedFix != "" {
		return fmt.Sprintf("%s: %s\n  files: %s\n  suggested fix: %s", e.Type, e.Message, strings.Join(paths, ", "), e.SuggestedFix)
	}
	return fmt.Sprintf("%s: %s\n  files: %s", e.Type, e.Message, strings.Join(paths, ", "))
}

// NewStateError creates a new StateError with the given type, message, and file paths.
// File paths should be provided as relative paths from the project root.
func NewStateError(errType, message string, filePaths ...string) *StateError {
	return &StateError{
		Type:      errType,
		Message:   message,
		FilePaths: filePaths,
	}
}

// NewInvalidStateError creates a StateError for invalid state conditions.
// For example, when both "open" and "closed" files exist in the same directory.
func NewInvalidStateError(message string, filePaths ...string) *StateError {
	return NewStateError(ErrInvalidState, message, filePaths...)
}

// NewInvalidTransitionError creates a StateError for invalid state transitions.
// For example, trying to close a phase that is already closed.
func NewInvalidTransitionError(message string, filePaths ...string) *StateError {
	return NewStateError(ErrInvalidTransition, message, filePaths...)
}

// NewHierarchyConstraintError creates a StateError for hierarchy constraint violations.
// For example, trying to close a phase when sprints are still open.
func NewHierarchyConstraintError(message string, filePaths ...string) *StateError {
	return NewStateError(ErrHierarchyConstraint, message, filePaths...)
}

// NewMissingFileError creates a StateError for missing required files.
// For example, a goal directory missing the "name" file.
func NewMissingFileError(message string, filePaths ...string) *StateError {
	return NewStateError(ErrMissingFile, message, filePaths...)
}

// ToRelPaths converts absolute file paths to relative paths from the project root.
// This is useful for converting internal absolute paths to user-friendly relative paths.
func ToRelPaths(projectRoot string, absolutePaths []string) []string {
	relPaths := make([]string, len(absolutePaths))
	for i, absPath := range absolutePaths {
		relPath, err := filepath.Rel(projectRoot, absPath)
		if err != nil {
			// If we can't convert to relative, use the absolute path
			relPaths[i] = absPath
		} else {
			relPaths[i] = relPath
		}
	}
	return relPaths
}

// ToRelPath converts a single absolute file path to a relative path from the project root.
func ToRelPath(projectRoot, absolutePath string) string {
	relPath, err := filepath.Rel(projectRoot, absolutePath)
	if err != nil {
		return absolutePath
	}
	return relPath
}

// IsStateError checks if an error is a StateError.
func IsStateError(err error) bool {
	_, ok := err.(*StateError)
	return ok
}

// AsStateError attempts to convert an error to a StateError.
// Returns nil if the error is not a StateError.
func AsStateError(err error) *StateError {
	if se, ok := err.(*StateError); ok {
		return se
	}
	return nil
}

// IsInvalidState checks if the error is an invalid state error.
func IsInvalidState(err error) bool {
	if se := AsStateError(err); se != nil {
		return se.Type == ErrInvalidState
	}
	return false
}

// IsInvalidTransition checks if the error is an invalid transition error.
func IsInvalidTransition(err error) bool {
	if se := AsStateError(err); se != nil {
		return se.Type == ErrInvalidTransition
	}
	return false
}

// IsHierarchyConstraint checks if the error is a hierarchy constraint error.
func IsHierarchyConstraint(err error) bool {
	if se := AsStateError(err); se != nil {
		return se.Type == ErrHierarchyConstraint
	}
	return false
}

// IsMissingFile checks if the error is a missing file error.
func IsMissingFile(err error) bool {
	if se := AsStateError(err); se != nil {
		return se.Type == ErrMissingFile
	}
	return false
}
