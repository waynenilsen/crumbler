package crumb

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// setupTestProject creates a temporary test project with .crumbler directory.
func setupTestProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	crumblerDir := filepath.Join(dir, CrumblerDir)
	if err := os.MkdirAll(crumblerDir, 0755); err != nil {
		t.Fatalf("failed to create .crumbler: %v", err)
	}
	// Create root README.md
	readmePath := filepath.Join(crumblerDir, ReadmeFile)
	if err := os.WriteFile(readmePath, []byte("# Project"), 0644); err != nil {
		t.Fatalf("failed to create README.md: %v", err)
	}
	return dir
}

// createCrumb creates a crumb directory with README.md.
func createCrumb(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("failed to create crumb dir: %v", err)
	}
	readmePath := filepath.Join(path, ReadmeFile)
	if err := os.WriteFile(readmePath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create README.md: %v", err)
	}
}

func TestKebabify(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected string
	}{
		{"Add User Auth", "add-user-auth"},
		{"Setup Database!", "setup-database"},
		{"UPPERCASE", "uppercase"},
		{"mixed_CASE_name", "mixed-case-name"},
		{"  spaces  ", "spaces"},
		{"multiple---hyphens", "multiple-hyphens"},
		{"special@#$chars", "specialchars"},
		{"123 numbers", "123-numbers"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := Kebabify(tt.input)
			if result != tt.expected {
				t.Errorf("Kebabify(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseDir(t *testing.T) {
	t.Parallel()
	tests := []struct {
		dirname      string
		expectedID   string
		expectedName string
	}{
		{"01-add-auth", "01", "add-auth"},
		{"10-final", "10", "final"},
		{"01", "01", ""},
		{"invalid", "", "invalid"},
		{"1-short", "", "1-short"},
		{"00-zero", "", "00-zero"},       // 00 is below MinID
		{"11-over", "", "11-over"},       // 11 is above MaxID
		{"01-multi-word-name", "01", "multi-word-name"},
	}

	for _, tt := range tests {
		t.Run(tt.dirname, func(t *testing.T) {
			id, name := ParseDir(tt.dirname)
			if id != tt.expectedID || name != tt.expectedName {
				t.Errorf("ParseDir(%q) = (%q, %q), want (%q, %q)",
					tt.dirname, id, name, tt.expectedID, tt.expectedName)
			}
		})
	}
}

func TestFormatDir(t *testing.T) {
	t.Parallel()
	tests := []struct {
		id       string
		name     string
		expected string
	}{
		{"01", "add-auth", "01-add-auth"},
		{"10", "final", "10-final"},
		{"01", "", "01"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatDir(tt.id, tt.name)
			if result != tt.expected {
				t.Errorf("FormatDir(%q, %q) = %q, want %q",
					tt.id, tt.name, result, tt.expected)
			}
		})
	}
}

func TestNextID(t *testing.T) {
	t.Parallel()

	t.Run("empty directory", func(t *testing.T) {
		dir := t.TempDir()
		id, err := NextID(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != "01" {
			t.Errorf("NextID() = %q, want %q", id, "01")
		}
	})

	t.Run("non-existent directory", func(t *testing.T) {
		id, err := NextID("/nonexistent/path")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != "01" {
			t.Errorf("NextID() = %q, want %q", id, "01")
		}
	})

	t.Run("with existing crumbs", func(t *testing.T) {
		dir := t.TempDir()
		createCrumb(t, filepath.Join(dir, "01-first"))
		createCrumb(t, filepath.Join(dir, "02-second"))

		id, err := NextID(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != "03" {
			t.Errorf("NextID() = %q, want %q", id, "03")
		}
	})

	t.Run("with gap in IDs", func(t *testing.T) {
		dir := t.TempDir()
		createCrumb(t, filepath.Join(dir, "01-first"))
		createCrumb(t, filepath.Join(dir, "03-third"))

		id, err := NextID(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != "02" {
			t.Errorf("NextID() = %q, want %q (should fill gap)", id, "02")
		}
	})

	t.Run("directory full", func(t *testing.T) {
		dir := t.TempDir()
		for i := 1; i <= 10; i++ {
			createCrumb(t, filepath.Join(dir, FormatDir(fmt.Sprintf("%02d", i), "crumb")))
		}

		_, err := NextID(dir)
		if err == nil {
			t.Error("expected error for full directory")
		}
	})
}

func TestGetCurrent(t *testing.T) {
	t.Parallel()

	t.Run("no children returns root crumb", func(t *testing.T) {
		root := setupTestProject(t)
		current, err := GetCurrent(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if current == nil {
			t.Fatal("expected root crumb when no children, got nil")
		}
		if current.ID != "" {
			t.Errorf("root crumb ID should be empty, got %q", current.ID)
		}
		if current.RelPath != CrumblerDir {
			t.Errorf("root crumb RelPath = %q, want %q", current.RelPath, CrumblerDir)
		}
	})

	t.Run("single crumb", func(t *testing.T) {
		root := setupTestProject(t)
		createCrumb(t, filepath.Join(root, CrumblerDir, "01-task"))

		current, err := GetCurrent(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if current == nil {
			t.Fatal("expected crumb, got nil")
		}
		if current.ID != "01" {
			t.Errorf("ID = %q, want %q", current.ID, "01")
		}
		if current.Name != "task" {
			t.Errorf("Name = %q, want %q", current.Name, "task")
		}
	})

	t.Run("nested crumbs returns deepest", func(t *testing.T) {
		root := setupTestProject(t)
		createCrumb(t, filepath.Join(root, CrumblerDir, "01-setup"))
		createCrumb(t, filepath.Join(root, CrumblerDir, "01-setup", "01-database"))
		createCrumb(t, filepath.Join(root, CrumblerDir, "01-setup", "01-database", "01-migrations"))

		current, err := GetCurrent(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if current == nil {
			t.Fatal("expected crumb, got nil")
		}
		if current.Name != "migrations" {
			t.Errorf("Name = %q, want %q", current.Name, "migrations")
		}
	})

	t.Run("returns first by ID", func(t *testing.T) {
		root := setupTestProject(t)
		createCrumb(t, filepath.Join(root, CrumblerDir, "02-second"))
		createCrumb(t, filepath.Join(root, CrumblerDir, "01-first"))

		current, err := GetCurrent(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if current == nil {
			t.Fatal("expected crumb, got nil")
		}
		if current.Name != "first" {
			t.Errorf("Name = %q, want %q", current.Name, "first")
		}
	})

	t.Run("no .crumbler returns nil", func(t *testing.T) {
		dir := t.TempDir()
		current, err := GetCurrent(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if current != nil {
			t.Error("expected nil when no .crumbler exists")
		}
	})
}

func TestCreate(t *testing.T) {
	t.Parallel()

	t.Run("create first crumb", func(t *testing.T) {
		root := setupTestProject(t)
		path, err := Create(root, "First Task")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedName := "01-first-task"
		if filepath.Base(path) != expectedName {
			t.Errorf("created dir = %q, want %q", filepath.Base(path), expectedName)
		}

		// Verify README.md exists
		readmePath := filepath.Join(path, ReadmeFile)
		if _, err := os.Stat(readmePath); os.IsNotExist(err) {
			t.Error("README.md not created")
		}
	})

	t.Run("create nested crumb", func(t *testing.T) {
		root := setupTestProject(t)
		createCrumb(t, filepath.Join(root, CrumblerDir, "01-parent"))

		path, err := Create(root, "Child Task")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should be created under 01-parent
		if filepath.Base(filepath.Dir(path)) != "01-parent" {
			t.Errorf("parent = %q, want %q", filepath.Base(filepath.Dir(path)), "01-parent")
		}
	})

	t.Run("auto-increments ID", func(t *testing.T) {
		root := setupTestProject(t)
		// Create two sibling crumbs at root level
		// First, create with no current crumb (root becomes current)
		path1, err := Create(root, "First Task")
		if err != nil {
			t.Fatalf("unexpected error creating first: %v", err)
		}
		if filepath.Base(path1) != "01-first-task" {
			t.Errorf("first created dir = %q, want %q", filepath.Base(path1), "01-first-task")
		}

		// Now 01-first-task is the current crumb (deepest leaf)
		// Creating "Second Task" should create UNDER 01-first-task
		path2, err := Create(root, "Second Task")
		if err != nil {
			t.Fatalf("unexpected error creating second: %v", err)
		}

		// Should be 01-first-task/01-second-task (nested, not sibling)
		if filepath.Base(path2) != "01-second-task" {
			t.Errorf("second created dir = %q, want %q", filepath.Base(path2), "01-second-task")
		}
		if filepath.Base(filepath.Dir(path2)) != "01-first-task" {
			t.Errorf("parent of second = %q, want %q", filepath.Base(filepath.Dir(path2)), "01-first-task")
		}
	})
}

func TestDelete(t *testing.T) {
	t.Parallel()

	t.Run("delete leaf crumb", func(t *testing.T) {
		root := setupTestProject(t)
		crumbPath := filepath.Join(root, CrumblerDir, "01-task")
		createCrumb(t, crumbPath)

		err := Delete(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, err := os.Stat(crumbPath); !os.IsNotExist(err) {
			t.Error("crumb directory should be deleted")
		}
	})

	t.Run("delete works depth-first", func(t *testing.T) {
		root := setupTestProject(t)
		parentPath := filepath.Join(root, CrumblerDir, "01-parent")
		childPath := filepath.Join(parentPath, "01-child")
		grandchildPath := filepath.Join(childPath, "01-grandchild")
		createCrumb(t, parentPath)
		createCrumb(t, childPath)
		createCrumb(t, grandchildPath)

		// Current should be grandchild (deepest)
		current, _ := GetCurrent(root)
		if current.Name != "grandchild" {
			t.Fatalf("current = %q, want grandchild", current.Name)
		}

		// Delete grandchild
		if err := Delete(root); err != nil {
			t.Fatalf("failed to delete grandchild: %v", err)
		}
		if _, err := os.Stat(grandchildPath); !os.IsNotExist(err) {
			t.Fatal("grandchild should be deleted")
		}

		// Current should now be child
		current, _ = GetCurrent(root)
		if current.Name != "child" {
			t.Fatalf("current = %q, want child", current.Name)
		}

		// Delete child
		if err := Delete(root); err != nil {
			t.Fatalf("failed to delete child: %v", err)
		}

		// Current should now be parent
		current, _ = GetCurrent(root)
		if current.Name != "parent" {
			t.Fatalf("current = %q, want parent", current.Name)
		}

		// Delete parent
		if err := Delete(root); err != nil {
			t.Fatalf("failed to delete parent: %v", err)
		}

		// After deleting all children, current is root
		current, _ = GetCurrent(root)
		if current == nil {
			t.Fatal("expected root as current after deleting children")
		}
		if current.RelPath != CrumblerDir {
			t.Errorf("expected root, got %q", current.RelPath)
		}

		// Project is not done yet (.crumbler still exists)
		done, _ := IsDone(root)
		if done {
			t.Error("project should not be done - .crumbler still exists")
		}

		// Delete root crumb (removes .crumbler entirely)
		if err := Delete(root); err != nil {
			t.Fatalf("failed to delete root: %v", err)
		}

		// Now project should be done (.crumbler deleted)
		done, _ = IsDone(root)
		if !done {
			t.Error("project should be done after deleting root")
		}
	})
}

func TestList(t *testing.T) {
	t.Parallel()

	t.Run("empty project", func(t *testing.T) {
		root := setupTestProject(t)
		tree, err := List(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tree == nil {
			t.Fatal("expected tree, got nil")
		}
		if len(tree.Children) != 0 {
			t.Errorf("expected no children, got %d", len(tree.Children))
		}
	})

	t.Run("nested structure", func(t *testing.T) {
		root := setupTestProject(t)
		createCrumb(t, filepath.Join(root, CrumblerDir, "01-setup"))
		createCrumb(t, filepath.Join(root, CrumblerDir, "01-setup", "01-database"))
		createCrumb(t, filepath.Join(root, CrumblerDir, "02-features"))

		tree, err := List(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(tree.Children) != 2 {
			t.Errorf("expected 2 children, got %d", len(tree.Children))
		}

		// First child should be setup with 1 child
		if tree.Children[0].Name != "setup" {
			t.Errorf("first child = %q, want %q", tree.Children[0].Name, "setup")
		}
		if len(tree.Children[0].Children) != 1 {
			t.Errorf("setup children = %d, want 1", len(tree.Children[0].Children))
		}
	})
}

func TestIsDone(t *testing.T) {
	t.Parallel()

	t.Run("no .crumbler is done", func(t *testing.T) {
		dir := t.TempDir()
		// No .crumbler directory

		done, err := IsDone(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !done {
			t.Error("no .crumbler should be done")
		}
	})

	t.Run(".crumbler exists is not done", func(t *testing.T) {
		root := setupTestProject(t)
		done, err := IsDone(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if done {
			t.Error(".crumbler exists should not be done")
		}
	})

	t.Run("project with crumbs not done", func(t *testing.T) {
		root := setupTestProject(t)
		createCrumb(t, filepath.Join(root, CrumblerDir, "01-task"))

		done, err := IsDone(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if done {
			t.Error("project with crumbs should not be done")
		}
	})
}

func TestCount(t *testing.T) {
	t.Parallel()

	t.Run("empty project", func(t *testing.T) {
		root := setupTestProject(t)
		count, err := Count(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if count != 0 {
			t.Errorf("count = %d, want 0", count)
		}
	})

	t.Run("nested structure", func(t *testing.T) {
		root := setupTestProject(t)
		createCrumb(t, filepath.Join(root, CrumblerDir, "01-setup"))
		createCrumb(t, filepath.Join(root, CrumblerDir, "01-setup", "01-database"))
		createCrumb(t, filepath.Join(root, CrumblerDir, "02-features"))

		count, err := Count(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if count != 3 {
			t.Errorf("count = %d, want 3", count)
		}
	})
}

func TestDisplayName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		crumb    Crumb
		expected string
	}{
		{Crumb{Name: "add-user-auth"}, "Add User Auth"},
		{Crumb{Name: "setup"}, "Setup"},
		{Crumb{Name: "", ID: "01"}, "01"},
		{Crumb{Name: "", ID: ""}, "root"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.crumb.DisplayName()
			if result != tt.expected {
				t.Errorf("DisplayName() = %q, want %q", result, tt.expected)
			}
		})
	}
}

var _ = fmt.Sprintf // use fmt
