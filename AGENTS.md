read the readme. that is all for now.

as you work take notes in here

you have to consider updating this document at every turn

## Implementation Notes

### Core Data Models (internal/models/)

Implemented the core data models in `internal/models/`:

1. **models.go** - Core data structures:
   - `Status` type with constants: `StatusOpen`, `StatusClosed`, `StatusDone`, `StatusUnknown`
   - `Goal` struct: ID, Path, Name (from `name` file), Status, Index
   - `Phase` struct: ID, Path, Goals, Sprints, Status, Index
   - `Sprint` struct: ID, Path, Goals, Tickets, PRDPath, ERDPath, Status, Index
   - `Ticket` struct: ID, Path, Goals, DescriptionPath, Status, Index
   - `Roadmap` struct: Path, Phases ([]PhaseDefinition), Metadata
   - `PhaseDefinition` struct: Name, Description, Goals ([]string)
   - Constants for status files, directory names, suffixes, and number format

2. **errors.go** - Custom error types:
   - `StateError` struct with: Type, FilePaths, Message
   - Error type constants: `ErrInvalidState`, `ErrInvalidTransition`, `ErrHierarchyConstraint`, `ErrMissingFile`
   - Constructor functions: `NewStateError`, `NewInvalidStateError`, `NewInvalidTransitionError`, `NewHierarchyConstraintError`, `NewMissingFileError`
   - Helper functions: `ToRelPaths`, `ToRelPath`, `IsStateError`, `AsStateError`
   - Type check functions: `IsInvalidState`, `IsInvalidTransition`, `IsHierarchyConstraint`, `IsMissingFile`

Key design decisions:
- All file paths in errors are stored as relative paths from project root for clear error messages
- `Index` field added to structs for extracting numeric index from ID (e.g., "0001" from "0001-goal")
- `StatusUnknown` added for handling invalid/unknown states
- `PhaseDefinition` is the roadmap template version; `Phase` is the created instance
- Constants defined for all file/directory names to avoid magic strings

### Agent Loop Query Functions (internal/query/)

Implemented state query functions for the AI agent loop in `internal/query/query.go`. These are read-only functions that check state and validate for errors.

**Phase-level queries:**
- `OpenPhaseExists(projectRoot string) (bool, error)` - checks if any phase has `open` file without `closed` file
- `RoadmapComplete(projectRoot string) (bool, error)` - checks if all phases have `closed` file

**Sprint-level queries:**
- `OpenSprintExists(phasePath string) (bool, error)` - checks if any sprint has `open` file without `closed` file
- `PhaseGoalsMet(phasePath string) (bool, error)` - checks if all phase goals are closed AND all sprints are closed

**Ticket-level queries:**
- `OpenTicketsExist(sprintPath string) (bool, error)` - checks if any ticket has `open` file without `done` file
- `SprintGoalsMet(sprintPath string) (bool, error)` - checks if all sprint goals are closed AND all tickets are done

**Ticket completion query:**
- `TicketComplete(ticketPath string) (bool, error)` - checks if ticket has `done` file AND all goals are closed

**Goals existence queries:**
- `PhaseGoalsExist(phasePath string) (bool, error)`
- `SprintGoalsExist(sprintPath string) (bool, error)`
- `TicketGoalsExist(ticketPath string) (bool, error)`

**Helper queries:**
- `SprintsExist(phasePath string) (bool, error)`
- `TicketsExist(sprintPath string) (bool, error)`

Key design decisions:
- All functions validate state and return errors with relative file paths on invalid states (e.g., both `open` and `closed` exist)
- `*GoalsMet` functions return false if no goals or children exist yet (allows CREATE_GOALS/CREATE_SPRINT/CREATE_TICKETS in flowchart)
- Uses `getProjectRoot()` helper to extract project root from paths within .crumbler directory
- Uses direct filesystem operations with `os.ReadDir` and `os.Stat`
- Helper functions: `checkGoalsStatus`, `checkSprintsStatus`, `checkTicketsStatus` for DRY code

### Test Infrastructure (internal/testutil/)

Implemented comprehensive test infrastructure in `internal/testutil/`:

1. **builder.go** - Fluent test builder API:
   - `TestProjectBuilder` struct with methods for building test project structures
   - `NewTestProject(t *testing.T)` - creates unique test directory under `.test/test-{timestamp}-{random}/`
   - `WithPhase(phaseID, status)` - adds phase with open/closed status
   - `WithSprint(phaseID, sprintID, status)` - adds sprint with open/closed status
   - `WithTicket(phaseID, sprintID, ticketID, status)` - adds ticket with open/done status
   - `WithRoadmap(content)` - sets custom roadmap content
   - `WithPhaseGoal/WithSprintGoal/WithTicketGoal` - adds goals with name and status
   - `WithPRD/WithERD/WithTicketDescription` - sets document content
   - `Build()` - creates file structure and returns project root path
   - Path helper methods: `PhasePath`, `SprintPath`, `TicketPath`, `GoalPath`
   - Automatic cleanup via `t.Cleanup()`

