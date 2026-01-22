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

// goalVerbs contains verbs commonly used in goal descriptions.
var goalVerbs = []string{
	"Implement", "Create", "Design", "Build", "Develop", "Configure", "Set up",
	"Integrate", "Test", "Validate", "Deploy", "Document", "Refactor", "Optimize",
	"Review", "Update", "Fix", "Add", "Remove", "Enable", "Disable", "Migrate",
}

// goalNouns contains nouns commonly used in goal descriptions.
var goalNouns = []string{
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

// GenerateRealisticMarkdown generates realistic markdown content for different document types.
// Supported docTypes: "README", "PRD", "ERD", "roadmap", "phase", "sprint", "ticket"
func GenerateRealisticMarkdown(docType string) string {
	switch strings.ToLower(docType) {
	case "readme":
		return generateReadme()
	case "prd":
		return generatePRD()
	case "erd":
		return generateERD()
	case "roadmap":
		return generateRoadmap()
	case "phase":
		return generatePhaseReadme()
	case "sprint":
		return generateSprintReadme()
	case "ticket":
		return generateTicketReadme()
	default:
		return fmt.Sprintf("# %s\n\n%s\n", docType, GenerateLoremIpsum(2))
	}
}

// GenerateGoalName generates a realistic goal name.
func GenerateGoalName() string {
	rng := rand.New(rand.NewSource(rand.Int63()))
	verb := goalVerbs[rng.Intn(len(goalVerbs))]
	noun := goalNouns[rng.Intn(len(goalNouns))]
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

// generateReadme creates a realistic project README.
func generateReadme() string {
	return `# Project Overview

This project implements a software development lifecycle (SDLC) automation system.

## Description

` + GenerateLoremIpsum(1) + `

## Features

- Feature-based development workflow
- Phase and sprint management
- Automated ticket tracking
- Goal-oriented progress tracking

## Getting Started

1. Initialize the project
2. Define your roadmap
3. Create phases and sprints
4. Track progress through tickets

## Status

This project is currently under active development.
`
}

// generatePRD creates a realistic Product Requirements Document.
func generatePRD() string {
	return `# Product Requirements Document

## Overview

` + GenerateLoremIpsum(1) + `

## Problem Statement

` + GenerateLoremIpsum(1) + `

## Goals

1. Improve user experience
2. Increase system reliability
3. Enable new capabilities
4. Support scalability requirements

## Requirements

### Functional Requirements

- FR-001: System shall provide user authentication
- FR-002: System shall support data persistence
- FR-003: System shall enable configuration management

### Non-Functional Requirements

- NFR-001: Response time under 200ms for 95th percentile
- NFR-002: 99.9% uptime availability
- NFR-003: Support for 10,000 concurrent users

## User Stories

### As a developer
- I want to track my progress
- I want to see clear goals
- I want automated validation

### As a project manager
- I want visibility into project status
- I want to manage phases and sprints
- I want to generate reports

## Success Metrics

- Deployment frequency increase by 50%
- Lead time reduction by 30%
- Defect rate reduction by 40%
`
}

// generateERD creates a realistic Entity Relationship Diagram document.
func generateERD() string {
	return `# Entity Relationship Diagram

## Overview

This document describes the data model for the system.

## Entities

### Phase
- id: string (primary key)
- name: string
- status: enum (open, closed)
- created_at: timestamp
- updated_at: timestamp

### Sprint
- id: string (primary key)
- phase_id: string (foreign key)
- name: string
- status: enum (open, closed)
- created_at: timestamp
- updated_at: timestamp

### Ticket
- id: string (primary key)
- sprint_id: string (foreign key)
- title: string
- description: text
- status: enum (open, done)
- created_at: timestamp
- updated_at: timestamp

### Goal
- id: string (primary key)
- parent_type: enum (phase, sprint, ticket)
- parent_id: string (foreign key)
- name: string
- status: enum (open, closed)
- order: integer

## Relationships

- Phase 1:N Sprint (one phase has many sprints)
- Sprint 1:N Ticket (one sprint has many tickets)
- Phase 1:N Goal (one phase has many goals)
- Sprint 1:N Goal (one sprint has many goals)
- Ticket 1:N Goal (one ticket has many goals)

## Indexes

- idx_sprint_phase_id ON Sprint(phase_id)
- idx_ticket_sprint_id ON Ticket(sprint_id)
- idx_goal_parent ON Goal(parent_type, parent_id)
`
}

// generateRoadmap creates a realistic roadmap document.
func generateRoadmap() string {
	return `# Project Roadmap

## Phase 1: Foundation

- Set up project infrastructure
- Establish development workflows
- Create initial architecture

## Phase 2: Core Features

- Implement user authentication
- Build data persistence layer
- Create API endpoints

## Phase 3: Enhancement

- Add advanced features
- Optimize performance
- Improve user experience

## Phase 4: Production

- Deploy to production
- Set up monitoring
- Document operational procedures

## Timeline

- Q1: Foundation and initial development
- Q2: Core features and testing
- Q3: Enhancement and optimization
- Q4: Production deployment and stabilization
`
}

// generatePhaseReadme creates a realistic phase README.
func generatePhaseReadme() string {
	return `# Phase Description

## Overview

` + GenerateLoremIpsum(1) + `

## Objectives

1. Establish project foundation
2. Define core architecture
3. Set up development environment
4. Create initial documentation

## Success Criteria

- All sprints completed successfully
- All phase goals met
- Documentation updated
- Code reviewed and approved

## Dependencies

- Infrastructure provisioning
- Team onboarding
- Tool configuration
`
}

// generateSprintReadme creates a realistic sprint README.
func generateSprintReadme() string {
	return `# Sprint Description

## Overview

` + GenerateLoremIpsum(1) + `

## Sprint Goals

1. Complete planned tickets
2. Meet acceptance criteria
3. Pass code review
4. Update documentation

## Deliverables

- Implemented features
- Unit tests
- Integration tests
- Updated documentation

## Notes

- Daily standups at 9:00 AM
- Sprint review on Friday
- Retrospective after review
`
}

// generateTicketReadme creates a realistic ticket README.
func generateTicketReadme() string {
	return `# Ticket Description

## Summary

` + GenerateLoremIpsum(1) + `

## Acceptance Criteria

- [ ] Implementation complete
- [ ] Unit tests passing
- [ ] Code reviewed
- [ ] Documentation updated

## Technical Notes

` + GenerateLoremIpsum(1) + `

## Testing Instructions

1. Run unit tests
2. Run integration tests
3. Verify functionality manually
4. Check for regressions
`
}
