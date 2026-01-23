package prompt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/waynenilsen/crumbler/internal/crumb"
)

// setupTestProject creates a test project with .crumbler directory.
func setupTestProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	crumblerDir := filepath.Join(dir, crumb.CrumblerDir)
	if err := os.MkdirAll(crumblerDir, 0755); err != nil {
		t.Fatalf("failed to create .crumbler: %v", err)
	}
	readmePath := filepath.Join(crumblerDir, crumb.ReadmeFile)
	if err := os.WriteFile(readmePath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create README.md: %v", err)
	}
	return dir
}

// createCrumb creates a crumb with optional README content.
func createCrumb(t *testing.T, path, readmeContent string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("failed to create crumb dir: %v", err)
	}
	readmePath := filepath.Join(path, crumb.ReadmeFile)
	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		t.Fatalf("failed to create README.md: %v", err)
	}
}

func TestGeneratePrompt(t *testing.T) {
	t.Parallel()

	t.Run("done prompt", func(t *testing.T) {
		// No .crumbler directory -> DONE prompt
		root := t.TempDir()

		prompt, err := GeneratePrompt(root, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(prompt, "DONE") {
			t.Error("expected DONE in prompt")
		}
	})

	t.Run("empty README shows warning", func(t *testing.T) {
		root := setupTestProject(t)
		createCrumb(t, filepath.Join(root, crumb.CrumblerDir, "01-task"), "")

		prompt, err := GeneratePrompt(root, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(prompt, "README is empty") {
			t.Error("expected empty README warning in prompt")
		}
		if !strings.Contains(prompt, "crumbler create") {
			t.Error("expected create command in prompt")
		}
	})

	t.Run("README with content shows content", func(t *testing.T) {
		root := setupTestProject(t)
		createCrumb(t, filepath.Join(root, crumb.CrumblerDir, "01-task"), "Do the thing")

		prompt, err := GeneratePrompt(root, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(prompt, "crumbler delete") {
			t.Error("expected delete command in prompt")
		}
		if !strings.Contains(prompt, "Do the thing") {
			t.Error("expected README content in prompt")
		}
	})

	t.Run("no preamble", func(t *testing.T) {
		root := setupTestProject(t)
		createCrumb(t, filepath.Join(root, crumb.CrumblerDir, "01-task"), "Task")

		prompt, err := GeneratePrompt(root, &Config{NoPreamble: true})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(prompt, "# Crumbler Agent Instructions") {
			t.Error("should not contain preamble")
		}
	})

	t.Run("no postamble", func(t *testing.T) {
		root := setupTestProject(t)
		createCrumb(t, filepath.Join(root, crumb.CrumblerDir, "01-task"), "Task")

		prompt, err := GeneratePrompt(root, &Config{NoPostamble: true})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(prompt, "## Next Steps") {
			t.Error("should not contain postamble")
		}
	})

	t.Run("minimal mode", func(t *testing.T) {
		root := setupTestProject(t)
		createCrumb(t, filepath.Join(root, crumb.CrumblerDir, "01-task"), "Task")

		prompt, err := GeneratePrompt(root, &Config{Minimal: true})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Minimal preamble is much shorter
		if strings.Contains(prompt, "# Crumbler Agent Instructions") {
			t.Error("should use minimal preamble")
		}
		if !strings.Contains(prompt, "# Crumbler") {
			t.Error("should contain minimal preamble")
		}
	})

	t.Run("no context", func(t *testing.T) {
		root := setupTestProject(t)
		createCrumb(t, filepath.Join(root, crumb.CrumblerDir, "01-task"), "Task content")

		prompt, err := GeneratePrompt(root, &Config{NoContext: true})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(prompt, "## Current Crumb") {
			t.Error("should not contain context section")
		}
	})

	t.Run("parent context shown for empty README", func(t *testing.T) {
		root := setupTestProject(t)
		// Create parent with content
		parentPath := filepath.Join(root, crumb.CrumblerDir, "01-parent")
		createCrumb(t, parentPath, "Parent task instructions")
		// Create child with empty README
		createCrumb(t, filepath.Join(parentPath, "01-child"), "")

		prompt, err := GeneratePrompt(root, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(prompt, "Parent Context") {
			t.Error("expected parent context in prompt")
		}
		if !strings.Contains(prompt, "Parent task instructions") {
			t.Error("expected parent README content in prompt")
		}
	})
}

func TestFormatTree(t *testing.T) {
	t.Parallel()

	t.Run("single crumb", func(t *testing.T) {
		root := &crumb.Crumb{
			RelPath: ".crumbler",
			IsLeaf:  true,
		}

		tree := FormatTreeWithCurrent(root, "", false, ".crumbler")
		if !strings.Contains(tree, ".crumbler/") {
			t.Error("expected .crumbler/ in tree")
		}
		if !strings.Contains(tree, "← current") {
			t.Error("expected current marker for leaf")
		}
	})

	t.Run("nested structure", func(t *testing.T) {
		root := &crumb.Crumb{
			RelPath: ".crumbler",
			Children: []crumb.Crumb{
				{
					ID:      "01",
					Name:    "setup",
					RelPath: ".crumbler/01-setup",
					Children: []crumb.Crumb{
						{
							ID:      "01",
							Name:    "database",
							RelPath: ".crumbler/01-setup/01-database",
							IsLeaf:  true,
						},
					},
				},
				{
					ID:      "02",
					Name:    "features",
					RelPath: ".crumbler/02-features",
					IsLeaf:  true,
				},
			},
		}

		tree := FormatTree(root, "", false)
		if !strings.Contains(tree, "01-setup/") {
			t.Error("expected setup in tree")
		}
		if !strings.Contains(tree, "01-database/") {
			t.Error("expected database in tree")
		}
		if !strings.Contains(tree, "02-features/") {
			t.Error("expected features in tree")
		}
	})
}

func TestWorkflow(t *testing.T) {
	t.Parallel()

	t.Run("full workflow", func(t *testing.T) {
		root := setupTestProject(t)

		// Initially, .crumbler exists with empty README
		prompt, err := GeneratePrompt(root, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(prompt, "README is empty") {
			t.Error("expected empty README warning initially")
		}

		// Create a child crumb
		_, err = crumb.Create(root, "Task 1")
		if err != nil {
			t.Fatalf("failed to create crumb: %v", err)
		}

		// Now child is current and empty
		prompt, _ = GeneratePrompt(root, nil)
		if !strings.Contains(prompt, "README is empty") {
			t.Error("expected empty README warning after create")
		}

		// Get current crumb and add content to its README
		current, _ := crumb.GetCurrent(root)
		readmePath := filepath.Join(current.Path, crumb.ReadmeFile)
		os.WriteFile(readmePath, []byte("Do this task"), 0644)

		// Now README has content
		prompt, _ = GeneratePrompt(root, nil)
		if strings.Contains(prompt, "README is empty") {
			t.Error("should not show empty warning after content added")
		}
		if !strings.Contains(prompt, "Do this task") {
			t.Error("expected README content in prompt")
		}

		// Delete the child crumb
		crumb.Delete(root)

		// Root crumb is now current again (with its original empty README)
		prompt, _ = GeneratePrompt(root, nil)
		if !strings.Contains(prompt, "README is empty") {
			t.Error("expected empty README warning after child delete")
		}

		// Delete the root crumb (removes .crumbler entirely)
		crumb.Delete(root)

		// No .crumbler -> DONE
		prompt, _ = GeneratePrompt(root, nil)
		if !strings.Contains(prompt, "DONE") {
			t.Error("expected DONE in prompt after all crumbs deleted")
		}
	})
}