2. **generator.go** - Content generation utilities:
   - `GenerateLoremIpsum(paragraphs int)` - deterministic lorem ipsum text
   - `GenerateRandomString(length int)` - random alphanumeric strings for unique IDs
   - `GenerateRealisticMarkdown(docType string)` - realistic content for README, PRD, ERD, etc.
   - `GenerateGoalName()` - realistic goal names (verb + noun)
   - `GenerateTestSeed(t *testing.T)` - deterministic seed based on test name
   - `NewSeededRandom(t *testing.T)` - seeded random generator for reproducible tests

3. **helpers.go** - Test helper/assertion functions:
   - File assertions: `AssertFileExists`, `AssertFileNotExists`, `AssertFileContent`, `AssertFileContains`
   - Directory assertions: `AssertDirExists`, `AssertDirNotExists`
   - Status assertions: `AssertStatus`, `AssertGoalStatus`, `AssertGoalName`
   - Error assertions: `AssertError`, `AssertErrorIs`, `AssertNoError`
   - File operations: `ReadFile`, `WriteFile`, `CreateFile`, `RemoveFile`, `CreateDir`
   - Directory operations: `ListDir`, `ListDirRecursive`
   - Status operations: `TouchStatusFile`, `SetStatus`
   - Goal helpers: `CountGoals`, `GetOpenGoals`, `GetClosedGoals`, `AssertAllGoalsClosed`, `AssertGoalCount`

Key design decisions:
- All tests create isolated directories under `.test/` (already in `.gitignore`)
- Unique directory names with timestamp + random suffix for parallel test support
- Automatic cleanup registered via `t.Cleanup()` for proper resource management
- Builder pattern auto-creates parent entities (e.g., `WithTicket` creates phase and sprint if missing)
- Generated content is deterministic for reproducibility where needed (lorem ipsum)
- Random strings are non-deterministic for unique test directory names
- Helper functions use `t.Helper()` for proper error location reporting

### Ticket Management (internal/ticket/)

Implemented ticket management functions in `internal/ticket/ticket.go`:

**Core Ticket Functions:**
- `GetOpenTickets(sprintPath string) ([]Ticket, error)` - scans tickets/ for dirs with `open` file (no `done` file)
- `GetNextTicketIndex(sprintPath string) (int, error)` - finds next ticket number (max index + 1, or 1 if none)
- `CreateTicket(sprintPath string, index int) (string, error)` - creates ticket structure:
  - `tickets/XXXX-ticket/` directory (4-digit zero-padded)
  - Empty `README.md` file (AI will populate)
  - `goals/` subdirectory
  - Touches `open` file
  - Returns the ticket path
- `GetTicket(sprintPath, ticketID string) (*Ticket, error)` - loads ticket by ID
- `ListTickets(sprintPath string) ([]Ticket, error)` - lists all tickets sorted by ID

**Ticket Goal Functions:**
- `GetTicketGoals(ticketPath string) ([]Goal, error)` - scans goals/ directory, returns sorted list
- `CreateTicketGoal(ticketPath string, index int, goalName string) (string, error)` - creates goal structure:
  - `goals/XXXX-goal/` directory
  - `name` file with goalName
  - Touches `open` file
- `CloseTicketGoal(ticketPath string, goalID string) error` - validates state, deletes `open`, touches `closed`
- `AreTicketGoalsMet(ticketPath string) (bool, error)` - checks if all goals have `closed` file

**Ticket State Functions:**
- `IsTicketComplete(ticketPath string) (bool, error)` - checks `done` file exists AND all goals closed
- `MarkTicketDone(ticketPath string) error` - validates, deletes `open`, touches `done`
  - Errors if goals still open (returns paths of open goals)
- `ValidateTicketState(ticketPath string) error` - checks for invalid state (both open and done exist)

**Key Design Decisions:**
- Tickets use `open` -> `done` state transition (not `closed`)
- All errors include file paths as relative paths (using .crumbler as anchor)
- Uses `models` package constants: `TicketsDir`, `GoalsDir`, `ReadmeFile`, `StatusFileOpen`, `StatusFileDone`, etc.
- Uses `state` package functions: `IsOpen`, `IsClosed`, `IsDone` (returning `(bool, error)`)
- Private helper functions: `loadTicket`, `loadGoal`, `validateGoalState`, `getOpenGoalPaths`, `getRelPath`, `touchFile`
- Follows same patterns as sprint.go for consistency

