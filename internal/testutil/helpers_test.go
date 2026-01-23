package testutil

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/waynenilsen/crumbler/internal/crumb"
)

func TestAssertFileExists(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.Build()

	// Test with existing file
	existingFile := filepath.Join(root, crumb.CrumblerDir, crumb.ReadmeFile)
	AssertFileExists(t, existingFile) // Should not fail
}

func TestAssertFileNotExists(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.Build()

	// Test with non-existing file
	nonExistingFile := filepath.Join(root, "non-existing.txt")
	AssertFileNotExists(t, nonExistingFile) // Should not fail
}

func TestAssertDirExists(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.Build()

	// Test with existing directory
	existingDir := filepath.Join(root, crumb.CrumblerDir)
	AssertDirExists(t, existingDir) // Should not fail
}

func TestAssertDirNotExists(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.Build()

	// Test with non-existing directory
	nonExistingDir := filepath.Join(root, "non-existing-dir")
	AssertDirNotExists(t, nonExistingDir) // Should not fail
}

func TestAssertFileContent(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.WithCrumb("01-task", "# Test Content\n").Build()

	contentFile := filepath.Join(root, crumb.CrumblerDir, "01-task", crumb.ReadmeFile)
	AssertFileContent(t, contentFile, "# Test Content\n") // Should not fail
}

func TestAssertFileContains(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.WithCrumb("01-task", "# Test Readme\nThis is a test.\n").Build()

	contentFile := filepath.Join(root, crumb.CrumblerDir, "01-task", crumb.ReadmeFile)
	AssertFileContains(t, contentFile, "Test Readme") // Should not fail
	AssertFileContains(t, contentFile, "test")        // Should not fail
}

func TestAssertError(t *testing.T) {
	t.Parallel()

	err := errors.New("this is a test error with specific text")
	AssertError(t, err, "test error")
	AssertError(t, err, "specific text")
}

func TestAssertErrorIs(t *testing.T) {
	t.Parallel()

	target := os.ErrNotExist
	err := os.ErrNotExist
	AssertErrorIs(t, err, target)
}

func TestAssertNoError(t *testing.T) {
	t.Parallel()

	AssertNoError(t, nil)
}

func TestReadFile(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.WithCrumb("01-task", "# Custom Content\n").Build()

	contentFile := filepath.Join(root, crumb.CrumblerDir, "01-task", crumb.ReadmeFile)
	content := ReadFile(t, contentFile)

	if content != "# Custom Content\n" {
		t.Errorf("ReadFile returned unexpected content: %s", content)
	}
}

func TestListDir(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.
		WithCrumb("01-first", "").
		WithCrumb("02-second", "").
		Build()

	crumblerDir := filepath.Join(root, crumb.CrumblerDir)
	entries := ListDir(t, crumblerDir)

	// Should have 01-first, 02-second, README.md
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d: %v", len(entries), entries)
	}
}

func TestListDirRecursive(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.
		WithCrumb("01-phase", "Phase content").
		WithCrumb("01-phase/01-task", "Task content").
		Build()

	phasePath := filepath.Join(root, crumb.CrumblerDir, "01-phase")
	entries := ListDirRecursive(t, phasePath)

	// Should include: README.md, 01-task/, 01-task/README.md
	if len(entries) < 2 {
		t.Errorf("expected at least 2 entries, got %d: %v", len(entries), entries)
	}
}

func TestWriteFile(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.Build()

	testFile := filepath.Join(root, "test.txt")
	WriteFile(t, testFile, "test content")

	content := ReadFile(t, testFile)
	if content != "test content" {
		t.Errorf("WriteFile/ReadFile mismatch: got %s", content)
	}
}

func TestCreateFile(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.Build()

	testFile := filepath.Join(root, "empty.txt")
	CreateFile(t, testFile)

	AssertFileExists(t, testFile)
	content := ReadFile(t, testFile)
	if content != "" {
		t.Errorf("CreateFile should create empty file, got: %s", content)
	}
}

func TestRemoveFile(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.Build()

	testFile := filepath.Join(root, "to-remove.txt")
	WriteFile(t, testFile, "content")
	AssertFileExists(t, testFile)

	RemoveFile(t, testFile)
	AssertFileNotExists(t, testFile)
}

func TestCreateDir(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.Build()

	testDir := filepath.Join(root, "new-dir", "nested")
	CreateDir(t, testDir)

	AssertDirExists(t, testDir)
}
