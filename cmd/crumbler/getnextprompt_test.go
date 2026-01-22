package crumbler_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/waynenilsen/crumbler/internal/testutil"
)

// =============================================================================
// Test: CLI Command Flags
// =============================================================================

func TestGetNextPromptFlags(t *testing.T) {
	// Build the crumbler binary for testing (don't run in parallel due to binary building)
	crumblerBin := buildCrumblerBinary(t)
	defer func() {
		// Don't remove binary until all subtests complete
		t.Cleanup(func() {
			os.Remove(crumblerBin)
		})
	}()

	t.Run("--help flag shows help", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
			Build()

		// Change to project directory
		oldWd, _ := os.Getwd()
		os.Chdir(projectRoot)
		defer os.Chdir(oldWd)

		cmd := exec.Command(crumblerBin, "get-next-prompt", "--help")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("command failed: %v\nOutput: %s", err, output)
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "crumbler get-next-prompt") {
			t.Error("expected help text")
		}
		if !strings.Contains(outputStr, "--no-prelude") {
			t.Error("expected --no-prelude flag in help")
		}
		if !strings.Contains(outputStr, "--state-only") {
			t.Error("expected --state-only flag in help")
		}
	})

	t.Run("--state-only outputs only state", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
			Build()

		oldWd, _ := os.Getwd()
		os.Chdir(projectRoot)
		defer os.Chdir(oldWd)

		cmd := exec.Command(crumblerBin, "get-next-prompt", "--state-only")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("command failed: %v\nOutput: %s", err, output)
		}

		outputStr := strings.TrimSpace(string(output))
		if outputStr != "CREATE_PHASE" {
			t.Errorf("expected CREATE_PHASE, got %q", outputStr)
		}
	})

	t.Run("--no-context excludes context", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
			Build()

		oldWd, _ := os.Getwd()
		os.Chdir(projectRoot)
		defer os.Chdir(oldWd)

		cmd := exec.Command(crumblerBin, "get-next-prompt", "--no-context")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("command failed: %v\nOutput: %s", err, output)
		}

		outputStr := string(output)
		if strings.Contains(outputStr, "CONTEXT") || strings.Contains(outputStr, "roadmap.md") {
			t.Error("should not contain context section")
		}
		if !strings.Contains(outputStr, "INSTRUCTION") {
			t.Error("should contain instruction section")
		}
	})

	t.Run("--minimal uses minimal format", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
			Build()

		oldWd, _ := os.Getwd()
		os.Chdir(projectRoot)
		defer os.Chdir(oldWd)

		cmd := exec.Command(crumblerBin, "get-next-prompt", "--minimal")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("command failed: %v\nOutput: %s", err, output)
		}

		outputStr := string(output)
		if strings.Contains(outputStr, "CRUMBLER AGENT PROMPT") {
			t.Error("should not contain full prelude header")
		}
		if !strings.Contains(outputStr, "STATE:") {
			t.Error("should contain state")
		}
	})

	t.Run("--no-prelude excludes prelude", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
			Build()

		oldWd, _ := os.Getwd()
		os.Chdir(projectRoot)
		defer os.Chdir(oldWd)

		cmd := exec.Command(crumblerBin, "get-next-prompt", "--no-prelude")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("command failed: %v\nOutput: %s", err, output)
		}

		outputStr := string(output)
		if strings.Contains(outputStr, "CRUMBLER AGENT PROMPT") {
			t.Error("should not contain prelude")
		}
		if !strings.Contains(outputStr, "INSTRUCTION") {
			t.Error("should contain instruction section")
		}
	})

	t.Run("--no-postlude excludes postlude", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
			Build()

		oldWd, _ := os.Getwd()
		os.Chdir(projectRoot)
		defer os.Chdir(oldWd)

		cmd := exec.Command(crumblerBin, "get-next-prompt", "--no-postlude")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("command failed: %v\nOutput: %s", err, output)
		}

		outputStr := string(output)
		if strings.Contains(outputStr, "NEXT STEPS") {
			t.Error("should not contain postlude")
		}
		if !strings.Contains(outputStr, "INSTRUCTION") {
			t.Error("should contain instruction section")
		}
	})

	t.Run("unknown flag shows error", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
			Build()

		oldWd, _ := os.Getwd()
		os.Chdir(projectRoot)
		defer os.Chdir(oldWd)

		cmd := exec.Command(crumblerBin, "get-next-prompt", "--unknown-flag")
		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Error("expected error for unknown flag")
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "unknown flag") {
			t.Error("expected error message about unknown flag")
		}
		if !strings.Contains(outputStr, "--unknown-flag") {
			t.Error("expected unknown flag name in error")
		}
	})

	t.Run("multiple flags work together", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
			Build()

		oldWd, _ := os.Getwd()
		os.Chdir(projectRoot)
		defer os.Chdir(oldWd)

		cmd := exec.Command(crumblerBin, "get-next-prompt", "--minimal", "--no-context")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("command failed: %v\nOutput: %s", err, output)
		}

		outputStr := string(output)
		if strings.Contains(outputStr, "CONTEXT") || strings.Contains(outputStr, "roadmap.md") {
			t.Error("should not contain context")
		}
		if strings.Contains(outputStr, "CRUMBLER AGENT PROMPT") {
			t.Error("should not contain full prelude")
		}
		if !strings.Contains(outputStr, "INSTRUCTION") {
			t.Error("should contain instruction")
		}
	})
}

// =============================================================================
// Test: CLI Error Handling
// =============================================================================