**Directory Structure Created:**
```
tickets/XXXX-ticket/
├── open              # empty file = ticket is open
├── README.md         # empty, AI populates
└── goals/            # ticket goals directory
    └── XXXX-goal/
        ├── name      # goal name (AI populates via CreateTicketGoal)
        ├── open      # or closed
```

**Note:** The existing codebase has compilation issues in the state package (redeclared functions, undefined error types). The ticket.go implementation follows correct patterns and will compile once the state package issues are resolved.

### Phase Management Tests (internal/phase/phase_test.go)

Implemented comprehensive phase management tests covering all scenarios from the README:

**Test Structure:**
- Uses `t.Parallel()` for all tests and subtests to enable parallel execution
- Uses table-driven tests with subtests (`t.Run()`) for thorough coverage
- Uses `testutil.NewTestProject(t)` fluent builder for test setup
- All tests clean up automatically via `t.Cleanup()`

**Test Scenarios Implemented:**

1. **TestCreatePhase** - Tests creating phase directory structure:
   - Creates directory, README.md, goals/, sprints/, and open file
   - Tests first phase (0001), second phase (0002), and high index (0099)
   - Verifies all files/directories exist and phase status is "open"

2. **TestCreatePhaseAlreadyExists** - Tests duplicate phase creation fails

3. **TestCreatePhaseGoal** - Tests creating phase goals:
   - Creates numbered goals (0001-goal, 0002-goal, etc.)
   - Verifies name file contains correct goal name
   - Verifies goal status is "open"

4. **TestClosePhaseGoal** - Tests closing phase goals:
   - Closes individual goals
   - Verifies status transitions from "open" to "closed"
   - Verifies other goals remain unaffected

5. **TestClosePhaseWhenAllSprintsAndGoalsClosed** - Tests closing phase when all conditions met:
   - All sprints closed AND all phase goals closed
   - Verifies phase status transitions to "closed"

6. **TestClosePhaseErrorWithOpenSprints** - Tests error when closing with open sprints:
   - Returns hierarchy constraint error
   - Error message includes "open sprints"

7. **TestClosePhaseErrorWithOpenGoals** - Tests error when closing with open goals:
   - Returns hierarchy constraint error
   - Error message includes "open goals"

8. **TestGetOpenPhase** - Tests finding the currently open phase:
   - Single open phase
   - Multiple phases with one open
   - No open phases (returns nil)
   - No phases at all (returns nil)

9. **TestArePhaseGoalsMet** - Tests phase goals met detection:
   - Returns false if no goals or sprints exist
   - Returns false if goals exist but not all closed
   - Returns false if sprints exist but not all closed
   - Returns true only when ALL goals closed AND ALL sprints closed

10. **TestGetNextPhaseIndex** - Tests next phase index calculation:
    - No phases returns 1
    - One phase returns 2
    - Multiple phases returns max + 1
    - Non-sequential phases returns max + 1

11. **Additional Tests:**
    - TestGetPhase - Load phase by ID
    - TestListPhases - List all phases sorted by index
    - TestGetPhaseGoals - Retrieve goals sorted by index
    - TestValidatePhaseState - State validation including invalid states
    - TestClosePhaseNotOpen - Error when closing already closed phase
    - TestPhaseGoalsWithNames - Goals created with proper names
    - TestClosePhaseGoalNotFound - Error when closing non-existent goal
    - TestPhaseWithMultipleGoalsAndSprints - Complex phase with multiple children
    - TestPhaseIndex - Phases have correct indices (1, 10, 100)

**Key Testing Patterns:**
- Use `testutil.AssertStatus(t, path, "open")` to verify status files
- Use `testutil.AssertGoalStatus(t, path, "closed")` for goal status
- Use `testutil.AssertGoalName(t, path, "name")` for goal names
- Use `models.IsHierarchyConstraint(err)` to check error types
- Builder methods chain: `WithPhase().WithPhaseGoal().WithSprint().Build()`

### Ticket Management Tests (internal/ticket/ticket_test.go)

Implemented comprehensive ticket management tests covering all scenarios from the README:

**Test Structure:**
- Uses `t.Parallel()` for all tests and subtests to enable parallel execution
- Uses table-driven tests with subtests (`t.Run()`) for thorough coverage
- Uses `testutil.NewTestProject(t)` fluent builder for test setup
- All tests clean up automatically via `t.Cleanup()`

**Test Scenarios Implemented:**

