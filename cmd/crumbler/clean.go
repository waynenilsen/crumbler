package crumbler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ariel-frischer/claude-clean/display"
	"github.com/ariel-frischer/claude-clean/parser"
)

// runClean handles the 'crumbler clean' command.
// It reads Claude Code streaming JSON and formats it beautifully.
func runClean(args []string) error {
	// Handle help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help") {
		printCleanHelp()
		return nil
	}

	// Parse flags
	cfg := &display.Config{
		Style:       display.StyleDefault,
		Verbose:     false,
		ShowLineNum: false,
	}
	var showUsage bool
	var inputFile string

	// Simple flag parsing
	for i, arg := range args {
		switch arg {
		case "-s", "--style":
			if i+1 < len(args) {
				cfg.Style = display.OutputStyle(args[i+1])
			}
		case "-v", "--verbose":
			cfg.Verbose = true
		case "-l", "--line-numbers":
			cfg.ShowLineNum = true
		case "-V", "--usage":
			showUsage = true
		default:
			// If it's not a flag and doesn't start with -, treat as input file
			if !strings.HasPrefix(arg, "-") && inputFile == "" {
				inputFile = arg
			}
		}
	}

	// Determine input source
	var input io.Reader
	if inputFile != "" {
		file, err := os.Open(inputFile)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()
		input = file
	} else {
		// Read from stdin
		input = os.Stdin
	}

	// Process input line by line
	scanner := bufio.NewScanner(input)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Parse JSON line
		var msg parser.StreamMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			// Skip invalid JSON lines
			continue
		}

		// Display the message
		display.DisplayMessage(&msg, lineNum, cfg)

		// Show usage if requested
		if showUsage && msg.Usage != nil {
			display.DisplayUsage(msg.Usage)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	return nil
}

// printCleanHelp prints help for the clean command.
func printCleanHelp() {
	fmt.Print(`crumbler clean - Format Claude Code streaming JSON output

USAGE:
    crumbler clean [options] [file]

DESCRIPTION:
    Reads Claude Code streaming JSON output (from stdin or a file) and formats
    it into beautiful, readable terminal output. This is useful for cleaning
    up verbose JSON logs from Claude Code commands.

    The command parses JSON lines and displays them in a formatted, colorized
    way, making it easy to read Claude Code's output.

INPUT:
    If a file is provided, reads from that file. Otherwise, reads from stdin.

OPTIONS:
    -s, --style STYLE       Output style: default, compact, minimal, or plain
                            (default: default)
    -v, --verbose           Show system reminders
    -l, --line-numbers      Show source line numbers
    -V, --usage             Show token usage statistics
    -h, --help              Show this help message

STYLES:
    default                 Boxed, colorful, easy to read (default)
    compact                 Single-line, quick scanning
    minimal                 No boxes, still colored
    plain                   No colors, great for logs

EXAMPLES:
    # Read from stdin (pipe Claude Code output)
    claude -p "your prompt" --verbose --output-format stream-json | crumbler clean

    # Read from a file
    crumbler clean logs.jsonl

    # Use compact style
    crumbler clean -s compact logs.jsonl

    # Show line numbers and usage stats
    crumbler clean -l -V logs.jsonl

    # Minimal style with verbose output
    crumbler clean -s minimal -v logs.jsonl

INTEGRATION WITH CLAUDE CODE:
    Claude Code can output streaming JSON format. Pipe it directly to crumbler clean:

    claude -p "your prompt" --verbose --output-format stream-json | crumbler clean

    This bypasses Claude's interactive UI and provides clean, readable output.

OUTPUT TYPES:
    The command recognizes and formats different message types:
    - SYSTEM      (Cyan)    Init, config, session info
    - ASSISTANT   (Green)   Claude's responses
    - TOOL        (Yellow)  Tool calls (Bash, Read, Write...)
    - RESULT      (Gray)    Tool output
    - ERROR       (Red)     Tool errors
    - USAGE       (Magenta) Token stats & costs

FOR AI AGENTS:
    Use 'crumbler clean' to format Claude Code JSON output when you need to
    read or debug Claude Code command output. This is especially useful when:

    1. Debugging Claude Code commands that output JSON
    2. Reading verbose logs from Claude Code
    3. Formatting output for better readability
    4. Extracting usage statistics from Claude Code sessions

    The command preserves all information while making it human-readable.
`)
}
