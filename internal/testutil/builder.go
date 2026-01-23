// Package testutil provides test infrastructure for the crumbler CLI tool.
package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/waynenilsen/crumbler/internal/crumb"
)

// TestProjectBuilder provides a fluent API for building test project structures.
type TestProjectBuilder struct {
	t           *testing.T
	projectRoot string
	crumbs      map[string]*crumbConfig
	cleaned     bool
}

// crumbConfig holds configuration for a crumb in the test project.
type crumbConfig struct {
	readme string
}

// NewTestProject creates a new TestProjectBuilder for building test project structures.
// It creates a unique test directory under .test/ with format: .test/test-{timestamp}-{random}/
func NewTestProject(t *testing.T) *TestProjectBuilder {
	t.Helper()

	// Generate unique test directory name
	timestamp := time.Now().UnixNano()
	random := GenerateRandomString(6)
	testDirName := fmt.Sprintf("test-%d-%s", timestamp, random)

	// Create test directory path relative to project root
	// Find project root by looking for go.mod
	projectRoot := findProjectRoot(t)
	testRoot := filepath.Join(projectRoot, ".test", testDirName)

	// Ensure .test directory exists
	testBaseDir := filepath.Join(projectRoot, ".test")
	if err := os.MkdirAll(testBaseDir, 0755); err != nil {
		t.Fatalf("failed to create .test directory: %v", err)
	}

	// Create the unique test directory
	if err := os.MkdirAll(testRoot, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	builder := &TestProjectBuilder{
		t:           t,
		projectRoot: testRoot,
		crumbs:      make(map[string]*crumbConfig),
	}

	// Register cleanup
	t.Cleanup(func() {
		if !builder.cleaned {
			builder.cleanup()
		}
	})

	return builder
}

// findProjectRoot finds the project root by looking for go.mod
func findProjectRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find project root (no go.mod found)")
		}
		dir = parent
	}
}

// cleanup removes the test project directory.
func (b *TestProjectBuilder) cleanup() {
	b.cleaned = true
	if err := os.RemoveAll(b.projectRoot); err != nil {
		b.t.Errorf("failed to cleanup test directory: %v", err)
	}
}

// WithCrumb adds a crumb at the specified relative path with optional README content.
	// Path should be relative to .crumbler/ (e.g., "01-task" or "01-setup/01-database")
func (b *TestProjectBuilder) WithCrumb(relPath, readme string) *TestProjectBuilder {
	b.crumbs[relPath] = &crumbConfig{readme: readme}
	return b
}

// Build creates the project structure and returns the project root path.
func (b *TestProjectBuilder) Build() string {
	b.t.Helper()

	// Create .crumbler directory
	crumblerDir := filepath.Join(b.projectRoot, crumb.CrumblerDir)
	if err := os.MkdirAll(crumblerDir, 0755); err != nil {
		b.t.Fatalf("failed to create .crumbler directory: %v", err)
	}

	// Create root README.md
	rootReadme := filepath.Join(crumblerDir, crumb.ReadmeFile)
	if err := os.WriteFile(rootReadme, []byte{}, 0644); err != nil {
		b.t.Fatalf("failed to create root README.md: %v", err)
	}

	// Create each crumb
	for relPath, config := range b.crumbs {
		crumbPath := filepath.Join(crumblerDir, relPath)
		if err := os.MkdirAll(crumbPath, 0755); err != nil {
			b.t.Fatalf("failed to create crumb directory %s: %v", relPath, err)
		}

		readmePath := filepath.Join(crumbPath, crumb.ReadmeFile)
		if err := os.WriteFile(readmePath, []byte(config.readme), 0644); err != nil {
			b.t.Fatalf("failed to create README.md for %s: %v", relPath, err)
		}
	}

	return b.projectRoot
}

// ProjectRoot returns the test project root path.
func (b *TestProjectBuilder) ProjectRoot() string {
	return b.projectRoot
}

// CrumblerDir returns the path to the .crumbler directory.
func (b *TestProjectBuilder) CrumblerDir() string {
	return filepath.Join(b.projectRoot, crumb.CrumblerDir)
}

// CrumbPath returns the full path to a crumb given its relative path.
func (b *TestProjectBuilder) CrumbPath(relPath string) string {
	return filepath.Join(b.projectRoot, crumb.CrumblerDir, relPath)
}
