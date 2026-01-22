package testutil

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// AssertFileExists asserts that a file exists at the given path.
func AssertFileExists(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		t.Errorf("expected file to exist: %s", path)
		return
	}
	if err != nil {
		t.Errorf("error checking file existence: %s: %v", path, err)
		return
	}
	if info.IsDir() {
		t.Errorf("expected file but got directory: %s", path)
	}
}

// AssertFileNotExists asserts that a file does not exist at the given path.
func AssertFileNotExists(t *testing.T, path string) {
	t.Helper()

	_, err := os.Stat(path)
	if err == nil {
		t.Errorf("expected file to not exist: %s", path)
		return
	}
	if !os.IsNotExist(err) {
		t.Errorf("error checking file non-existence: %s: %v", path, err)
	}
}

// AssertDirExists asserts that a directory exists at the given path.
func AssertDirExists(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		t.Errorf("expected directory to exist: %s", path)
		return
	}
	if err != nil {
		t.Errorf("error checking directory existence: %s: %v", path, err)
		return
	}
	if !info.IsDir() {
		t.Errorf("expected directory but got file: %s", path)
	}
}

// AssertDirNotExists asserts that a directory does not exist at the given path.
func AssertDirNotExists(t *testing.T, path string) {
	t.Helper()

	_, err := os.Stat(path)
	if err == nil {
		t.Errorf("expected directory to not exist: %s", path)
		return
	}
	if !os.IsNotExist(err) {
		t.Errorf("error checking directory non-existence: %s: %v", path, err)
	}
}

// AssertFileContent asserts that a file contains the expected content.
func AssertFileContent(t *testing.T, path, expected string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("failed to read file %s: %v", path, err)
		return
	}

	if string(content) != expected {
		t.Errorf("file content mismatch in %s:\nexpected:\n%s\n\ngot:\n%s", path, expected, string(content))
	}
}

// AssertFileContains asserts that a file contains the expected substring.
func AssertFileContains(t *testing.T, path, substring string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("failed to read file %s: %v", path, err)
		return
	}

	if !strings.Contains(string(content), substring) {
		t.Errorf("file %s does not contain expected substring:\nexpected to contain: %s\nactual content:\n%s", path, substring, string(content))
	}
}

// AssertStatus checks the status of a phase, sprint, or ticket by verifying status file existence.
// expected should be "open", "closed", or "done".
func AssertStatus(t *testing.T, path string, expected string) {
	t.Helper()

	validStatuses := map[string][]string{
		"open":   {"closed", "done"},
		"closed": {"open", "done"},
		"done":   {"open", "closed"},
	}

	invalidStatuses, ok := validStatuses[expected]
	if !ok {
		t.Errorf("invalid expected status: %s (must be open, closed, or done)", expected)
		return
	}

	// Check expected status file exists
	expectedFile := filepath.Join(path, expected)
	info, err := os.Stat(expectedFile)
	if os.IsNotExist(err) {
		t.Errorf("expected status file to exist: %s", expectedFile)
		return
	}
	if err != nil {
		t.Errorf("error checking status file: %s: %v", expectedFile, err)
		return
	}
	if info.IsDir() {
		t.Errorf("status file should not be a directory: %s", expectedFile)
		return
	}

	// Check that invalid status files do not exist
	for _, invalid := range invalidStatuses {
		invalidFile := filepath.Join(path, invalid)
		if _, err := os.Stat(invalidFile); err == nil {
			t.Errorf("conflicting status file exists: %s (expected only %s)", invalidFile, expected)
		}
	}
}

