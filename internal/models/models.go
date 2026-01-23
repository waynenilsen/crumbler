// Package models defines shared constants for the crumbler CLI tool.
// v2 uses a simplified model where crumbs are directories with README.md files.
package models

// Core constants for crumbler v2.
const (
	// CrumblerDir is the name of the crumbler state directory.
	CrumblerDir = ".crumbler"

	// ReadmeFile is the filename for README documentation in each crumb.
	ReadmeFile = "README.md"

	// MaxChildren is the maximum number of child crumbs (01-10).
	MaxChildren = 10
)
