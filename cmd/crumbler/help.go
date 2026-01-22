package crumbler

import (
	"fmt"
	"os"
)

// runHelp handles the 'crumbler help' command.
func runHelp(args []string) error {
	if len(args) == 0 {
		printTopLevelHelp()
		return nil
	}

	// Route to command-specific help
	switch args[0] {
	case "init":
		printInitHelp()
	case "status":
		printStatusHelp()
	case "phase":
		if len(args) > 1 {
			return runPhaseHelpSubcommand(args[1:])
		}
		printPhaseHelp()
	case "sprint":
		if len(args) > 1 {
			return runSprintHelpSubcommand(args[1:])
		}
		printSprintHelp()
	case "ticket":
		if len(args) > 1 {
			return runTicketHelpSubcommand(args[1:])
		}
		printTicketHelp()
	case "roadmap":
		if len(args) > 1 {
			return runRoadmapHelpSubcommand(args[1:])
		}
		printRoadmapHelp()
	case "clean":
		printCleanHelp()
	case "help":
		printHelpHelp()
	default:
		fmt.Fprintf(os.Stderr, "error: unknown command '%s'\n\n", args[0])
		printTopLevelHelp()
		return fmt.Errorf("unknown command: %s", args[0])
	}

	return nil
}

// runPhaseHelpSubcommand handles 'crumbler help phase <subcommand>'.
func runPhaseHelpSubcommand(args []string) error {
	switch args[0] {
	case "list":
		printPhaseListHelp()
	case "create":
		printPhaseCreateHelp()
	case "close":
		printPhaseCloseHelp()
	case "goal":
		if len(args) > 1 {
			return runPhaseGoalHelpSubcommand(args[1:])
		}
		printPhaseGoalHelp()
	default:
		fmt.Fprintf(os.Stderr, "error: unknown phase subcommand '%s'\n\n", args[0])
		printPhaseHelp()
		return fmt.Errorf("unknown phase subcommand: %s", args[0])
	}
	return nil
}

// runPhaseGoalHelpSubcommand handles 'crumbler help phase goal <subcommand>'.
func runPhaseGoalHelpSubcommand(args []string) error {
	switch args[0] {
	case "create":
		printPhaseGoalCreateHelp()
	case "close":
		printPhaseGoalCloseHelp()
	default:
		fmt.Fprintf(os.Stderr, "error: unknown phase goal subcommand '%s'\n\n", args[0])
		printPhaseGoalHelp()
		return fmt.Errorf("unknown phase goal subcommand: %s", args[0])
	}
	return nil
}

// runSprintHelpSubcommand handles 'crumbler help sprint <subcommand>'.
func runSprintHelpSubcommand(args []string) error {
	switch args[0] {
	case "list":
		printSprintListHelp()
	case "create":
		printSprintCreateHelp()
	case "close":
		printSprintCloseHelp()
	case "goal":
		if len(args) > 1 {
			return runSprintGoalHelpSubcommand(args[1:])
		}
		printSprintGoalHelp()
	default:
		fmt.Fprintf(os.Stderr, "error: unknown sprint subcommand '%s'\n\n", args[0])
		printSprintHelp()
		return fmt.Errorf("unknown sprint subcommand: %s", args[0])
	}
	return nil
}

// runSprintGoalHelpSubcommand handles 'crumbler help sprint goal <subcommand>'.
func runSprintGoalHelpSubcommand(args []string) error {
	switch args[0] {
	case "create":
		printSprintGoalCreateHelp()
	case "close":
		printSprintGoalCloseHelp()
	default:
		fmt.Fprintf(os.Stderr, "error: unknown sprint goal subcommand '%s'\n\n", args[0])
		printSprintGoalHelp()
		return fmt.Errorf("unknown sprint goal subcommand: %s", args[0])
	}
	return nil
}

