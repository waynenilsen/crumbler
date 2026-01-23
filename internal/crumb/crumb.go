package crumb

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// CrumblerDir is the name of the crumbler directory
	CrumblerDir = ".crumbler"
	// ReadmeFile is the name of the README file in each crumb
	ReadmeFile = "README.md"
)

// Crumb represents a unit of work in the crumbler system.
type Crumb struct {
	Path     string  // Full filesystem path
	RelPath  string  // Relative path from project root
	Name     string  // Human-readable name (from dirname)
	ID       string  // Two-digit ID (01-10)
	IsLeaf   bool    // True if no children
	Children []Crumb // Child crumbs (if branch)
}

// GetCurrent finds the current crumb using depth-first traversal.
// Returns nil if project is done (.crumbler doesn't exist).
func GetCurrent(root string) (*Crumb, error) {
	crumblerPath := filepath.Join(root, CrumblerDir)

	// Check if .crumbler exists
	if _, err := os.Stat(crumblerPath); os.IsNotExist(err) {
		return nil, nil // Project is done (no .crumbler)
	}

	// Check for child crumbs
	children, err := ListChildDirs(crumblerPath)
	if err != nil {
		return nil, err
	}

	// If has children, traverse to find current (deepest first) crumb
	if len(children) > 0 {
		crumb, err := traverse(children[0])
		if err != nil {
			return nil, err
		}
		if crumb != nil {
			crumb.RelPath = relPath(root, crumb.Path)
		}
		return crumb, nil
	}

	// No children - root crumb is the current crumb
	crumb := &Crumb{
		Path:    crumblerPath,
		RelPath: CrumblerDir,
		Name:    "",
		ID:      "",
		IsLeaf:  true,
	}
	return crumb, nil
}

// Create creates a new sub-crumb under the current crumb.
// If .crumbler doesn't exist, creates it first (auto-init).
// Returns the path to the created crumb directory.
func Create(root string, name string) (string, error) {
	crumblerPath := filepath.Join(root, CrumblerDir)

	// Auto-init: create .crumbler with README if it doesn't exist
	if _, err := os.Stat(crumblerPath); os.IsNotExist(err) {
		if err := os.MkdirAll(crumblerPath, 0755); err != nil {
			return "", fmt.Errorf("failed to create .crumbler directory: %w", err)
		}
		readmePath := filepath.Join(crumblerPath, ReadmeFile)
		if err := os.WriteFile(readmePath, []byte{}, 0644); err != nil {
			return "", fmt.Errorf("failed to create root README.md: %w", err)
		}
	}

	current, err := GetCurrent(root)
	if err != nil {
		return "", err
	}

	var parentPath string
	if current == nil {
		// No current crumb (shouldn't happen after auto-init, but be safe)
		parentPath = crumblerPath
	} else {
		parentPath = current.Path
	}

	// Get next available ID
	id, err := NextID(parentPath)
	if err != nil {
		return "", err
	}

	// Create directory name
	kebabName := Kebabify(name)
	dirname := FormatDir(id, kebabName)
	crumbPath := filepath.Join(parentPath, dirname)

	// Create directory
	if err := os.MkdirAll(crumbPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create crumb directory: %w", err)
	}

	// Create empty README.md
	readmePath := filepath.Join(crumbPath, ReadmeFile)
	if err := os.WriteFile(readmePath, []byte{}, 0644); err != nil {
		return "", fmt.Errorf("failed to create README.md: %w", err)
	}

	return crumbPath, nil
}

// Delete removes the current crumb.
// Fails if the crumb has children.
// Deleting the root crumb removes the entire .crumbler directory.
func Delete(root string) error {
	current, err := GetCurrent(root)
	if err != nil {
		return err
	}

	if current == nil {
		return fmt.Errorf("no crumb to delete")
	}

	// Check if crumb has children
	children, err := ListChildDirs(current.Path)
	if err != nil {
		return err
	}
	if len(children) > 0 {
		return fmt.Errorf("cannot delete crumb with children (has %d children)", len(children))
	}

	// Remove the directory (including root .crumbler)
	if err := os.RemoveAll(current.Path); err != nil {
		return fmt.Errorf("failed to delete crumb: %w", err)
	}

	return nil
}

// List returns all crumbs as a tree structure.
// Returns nil if .crumbler doesn't exist (project is done).
func List(root string) (*Crumb, error) {
	crumblerPath := filepath.Join(root, CrumblerDir)

	// Check if .crumbler exists
	if _, err := os.Stat(crumblerPath); os.IsNotExist(err) {
		return nil, nil // No crumbs
	}

	return listCrumb(root, crumblerPath)
}

// listCrumb recursively builds the crumb tree.
func listCrumb(root, path string) (*Crumb, error) {
	// Check for README.md
	readmePath := filepath.Join(path, ReadmeFile)
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		return nil, nil
	}

	dirname := filepath.Base(path)
	id, name := ParseDir(dirname)

	crumb := &Crumb{
		Path:    path,
		RelPath: relPath(root, path),
		Name:    name,
		ID:      id,
	}

	// Get children
	children, err := ListChildDirs(path)
	if err != nil {
		return nil, err
	}

	for _, childPath := range children {
		child, err := listCrumb(root, childPath)
		if err != nil {
			return nil, err
		}
		if child != nil {
			crumb.Children = append(crumb.Children, *child)
		}
	}

	crumb.IsLeaf = len(crumb.Children) == 0

	return crumb, nil
}

// IsDone returns true if no work remains in the project.
// Project is done when .crumbler directory doesn't exist.
// The filesystem IS the state - existence = work to do, deleted = done.
func IsDone(root string) (bool, error) {
	crumblerPath := filepath.Join(root, CrumblerDir)

	// Check if .crumbler exists
	if _, err := os.Stat(crumblerPath); os.IsNotExist(err) {
		return true, nil // Project is done (no .crumbler directory)
	}

	return false, nil // Project has work to do
}

// Count returns the total number of crumbs (excluding root).
// Returns 0 if .crumbler doesn't exist.
func Count(root string) (int, error) {
	crumblerPath := filepath.Join(root, CrumblerDir)

	// Check if .crumbler exists
	if _, err := os.Stat(crumblerPath); os.IsNotExist(err) {
		return 0, nil // No crumbs
	}

	return countCrumbs(crumblerPath)
}

// GetReadme returns the contents of the crumb's README.md.
func (c *Crumb) GetReadme() (string, error) {
	readmePath := filepath.Join(c.Path, ReadmeFile)
	content, err := os.ReadFile(readmePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// DisplayName returns a human-readable name for the crumb.
func (c *Crumb) DisplayName() string {
	if c.Name != "" {
		// Convert kebab-case to title case
		words := strings.Split(c.Name, "-")
		for i, word := range words {
			if len(word) > 0 {
				words[i] = strings.ToUpper(word[:1]) + word[1:]
			}
		}
		return strings.Join(words, " ")
	}
	if c.ID != "" {
		return c.ID
	}
	return "root"
}

// relPath returns a path relative to the project root.
func relPath(root, fullPath string) string {
	rel, err := filepath.Rel(root, fullPath)
	if err != nil {
		return fullPath
	}
	return rel
}