1. **TestCreateTicket** - Tests creating ticket in sprint:
   - Creates directory structure: `tickets/XXXX-ticket/`, `README.md`, `goals/`, `open` file
   - Tests first ticket (0001), high index (0042), and max 4-digit index (9999)
   - Verifies all files/directories exist and ticket status is "open"

2. **TestCreateTicketGoal** - Tests creating ticket goals:
   - Creates numbered goals (0001-goal, 0002-goal, etc.)
   - Verifies name file contains correct goal name
   - Verifies goal status is "open"
   - Tests goals with long names

3. **TestMarkTicketDone** - Tests marking ticket as done:
   - Marks ticket done with no goals (succeeds)
   - Marks ticket done with all goals closed (succeeds)
   - Error when marking ticket done with open goals (fails with "goals still open")
   - Error when marking already done ticket (fails with "already done")

4. **TestGetOpenTickets** - Tests getting open tickets:
   - No tickets returns empty slice
   - All tickets open returns all tickets sorted by ID
   - Mixed open and done tickets returns only open tickets
   - All tickets done returns empty slice

5. **TestIsTicketComplete** - Tests ticket complete detection:
   - Open ticket is not complete
   - Done ticket with no goals is complete
   - Done ticket with all goals closed is complete
   - Done ticket with open goals is not complete

6. **TestGetNextTicketIndex** - Tests next ticket index calculation:
   - No tickets returns 1
   - One ticket returns 2
   - Multiple tickets returns max + 1
   - Gap in ticket numbers returns max + 1

7. **TestAreTicketGoalsMet** - Tests ticket goals met detection:
   - No goals means goals are met (vacuously true)
   - All goals closed means goals are met
   - One open goal means goals are not met
   - All goals open means goals are not met

8. **TestValidateTicketState** - Tests ticket state validation:
   - Valid open state (passes)
   - Valid done state (passes)
   - Invalid state with both open and done (errors with "both 'open' and 'done'")

9. **TestCloseTicketGoal** - Tests closing ticket goals:
   - Close open goal (succeeds)
   - Error closing already closed goal (fails with "already closed")
   - Error closing non-existent goal (fails with "not found")

10. **TestGetTicketGoals** - Tests getting ticket goals:
    - No goals returns empty slice
    - Single goal returns one goal
    - Multiple goals sorted by index

11. **TestGetTicket** - Tests loading ticket by ID:
    - Get open ticket (returns with StatusOpen)
    - Get done ticket (returns with StatusDone)
    - Error for non-existent ticket (fails with "not found")

12. **TestListTickets** - Tests listing all tickets:
    - No tickets returns empty slice
    - Single ticket returns one ticket
    - Multiple tickets sorted by ID

13. **Integration Tests:**
    - `TestTicketGoalIntegration` - Full workflow: create goals, close them, mark ticket done
    - `TestTicketCreationIntegration` - Create multiple tickets in sequence, verify indexing
    - `TestTicketWithGoalsWorkflow` - Complex workflow with pre-existing goals

14. **Edge Cases:**
    - `TestTicketEdgeCases` - Non-existent paths return empty slices/index 1

**Key Testing Pattern Note:**
When using `WithTicketGoal`, it internally calls `WithTicket` with "open" status. To create a "done" ticket with goals, set goals first, then set ticket status:
```go
b.WithTicketGoal("0001", "0001", "0001", 1, "Goal 1", "closed").
    WithTicket("0001", "0001", "0001", "done")  // Set status last
```

### Sprint Management Tests (internal/sprint/sprint_test.go)

Implemented comprehensive sprint management tests covering all scenarios from the README:

**Test Structure:**
- Uses `t.Parallel()` for all tests and subtests to enable parallel execution
- Uses table-driven tests with subtests (`t.Run()`) for thorough coverage
- Uses `testutil.NewTestProject(t)` fluent builder for test setup
- All tests clean up automatically via `t.Cleanup()`

**Test Scenarios Implemented:**

1. **TestCreateSprint** - Tests creating sprint directory structure:
   - Creates directory, README.md, PRD.md, ERD.md, goals/, tickets/, and open file
   - Tests first sprint (0001), high index (0042), and max 4-digit index (9999)
   - Verifies all files/directories exist and sprint status is "open"

2. **TestCreateSprintGoal** - Tests creating sprint goals:
   - Creates numbered goals (0001-goal, 0100-goal, etc.)
   - Verifies name file contains correct goal name
   - Verifies goal status is "open"

3. **TestCloseSprintSuccess** - Tests closing sprint when all conditions met:
   - All tickets done AND all sprint goals closed
   - Verifies sprint status transitions to "closed"

