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
- Populating README.md, PRD.md, ERD.md files (ERD = Engineering Requirements Document)
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
│               ├── README.md, PRD.md, ERD.md  # (you populate; ERD = Engineering Requirements Document)
│               ├── goals/  # Sprint-level goals
│               └── tickets/
│                   └── XXXX-ticket/
│                       ├── open|done
│                       ├── README.md  # (you populate)
│                       └── goals/     # Ticket-level goals`

// AgentWorkflowInstructions provides guidance on how the agent should work.
const AgentWorkflowInstructions = `How to work (SINGLE-SHOT EXECUTION):
1. Read the current state and understand context
2. Execute ONLY the SINGLE instruction provided below
3. After completing the instruction, run 'crumbler get-next-prompt' to get the next instruction
4. DO NOT loop or repeat - execute ONE action and stop

Important:
- This is a SINGLE-SHOT prompt - execute ONE action only
- Do not skip steps or combine multiple actions
- Do not loop or check for EXIT state yourself
- Always use crumbler commands for state transitions
- Populate empty files with meaningful content before proceeding
- The next call to 'crumbler get-next-prompt' will determine if EXIT state is reached`

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

// HierarchyDetailLevels explains the proper hierarchy and detail levels.
const HierarchyDetailLevels = `CRITICAL: Hierarchy and Detail Levels

The structure is: Phase -> Sprint -> Ticket

Detail level INCREASES as you go down the hierarchy:
- Phase: High-level objectives, multiple sprints required
- Sprint: Detailed requirements (PRD = Product Requirements Document, ERD = Engineering Requirements Document), multiple tickets required  
- Ticket: Implementation-level tasks, specific code changes

REQUIREMENTS:
- A phase MUST have MORE THAN ONE sprint (typically 3-5+ sprints)
- A sprint MUST have MORE THAN ONE ticket (typically 3-10+ tickets)
- Do NOT declare ticket-level detail in phase goals
- Do NOT create a phase with 1 sprint that maps 1-1 to phase goals
- Do NOT create a sprint with 1 ticket that maps 1-1 to sprint goals

GOOD EXAMPLES:

Phase "User Authentication":
  Goal: "Implement secure user authentication system"
  → Sprint 1: "User registration and login"
  → Sprint 2: "Password reset flow"
  → Sprint 3: "OAuth integration"
  → Sprint 4: "Session management"
  (Each sprint has multiple tickets)

Sprint "User registration and login":
  Goal: "Users can register and log in"
  → Ticket 1: "Create user model and database schema"
  → Ticket 2: "Implement registration API endpoint"
  → Ticket 3: "Implement login API endpoint"
  → Ticket 4: "Add input validation and error handling"
  → Ticket 5: "Write tests for auth endpoints"
  (Each ticket has specific implementation goals)

BAD EXAMPLES (DO NOT DO THIS):

❌ Phase with ticket-level detail:
  Phase "User Authentication":
    Goal: "Create user model and database schema"
    Goal: "Implement registration API endpoint"
    Goal: "Implement login API endpoint"
  → Sprint 1: (maps 1-1 to phase goals)
    → Ticket 1: "Create user model"
    → Ticket 2: "Registration API"
    → Ticket 3: "Login API"
  This defeats the purpose - phase should be higher level!

❌ Sprint with 1 ticket:
  Sprint "User Authentication":
    Goal: "Implement user authentication"
  → Ticket 1: "Do all authentication work"
  This is too coarse - break it down!

❌ Phase with 1 sprint:
  Phase "Build Todo App":
  → Sprint 1: "Build everything"
  This defeats the purpose - phases should span multiple sprints!

Remember: Each level should decompose the level above into smaller, more detailed pieces.`
