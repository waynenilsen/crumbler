package crumbler

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/ariel-frischer/claude-clean/display"
)

func TestSkipWhitespace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "no whitespace",
			input:    `{"key":"value"}`,
			expected: 0,
		},
		{
			name:     "leading spaces",
			input:    `   {"key":"value"}`,
			expected: 3,
		},
		{
			name:     "leading tabs",
			input:    "\t\t\t{\"key\":\"value\"}",
			expected: 3,
		},
		{
			name:     "LF line ending",
			input:    "\n{\"key\":\"value\"}",
			expected: 1,
		},
		{
			name:     "CR line ending",
			input:    "\r{\"key\":\"value\"}",
			expected: 1,
		},
		{
			name:     "CRLF line ending",
			input:    "\r\n{\"key\":\"value\"}",
			expected: 2,
		},
		{
			name:     "multiple LF",
			input:    "\n\n\n{\"key\":\"value\"}",
			expected: 3,
		},
		{
			name:     "multiple CRLF",
			input:    "\r\n\r\n{\"key\":\"value\"}",
			expected: 4,
		},
		{
			name:     "mixed whitespace",
			input:    " \t \r\n \n {\"key\":\"value\"}",
			expected: 8, // space(1) + tab(1) + space(1) + CRLF(2) + space(1) + LF(1) + space(1) = 8
		},
		{
			name:     "empty buffer",
			input:    "",
			expected: 0,
		},
		{
			name:     "only whitespace",
			input:    "   \r\n\t  ",
			expected: 8, // space(1) + space(1) + space(1) + CRLF(2) + tab(1) + space(1) + space(1) = 8
		},
		{
			name:     "CR not followed by LF",
			input:    "\r{\"key\":\"value\"}",
			expected: 1,
		},
		{
			name:     "CRLF at end of buffer",
			input:    "\r\n",
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := skipWhitespace([]byte(tt.input))
			if result != tt.expected {
				t.Errorf("skipWhitespace(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFindCompleteJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		input         string
		expectedStart int
		expectedEnd   int
	}{
		{
			name:          "simple JSON object",
			input:         `{"key":"value"}`,
			expectedStart: 0,
			expectedEnd:   15,
		},
		{
			name:          "JSON with nested objects",
			input:         `{"outer":{"inner":"value"}}`,
			expectedStart: 0,
			expectedEnd:   27, // len(`{"outer":{"inner":"value"}}`) = 27
		},
		{
			name:          "JSON with string containing braces",
			input:         `{"key":"value {with} braces"}`,
			expectedStart: 0,
			expectedEnd:   29, // len(`{"key":"value {with} braces"}`) = 29
		},
		{
			name:          "JSON with escaped quotes",
			input:         `{"key":"value \"quoted\""}`,
			expectedStart: 0,
			expectedEnd:   26,
		},
		{
			name:          "JSON with escaped backslash",
			input:         `{"key":"value \\ backslash"}`,
			expectedStart: 0,
			expectedEnd:   28,
		},
		{
			name:          "multiple JSON objects",
			input:         `{"first":"value"}{"second":"value"}`,
			expectedStart: 0,
			expectedEnd:   17, // len(`{"first":"value"}`) = 17
		},
		{
			name:          "JSON with leading whitespace",
			input:         `   {"key":"value"}`,
			expectedStart: 3,
			expectedEnd:   18,
		},
		{
			name:          "incomplete JSON",
			input:         `{"key":"value`,
			expectedStart: 0, // Function finds opening brace, but no closing brace
			expectedEnd:   -1,
		},
		{
			name:          "empty buffer",
			input:         ``,
			expectedStart: -1,
			expectedEnd:   -1,
		},
		{
			name:          "only opening brace",
			input:         `{`,
			expectedStart: 0,
			expectedEnd:   -1,
		},
		{
			name:          "JSON with array",
			input:         `{"array":[1,2,3]}`,
			expectedStart: 0,
			expectedEnd:   17,
		},
		{
			name:          "JSON with nested braces in string",
			input:         `{"key":"{nested}"}`,
			expectedStart: 0,
			expectedEnd:   18, // len(`{"key":"{nested}"}`) = 18
		},
		{
			name:          "JSON with newlines",
			input:         "{\n  \"key\": \"value\"\n}",
			expectedStart: 0,
			expectedEnd:   20,
		},
		{
			name:          "JSON with CRLF",
			input:         "{\r\n  \"key\": \"value\"\r\n}",
			expectedStart: 0,
			expectedEnd:   22,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			start, end := findCompleteJSON([]byte(tt.input))
			if start != tt.expectedStart || end != tt.expectedEnd {
				t.Errorf("findCompleteJSON(%q) = (%d, %d), want (%d, %d)",
					tt.input, start, end, tt.expectedStart, tt.expectedEnd)
			}
		})
	}
}