4. **TestCloseSprintWithOpenTickets** - Tests error when closing with open tickets:
   - Returns error containing "tickets still open"
   - Sprint remains in "open" state

5. **TestCloseSprintWithOpenGoals** - Tests error when closing with open goals:
   - Returns error containing "goals still open"
   - Sprint remains in "open" state

6. **TestGetOpenSprint** - Tests retrieving the open sprint:
   - Finds single open sprint
   - Returns nil when all sprints closed
   - Finds open sprint among multiple sprints
   - Returns nil when no sprints exist

7. **TestAreSprintGoalsMet** - Tests sprint goals met detection:
   - Returns true when all goals closed AND all tickets done
   - Returns false if any goal is still open
   - Returns false if any ticket is still open
   - Returns false if no goals exist (allows CREATE_GOALS)
   - Returns false if no tickets exist (allows CREATE_TICKETS)
   - Returns false if empty sprint (no goals and no tickets)

8. **TestPRDAndERDFileCreation** - Tests PRD.md and ERD.md creation:
   - Verifies both files are created when sprint is created

9. **TestGetNextSprintIndex** - Tests next sprint index calculation:
   - No sprints returns 1
   - One sprint returns 2
   - Multiple sprints (with gaps) returns max + 1

10. **Additional Tests:**
    - TestGetSprintGoals - Retrieve goals sorted by index
    - TestCloseSprintGoal - Close individual sprint goals
    - TestCloseSprintGoalNotFound - Error when closing non-existent goal
    - TestValidateSprintState - State validation including invalid states (both open and closed)
    - TestGetSprint - Load sprint by ID
    - TestListSprints - List all sprints sorted by ID
    - TestCloseSprintAlreadyClosed - Error when closing already closed sprint
    - TestSprintWithGoalsAndTickets - Complex sprint with goals and tickets with their own goals
    - TestMultipleSprintsInPhase - Managing multiple sprints in different states

**Key Testing Patterns:**
- Use `testutil.AssertStatus(t, path, "open")` to verify sprint/ticket status files
- Use `testutil.AssertGoalStatus(t, path, "closed")` for goal status
- Use `testutil.AssertGoalName(t, path, "name")` for goal names
- Use `testutil.AssertFileExists(t, path)` for PRD.md/ERD.md/README.md verification
- Use `testutil.AssertDirExists(t, path)` for directory verification
- Builder methods chain: `WithPhase().WithSprint().WithSprintGoal().WithTicket().Build()`

**Note on Builder Behavior:**
- `WithTicketGoal()` internally calls `WithTicket()` with "open" status
- To set a ticket as "done" with goals, add goals first, then call `WithTicket()` with "done" status last
- Example: `WithTicketGoal(...).WithTicket("phaseID", "sprintID", "ticketID", "done")`

### State Machine Validation Tests (internal/state/state_test.go)

Implemented comprehensive state machine validation tests covering all scenarios from the README:

**Test Structure:**
- Uses `t.Parallel()` for all tests and subtests to enable parallel execution
- Uses table-driven tests with subtests (`t.Run()`) for thorough coverage
- Uses `testutil.NewTestProject(t)` fluent builder for test setup
- All tests clean up automatically via `t.Cleanup()`

**Test Scenarios Implemented:**

1. **TestValidStateTransitions** - Tests valid state transitions:
   - Phase: `open` -> `closed` (succeeds)
   - Sprint: `open` -> `closed` (succeeds)
   - Ticket: `open` -> `done` (succeeds)
   - Verifies status changes correctly after transition

2. **TestInvalidStateTransitions** - Tests invalid state transitions are rejected:
   - Phase `closed` -> `closed` again (fails with "cannot transition")
   - Sprint `closed` -> `closed` again (fails with "cannot transition")
   - Ticket `done` -> `done` again (fails with "cannot transition")

3. **TestMutuallyExclusiveStates** - Tests detection of conflicting state files:
   - Phase with both `open` and `closed` files (errors with "'open' and 'closed'")
   - Sprint with both `open` and `closed` files (errors)
   - Ticket with both `open` and `done` files (errors with "'open' and 'done'")

4. **TestGoalStateTransitions** - Tests goal state transitions at all levels:
   - Phase goal: `open` -> `closed` (succeeds)
   - Sprint goal: `open` -> `closed` (succeeds)
   - Ticket goal: `open` -> `closed` (succeeds)
   - Verifies `open` file removed, `closed` file created

5. **TestGoalMutuallyExclusiveStates** - Tests goal mutually exclusive states:
   - Phase goal with both `open` and `closed` (errors)
   - Sprint goal with both `open` and `closed` (errors)
   - Ticket goal with both `open` and `closed` (errors)