// AssertGoalStatus checks the status of a goal by verifying the status file in the goal directory.
// goalPath should be the path to the goal directory (e.g., .../goals/0001-goal).
// expected should be "open" or "closed".
func AssertGoalStatus(t *testing.T, goalPath string, expected string) {
	t.Helper()

	if expected != "open" && expected != "closed" {
		t.Errorf("invalid expected goal status: %s (must be open or closed)", expected)
		return
	}

	// Check expected status file exists
	expectedFile := filepath.Join(goalPath, expected)
	info, err := os.Stat(expectedFile)
	if os.IsNotExist(err) {
		t.Errorf("expected goal status file to exist: %s", expectedFile)
		return
	}
	if err != nil {
		t.Errorf("error checking goal status file: %s: %v", expectedFile, err)
		return
	}
	if info.IsDir() {
		t.Errorf("goal status file should not be a directory: %s", expectedFile)
		return
	}

	// Check that conflicting status file does not exist
	conflicting := "closed"
	if expected == "closed" {
		conflicting = "open"
	}
	conflictingFile := filepath.Join(goalPath, conflicting)
	if _, err := os.Stat(conflictingFile); err == nil {
		t.Errorf("conflicting goal status file exists: %s (expected only %s)", conflictingFile, expected)
	}

	// Check that name file exists
	nameFile := filepath.Join(goalPath, "name")
	if _, err := os.Stat(nameFile); os.IsNotExist(err) {
		t.Errorf("goal name file does not exist: %s", nameFile)
	}
}

// AssertGoalName checks that a goal has the expected name.
func AssertGoalName(t *testing.T, goalPath, expectedName string) {
	t.Helper()

	nameFile := filepath.Join(goalPath, "name")
	content, err := os.ReadFile(nameFile)
	if err != nil {
		t.Errorf("failed to read goal name file %s: %v", nameFile, err)
		return
	}

	if string(content) != expectedName {
		t.Errorf("goal name mismatch in %s:\nexpected: %s\ngot: %s", nameFile, expectedName, string(content))
	}
}

// AssertError asserts that an error matches the expected type.
// expectedType can be a string that the error message should contain,
// or a specific error type name.
func AssertError(t *testing.T, err error, expectedType string) {
	t.Helper()

	if err == nil {
		t.Errorf("expected error containing %q, got nil", expectedType)
		return
	}

	if !strings.Contains(err.Error(), expectedType) {
		t.Errorf("expected error containing %q, got: %v", expectedType, err)
	}
}

// AssertErrorIs asserts that an error matches a specific error using errors.Is.
func AssertErrorIs(t *testing.T, err, target error) {
	t.Helper()

	if !errors.Is(err, target) {
		t.Errorf("expected error %v, got: %v", target, err)
	}
}

// AssertNoError asserts that an error is nil.
func AssertNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// ReadFile reads and returns the content of a file.
// Fails the test if the file cannot be read.
func ReadFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file %s: %v", path, err)
	}
	return string(content)
}

// ListDir lists the contents of a directory and returns file/directory names.
// Fails the test if the directory cannot be read.
func ListDir(t *testing.T, path string) []string {
	t.Helper()

	entries, err := os.ReadDir(path)
	if err != nil {
		t.Fatalf("failed to read directory %s: %v", path, err)
	}

	names := make([]string, len(entries))
	for i, entry := range entries {
		names[i] = entry.Name()
	}
	return names
}

