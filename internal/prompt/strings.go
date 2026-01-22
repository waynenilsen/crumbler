// Package prompt provides functionality for generating AI agent prompts
// based on the current crumbler state machine state.
package prompt

// Shared strings used both in CLI help and prompt generation to prevent doc drift.

// CrumblerDescription provides a brief description of what crumbler is.
const CrumblerDescription = `crumbler is a lightweight state machine manager for agentic software development
lifecycle (SDLC) automation. It manages:
- State transitions (open -> closed, open -> done)
- Goal state tracking (numbered, named, open/closed)
- Directory structure creation
- State machine integrity validation

crumbler does NOT generate content - YOU (the AI agent) are responsible for:
- Populating README.md, PRD.md, ERD.md files
- Creating goal names
- Writing ticket descriptions
- Making development decisions`

// StateExplanation explains how state is tracked in crumbler.
const StateExplanation = `State is tracked via empty files in each entity directory:
- 'open' file = entity is active/in progress
- 'closed' file = entity is complete (phases, sprints, goals)
- 'done' file = ticket is complete (tickets only)

Files are mutually exclusive: having both 'open' and 'closed' is an error.`

// GoalExplanation explains how goals work.
const GoalExplanation = `Goals are tracked in goals/ directories at phase, sprint, and ticket levels:
- goals/XXXX-goal/ (4-digit zero-padded)
- 'name' file contains the goal description (you populate this)
- 'open' or 'closed' status files`

// HierarchyRules explains the state machine hierarchy constraints.
const HierarchyRules = `Hierarchy constraints:
- Phase can close only when: all sprints closed AND all phase goals closed
- Sprint can close only when: all tickets done AND all sprint goals closed
- Ticket can be marked done only when: all ticket goals closed`

// DirectoryStructure shows the crumbler directory structure.
const DirectoryStructure = `Directory structure:
.crumbler/
├── roadmap.md              # Project roadmap (phases and goals)
├── phases/
│   └── XXXX-phase/
│       ├── open|closed     # Phase status
│       ├── README.md       # Phase description (you populate)
│       ├── goals/          # Phase-level goals
│       │   └── XXXX-goal/
│       │       ├── name    # Goal name (you populate)
│       │       └── open|closed
│       └── sprints/
│           └── XXXX-sprint/
│               ├── open|closed
│               ├── README.md, PRD.md, ERD.md  # (you populate)
│               ├── goals/  # Sprint-level goals
│               └── tickets/
│                   └── XXXX-ticket/
│                       ├── open|done
│                       ├── README.md  # (you populate)
│                       └── goals/     # Ticket-level goals`

// AgentWorkflowInstructions provides guidance on how the agent should work.
const AgentWorkflowInstructions = `How to work:
1. Read the current state and understand context
2. Execute the SINGLE instruction provided below
3. Run 'crumbler get-next-prompt' for your next instruction
4. Repeat until EXIT state is reached

Important:
- Execute ONE action at a time
- Do not skip steps or combine multiple actions
- Always use crumbler commands for state transitions
- Populate empty files with meaningful content before proceeding`

// PhaseClosingRules describes when a phase can be closed.
const PhaseClosingRules = `A phase can be closed when:
1. All sprints in the phase have 'closed' status
2. All phase goals have 'closed' status
Command: crumbler phase close`

// SprintClosingRules describes when a sprint can be closed.
const SprintClosingRules = `A sprint can be closed when:
1. All tickets in the sprint have 'done' status
2. All sprint goals have 'closed' status
Command: crumbler sprint close`

// TicketDoneRules describes when a ticket can be marked done.
const TicketDoneRules = `A ticket can be marked done when:
1. All ticket goals have 'closed' status
2. The ticket's work has been completed
Command: crumbler ticket done <ticket-id>`

// GoalClosingRule describes when a goal can be closed.
const GoalClosingRule = `A goal can be closed when its objective has been achieved.
Commands:
- crumbler phase goal close <goal-id>
- crumbler sprint goal close <goal-id>
- crumbler ticket goal close <ticket-id> <goal-id>`