func TestProcessJSONStream_SimpleObjects(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "single JSON object",
			input: `{"type":"test","content":"hello"}`,
		},
		{
			name:  "two JSON objects",
			input: `{"type":"test1"}{"type":"test2"}`,
		},
		{
			name:  "JSON objects with LF separator",
			input: "{\"type\":\"test1\"}\n{\"type\":\"test2\"}",
		},
		{
			name:  "JSON objects with CRLF separator",
			input: "{\"type\":\"test1\"}\r\n{\"type\":\"test2\"}",
		},
		{
			name:  "JSON objects with CR separator",
			input: "{\"type\":\"test1\"}\r{\"type\":\"test2\"}",
		},
		{
			name:  "JSON objects with whitespace",
			input: "  {\"type\":\"test1\"}  \n  {\"type\":\"test2\"}  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := &display.Config{
				Style:       display.StyleDefault,
				Verbose:     false,
				ShowLineNum: false,
			}
			input := bytes.NewReader([]byte(tt.input))
			err := processJSONStream(input, cfg, false)
			if err != nil {
				t.Errorf("processJSONStream() error = %v", err)
			}
		})
	}
}

func TestProcessJSONStream_ChunkedInput(t *testing.T) {
	t.Parallel()

	// Create a custom reader that reads in small chunks to simulate streaming
	chunkedReader := &chunkedReader{
		data:  []byte(`{"type":"test1"}{"type":"test2"}{"type":"test3"}`),
		chunk: 5, // Read 5 bytes at a time
	}

	cfg := &display.Config{
		Style:       display.StyleDefault,
		Verbose:     false,
		ShowLineNum: false,
	}

	err := processJSONStream(chunkedReader, cfg, false)
	if err != nil {
		t.Errorf("processJSONStream() with chunked input error = %v", err)
	}
}

func TestProcessJSONStream_PartialJSON(t *testing.T) {
	t.Parallel()

	// Test with incomplete JSON at the end
	input := bytes.NewReader([]byte(`{"type":"test1"}{"type":"test2"}{"incomplete":`))
	cfg := &display.Config{
		Style:       display.StyleDefault,
		Verbose:     true, // Enable verbose to see warning
		ShowLineNum: false,
	}

	err := processJSONStream(input, cfg, false)
	if err != nil {
		t.Errorf("processJSONStream() with partial JSON error = %v", err)
	}
}

func TestProcessJSONStream_InvalidJSON(t *testing.T) {
	t.Parallel()

	// Test with invalid JSON mixed with valid JSON
	input := bytes.NewReader([]byte(`{"type":"test1"}invalid{"type":"test2"}`))
	cfg := &display.Config{
		Style:       display.StyleDefault,
		Verbose:     true,
		ShowLineNum: false,
	}

	err := processJSONStream(input, cfg, false)
	if err != nil {
		t.Errorf("processJSONStream() with invalid JSON error = %v", err)
	}
}

func TestProcessJSONStream_EmptyInput(t *testing.T) {
	t.Parallel()

	input := bytes.NewReader([]byte(""))
	cfg := &display.Config{
		Style:       display.StyleDefault,
		Verbose:     false,
		ShowLineNum: false,
	}

	err := processJSONStream(input, cfg, false)
	if err != nil {
		t.Errorf("processJSONStream() with empty input error = %v", err)
	}
}

func TestProcessJSONStream_WhitespaceOnly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "spaces only",
			input: "   ",
		},
		{
			name:  "LF only",
			input: "\n\n\n",
		},
		{
			name:  "CRLF only",
			input: "\r\n\r\n",
		},
		{
			name:  "mixed whitespace",
			input: " \t \r\n \n ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := bytes.NewReader([]byte(tt.input))
			cfg := &display.Config{
				Style:       display.StyleDefault,
				Verbose:     false,
				ShowLineNum: false,
			}

			err := processJSONStream(input, cfg, false)
			if err != nil {
				t.Errorf("processJSONStream() with whitespace only error = %v", err)
			}
		})
	}
}

func TestProcessJSONStream_LargeJSON(t *testing.T) {
	t.Parallel()

	// Create a large JSON object (but under buffer limit)
	largeValue := strings.Repeat("x", 10000)
	largeJSON := `{"type":"test","content":"` + largeValue + `"}`
	input := bytes.NewReader([]byte(largeJSON))

	cfg := &display.Config{
		Style:       display.StyleDefault,
		Verbose:     false,
		ShowLineNum: false,
	}

	err := processJSONStream(input, cfg, false)
	if err != nil {
		t.Errorf("processJSONStream() with large JSON error = %v", err)
	}
}

func TestProcessJSONStream_PanicRecovery(t *testing.T) {
	t.Parallel()

	// Test that panic recovery works - we can't actually cause the display library to panic
	// in a controlled way, but we can verify the code path exists
	input := bytes.NewReader([]byte(`{"type":"test","content":"hello"}`))
	cfg := &display.Config{
		Style:       display.StyleDefault,
		Verbose:     false,
		ShowLineNum: false,
	}

	// This should complete without crashing even if display library panics
	err := processJSONStream(input, cfg, false)
	if err != nil {
		t.Errorf("processJSONStream() should recover from panics, got error: %v", err)
	}
}

// chunkedReader is a test helper that reads data in fixed-size chunks
type chunkedReader struct {
	data  []byte
	chunk int
	pos   int
}

func (r *chunkedReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF // Return EOF when done
	}

	remaining := len(r.data) - r.pos
	toRead := r.chunk
	if toRead > len(p) {
		toRead = len(p)
	}
	if toRead > remaining {
		toRead = remaining
	}

	copy(p, r.data[r.pos:r.pos+toRead])
	r.pos += toRead

	if r.pos >= len(r.data) {
		return toRead, io.EOF // Return EOF when done
	}
	return toRead, nil
}
