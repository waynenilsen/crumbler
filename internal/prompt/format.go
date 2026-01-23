package prompt

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/waynenilsen/crumbler/internal/crumb"
)

// formatContext generates the context section showing the current crumb.
func formatContext(root string, current *crumb.Crumb, readme string) string {
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

	// Show README content
	sb.WriteString("### README.md\n\n")
	if strings.TrimSpace(readme) == "" {
		sb.WriteString("*(empty - needs decomposition)*\n")
	} else {
		// Indent README content
		sb.WriteString("```markdown\n")
		sb.WriteString(readme)
		if !strings.HasSuffix(readme, "\n") {
			sb.WriteString("\n")
		}
		sb.WriteString("```\n")
	}

	return sb.String()
}

// formatInstructions generates state-specific instructions.
func formatInstructions(state State, current *crumb.Crumb) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## STATE: %s\n\n", state))

	switch state {
	case StateDecompose:
		sb.WriteString(formatDecomposeInstructions(current))
	case StateExecute:
		sb.WriteString(formatExecuteInstructions(current))
	case StateDone:
		sb.WriteString("All work is complete. No further action needed.\n")
	}

	return sb.String()
}

// formatDecomposeInstructions generates instructions for DECOMPOSE state.
func formatDecomposeInstructions(current *crumb.Crumb) string {
	var sb strings.Builder

	sb.WriteString("The README is empty. You need to plan this work.\n\n")

	sb.WriteString("### Your Task\n\n")
	sb.WriteString("1. **Understand the context**: Look at parent crumb READMEs for context\n")
	sb.WriteString("2. **Plan the work**: Decide if this is a single task or needs breakdown\n")
	sb.WriteString("3. **Either:**\n")
	sb.WriteString("   - Write a clear task description in the README (if it's a single task)\n")
	sb.WriteString("   - Create sub-crumbs for each sub-task (if it needs breakdown)\n\n")

	sb.WriteString("### Commands\n\n")
	sb.WriteString("To create sub-crumbs:\n")
	sb.WriteString("```bash\n")
	sb.WriteString("crumbler create \"Sub-task 1\"\n")
	sb.WriteString("crumbler create \"Sub-task 2\"\n")
	sb.WriteString("```\n\n")

	sb.WriteString("To write the README (if single task):\n")
	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("# Edit: %s/README.md\n", current.RelPath))
	sb.WriteString("```\n\n")

	sb.WriteString("After planning, run: `crumbler prompt`\n")

	return sb.String()
}

// formatExecuteInstructions generates instructions for EXECUTE state.
func formatExecuteInstructions(current *crumb.Crumb) string {
	var sb strings.Builder

	sb.WriteString("The README has instructions. Execute the work.\n\n")

	sb.WriteString("### Your Task\n\n")
	sb.WriteString("1. **Read the README carefully** (shown above)\n")
	sb.WriteString("2. **Execute the work**: Make the necessary code changes\n")
	sb.WriteString("3. **Verify**: Ensure the work is complete and correct\n")
	sb.WriteString("4. **Delete the crumb**: Mark this task as done\n\n")

	sb.WriteString("### Commands\n\n")
	sb.WriteString("When work is complete:\n")
	sb.WriteString("```bash\n")
	sb.WriteString("crumbler delete\n")
	sb.WriteString("crumbler prompt\n")
	sb.WriteString("```\n")

	return sb.String()
}

// FormatTree formats a crumb tree for display.
func FormatTree(root *crumb.Crumb, indent string, isLast bool) string {
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

	if root.IsLeaf {
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
			if child.IsLeaf {
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
				sb.WriteString(FormatTree(&grandchild, grandchildIndent, isGrandchildLast))
			}
		} else {
			sb.WriteString(FormatTree(&child, childIndent, isChildLast))
		}
	}

	return sb.String()
}