6. **TestHierarchyConstraintPhaseWithOpenSprints** - Tests phase hierarchy constraints:
   - Cannot close phase with open sprint (errors with "open sprints")
   - Cannot close phase with multiple open sprints (errors)
   - Cannot close phase with open phase goals (errors with "open goals")
   - Can close phase with closed sprints and closed goals (succeeds)

7. **TestHierarchyConstraintSprintWithOpenTickets** - Tests sprint hierarchy constraints:
   - Cannot close sprint with open ticket (errors with "open tickets")
   - Cannot close sprint with multiple open tickets (errors)
   - Cannot close sprint with open sprint goals (errors with "open goals")
   - Can close sprint with done tickets and closed goals (succeeds)

8. **TestHierarchyConstraintTicketWithOpenGoals** - Tests ticket hierarchy constraints:
   - Cannot mark ticket done with open goal (errors with "open goals")
   - Cannot mark ticket done with multiple open goals (errors)
   - Can mark ticket done with all goals closed (succeeds)
   - Can mark ticket done with no goals (succeeds - vacuously true)

9. **TestValidateStateMachine** - Tests full state machine validation:
   - Valid state machine with open phase (passes)
   - Valid state machine with closed phase and closed sprints (passes)
   - Invalid: closed phase with open sprint (fails)
   - Invalid: closed sprint with open tickets (fails)
   - Invalid: done ticket with open goals (fails)
   - Invalid: mutually exclusive state in phase (fails)
   - Invalid: closed phase with open phase goals (fails)
   - Invalid: closed sprint with open sprint goals (fails)

10. **TestValidateHierarchy** - Tests hierarchy validation specifically:
    - Valid hierarchy with all entities open (passes)
    - Valid hierarchy with properly closed entities (passes)
    - Invalid: closed phase with open sprint (fails)
    - Invalid: closed phase with open phase goal (fails)
    - Invalid: closed sprint with open ticket (fails)
    - Invalid: closed sprint with open sprint goal (fails)
    - Invalid: done ticket with open ticket goal (fails)

11. **TestValidatePhaseState** / **TestValidateSprintState** / **TestValidateTicketState** - Entity-specific validation:
    - Valid open state (passes)
    - Valid closed/done state (passes)
    - Invalid state with both status files (fails)

12. **TestCollectAllErrors** - Tests collecting multiple validation errors:
    - No errors in valid project (returns empty slice)
    - Multiple mutually exclusive errors are all collected
    - Hierarchy and goal errors are all collected

13. **TestGoalMissingNameFile** - Tests goal validation:
    - Goal without name file is detected as error

14. **TestEmptyProject** - Tests validation of empty project:
    - Empty .crumbler directory passes validation

15. **TestMultiplePhases** - Tests validation with multiple phases:
    - Mix of closed and open phases validates correctly

16. **TestComplexHierarchy** - Tests validation of complex hierarchy:
    - Multiple phases, sprints, tickets, and goals at all levels
    - Properly closed phase 1 with open phase 2
    - All hierarchy constraints satisfied

**Key Testing Patterns:**
- Use `state.NewStateValidator(projectRoot)` to create validator
- Use `validator.CanClosePhase/CanCloseSprint/CanMarkTicketDone` to check transitions
- Use `validator.ValidateStateMachine(projectRoot)` for full validation
- Use `validator.ValidateHierarchy(projectRoot)` for hierarchy-only validation
- Use `state.GetStatus(path)` to check current status
- Use `state.SetClosedValidated(path)` / `state.SetDoneValidated(path)` for transitions
- Create invalid states manually with `testutil.CreateFile(t, path)` and `os.Remove(path)`
- Error messages include relative file paths (e.g., `.crumbler/phases/0001-phase/open`)

**Note on Test Setup:**
When creating invalid hierarchy states for testing, the builder creates entities with their initial status. To create an invalid state (e.g., closed phase with open sprint), you need to:
1. Build with the desired open status initially
2. Manually change the parent status using `os.Remove()` and `testutil.CreateFile()`

Example:
```go
root := testutil.NewTestProject(t).
    WithPhase("0001", "open").
    WithSprint("0001", "0001", "open").
    Build()
// Manually close the phase to create invalid hierarchy
phasePath := filepath.Join(root, ".crumbler", "phases", "0001-phase")
os.Remove(filepath.Join(phasePath, "open"))
testutil.CreateFile(t, filepath.Join(phasePath, "closed"))
```

### Agent Loop Query Tests (internal/query/query_test.go)

Implemented comprehensive agent loop query tests covering all scenarios from the README:

