package crumbler

import (
	"fmt"

	"github.com/waynenilsen/crumbler/internal/prompt"
)

// runPrompt handles the 'crumbler prompt' command.
// It generates the AI agent prompt based on current state.
func runPrompt(args []string) error {
	config := &prompt.Config{}

	// Parse flags
	for _, arg := range args {
		switch arg {
		case "--help", "-h", "help":
			printPromptHelp()
			return nil
		case "--no-preamble":
			config.NoPreamble = true
		case "--no-postamble":
			config.NoPostamble = true
		case "--no-context":
			config.NoContext = true
		case "--minimal":
			config.Minimal = true
		case "--state-only":
			config.StateOnly = true
		default:
			return fmt.Errorf("unknown flag: %s\n\nRun 'crumbler prompt --help' for usage", arg)
		}
	}

	projectRoot, err := getProjectRoot()
	if err != nil {
		return err
	}

	// Generate the prompt
	output, err := prompt.GeneratePrompt(projectRoot, config)
	if err != nil {
		return fmt.Errorf("failed to generate prompt: %w", err)
	}

	fmt.Print(output)
	return nil
}

// printPromptHelp prints help for the prompt command.
func printPromptHelp() {
	fmt.Print(`crumbler prompt - Generate AI agent prompt

USAGE:
    crumbler prompt [flags]

DESCRIPTION:
    Generates a structured prompt for AI agents based on the current project
    state. The prompt includes context about the current crumb and instructions
    for what to do next.

FLAGS:
    --no-preamble    Skip the preamble section (crumbler explanation)
    --no-postamble   Skip the postamble section (next steps)
    --no-context     Skip the context section (README contents)
    --minimal        Use minimal preamble/postamble
    --state-only     Output only the state name (DECOMPOSE, EXECUTE, or DONE)

STATES:
    DONE        No crumbs remain - project is complete
    DECOMPOSE   Current crumb's README is empty - plan the work
    EXECUTE     Current crumb's README has content - do the work

EXAMPLES:
    # Get full prompt
    crumbler prompt

    # Get just the state
    crumbler prompt --state-only

    # Get minimal prompt
    crumbler prompt --minimal

    # Get prompt without context
    crumbler prompt --no-context

AGENT LOOP:
    The typical agent loop is:
    1. crumbler prompt          # Get instructions
    2. [Do the work]            # Follow instructions
    3. crumbler delete          # If work is done
    4. crumbler prompt          # Get next instructions

OUTPUT FORMAT:
    The prompt includes:
    - Preamble: Explanation of crumbler system
    - Context: Current crumb path and README contents
    - Instructions: State-specific guidance
    - Postamble: Next steps to run
`)
}
