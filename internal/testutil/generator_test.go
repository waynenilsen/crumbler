package testutil

import (
	"strings"
	"testing"
)

func TestGenerateLoremIpsum(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		paragraphs int
		wantMin    int // minimum expected length
	}{
		{"one paragraph", 1, 100},
		{"three paragraphs", 3, 300},
		{"zero defaults to one", 0, 100},
		{"negative defaults to one", -1, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := GenerateLoremIpsum(tt.paragraphs)
			if len(result) < tt.wantMin {
				t.Errorf("GenerateLoremIpsum(%d) returned %d chars, want at least %d",
					tt.paragraphs, len(result), tt.wantMin)
			}
			// Check it starts with a capital letter (first word of sentence)
			if len(result) > 0 && (result[0] < 'A' || result[0] > 'Z') {
				t.Errorf("GenerateLoremIpsum(%d) should start with capital letter, got: %s...",
					tt.paragraphs, result[:min(20, len(result))])
			}
			// Check it ends with a period (end of sentence)
			if len(result) > 0 && result[len(result)-1] != '.' {
				t.Errorf("GenerateLoremIpsum(%d) should end with period, got: ...%s",
					tt.paragraphs, result[max(0, len(result)-20):])
			}
		})
	}
}

func TestGenerateLoremIpsumReproducible(t *testing.T) {
	t.Parallel()

	// Same input should produce same output (deterministic)
	result1 := GenerateLoremIpsum(2)
	result2 := GenerateLoremIpsum(2)

	if result1 != result2 {
		t.Errorf("GenerateLoremIpsum should be deterministic")
	}
}

func TestGenerateRandomString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		length int
		want   int
	}{
		{"standard length", 10, 10},
		{"short", 3, 3},
		{"long", 50, 50},
		{"zero defaults to 8", 0, 8},
		{"negative defaults to 8", -1, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := GenerateRandomString(tt.length)
			if len(result) != tt.want {
				t.Errorf("GenerateRandomString(%d) returned length %d, want %d",
					tt.length, len(result), tt.want)
			}
			// Check it only contains alphanumeric characters
			for _, c := range result {
				if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
					t.Errorf("GenerateRandomString(%d) contains invalid character: %c",
						tt.length, c)
				}
			}
		})
	}
}

func TestGenerateRandomStringUnique(t *testing.T) {
	t.Parallel()

	// Generate multiple random strings and verify they're different
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		result := GenerateRandomString(10)
		if seen[result] {
			t.Errorf("GenerateRandomString produced duplicate: %s", result)
		}
		seen[result] = true
	}
}

func TestGenerateRealisticMarkdown(t *testing.T) {
	t.Parallel()

	docTypes := []struct {
		docType       string
		expectedTitle string
	}{
		{"README", "# Project Overview"},
		{"PRD", "# Product Requirements Document"},
		{"ERD", "# Entity Relationship Diagram"},
		{"roadmap", "# Project Roadmap"},
		{"phase", "# Phase Description"},
		{"sprint", "# Sprint Description"},
		{"ticket", "# Ticket Description"},
		{"unknown", "# unknown"},
	}

	for _, tt := range docTypes {
		t.Run(tt.docType, func(t *testing.T) {
			t.Parallel()
			result := GenerateRealisticMarkdown(tt.docType)

			if len(result) == 0 {
				t.Errorf("GenerateRealisticMarkdown(%q) returned empty string", tt.docType)
			}

			if !strings.Contains(result, tt.expectedTitle) {
				t.Errorf("GenerateRealisticMarkdown(%q) should contain %q, got:\n%s",
					tt.docType, tt.expectedTitle, result[:min(100, len(result))])
			}
		})
	}
}

func TestGenerateRealisticMarkdownCaseInsensitive(t *testing.T) {
	t.Parallel()

	// Test that docType is case insensitive
	result1 := GenerateRealisticMarkdown("README")
	result2 := GenerateRealisticMarkdown("readme")
	result3 := GenerateRealisticMarkdown("ReadMe")

	if result1 != result2 || result2 != result3 {
		t.Errorf("GenerateRealisticMarkdown should be case insensitive")
	}
}

func TestGenerateGoalName(t *testing.T) {
	t.Parallel()

	// Generate multiple goal names and check they're valid
	for i := 0; i < 20; i++ {
		result := GenerateGoalName()

		if len(result) == 0 {
			t.Errorf("GenerateGoalName returned empty string")
		}

		// Check that it contains a space (verb + noun)
		if !strings.Contains(result, " ") {
			t.Errorf("GenerateGoalName should contain a space: %s", result)
		}

		// Check that it starts with a capital letter (verb)
		if result[0] < 'A' || result[0] > 'Z' {
			t.Errorf("GenerateGoalName should start with capital letter: %s", result)
		}
	}
}

func TestGenerateGoalNameUnique(t *testing.T) {
	t.Parallel()

	// While not guaranteed unique, we should get some variety
	seen := make(map[string]int)
	for i := 0; i < 50; i++ {
		result := GenerateGoalName()
		seen[result]++
	}

	// We should have at least 5 different goal names out of 50
	if len(seen) < 5 {
		t.Errorf("GenerateGoalName should produce variety, only got %d unique names", len(seen))
	}
}

func TestGenerateTestSeed(t *testing.T) {
	t.Parallel()

	// Same test should produce same seed
	seed1 := GenerateTestSeed(t)
	seed2 := GenerateTestSeed(t)

	if seed1 != seed2 {
		t.Errorf("GenerateTestSeed should be deterministic for same test")
	}
}

func TestGenerateTestSeedDifferentTests(t *testing.T) {
	t.Parallel()

	t.Run("subtest1", func(t *testing.T) {
		t.Parallel()
		seed1 := GenerateTestSeed(t)

		t.Run("nested", func(t *testing.T) {
			t.Parallel()
			seed2 := GenerateTestSeed(t)
			if seed1 == seed2 {
				t.Errorf("Different tests should have different seeds")
			}
		})
	})
}

func TestNewSeededRandom(t *testing.T) {
	t.Parallel()

	rng1 := NewSeededRandom(t)
	rng2 := NewSeededRandom(t)

	// Should produce the same sequence
	vals1 := make([]int64, 5)
	vals2 := make([]int64, 5)

	for i := 0; i < 5; i++ {
		vals1[i] = rng1.Int63()
	}
	for i := 0; i < 5; i++ {
		vals2[i] = rng2.Int63()
	}

	for i := range vals1 {
		if vals1[i] != vals2[i] {
			t.Errorf("NewSeededRandom should be deterministic: got different values at index %d", i)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