**Test Structure:**
- Uses `t.Parallel()` for all tests and subtests to enable parallel execution
- Uses table-driven tests with subtests (`t.Run()`) for thorough coverage
- Uses `testutil.NewTestProject(t)` fluent builder for test setup
- All tests clean up automatically via `t.Cleanup()`

**Test Scenarios Implemented:**

1. **TestOpenPhaseExists** - Returns true when open phase exists, false when none:
   - True when open phase exists
   - False when no phases exist
   - False when all phases are closed
   - True when at least one phase is open among closed phases
   - Error when phase has both open and closed files (invalid state)

2. **TestRoadmapComplete** - Returns true when all phases closed:
   - True when all phases are closed
   - False when no phases exist
   - False when at least one phase is open
   - False when any phase is not closed
   - Error when phase has invalid state

3. **TestPhaseGoalsMet** - Returns true when all phase goals closed AND all sprints closed:
   - True when all conditions met
   - False if no goals exist (allows CREATE_GOALS)
   - False if no sprints exist (allows CREATE_SPRINT)
   - False if goals exist but not all closed
   - False if sprints exist but not all closed
   - Error when goal or sprint has invalid state

4. **TestOpenSprintExists** - Returns true when open sprint exists, false when none:
   - True when open sprint exists
   - False when no sprints exist
   - False when all sprints are closed
   - True when at least one sprint is open among closed
   - Error when sprint has invalid state

5. **TestSprintGoalsMet** - Returns true when all sprint goals closed AND all tickets done:
   - True when all conditions met
   - False if no goals exist (allows CREATE_GOALS)
   - False if no tickets exist (allows CREATE_TICKETS)
   - False if goals exist but not all closed
   - False if tickets exist but not all done
   - Error when ticket has invalid state

6. **TestOpenTicketsExist** - Returns true when open tickets exist, false when none:
   - True when open tickets exist
   - False when no tickets exist
   - False when all tickets are done
   - True when at least one ticket is open among done
   - Error when ticket has invalid state

7. **TestTicketComplete** - Returns true when done file exists AND all goals closed:
   - True when done file exists AND all goals closed
   - True when done file exists and no goals exist
   - False when done file does not exist
   - False when done file exists but goals not all closed
   - Error when ticket has invalid state
   - Error when goal has invalid state

8. **TestPhaseGoalsExist, TestSprintGoalsExist, TestTicketGoalsExist** - Test existence checks for goals at each level

9. **TestSprintsExist, TestTicketsExist** - Test existence checks for children

10. **TestCompleteWorkflow** - Tests the complete workflow:
    - Roadmap -> phases -> sprints -> tickets -> done
    - Simulates manual creation/status transitions through the full agent loop

11. **TestDecisionPointLogic** - Tests decision point logic from flowchart:
    - **CHECK_PHASE flow**: No open phase -> check roadmap complete -> CREATE_PHASE or EXIT
    - **CHECK_SPRINT flow**: No open sprint -> check phase goals met -> CLOSE_PHASE or CREATE_GOALS/CREATE_SPRINT
    - **CHECK_TICKETS flow**: No open tickets -> check sprint goals met -> CLOSE_SPRINT or CREATE_GOALS/CREATE_TICKETS
    - **TICKET_DONE decision**: Ticket not complete -> continue EXECUTE; Ticket complete -> MARK_DONE

12. **TestEdgeCases** - Tests edge cases:
    - Non-existent project/phase/sprint/ticket paths return false (no error)
    - Empty goals directory returns false for goals exist
    - Files in goals directory don't count as goals
    - Multiple phases with mixed states
    - Deep hierarchy with all levels
### Edge Case and Error Handling Tests (internal/state/edge_cases_test.go)

Implemented comprehensive edge case and error handling tests covering all scenarios from the README:

**Test Structure:**
- Uses `t.Parallel()` for all tests and subtests to enable parallel execution
- Uses table-driven tests with subtests (`t.Run()`) for thorough coverage
- Uses `testutil.NewTestProject(t)` fluent builder for test setup
- All tests clean up automatically via `t.Cleanup()`

**Edge Case Scenarios Implemented:**

1. **TestEmptyProject** - Tests empty project states:
   - No phases directory (validation passes, OpenPhaseExists returns false)
   - Empty phases directory (OpenPhaseExists returns false)

2. **TestPhaseWithNoSprints** - Tests phase with no sprints:
   - Phase exists and is open
   - PhaseGoalsMet returns false (no sprints exist)
   - OpenSprintExists returns false
   - SprintsExist returns false

3. **TestPhaseWithNoGoals** - Tests phase with no goals:
   - PhaseGoalsExist returns false
   - PhaseGoalsMet returns false (no goals exist)

