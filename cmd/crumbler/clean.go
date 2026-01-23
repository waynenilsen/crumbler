package crumbler

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ariel-frischer/claude-clean/display"
	"github.com/ariel-frischer/claude-clean/parser"
)

const (
	defaultChunkSize = 8 * 1024  // 8KB chunks
	maxBufferSize    = 10 * 1024 * 1024 // 10MB max buffer
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

	// Process input stream
	return processJSONStream(input, cfg, showUsage)
}

// skipWhitespace removes leading whitespace from buffer, handling CRLF, LF, CR, spaces, and tabs.
// Returns the number of bytes skipped.
func skipWhitespace(buffer []byte) int {
	start := 0
	for start < len(buffer) {
		b := buffer[start]
		if b == ' ' || b == '\t' {
			start++
		} else if b == '\r' {
			// Handle CRLF (\r\n) as a unit, or standalone CR
			if start+1 < len(buffer) && buffer[start+1] == '\n' {
				start += 2 // Skip CRLF
			} else {
				start++ // Skip standalone CR
			}
		} else if b == '\n' {
			start++ // Skip LF
		} else {
			break // Non-whitespace found
		}
	}
	return start
}

// findCompleteJSON finds the first complete JSON object in the buffer by tracking brace depth.
// Returns the start and end indices of the JSON object, or -1, -1 if no complete object found.
// Handles strings and escape sequences correctly.
func findCompleteJSON(buffer []byte) (start, end int) {
	start = -1
	end = -1
	depth := 0
	inString := false
	escapeNext := false

	for i := 0; i < len(buffer); i++ {
		b := buffer[i]

		if escapeNext {
			escapeNext = false
			continue
		}

		if b == '\\' {
			escapeNext = true
			continue
		}

		if b == '"' {
			inString = !inString
			continue
		}

		if inString {
			continue
		}

		// Track brace depth to find complete objects
		if b == '{' {
			if start == -1 {
				start = i
			}
			depth++
		} else if b == '}' {
			depth--
			if depth == 0 && start != -1 {
				end = i + 1
				break
			}
		}
	}

	return start, end
}

// processJSONStream processes a stream of JSON objects from the input reader.
// It reads in chunks, accumulates complete JSON objects, and processes them.
func processJSONStream(input io.Reader, cfg *display.Config, showUsage bool) (err error) {
	// Top-level panic recovery to ensure we never crash with nonzero exit
	// Even if the library panics, we want to exit successfully (exit code 0)
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "warning: recovered from panic in processJSONStream: %v\n", r)
			// Return nil to ensure successful exit code
			err = nil
		}
	}()

	buffer := make([]byte, 0, defaultChunkSize*2)
	chunk := make([]byte, defaultChunkSize)
	lineNum := 0

	var readErr error
	for {
		// Read chunk from input
		n, err := input.Read(chunk)
		if n > 0 {
			// Append chunk to buffer
			buffer = append(buffer, chunk[:n]...)
		}
		
		// Track EOF separately - Read can return data AND EOF
		if err == io.EOF {
			readErr = io.EOF
		} else if err != nil {
			return fmt.Errorf("error reading input: %w", err)
		}

		// Process complete JSON objects from buffer
		processed := false
		for {
			// Skip leading whitespace
			skip := skipWhitespace(buffer)
			if skip > 0 {
				buffer = buffer[skip:]
			}

			// Find complete JSON object
			jsonStart, jsonEnd := findCompleteJSON(buffer)

			// If we found a complete JSON object, parse and display it
			if jsonEnd > 0 && jsonStart >= 0 {
				jsonData := buffer[jsonStart:jsonEnd]

				// Wrap processing in panic recovery to prevent crashes
				func() {
					defer func() {
						if r := recover(); r != nil {
							// Always log panic recovery - library crashes should be visible
							fmt.Fprintf(os.Stderr, "warning: recovered from panic at offset %d: %v\n", jsonStart, r)
							if cfg.Verbose {
								// In verbose mode, also show the JSON that caused the panic
								fmt.Fprintf(os.Stderr, "  JSON data: %s\n", string(jsonData))
							}
						}
					}()

					// Parse JSON object
					var msg parser.StreamMessage
					if err := json.Unmarshal(jsonData, &msg); err != nil {
						// Skip invalid JSON
						if cfg.Verbose {
							fmt.Fprintf(os.Stderr, "warning: skipped invalid JSON at offset %d: %v\n", jsonStart, err)
						}
					} else {
						lineNum++
						// Display the message (may panic on unexpected formats)
						display.DisplayMessage(&msg, lineNum, cfg)

						// Show usage if requested (may panic)
						if showUsage && msg.Usage != nil {
							display.DisplayUsage(msg.Usage)
						}
					}
				}()

				// Remove processed JSON from buffer
				buffer = buffer[jsonEnd:]
				processed = true
				continue // Try to find another complete object
			}

			// No complete object found, break and read more data
			break
		}

		// If we've hit EOF and processed everything we can, exit
		if readErr == io.EOF {
			// Check if there's remaining data in buffer (incomplete JSON)
			remaining := strings.TrimSpace(string(buffer))
			if remaining != "" {
				if cfg.Verbose {
					fmt.Fprintf(os.Stderr, "warning: incomplete JSON at end of input (%d bytes)\n", len(buffer))
				}
			}
			break
		}

		// If we didn't process anything and didn't read anything, we're stuck
		if !processed && n == 0 && readErr == nil {
			// This shouldn't happen with a well-behaved reader, but protect against infinite loops
			break
		}

		// Prevent buffer from growing unbounded
		if len(buffer) > maxBufferSize {
			return fmt.Errorf("error: buffer too large (max %dMB), possible malformed input", maxBufferSize/(1024*1024))
		}
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
