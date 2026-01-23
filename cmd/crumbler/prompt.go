package crumbler

import (
	"fmt"

	"github.com/waynenilsen/crumbler/internal/prompt"
)

// runPrompt handles the 'crumbler prompt' command.
// It generates the AI agent prompt based on current crumb.
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
    Generates a structured prompt for AI agents based on the current crumb.
    The prompt includes context about the current crumb and instructions
    for what to do next. The agent decides whether to execute or decompose.

FLAGS:
    --no-preamble    Skip the preamble section (crumbler explanation)
    --no-postamble   Skip the postamble section (next steps)
    --no-context     Skip the context section (README contents)
    --minimal        Use minimal preamble/postamble

EXAMPLES:
    # Get full prompt
    crumbler prompt

    # Get minimal prompt
    crumbler prompt --minimal

    # Get prompt without context
    crumbler prompt --no-context

AGENT LOOP:
    The typical agent loop is:
    1. crumbler prompt          # Get instructions
    2. [Decide: execute or decompose]
    3. [Do the work or create sub-crumbs]
    4. crumbler delete          # If work is done
    5. exit                     # Context resets, loop continues

OUTPUT FORMAT:
    The prompt includes:
    - Preamble: Explanation of crumbler system and decision options
    - Context: Current crumb path and README contents
    - Postamble: Reminder to exit when done
`)
}
