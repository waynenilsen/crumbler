read the readme. that is all for now.

as you work take notes in here

you have to consider updating this document at every turn

## Implementation Notes (v2)

### Architecture Overview

v2 is a radical simplification. The filesystem IS the state - no status files, no checkboxes, no tracking.

**Key packages:**
- `internal/crumb/` - Core crumb operations (traverse, create, delete, list)
- `internal/prompt/` - Prompt generation (agent decides execute vs decompose)
- `internal/models/` - Just constants (CrumblerDir, ReadmeFile, MaxChildren)
- `internal/testutil/` - Test infrastructure (builder, helpers, generator)
- `cmd/crumbler/` - CLI commands (prompt, create, delete, status, clean)

### Core Data Model

```go
// internal/crumb/crumb.go
type Crumb struct {
    Path     string   // Full filesystem path
    RelPath  string   // Relative path from .crumbler/
    Name     string   // Human-readable name (from dirname)
    ID       string   // Two-digit ID (01-10)
    IsLeaf   bool     // True if no children
    Children []Crumb  // Child crumbs (if branch)
}
```

### Traversal Algorithm (internal/crumb/traverse.go)

Depth-first traversal to find current crumb:
1. Start at `.crumbler/`
2. If has children (`01-*/`), recurse into first (sorted by ID)
3. If no children, this is the current crumb (leaf)
4. If no directories remain at root, return nil (DONE)

### Naming Utilities (internal/crumb/naming.go)

- `Kebabify(name)` - "Add User Auth" → "add-user-auth"
- `NextID(parent)` - Returns "01"-"10" or error if full
- `FormatDir(id, name)` - "01" + "add-auth" → "01-add-auth"
- `ParseDir(dirname)` - "01-add-auth" → "01", "add-auth"

### Prompt Generation (internal/prompt/)

**Stateless design** - The agent decides what to do based on README content:
- If README is empty: shows warning + parent context for inheritance
- If README has content: shows the content
- If no `.crumbler/`: shows DONE message

The preamble/postamble are static instructions explaining the execute vs decompose decision.

### CLI Commands

| Command | File | Description |
|---------|------|-------------|
| `prompt` | prompt.go | Outputs structured prompt |
| `create` | create.go | Creates sub-crumbs under current (supports multiple names) |
| `delete` | delete.go | Deletes current crumb (must be leaf) |
| `status` | status.go | Shows tree with crumb count |
| `clean` | clean.go | JSON cleaning utility (unchanged) |

### Test Infrastructure (internal/testutil/)

```go
// Builder pattern for tests
builder := NewTestProject(t)
root := builder.
    WithCrumb("01-task", "Task content").
    WithCrumb("01-task/01-subtask", "Subtask content").
    Build()

// Assertions
AssertFileExists(t, path)
AssertFileContent(t, path, expected)
AssertDirExists(t, path)
```

### Design Decisions

1. **Filesystem as state**: Directory exists = work in progress, deleted = complete
2. **No status files**: Eliminates sync problems
3. **Depth-first traversal**: Deterministic current crumb without tracking
4. **10-item limit**: Forces meaningful decomposition
5. **2-digit IDs**: Simpler than v1's 4-digit
6. **Auto-kebabification**: `create` transforms names automatically
7. **Auto-init**: `create` auto-initializes .crumbler/ if it doesn't exist
8. **Batch create**: `create` accepts multiple names to create sibling crumbs
9. **No state machine**: Agent decides execute vs decompose based on README content

### Key Files

```
internal/
├── crumb/
│   ├── crumb.go       # Core operations: GetCurrent, Create, Delete, List, Count
│   ├── traverse.go    # Depth-first traversal algorithm
│   ├── naming.go      # Kebabify, NextID, FormatDir, ParseDir
│   └── crumb_test.go  # Tests
├── prompt/
│   ├── prompt.go      # GeneratePrompt (stateless)
│   ├── templates.go   # Static preamble/postamble templates
│   ├── format.go      # FormatTree, formatContext, getParentReadmeContext
│   └── prompt_test.go # Tests
├── models/
│   └── models.go      # Constants only (CrumblerDir, ReadmeFile, MaxChildren)
└── testutil/
    ├── builder.go     # TestProjectBuilder
    ├── helpers.go     # Assertions and file operations
    └── generator.go   # Lorem ipsum, random strings, task names
```

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o crumbler
```

### Manual Testing

Use `.test/` directory for manual testing (gitignored):

```bash
mkdir -p .test && cd .test
../crumbler create "Test Task"
../crumbler prompt
../crumbler status
```