// ListDirRecursive lists all files and directories recursively under a path.
// Returns paths relative to the given path.
func ListDirRecursive(t *testing.T, path string) []string {
	t.Helper()

	var paths []string
	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(path, p)
		if err != nil {
			return err
		}
		if rel != "." {
			paths = append(paths, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk directory %s: %v", path, err)
	}
	return paths
}

// WriteFile writes content to a file.
// Fails the test if the file cannot be written.
func WriteFile(t *testing.T, path, content string) {
	t.Helper()

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create directory %s: %v", dir, err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}

// CreateFile creates an empty file at the given path.
// Fails the test if the file cannot be created.
func CreateFile(t *testing.T, path string) {
	t.Helper()
	WriteFile(t, path, "")
}

// RemoveFile removes a file at the given path.
// Fails the test if the file cannot be removed.
func RemoveFile(t *testing.T, path string) {
	t.Helper()

	if err := os.Remove(path); err != nil {
		t.Fatalf("failed to remove file %s: %v", path, err)
	}
}

// CreateDir creates a directory at the given path.
// Fails the test if the directory cannot be created.
func CreateDir(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("failed to create directory %s: %v", path, err)
	}
}

// TouchStatusFile creates an empty status file (open, closed, or done).
func TouchStatusFile(t *testing.T, parentPath, status string) {
	t.Helper()
	CreateFile(t, filepath.Join(parentPath, status))
}

// SetStatus sets the status of a phase, sprint, or ticket by managing status files.
// Removes conflicting status files and creates the new status file.
func SetStatus(t *testing.T, path, status string) {
	t.Helper()

	allStatuses := []string{"open", "closed", "done"}
	for _, s := range allStatuses {
		statusFile := filepath.Join(path, s)
		if _, err := os.Stat(statusFile); err == nil {
			if err := os.Remove(statusFile); err != nil {
				t.Fatalf("failed to remove status file %s: %v", statusFile, err)
			}
		}
	}

	CreateFile(t, filepath.Join(path, status))
}

// CountGoals counts the number of goal directories in a goals directory.
func CountGoals(t *testing.T, parentPath string) int {
	t.Helper()

	goalsDir := filepath.Join(parentPath, "goals")
	entries, err := os.ReadDir(goalsDir)
	if os.IsNotExist(err) {
		return 0
	}
	if err != nil {
		t.Fatalf("failed to read goals directory %s: %v", goalsDir, err)
	}

	count := 0
	for _, entry := range entries {
		if entry.IsDir() && strings.HasSuffix(entry.Name(), "-goal") {
			count++
		}
	}
	return count
}

// GetOpenGoals returns the paths of all open goals in a parent directory.
func GetOpenGoals(t *testing.T, parentPath string) []string {
	t.Helper()

	goalsDir := filepath.Join(parentPath, "goals")
	entries, err := os.ReadDir(goalsDir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatalf("failed to read goals directory %s: %v", goalsDir, err)
	}

	var openGoals []string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasSuffix(entry.Name(), "-goal") {
			goalPath := filepath.Join(goalsDir, entry.Name())
			openFile := filepath.Join(goalPath, "open")
			if _, err := os.Stat(openFile); err == nil {
				openGoals = append(openGoals, goalPath)
			}
		}
	}
	return openGoals
}

// GetClosedGoals returns the paths of all closed goals in a parent directory.
func GetClosedGoals(t *testing.T, parentPath string) []string {
	t.Helper()

	goalsDir := filepath.Join(parentPath, "goals")
	entries, err := os.ReadDir(goalsDir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatalf("failed to read goals directory %s: %v", goalsDir, err)
	}

	var closedGoals []string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasSuffix(entry.Name(), "-goal") {
			goalPath := filepath.Join(goalsDir, entry.Name())
			closedFile := filepath.Join(goalPath, "closed")
			if _, err := os.Stat(closedFile); err == nil {
				closedGoals = append(closedGoals, goalPath)
			}
		}
	}
	return closedGoals
}

// AssertAllGoalsClosed asserts that all goals in a parent directory are closed.
func AssertAllGoalsClosed(t *testing.T, parentPath string) {
	t.Helper()

	openGoals := GetOpenGoals(t, parentPath)
	if len(openGoals) > 0 {
		t.Errorf("expected all goals to be closed, but found open goals: %v", openGoals)
	}
}

// AssertGoalCount asserts that the parent has the expected number of goals.
func AssertGoalCount(t *testing.T, parentPath string, expected int) {
	t.Helper()

	actual := CountGoals(t, parentPath)
	if actual != expected {
		t.Errorf("expected %d goals in %s, got %d", expected, parentPath, actual)
	}
}
