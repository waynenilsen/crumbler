// Package crumb provides core crumb operations for crumbler v2.
package crumb

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

const (
	// MaxChildren is the maximum number of child crumbs (01-10)
	MaxChildren = 10
	// MinID is the minimum crumb ID
	MinID = 1
	// MaxID is the maximum crumb ID
	MaxID = 10
)

// Kebabify converts a human-readable name to kebab-case.
// "Add User Auth" → "add-user-auth"
// "Setup Database!" → "setup-database"
func Kebabify(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)

	// Replace spaces and underscores with hyphens
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")

	// Remove non-alphanumeric characters except hyphens
	var result strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			result.WriteRune(r)
		}
	}

	// Collapse multiple hyphens
	re := regexp.MustCompile(`-+`)
	cleaned := re.ReplaceAllString(result.String(), "-")

	// Trim leading/trailing hyphens
	return strings.Trim(cleaned, "-")
}

// NextID returns the next available ID (01-10) in the given directory.
// Returns error if directory is full (already has 10 children).
func NextID(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return "01", nil
		}
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	// Find existing IDs
	usedIDs := make(map[int]bool)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		id, _ := ParseDir(entry.Name())
		if id != "" {
			if num, err := strconv.Atoi(id); err == nil {
				usedIDs[num] = true
			}
		}
	}

	// Find first available ID
	for i := MinID; i <= MaxID; i++ {
		if !usedIDs[i] {
			return fmt.Sprintf("%02d", i), nil
		}
	}

	return "", fmt.Errorf("directory is full (max %d children)", MaxChildren)
}

// FormatDir combines an ID and name into a directory name.
// "01" + "add-auth" → "01-add-auth"
func FormatDir(id, name string) string {
	if name == "" {
		return id
	}
	return id + "-" + name
}

// ParseDir extracts the ID and name from a directory name.
// "01-add-auth" → "01", "add-auth"
// "01" → "01", ""
// "invalid" → "", "invalid"
func ParseDir(dirname string) (id, name string) {
	// Check if starts with 2 digits
	if len(dirname) < 2 {
		return "", dirname
	}

	// Extract potential ID
	potentialID := dirname[:2]
	if _, err := strconv.Atoi(potentialID); err != nil {
		return "", dirname
	}

	// Check if ID is in valid range
	num, _ := strconv.Atoi(potentialID)
	if num < MinID || num > MaxID {
		return "", dirname
	}

	// Extract name (after the hyphen)
	if len(dirname) > 2 && dirname[2] == '-' {
		return potentialID, dirname[3:]
	}

	return potentialID, ""
}

// ListChildDirs returns sorted child directories that match the ID pattern.
func ListChildDirs(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var children []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		id, _ := ParseDir(entry.Name())
		if id != "" {
			children = append(children, filepath.Join(dir, entry.Name()))
		}
	}

	// Sort by directory name (which sorts by ID due to zero-padding)
	sort.Strings(children)
	return children, nil
}
