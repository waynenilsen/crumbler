// Package models defines the core data structures for the crumbler CLI tool.
// These structures represent the hierarchical state machine for managing
// software development lifecycle automation.
package models

// Status represents the state of a phase, sprint, ticket, or goal.
// Status is determined by file existence in the filesystem:
// - "open" file exists = StatusOpen
// - "closed" file exists = StatusClosed
// - "done" file exists = StatusDone
type Status string

const (
	// StatusOpen indicates the entity is currently active/in progress.
	// Represented by an "open" file in the entity's directory.
	StatusOpen Status = "open"

	// StatusClosed indicates the entity has been completed and closed.
	// Used for phases, sprints, and goals.
	// Represented by a "closed" file in the entity's directory.
	StatusClosed Status = "closed"

	// StatusDone indicates a ticket has been completed.
	// Only used for tickets (not phases, sprints, or goals).
	// Represented by a "done" file in the entity's directory.
	StatusDone Status = "done"

	// StatusUnknown indicates an unknown or invalid state.
	StatusUnknown Status = "unknown"
)

// String returns the string representation of the status.
func (s Status) String() string {
	return string(s)
}

// IsValid returns true if the status is a known valid status.
func (s Status) IsValid() bool {
	switch s {
	case StatusOpen, StatusClosed, StatusDone:
		return true
	default:
		return false
	}
}

// Goal represents a goal at any level (phase, sprint, or ticket).
// Goals are stored in goals/XXXX-goal/ directories with:
// - name file containing the goal name (AI populates)
// - open/closed status files
type Goal struct {
	// ID is the directory name (e.g., "0001-goal", "0002-goal").
	// Uses 4-digit zero-padded numbering.
	ID string

	// Path is the absolute path to the goal directory.
	Path string

	// Name is the goal name read from the "name" file.
	// This content is populated by the AI agent.
	Name string

	// Status is determined by file existence (open or closed).
	Status Status

	// Index is the numeric index of the goal (extracted from ID).
	Index int
}

// Phase represents a development phase in the roadmap.
// Phases are stored in .crumbler/phases/XXXX-phase/ directories.
type Phase struct {
	// ID is the directory name (e.g., "0001-phase", "0002-phase").
	// Uses 4-digit zero-padded numbering.
	ID string

	// Path is the absolute path to the phase directory.
	Path string

	// Goals is the list of goals for this phase.
	// Read from the goals/ subdirectory.
	Goals []Goal

	// Sprints is the list of sprints in this phase.
	// Read from the sprints/ subdirectory.
	Sprints []Sprint

	// Status is determined by file existence (open or closed).
	// A phase can only be closed when all sprints are closed
	// and all phase goals are closed.
	Status Status

	// Index is the numeric index of the phase (extracted from ID).
	Index int
}

// Sprint represents a sprint within a phase.
// Sprints are stored in .crumbler/phases/XXXX-phase/sprints/XXXX-sprint/ directories.
type Sprint struct {
	// ID is the directory name (e.g., "0001-sprint", "0002-sprint").
	// Uses 4-digit zero-padded numbering.
	ID string

	// Path is the absolute path to the sprint directory.
	Path string

	// Goals is the list of goals for this sprint.
	// Read from the goals/ subdirectory.
	Goals []Goal

	// Tickets is the list of tickets in this sprint.
	// Read from the tickets/ subdirectory.
	Tickets []Ticket

	// PRDPath is the path to the PRD.md file (AI populates content).
	PRDPath string

	// ERDPath is the path to the ERD.md file (AI populates content).
	ERDPath string

	// Status is determined by file existence (open or closed).
	// A sprint can only be closed when all tickets are done
	// and all sprint goals are closed.
	Status Status

	// Index is the numeric index of the sprint (extracted from ID).
	Index int
}

// Ticket represents a ticket within a sprint.
// Tickets are stored in .crumbler/phases/XXXX-phase/sprints/XXXX-sprint/tickets/XXXX-ticket/ directories.
type Ticket struct {
	// ID is the directory name (e.g., "0001-ticket", "0002-ticket").
	// Uses 4-digit zero-padded numbering.
	ID string

	// Path is the absolute path to the ticket directory.
	Path string

	// Goals is the list of goals for this ticket.
	// Read from the goals/ subdirectory.
	Goals []Goal

	// DescriptionPath is the path to the README.md file (AI populates content).
	DescriptionPath string

	// Status is determined by file existence (open or done).
	// A ticket can only be marked done when all ticket goals are closed.
	Status Status

	// Index is the numeric index of the ticket (extracted from ID).
	Index int
}

// Roadmap represents the project roadmap loaded from .crumbler/roadmap.md.
// The roadmap defines the phases and their descriptions/goals.
type Roadmap struct {
	// Path is the absolute path to the roadmap file.
	Path string

	// Phases is the list of phase definitions from the roadmap.
	Phases []PhaseDefinition

	// Metadata contains additional key-value pairs from the roadmap.
	Metadata map[string]string
}

// PhaseDefinition represents a phase as defined in the roadmap markdown.
// This is the planned/templated version before a phase directory is created.
type PhaseDefinition struct {
	// Name is the phase name from the roadmap.
	Name string

	// Description is the phase description from the roadmap.
	Description string

	// Goals is the list of goal names for this phase.
	Goals []string
}

// StatusFile constants for file names used to track status.
const (
	// StatusFileOpen is the filename for open status.
	StatusFileOpen = "open"

	// StatusFileClosed is the filename for closed status.
	StatusFileClosed = "closed"

	// StatusFileDone is the filename for done status (tickets only).
	StatusFileDone = "done"

	// GoalNameFile is the filename containing the goal name.
	GoalNameFile = "name"
)

// Directory and file naming constants.
const (
	// CrumblerDir is the name of the crumbler state directory.
	CrumblerDir = ".crumbler"

	// PhasesDir is the subdirectory name for phases.
	PhasesDir = "phases"

	// SprintsDir is the subdirectory name for sprints within a phase.
	SprintsDir = "sprints"

	// TicketsDir is the subdirectory name for tickets within a sprint.
	TicketsDir = "tickets"

	// GoalsDir is the subdirectory name for goals at any level.
	GoalsDir = "goals"

	// RoadmapsDir is the subdirectory name for roadmap templates/archives.
	RoadmapsDir = "roadmaps"

	// RoadmapFile is the filename for the roadmap.
	RoadmapFile = "roadmap.md"

	// ReadmeFile is the filename for README documentation.
	ReadmeFile = "README.md"

	// PRDFile is the filename for Product Requirements Document.
	PRDFile = "PRD.md"

	// ERDFile is the filename for Entity Relationship Diagram.
	ERDFile = "ERD.md"

	// LockFile is the filename for the lock file used for concurrent execution.
	LockFile = ".lock"
)

// Suffix constants for entity directory names.
const (
	// PhaseSuffix is appended to phase directory names (e.g., "0001-phase").
	PhaseSuffix = "-phase"

	// SprintSuffix is appended to sprint directory names (e.g., "0001-sprint").
	SprintSuffix = "-sprint"

	// TicketSuffix is appended to ticket directory names (e.g., "0001-ticket").
	TicketSuffix = "-ticket"

	// GoalSuffix is appended to goal directory names (e.g., "0001-goal").
	GoalSuffix = "-goal"
)

// NumberFormat is the format string for zero-padded numbering (4 digits).
const NumberFormat = "%04d"
