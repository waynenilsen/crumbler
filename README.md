# crumbler

A lightweight Go CLI tool for AI agent task decomposition and execution. Zero dependencies.

Crumbler is a task decomposition and execution framework for AI agents running in a loop. Designed for Claude Code CLI (`claude` command) with cheap, high-volume execution.

## The Core Loop

```bash
while true; do
    prompt=$(crumbler prompt)
    if echo "$prompt" | grep -q "STATE: DONE"; then
        echo "Project complete!"
        break
    fi
    claude --print "$prompt"
done
```

The agent decomposes work into crumbs, executes leaf crumbs, and deletes completed work. Crumbler manages the structure; the agent does the thinking.

## Core Concepts

### Everything is a Crumb

A crumb is a unit of work represented by a directory containing `README.md`. Crumbs can contain sub-crumbs. No phase/sprint/ticket—just crumbs at any depth.

### Filesystem IS the State

**No checkboxes. No status files. No tracking.**

The directory structure is the only state:
- Directory exists → work in progress
- Directory has children → it's a branch (traverse into children)
- Directory has no children → it's a leaf (execute the work)
- Directory deleted → work complete

### Leaf vs Branch

| Type | Children? | Agent Action |
|------|-----------|--------------|
| **Leaf** | No `01-*/` dirs | Execute work, then `crumbler delete` |
| **Branch** | Has `01-*/` dirs | Traverse into first incomplete child |

A crumb becomes a branch when you create a child. A branch becomes deletable when all children are deleted.

### Completion = Deletion

When work is complete, delete the crumb. Git provides recovery.

- No done/closed states
- Only "exists" or "doesn't exist"
- Tree shrinks as work completes

### The 10-Item Limit

Maximum 10 sub-crumbs per parent (IDs `01`-`10`). This constraint:
- Forces meaningful decomposition (not micro-tasks)
- Prevents infinite creation loops
- If you need 11+ items, your parent crumb is too broad—split it

## File Structure

```
.crumbler/
├── README.md                    # Root crumb
├── 01-setup/
│   ├── README.md                # Branch (has children)
│   ├── 01-database/
│   │   └── README.md            # Leaf (no children)
│   └── 02-api/
│       └── README.md            # Leaf
└── 02-features/
    └── README.md                # Leaf or branch depending on children
```

### Naming Convention

- IDs: Two-digit zero-padded (`01`, `02`, ... `10`)
- Names: kebab-case after ID (`01-api-design`, `02-auth-flow`)
- Auto-kebabification: `crumbler create "Add User Auth"` → `01-add-user-auth`
- Format: `{ID}-{name}/README.md`

## The Agent Loop

Each iteration:

1. **`crumbler prompt`** traverses the tree (depth-first) to find the current crumb
2. **`crumbler prompt`** outputs a structured prompt with context and instructions
3. **Agent** reads the prompt and decides:
   - **DECOMPOSE**: Work is too big → `crumbler create "child-name"`
   - **EXECUTE**: Work is doable → do it → `crumbler delete`
4. **Loop repeats** until no crumbs remain

### Prompt States

| State | Condition | Agent Action |
|-------|-----------|--------------|
| `DECOMPOSE` | README is empty | Plan work, create sub-crumbs or write README |
| `EXECUTE` | README has content, no children | Do the work, then delete |
| `DONE` | No crumbs remain | Exit loop |

### Traversal Algorithm

`crumbler prompt` finds the current crumb via depth-first traversal:

1. Start at `.crumbler/`
2. If directory has children (`01-*/`), recurse into first child (sorted by ID)
3. If directory has no children, this is the current crumb (a leaf)
4. If no directories remain, project is complete (`STATE: DONE`)

No position tracking needed—the algorithm always finds the same crumb given the same filesystem state.

## Commands

Minimal command set (5 commands):

| Command | Description |
|---------|-------------|
| `crumbler init` | Create `.crumbler/` with root README.md |
| `crumbler prompt` | Output structured prompt for agent |
| `crumbler create {name}` | Create sub-crumb under current crumb |
| `crumbler delete` | Delete current crumb (must be leaf) |
| `crumbler status` | Show tree structure and progress |

### Command Details

**`crumbler init`**
- Creates `.crumbler/README.md`
- Ready for agent to start decomposing

**`crumbler create "Some Name Here"`**
- Auto-assigns next available ID (01-10)
- Auto-kebabifies: "Some Name Here" → `some-name-here`
- Creates directory with empty README.md under current crumb
- Returns full relative path

**`crumbler delete`**
- Finds current crumb via traversal
- Fails if crumb has children (must delete children first)
- Removes directory and contents

**`crumbler prompt`**
- Traverses tree to find current crumb
- Outputs structured prompt with preamble, README content, instructions
- If no crumbs remain, outputs `STATE: DONE`

**`crumbler status`**
- Shows tree structure with crumb count
- Current crumb marked with `← current`

## Installation

### From Source

```bash
git clone https://github.com/waynenilsen/crumbler.git
cd crumbler
go build -o crumbler
```

### Using Go Install

```bash
go install github.com/waynenilsen/crumbler@latest
```

## Usage Example

```bash
# Initialize
crumbler init

# Create initial crumbs
crumbler create "Setup"
crumbler create "Features"

# Check status
crumbler status
# .crumbler/ (2 crumbs remaining)
# ├── 01-setup/ ← current
# └── 02-features/

# Get the prompt for the agent
crumbler prompt

# After agent completes work
crumbler delete

# Continue loop...
crumbler prompt
```

## Decomposition Guidelines

### When to Decompose

Decompose when ANY of these are true:
- Work spans multiple unrelated areas
- You can't hold the full context in one session
- Steps have no shared state or dependencies

### When NOT to Decompose

Stay as leaf when ALL of these are true:
- Work is a "full stack slice" (DB + API + tests for one feature)
- Steps share context (same files, same domain concepts)
- You can complete it in one agent session

### The "Full Stack Unit" Principle

Each leaf crumb should be a **full stack unit of work**—something that:
- Touches all relevant layers (DB, API, UI, tests)
- Can be tested end-to-end
- Delivers a complete, working increment

## Progress Tracking

Since completed work is deleted:

```bash
# Count remaining crumbs
crumbler status

# See what's been completed (via git)
git log --oneline --diff-filter=D -- '.crumbler/**/README.md'

# See current structure
tree .crumbler
```

## Why This Works

### Fully Deterministic
- No state to sync (filesystem IS state)
- No position file to track
- Traversal always finds same crumb given same filesystem

### Token Efficient
- One file per crumb
- Tree shrinks as work completes
- No status files or tracking overhead

### Agent-Friendly
- Binary decision: decompose or execute+delete
- Structured prompt guides agent clearly
- Can't get into invalid state

### Human Compatible
- Markdown renders on GitHub
- Git history preserves completed work
- `tree .crumbler` shows current state

## License

MIT
