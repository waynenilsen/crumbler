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

// WriteFile writes content to a file.
// Fails the test if the file cannot be written.
func WriteFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}

// CreateFile creates an empty file at the given path.
// Fails the test if the file cannot be created.
func CreateFile(t *testing.T, path string) {
	t.Helper()

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create file %s: %v", path, err)
	}
	file.Close()
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