4. **TestSprintWithNoTickets** - Tests sprint with no tickets:
   - OpenTicketsExist returns false
   - TicketsExist returns false
   - SprintGoalsMet returns false (no tickets exist)

5. **TestSprintWithNoGoals** - Tests sprint with no goals:
   - SprintGoalsExist returns false
   - SprintGoalsMet returns false (no goals exist)

6. **TestTicketWithNoGoals** - Tests ticket with no goals:
   - TicketGoalsExist returns false
   - Can mark ticket done (vacuously true - no goals to be open)

7. **TestMultiplePhasesSprintsTicketsGoals** - Tests complex hierarchy:
   - Multiple phases with different states
   - Multiple sprints per phase
   - Multiple tickets per sprint
   - Multiple goals at all levels
   - Validates state machine passes with valid hierarchy

8. **TestRunningOutsideManagedProject** - Tests running outside managed project:
   - No .crumbler directory returns "not a crumbler project" error

9. **TestMissingRequiredFiles** - Tests missing required files:
   - Missing goal name file causes validation error with path
   - Missing status file causes GetStatus error

10. **TestGoalNameFileReading** - Tests goal name file reading (AI-populated content):
    - Simple goal name
    - Goal name with whitespace (trimmed)
    - Multiline goal name
    - Empty goal name after clearing
    - Goal name with special characters

**Error Handling Scenarios Implemented:**

1. **TestErrorMessagesIncludeCorrectFilePaths** - Tests error paths:
   - Invalid phase state error includes relative path
   - Invalid sprint state error includes relative path
   - Invalid ticket state error includes relative path
   - Invalid goal state error includes relative path

2. **TestInvalidStateConflictDetection** - Tests conflicting state detection:
   - Phase with both open and closed files
   - Sprint with both open and closed files
   - Ticket with both open and done files
   - Goal with both open and closed files
   - Each case validates error contains entity path

3. **TestStateValidationErrorsContainProperPaths** - Tests error paths contain proper paths:
   - Hierarchy constraint error with phase paths
   - Hierarchy constraint error with sprint paths
   - Hierarchy constraint error with ticket goal paths
   - Missing goal name error with proper path

4. **TestStateErrorTypes** - Tests StateError types are correct:
   - Invalid state error has type `ErrorTypeMutuallyExclusiveState`
   - Hierarchy constraint error has type `ErrorTypeHierarchyConstraint`
   - Missing goal name error has type `ErrorTypeMissingGoalName`
   - Project not found has type `ErrorTypeOrphanedState`

5. **TestCollectAllErrorsEdgeCases** - Tests collecting multiple errors:
   - Creates project with multiple error types
   - Verifies both mutually exclusive and hierarchy errors are collected

6. **TestQueryFunctionErrorsWithConflictingState** - Tests query function error handling:
   - OpenPhaseExists with conflicting state returns error
   - OpenSprintExists with conflicting state returns error
   - OpenTicketsExist with conflicting state returns error
   - TicketComplete with conflicting state returns error
   - PhaseGoalsMet with conflicting goal state returns error

7. **TestGetStatusFunction** - Tests GetStatus edge cases:
   - Returns correct status for open, closed, done
   - Returns error for no status file
   - Returns error for conflicting status files (open+closed, open+done, closed+done)

8. **TestSetClosedValidated / TestSetDoneValidated** - Tests state transition validation:
   - Valid transition from open succeeds
   - Invalid transition from closed/done fails

9. **TestListGoals** - Tests goal listing:
   - Goals sorted by index
   - Empty directory returns empty slice
   - Non-existent directory returns nil

10. **TestGetNextGoalIndex** - Tests next goal index calculation:
    - First goal returns 1
    - Returns max + 1 after existing goals
    - Handles gaps in numbering (returns max + 1, doesn't fill gaps)

11. **TestAreAllGoalsClosed** - Tests all goals closed detection:
    - No goals = vacuously true
    - All goals closed = true
    - One goal open = false
    - All goals open = false

**Key Testing Patterns:**
- Create invalid states manually after builder (builder creates valid states)
- Use `testutil.RemoveFile()` and `testutil.CreateFile()` to manipulate status
- Error messages checked with `strings.Contains()` for path verification
- Use `models.IsStateError()` and `models.AsStateError()` for error type checking
- Use `models.ErrorTypeMutuallyExclusiveState`, `models.ErrorTypeHierarchyConstraint` for type verification

**Note on Builder Behavior:**
- `WithTicket()` internally calls `WithSprint()` with "open" status
- `WithTicketGoal()` internally calls `WithTicket()` with "open" status
- To create closed/done parents with open children (invalid hierarchy), build with open status first, then manually change parent status after build
