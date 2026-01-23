package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/waynenilsen/crumbler/internal/crumb"
)

func TestNewTestProject(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.Build()

	// Verify .crumbler directory was created
	crumblerDir := filepath.Join(root, crumb.CrumblerDir)
	info, err := os.Stat(crumblerDir)
	if err != nil {
		t.Fatalf("expected .crumbler directory to exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected .crumbler to be a directory")
	}

	// Verify root README.md exists
	readmePath := filepath.Join(crumblerDir, crumb.ReadmeFile)
	if _, err := os.Stat(readmePath); err != nil {
		t.Fatalf("expected root README.md to exist: %v", err)
	}
}

func TestWithCrumb(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.
		WithCrumb("01-task", "Task content").
		WithCrumb("02-other", "").
		Build()

	// Verify first crumb
	task1Path := filepath.Join(root, crumb.CrumblerDir, "01-task")
	if _, err := os.Stat(task1Path); err != nil {
		t.Fatalf("expected crumb directory to exist: %v", err)
	}

	// Verify README content
	readmePath := filepath.Join(task1Path, crumb.ReadmeFile)
	content, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("failed to read README: %v", err)
	}
	if string(content) != "Task content" {
		t.Errorf("expected README content %q, got %q", "Task content", string(content))
	}

	// Verify second crumb with empty README
	task2Path := filepath.Join(root, crumb.CrumblerDir, "02-other")
	if _, err := os.Stat(task2Path); err != nil {
		t.Fatalf("expected crumb directory to exist: %v", err)
	}

	readme2Path := filepath.Join(task2Path, crumb.ReadmeFile)
	content2, err := os.ReadFile(readme2Path)
	if err != nil {
		t.Fatalf("failed to read README: %v", err)
	}
	if string(content2) != "" {
		t.Errorf("expected empty README, got %q", string(content2))
	}
}

func TestNestedCrumbs(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	root := builder.
		WithCrumb("01-setup", "Setup description").
		WithCrumb("01-setup/01-database", "Database description").
		Build()

	// Verify nested structure
	setupPath := filepath.Join(root, crumb.CrumblerDir, "01-setup")
	databasePath := filepath.Join(setupPath, "01-database")

	if _, err := os.Stat(setupPath); err != nil {
		t.Fatalf("expected setup directory to exist: %v", err)
	}
	if _, err := os.Stat(databasePath); err != nil {
		t.Fatalf("expected database directory to exist: %v", err)
	}

	// Verify database README
	readmePath := filepath.Join(databasePath, crumb.ReadmeFile)
	content, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("failed to read README: %v", err)
	}
	if string(content) != "Database description" {
		t.Errorf("expected README content %q, got %q", "Database description", string(content))
	}
}

func TestCrumbPath(t *testing.T) {
	t.Parallel()

	builder := NewTestProject(t)
	builder.WithCrumb("01-task", "").Build()

	expected := filepath.Join(builder.ProjectRoot(), crumb.CrumblerDir, "01-task")
	actual := builder.CrumbPath("01-task")

	if actual != expected {
		t.Errorf("CrumbPath() = %q, want %q", actual, expected)
	}
}

func TestParallelExecution(t *testing.T) {
	t.Parallel()

	// Run multiple subtests in parallel to verify isolation
	for i := 0; i < 5; i++ {
		i := i
		t.Run("parallel", func(t *testing.T) {
			t.Parallel()

			builder := NewTestProject(t)
			root := builder.
				WithCrumb("01-task", "Test content").
				Build()

			// Verify each test has its own isolated directory
			AssertDirExists(t, root)
			crumbPath := filepath.Join(root, crumb.CrumblerDir, "01-task")
			AssertDirExists(t, crumbPath)

			// Log for debugging
			info, _ := os.Stat(root)
			t.Logf("Parallel test %d: created directory %s", i, info.Name())
		})
	}
}

func TestCleanup(t *testing.T) {
	builder := NewTestProject(t)
	root := builder.Build()

	// Directory should exist during the test
	AssertDirExists(t, root)

	// The cleanup function is registered via t.Cleanup(), which will
	// be called after this test function completes
}