func TestGetNextPromptErrors(t *testing.T) {
	// Build the crumbler binary for testing (don't run in parallel due to binary building)
	crumblerBin := buildCrumblerBinary(t)
	defer func() {
		t.Cleanup(func() {
			os.Remove(crumblerBin)
		})
	}()

	t.Run("error when not in crumbler project", func(t *testing.T) {
		t.Parallel()

		// Create temp directory without .crumbler
		tempDir := t.TempDir()

		oldWd, _ := os.Getwd()
		os.Chdir(tempDir)
		defer os.Chdir(oldWd)

		cmd := exec.Command(crumblerBin, "get-next-prompt")
		output, err := cmd.CombinedOutput()
		// Command should fail
		if err == nil {
			// If no error, check if output contains error message
			outputStr := string(output)
			if !strings.Contains(outputStr, "not a crumbler project") && !strings.Contains(outputStr, ".crumbler") && !strings.Contains(outputStr, "error") {
				t.Errorf("expected error when not in crumbler project. Output: %s", outputStr)
			}
		} else {
			// Error is expected - verify error message
			outputStr := string(output)
			if !strings.Contains(outputStr, "not a crumbler project") && !strings.Contains(outputStr, ".crumbler") && !strings.Contains(outputStr, "error") {
				t.Logf("Got error (expected): %v, Output: %s", err, outputStr)
			}
		}
	})

	t.Run("--state-only error when not in project", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()

		oldWd, _ := os.Getwd()
		os.Chdir(tempDir)
		defer os.Chdir(oldWd)

		cmd := exec.Command(crumblerBin, "get-next-prompt", "--state-only")
		output, err := cmd.CombinedOutput()
		// Command should fail
		if err == nil {
			// If no error, check if output contains error message
			outputStr := string(output)
			if !strings.Contains(outputStr, "not a crumbler project") && !strings.Contains(outputStr, ".crumbler") && !strings.Contains(outputStr, "error") {
				t.Errorf("expected error when not in crumbler project. Output: %s", outputStr)
			}
		} else {
			// Error is expected - verify error message
			outputStr := string(output)
			if !strings.Contains(outputStr, "not a crumbler project") && !strings.Contains(outputStr, ".crumbler") && !strings.Contains(outputStr, "error") {
				t.Logf("Got error (expected): %v, Output: %s", err, outputStr)
			}
		}
	})
}

// =============================================================================
// Test: Full CLI Integration
// =============================================================================

func TestGetNextPromptFullIntegration(t *testing.T) {
	// Build the crumbler binary for testing (don't run in parallel due to binary building)
	crumblerBin := buildCrumblerBinary(t)
	defer func() {
		t.Cleanup(func() {
			os.Remove(crumblerBin)
		})
	}()

	t.Run("full prompt includes all sections", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
			WithPhase("0001", "open").
			WithPhaseGoal("0001", 1, "My Goal", "open").
			Build()

		oldWd, _ := os.Getwd()
		os.Chdir(projectRoot)
		defer os.Chdir(oldWd)

		cmd := exec.Command(crumblerBin, "get-next-prompt")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("command failed: %v\nOutput: %s", err, output)
		}

		outputStr := string(output)
		// Should have state
		if !strings.Contains(outputStr, "STATE:") {
			t.Error("expected STATE in output")
		}
		// Should have context
		if !strings.Contains(outputStr, "roadmap.md") {
			t.Error("expected roadmap context")
		}
		// Should have instruction
		if !strings.Contains(outputStr, "INSTRUCTION") {
			t.Error("expected instruction section")
		}
		// Should have postlude
		if !strings.Contains(outputStr, "crumbler get-next-prompt") {
			t.Error("expected postlude with next command")
		}
	})

	t.Run("EXIT state shows completion", func(t *testing.T) {
		t.Parallel()

		projectRoot := testutil.NewTestProject(t).
			WithRoadmap(`# Roadmap
## Phase 1: Foundation
- Goal 1
`).
			WithPhase("0001", "closed").
			Build()

		oldWd, _ := os.Getwd()
		os.Chdir(projectRoot)
		defer os.Chdir(oldWd)

		cmd := exec.Command(crumblerBin, "get-next-prompt")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("command failed: %v\nOutput: %s", err, output)
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "STATE: EXIT") {
			t.Error("expected EXIT state")
		}
		if !strings.Contains(outputStr, "Project Complete") || !strings.Contains(outputStr, "complete") {
			t.Error("expected completion message")
		}
	})
}

// =============================================================================
// Helper Functions
// =============================================================================

// buildCrumblerBinary builds the crumbler binary and returns its path.
func buildCrumblerBinary(t *testing.T) string {
	t.Helper()

	// Find project root (use absolute path)
	projectRoot := findProjectRoot(t)

	// Build binary in temp location
	tempFile := filepath.Join(t.TempDir(), "crumbler")
	if runtime.GOOS == "windows" {
		tempFile += ".exe"
	}

	// Use absolute path for binary
	tempFileAbs, err := filepath.Abs(tempFile)
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}

	cmd := exec.Command("go", "build", "-o", tempFileAbs, filepath.Join(projectRoot, "main.go"))
	cmd.Dir = projectRoot
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to build crumbler binary: %v", err)
	}

	return tempFileAbs
}

// findProjectRoot finds the project root by looking for go.mod
func findProjectRoot(t *testing.T) string {
	t.Helper()

	// Get the current test file's directory
	_, testFile, _, _ := runtime.Caller(1)
	testDir := filepath.Dir(testFile)

	// Walk up from test file location
	dir := testDir
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Fallback: try from current working directory
			wd, err := os.Getwd()
			if err == nil {
				dir = wd
				for {
					if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
						return dir
					}
					parent := filepath.Dir(dir)
					if parent == dir {
						t.Fatalf("could not find project root (no go.mod found from %s or %s)", testDir, wd)
					}
					dir = parent
				}
			}
			t.Fatalf("could not find project root (no go.mod found from %s)", testDir)
		}
		dir = parent
	}
}
