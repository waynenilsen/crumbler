package crumbler

import (
	"fmt"
	"os"

	"github.com/waynenilsen/crumbler/internal/prompt"
)

// runGetNextPrompt handles the get-next-prompt command.
func runGetNextPrompt(args []string) error {
	// Parse flags
	config := prompt.DefaultConfig()
	stateOnly := false

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--help", "-h":
			printGetNextPromptHelp()
			return nil
		case "--no-prelude":
			config.IncludePrelude = false
		case "--no-postlude":
			config.IncludePostlude = false
		case "--no-context":
			config.IncludeContext = false
		case "--minimal":
			config.Minimal = true
		case "--state-only":
			stateOnly = true
		default:
			fmt.Fprintf(os.Stderr, "error: unknown flag '%s'\n\n", args[i])
			printGetNextPromptHelp()
			return fmt.Errorf("unknown flag: %s", args[i])
		}
	}

	// Find project root
	projectRoot, err := requireProject()
	if err != nil {
		return err
	}

	// If state-only, just print the state
	if stateOnly {
		state, err := prompt.GetStateString(projectRoot)
		if err != nil {
			return fmt.Errorf("failed to determine state: %w", err)
		}
		fmt.Println(state)
		return nil
	}

	// Generate full prompt
	output, err := prompt.GeneratePrompt(projectRoot, config)
	if err != nil {
		return fmt.Errorf("failed to generate prompt: %w", err)
	}

	fmt.Print(output)
	return nil
}

// printGetNextPromptHelp prints help for the get-next-prompt command.
func printGetNextPromptHelp() {
	fmt.Print(`crumbler get-next-prompt - Generate AI agent prompt

DESCRIPTION:
    Generates a structured prompt telling an AI agent exactly what to do next
    based on the current state machine state. This enables hands-off development
    where the agent simply executes instructions in a loop.

    The prompt includes:
    - PRELUDE: What crumbler is, how state works, current position
    - CONTEXT: Relevant file contents (roadmap, README, PRD, ERD, goals)
    - INSTRUCTION: State-specific steps to execute
    - POSTLUDE: Commands to run next, helpful references

USAGE:
    crumbler get-next-prompt [flags]

FLAGS:
    -h, --help       Show this help message
    --no-prelude     Skip the prelude section
    --no-postlude    Skip the postlude section
    --no-context     Skip the context section (file contents)
    --minimal        Use minimal prelude/postlude (less verbose)
    --state-only     Only output the current state name (e.g., "CREATE_PHASE")

STATES:
    EXIT               - Project complete, all phases closed
    CREATE_PHASE       - Create next phase from roadmap
    CREATE_PHASE_GOALS - Create goals for current phase
    CREATE_SPRINT      - Create sprint in current phase
    CLOSE_PHASE        - Close current phase (all sprints/goals done)
    CREATE_SPRINT_GOALS - Create goals for current sprint
    CREATE_TICKETS     - Decompose sprint into tickets
    CLOSE_SPRINT       - Close current sprint (all tickets/goals done)
    CREATE_TICKET_GOALS - Create goals for current ticket
    EXECUTE_TICKET     - Execute current ticket (work on goals)
    MARK_TICKET_DONE   - Mark current ticket as done

EXAMPLES:
    # Generate full prompt
    crumbler get-next-prompt

    # Generate minimal prompt (less verbose)
    crumbler get-next-prompt --minimal

    # Generate prompt without context files
    crumbler get-next-prompt --no-context

    # Just get the current state
    crumbler get-next-prompt --state-only

    # Agent loop pattern
    while true; do
        prompt=$(crumbler get-next-prompt)
        if echo "$prompt" | grep -q "STATE: EXIT"; then
            echo "Project complete!"
            break
        fi
        # Pass prompt to AI agent
        claude --prompt "$prompt"
    done

VALUE PROPOSITION:
    Claude Max 20x subscription ($200/month for ~$3600 of API tokens) requires
    using Claude CLI. This command maximizes that value by encoding workflow
    logic in Go, minimizing agent decision-making overhead and preventing
    hallucination about workflow state.

SEE ALSO:
    crumbler status     - Show current project status
    crumbler help       - General help
`)
}