// runTicketHelpSubcommand handles 'crumbler help ticket <subcommand>'.
func runTicketHelpSubcommand(args []string) error {
	switch args[0] {
	case "list":
		printTicketListHelp()
	case "create":
		printTicketCreateHelp()
	case "done":
		printTicketDoneHelp()
	case "goal":
		if len(args) > 1 {
			return runTicketGoalHelpSubcommand(args[1:])
		}
		printTicketGoalHelp()
	default:
		fmt.Fprintf(os.Stderr, "error: unknown ticket subcommand '%s'\n\n", args[0])
		printTicketHelp()
		return fmt.Errorf("unknown ticket subcommand: %s", args[0])
	}
	return nil
}

// runTicketGoalHelpSubcommand handles 'crumbler help ticket goal <subcommand>'.
func runTicketGoalHelpSubcommand(args []string) error {
	switch args[0] {
	case "create":
		printTicketGoalCreateHelp()
	case "close":
		printTicketGoalCloseHelp()
	default:
		fmt.Fprintf(os.Stderr, "error: unknown ticket goal subcommand '%s'\n\n", args[0])
		printTicketGoalHelp()
		return fmt.Errorf("unknown ticket goal subcommand: %s", args[0])
	}
	return nil
}

// runRoadmapHelpSubcommand handles 'crumbler help roadmap <subcommand>'.
func runRoadmapHelpSubcommand(args []string) error {
	switch args[0] {
	case "load":
		printRoadmapLoadHelp()
	case "show":
		printRoadmapShowHelp()
	default:
		fmt.Fprintf(os.Stderr, "error: unknown roadmap subcommand '%s'\n\n", args[0])
		printRoadmapHelp()
		return fmt.Errorf("unknown roadmap subcommand: %s", args[0])
	}
	return nil
}

// printHelpHelp prints help for the help command itself.
func printHelpHelp() {
	fmt.Print(`crumbler help - Show help for commands

USAGE:
    crumbler help [command] [subcommand]

DESCRIPTION:
    Shows help documentation for crumbler commands. Help is available at
    multiple levels:

    Top-level:     crumbler help
    Command:       crumbler help <command>
    Subcommand:    crumbler help <command> <subcommand>

    Alternatively, use --help or -h flag with any command:

    crumbler --help
    crumbler phase --help
    crumbler phase create --help

AVAILABLE COMMANDS:
    init      Initialize a new crumbler project
    status    Show current project state
    phase     Manage phases (list, create, close, goal)
    sprint    Manage sprints (list, create, close, goal)
    ticket    Manage tickets (list, create, done, goal)
    roadmap   Manage roadmap (load, show)
    clean     Format Claude Code streaming JSON output

EXAMPLES:
    crumbler help                         Show top-level help
    crumbler help init                    Show help for init command
    crumbler help phase                   Show help for phase command
    crumbler help phase create            Show help for phase create
    crumbler help phase goal              Show help for phase goal commands
    crumbler help phase goal create       Show help for phase goal create

FOR AI AGENTS:
    The help system is designed to provide self-contained documentation at
    each level. When you need to understand a command:

    1. Start with command-level help:
       crumbler help <command>

    2. Then get subcommand-level help for specific operations:
       crumbler help <command> <subcommand>

    Each help document includes:
    - Usage syntax
    - Description of what the command does
    - Arguments and flags
    - File paths created/modified
    - Examples
    - Tips for AI agents

    This hierarchical help structure prevents context rot - you only load
    the documentation you need for the current task.

QUICK REFERENCE:

    Initialize project:
        crumbler init

    Check status:
        crumbler status

    Phase workflow:
        crumbler phase create
        crumbler phase goal create <phase-id> <goal-name>
        crumbler phase goal close <phase-id> <goal-id>
        crumbler phase close <phase-id>

    Sprint workflow:
        crumbler sprint create [phase-id]
        crumbler sprint goal create <sprint-id> <goal-name>
        crumbler sprint goal close <sprint-id> <goal-id>
        crumbler sprint close <sprint-id>

    Ticket workflow:
        crumbler ticket create [sprint-id]
        crumbler ticket goal create <ticket-id> <goal-name>
        crumbler ticket goal close <ticket-id> <goal-id>
        crumbler ticket done <ticket-id>

    Roadmap:
        crumbler roadmap load <file>
        crumbler roadmap show

    Clean:
        claude -p "prompt" --output-format stream-json | crumbler clean
        crumbler clean logs.jsonl
`)
}
