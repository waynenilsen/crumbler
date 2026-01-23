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

func TestGetState(t *testing.T) {
	t.Parallel()

	t.Run("done when no .crumbler", func(t *testing.T) {
		// No .crumbler directory -> DONE
		root := t.TempDir()

		state, err := GetState(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state != StateDone {
			t.Errorf("state = %q, want %q", state, StateDone)
		}
	})

	t.Run("decompose when root crumb empty", func(t *testing.T) {
		// .crumbler exists with empty README -> DECOMPOSE (root crumb is current)
		root := setupTestProject(t)

		state, err := GetState(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state != StateDecompose {
			t.Errorf("state = %q, want %q", state, StateDecompose)
		}
	})

	t.Run("decompose when README empty", func(t *testing.T) {
		root := setupTestProject(t)
		createCrumb(t, filepath.Join(root, crumb.CrumblerDir, "01-task"), "")

		state, err := GetState(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state != StateDecompose {
			t.Errorf("state = %q, want %q", state, StateDecompose)
		}
	})

	t.Run("execute when README has content", func(t *testing.T) {
		root := setupTestProject(t)
		createCrumb(t, filepath.Join(root, crumb.CrumblerDir, "01-task"), "Do this task")

		state, err := GetState(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state != StateExecute {
			t.Errorf("state = %q, want %q", state, StateExecute)
		}
	})

	t.Run("whitespace-only README is decompose", func(t *testing.T) {
		root := setupTestProject(t)
		createCrumb(t, filepath.Join(root, crumb.CrumblerDir, "01-task"), "   \n\t  \n")

		state, err := GetState(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state != StateDecompose {
			t.Errorf("state = %q, want %q", state, StateDecompose)
		}
	})
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
		if !strings.Contains(prompt, "STATE: DONE") {
			t.Error("expected STATE: DONE in prompt")
		}
	})

	t.Run("decompose prompt", func(t *testing.T) {
		root := setupTestProject(t)
		createCrumb(t, filepath.Join(root, crumb.CrumblerDir, "01-task"), "")

		prompt, err := GeneratePrompt(root, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(prompt, "STATE: DECOMPOSE") {
			t.Error("expected STATE: DECOMPOSE in prompt")
		}
		if !strings.Contains(prompt, "crumbler create") {
			t.Error("expected create command in prompt")
		}
	})

	t.Run("execute prompt", func(t *testing.T) {
		root := setupTestProject(t)
		createCrumb(t, filepath.Join(root, crumb.CrumblerDir, "01-task"), "Do the thing")

		prompt, err := GeneratePrompt(root, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(prompt, "STATE: EXECUTE") {
			t.Error("expected STATE: EXECUTE in prompt")
		}
		if !strings.Contains(prompt, "crumbler delete") {
			t.Error("expected delete command in prompt")
		}
		if !strings.Contains(prompt, "Do the thing") {
			t.Error("expected README content in prompt")
		}
	})

	t.Run("state only mode", func(t *testing.T) {
		root := setupTestProject(t)
		createCrumb(t, filepath.Join(root, crumb.CrumblerDir, "01-task"), "Do it")

		prompt, err := GeneratePrompt(root, &Config{StateOnly: true})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if prompt != "EXECUTE" {
			t.Errorf("prompt = %q, want %q", prompt, "EXECUTE")
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
}

func TestFormatTree(t *testing.T) {
	t.Parallel()

	t.Run("single crumb", func(t *testing.T) {
		root := &crumb.Crumb{
			RelPath: ".crumbler",
			IsLeaf:  true,
		}

		tree := FormatTree(root, "", false)
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
					Name:    "phase1",
					RelPath: ".crumbler/01-phase1",
					Children: []crumb.Crumb{
						{
							ID:      "01",
							Name:    "task",
							RelPath: ".crumbler/01-phase1/01-task",
							IsLeaf:  true,
						},
					},
				},
				{
					ID:      "02",
					Name:    "phase2",
					RelPath: ".crumbler/02-phase2",
					IsLeaf:  true,
				},
			},
		}

		tree := FormatTree(root, "", false)
		if !strings.Contains(tree, "01-phase1/") {
			t.Error("expected phase1 in tree")
		}
		if !strings.Contains(tree, "01-task/") {
			t.Error("expected task in tree")
		}
		if !strings.Contains(tree, "02-phase2/") {
			t.Error("expected phase2 in tree")
		}
	})
}

func TestWorkflow(t *testing.T) {
	t.Parallel()

	t.Run("full workflow", func(t *testing.T) {
		root := setupTestProject(t)

		// Initially, .crumbler exists with empty README -> root is current -> DECOMPOSE
		state, _ := GetState(root)
		if state != StateDecompose {
			t.Errorf("initial state = %q, want DECOMPOSE", state)
		}

		// Create a child crumb
		_, err := crumb.Create(root, "Task 1")
		if err != nil {
			t.Fatalf("failed to create crumb: %v", err)
		}

		// Now child is current and empty -> DECOMPOSE
		state, _ = GetState(root)
		if state != StateDecompose {
			t.Errorf("after create state = %q, want DECOMPOSE", state)
		}

		// Get current crumb and add content to its README
		current, _ := crumb.GetCurrent(root)
		readmePath := filepath.Join(current.Path, crumb.ReadmeFile)
		os.WriteFile(readmePath, []byte("Do this task"), 0644)

		// Now it should be EXECUTE
		state, _ = GetState(root)
		if state != StateExecute {
			t.Errorf("after content state = %q, want EXECUTE", state)
		}

		// Delete the child crumb
		crumb.Delete(root)

		// Root crumb is now current again (with its original empty README) -> DECOMPOSE
		state, _ = GetState(root)
		if state != StateDecompose {
			t.Errorf("after delete child state = %q, want DECOMPOSE", state)
		}

		// Delete the root crumb (removes .crumbler entirely)
		crumb.Delete(root)

		// No .crumbler -> DONE
		state, _ = GetState(root)
		if state != StateDone {
			t.Errorf("after delete root state = %q, want DONE", state)
		}
	})
}
