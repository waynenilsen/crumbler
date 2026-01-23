package crumb

import (
	"os"
	"path/filepath"
)

// traverse performs depth-first traversal to find the current (leaf) crumb.
// Algorithm:
// 1. Start at given directory
// 2. If has children (01-*/), recurse into first (sorted by ID)
// 3. If no children, this is the current crumb (leaf)
// 4. If directory doesn't exist or has no README.md, return nil
func traverse(dir string) (*Crumb, error) {
	// Check if directory exists
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if !info.IsDir() {
		return nil, nil
	}

	// Check if this is a valid crumb (has README.md)
	readmePath := filepath.Join(dir, ReadmeFile)
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		return nil, nil
	}

	// Get child directories
	children, err := ListChildDirs(dir)
	if err != nil {
		return nil, err
	}

	// If has children, recurse into first child
	if len(children) > 0 {
		return traverse(children[0])
	}

	// This is a leaf crumb - build and return it
	return buildCrumb(dir)
}

// traverseAll collects all crumbs in the tree.
func traverseAll(dir string) ([]Crumb, error) {
	// Check if directory exists
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if !info.IsDir() {
		return nil, nil
	}

	// Check if this is a valid crumb (has README.md)
	readmePath := filepath.Join(dir, ReadmeFile)
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		return nil, nil
	}

	// Get child directories
	children, err := ListChildDirs(dir)
	if err != nil {
		return nil, err
	}

	var crumbs []Crumb

	// Recursively collect children
	for _, childPath := range children {
		childCrumbs, err := traverseAll(childPath)
		if err != nil {
			return nil, err
		}
		crumbs = append(crumbs, childCrumbs...)
	}

	// Build crumb for this directory
	crumb, err := buildCrumb(dir)
	if err != nil {
		return nil, err
	}
	if crumb != nil {
		crumb.IsLeaf = len(children) == 0
		crumbs = append(crumbs, *crumb)
	}

	return crumbs, nil
}

// countCrumbs counts all crumbs in the tree (excluding root).
func countCrumbs(dir string) (int, error) {
	children, err := ListChildDirs(dir)
	if err != nil {
		return 0, err
	}

	count := len(children)
	for _, child := range children {
		childCount, err := countCrumbs(child)
		if err != nil {
			return 0, err
		}
		count += childCount
	}

	return count, nil
}

// buildCrumb creates a Crumb from a directory path.
func buildCrumb(dir string) (*Crumb, error) {
	// Get child directories to determine if leaf
	children, err := ListChildDirs(dir)
	if err != nil {
		return nil, err
	}

	dirname := filepath.Base(dir)
	id, name := ParseDir(dirname)

	return &Crumb{
		Path:   dir,
		Name:   name,
		ID:     id,
		IsLeaf: len(children) == 0,
	}, nil
}
