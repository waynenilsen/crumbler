package prompt

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/waynenilsen/crumbler/internal/crumb"
)

// formatContext generates the context section showing the current crumb.
func formatContext(root string, current *crumb.Crumb) string {
	var sb strings.Builder

	sb.WriteString("## Current Crumb\n\n")

	// Show path
	relPath := current.RelPath
	if relPath == "" {
		relPath = current.Path
		if rel, err := filepath.Rel(root, current.Path); err == nil {
			relPath = rel
		}
	}
	sb.WriteString(fmt.Sprintf("**Path:** %s\n", relPath))

	// Show display name
	if current.Name != "" {
		sb.WriteString(fmt.Sprintf("**Name:** %s\n", current.DisplayName()))
	}

	sb.WriteString("\n")

	// Get README content
	readme, _ := current.GetReadme()

	// Show README content
	sb.WriteString("### README.md\n\n")
	if strings.TrimSpace(readme) == "" {
		sb.WriteString("**⚠️ README is empty**\n\n")

		// Traverse up and show parent README(s) for context
		parentContext := getParentReadmeContext(root, current)
		if parentContext != "" {
			sb.WriteString("### Parent Context\n\n")
			sb.WriteString(parentContext)
		}
	} else {
		sb.WriteString("```markdown\n")
		sb.WriteString(readme)
		if !strings.HasSuffix(readme, "\n") {
			sb.WriteString("\n")
		}
		sb.WriteString("```\n")
	}

	return sb.String()
}

// getParentReadmeContext traverses up the crumb hierarchy and returns
// parent README content for context when the current README is empty.
func getParentReadmeContext(root string, current *crumb.Crumb) string {
	var sb strings.Builder
	crumblerPath := filepath.Join(root, crumb.CrumblerDir)

	// Get current path relative to .crumbler
	currentPath := current.Path

	// Walk up the directory tree until we hit .crumbler
	for {
		parentPath := filepath.Dir(currentPath)

		// Stop if we've gone above .crumbler
		if !strings.HasPrefix(parentPath, crumblerPath) {
			break
		}

		// Try to read parent README
		parentReadmePath := filepath.Join(parentPath, crumb.ReadmeFile)
		content, err := os.ReadFile(parentReadmePath)
		if err != nil {
			break
		}

		readmeContent := string(content)
		if strings.TrimSpace(readmeContent) != "" {
			// Get relative path for display
			relPath, _ := filepath.Rel(root, parentPath)
			sb.WriteString(fmt.Sprintf("**From %s/README.md:**\n", relPath))
			sb.WriteString("```markdown\n")
			sb.WriteString(readmeContent)
			if !strings.HasSuffix(readmeContent, "\n") {
				sb.WriteString("\n")
			}
			sb.WriteString("```\n\n")
			// Only show the first parent with content
			break
		}

		currentPath = parentPath
	}

	return sb.String()
}

// FormatTree formats a crumb tree for display.
// currentPath is the path of the actual current crumb (to mark with arrow).
func FormatTree(root *crumb.Crumb, indent string, isLast bool) string {
	return FormatTreeWithCurrent(root, indent, isLast, "")
}

// FormatTreeWithCurrent formats a crumb tree for display, marking the current crumb.
func FormatTreeWithCurrent(root *crumb.Crumb, indent string, isLast bool, currentPath string) string {
	var sb strings.Builder

	// Determine if this is the root node (indent is empty and ID is empty)
	isRoot := indent == "" && root.ID == ""

	// Determine prefix
	prefix := indent
	if !isRoot && indent != "" {
		if isLast {
			prefix += "└── "
		} else {
			prefix += "├── "
		}
	}

	// Format this crumb
	var name string
	if isRoot {
		name = ".crumbler/"
	} else if root.Name != "" {
		name = fmt.Sprintf("%s-%s/", root.ID, root.Name)
	} else if root.ID != "" {
		name = root.ID + "/"
	} else {
		name = root.RelPath
	}

	if currentPath != "" && (root.Path == currentPath || root.RelPath == currentPath) {
		name += " ← current"
	}

	sb.WriteString(prefix + name + "\n")

	// Format children
	childIndent := indent
	if isRoot {
		// Root's children start with tree structure
		childIndent = ""
	} else if indent != "" {
		if isLast {
			childIndent += "    "
		} else {
			childIndent += "│   "
		}
	}

	for i, child := range root.Children {
		isChildLast := i == len(root.Children)-1
		// For root's children, use tree structure
		if isRoot {
			if isChildLast {
				sb.WriteString("└── ")
			} else {
				sb.WriteString("├── ")
			}
			childName := fmt.Sprintf("%s-%s/", child.ID, child.Name)
			if currentPath != "" && (child.Path == currentPath || child.RelPath == currentPath) {
				childName += " ← current"
			}
			sb.WriteString(childName + "\n")
			// Recursively format grandchildren
			grandchildIndent := "    "
			if !isChildLast {
				grandchildIndent = "│   "
			}
			for j, grandchild := range child.Children {
				isGrandchildLast := j == len(child.Children)-1
				sb.WriteString(FormatTreeWithCurrent(&grandchild, grandchildIndent, isGrandchildLast, currentPath))
			}
		} else {
			sb.WriteString(FormatTreeWithCurrent(&child, childIndent, isChildLast, currentPath))
		}
	}

	return sb.String()
}
