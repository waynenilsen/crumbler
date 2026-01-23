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

func TestGenerateCrumbReadme(t *testing.T) {
	t.Parallel()

	result := GenerateCrumbReadme()

	if len(result) == 0 {
		t.Error("GenerateCrumbReadme returned empty string")
	}

	if !strings.Contains(result, "# Task Description") {
		t.Errorf("GenerateCrumbReadme should contain title, got:\n%s", result[:min(100, len(result))])
	}

	if !strings.Contains(result, "## Acceptance Criteria") {
		t.Errorf("GenerateCrumbReadme should contain acceptance criteria section")
	}
}

func TestGenerateTaskName(t *testing.T) {
	t.Parallel()

	// Generate multiple task names and check they're valid
	for i := 0; i < 20; i++ {
		result := GenerateTaskName()

		if len(result) == 0 {
			t.Errorf("GenerateTaskName returned empty string")
		}

		// Check that it contains a space (verb + noun)
		if !strings.Contains(result, " ") {
			t.Errorf("GenerateTaskName should contain a space: %s", result)
		}

		// Check that it starts with a capital letter (verb)
		if result[0] < 'A' || result[0] > 'Z' {
			t.Errorf("GenerateTaskName should start with capital letter: %s", result)
		}
	}
}

func TestGenerateTaskNameUnique(t *testing.T) {
	t.Parallel()

	// While not guaranteed unique, we should get some variety
	seen := make(map[string]int)
	for i := 0; i < 50; i++ {
		result := GenerateTaskName()
		seen[result]++
	}

	// We should have at least 5 different task names out of 50
	if len(seen) < 5 {
		t.Errorf("GenerateTaskName should produce variety, only got %d unique names", len(seen))
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
