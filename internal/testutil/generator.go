package testutil

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

// loremIpsumWords contains standard lorem ipsum words for generating text.
var loremIpsumWords = []string{
	"lorem", "ipsum", "dolor", "sit", "amet", "consectetur", "adipiscing", "elit",
	"sed", "do", "eiusmod", "tempor", "incididunt", "ut", "labore", "et", "dolore",
	"magna", "aliqua", "enim", "ad", "minim", "veniam", "quis", "nostrud",
	"exercitation", "ullamco", "laboris", "nisi", "aliquip", "ex", "ea", "commodo",
	"consequat", "duis", "aute", "irure", "in", "reprehenderit", "voluptate",
	"velit", "esse", "cillum", "fugiat", "nulla", "pariatur", "excepteur", "sint",
	"occaecat", "cupidatat", "non", "proident", "sunt", "culpa", "qui", "officia",
	"deserunt", "mollit", "anim", "id", "est", "laborum",
}

// taskVerbs contains verbs commonly used in task descriptions.
var taskVerbs = []string{
	"Implement", "Create", "Design", "Build", "Develop", "Configure", "Set up",
	"Integrate", "Test", "Validate", "Deploy", "Document", "Refactor", "Optimize",
	"Review", "Update", "Fix", "Add", "Remove", "Enable", "Disable", "Migrate",
}

// taskNouns contains nouns commonly used in task descriptions.
var taskNouns = []string{
	"authentication", "authorization", "API endpoint", "database schema", "user interface",
	"caching layer", "logging system", "error handling", "unit tests", "integration tests",
	"documentation", "configuration", "deployment pipeline", "monitoring", "alerting",
	"performance metrics", "security audit", "data validation", "input sanitization",
	"rate limiting", "load balancing", "backup system", "recovery procedures",
}

// alphanumericChars contains characters used for random string generation.
const alphanumericChars = "abcdefghijklmnopqrstuvwxyz0123456789"

// GenerateLoremIpsum generates lorem ipsum text with the specified number of paragraphs.
// Each paragraph contains 5-10 sentences.
func GenerateLoremIpsum(paragraphs int) string {
	if paragraphs <= 0 {
		paragraphs = 1
	}

	rng := rand.New(rand.NewSource(42)) // Use fixed seed for reproducibility
	var result strings.Builder

	for p := 0; p < paragraphs; p++ {
		if p > 0 {
			result.WriteString("\n\n")
		}

		// Generate 5-10 sentences per paragraph
		sentences := 5 + rng.Intn(6)
		for s := 0; s < sentences; s++ {
			if s > 0 {
				result.WriteString(" ")
			}

			// Generate 8-15 words per sentence
			wordCount := 8 + rng.Intn(8)
			for w := 0; w < wordCount; w++ {
				if w > 0 {
					result.WriteString(" ")
				}
				word := loremIpsumWords[rng.Intn(len(loremIpsumWords))]
				if w == 0 {
					// Capitalize first word
					word = strings.ToUpper(word[:1]) + word[1:]
				}
				result.WriteString(word)
			}
			result.WriteString(".")
		}
	}

	return result.String()
}

// GenerateRandomString generates a random alphanumeric string of the specified length.
// Uses a time-based seed to ensure uniqueness across test runs.
func GenerateRandomString(length int) string {
	if length <= 0 {
		length = 8
	}

	rng := rand.New(rand.NewSource(rand.Int63()))
	result := make([]byte, length)
	for i := range result {
		result[i] = alphanumericChars[rng.Intn(len(alphanumericChars))]
	}
	return string(result)
}

// GenerateCrumbReadme generates realistic README content for a crumb.
func GenerateCrumbReadme() string {
	return `# Task Description

## Summary

` + GenerateLoremIpsum(1) + `

## Acceptance Criteria

- Implementation complete
- Tests passing
- Code reviewed

## Notes

` + GenerateLoremIpsum(1) + `
`
}

// GenerateTaskName generates a realistic task name.
func GenerateTaskName() string {
	rng := rand.New(rand.NewSource(rand.Int63()))
	verb := taskVerbs[rng.Intn(len(taskVerbs))]
	noun := taskNouns[rng.Intn(len(taskNouns))]
	return fmt.Sprintf("%s %s", verb, noun)
}

// GenerateTestSeed generates a deterministic seed based on the test name.
// This ensures reproducible random values within the same test.
func GenerateTestSeed(t *testing.T) int64 {
	t.Helper()
	hash := sha256.Sum256([]byte(t.Name()))
	return int64(binary.BigEndian.Uint64(hash[:8]))
}

// NewSeededRandom creates a new random generator with a test-based seed.
// This provides deterministic random values for reproducible tests.
func NewSeededRandom(t *testing.T) *rand.Rand {
	t.Helper()
	return rand.New(rand.NewSource(GenerateTestSeed(t)))
}
