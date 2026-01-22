# Product Requirements Document (PRD) Guide

## Philosophy and Core Principles

The best PRDs focus on outcomes, not outputs. They articulate the problem to be solved and why it matters, rather than prescribing specific solutions. Great PRDs serve as a shared understanding between product, engineering, design, and stakeholders—enabling teams to make informed decisions while maintaining flexibility in implementation.

### Key Principles

**Problem-First Thinking**: Start with understanding the user problem or business need. The solution comes later through collaboration between product, engineering, and design.

**Outcome-Oriented**: Define success in terms of measurable outcomes—user behavior changes, business metrics, or system improvements—not just feature completion.

**Clear Constraints**: Explicitly state what cannot be done, what must be preserved, and what trade-offs are acceptable. Constraints enable creativity within boundaries.

**Living Document**: PRDs evolve as understanding deepens. They capture decisions and rationale, serving as historical context for future work.

**Shared Understanding**: A PRD is a communication tool. It should be readable by engineers, designers, product managers, and stakeholders—avoiding jargon while maintaining technical accuracy.

## Essential PRD Structure

### 1. Problem Statement

**What to Include:**
- Clear articulation of the problem being solved
- Who experiences this problem (users, developers, business)
- Why this problem matters now (context, urgency, impact)
- Current state vs. desired state

**Best Practices:**
- Use specific examples or user stories to illustrate the problem
- Quantify the impact where possible (time saved, errors reduced, revenue affected)
- Avoid solution language in the problem statement
- Reference user research, data, or stakeholder input that validates the problem

**Example Structure:**
- Problem: What is broken or missing?
- Users Affected: Who experiences this?
- Impact: What happens if we don't solve this?
- Context: Why is this important now?

### 2. Goals and Success Metrics

**What to Include:**
- Primary goal: The main outcome this work should achieve
- Success metrics: How will we know we succeeded?
- Leading indicators: Early signals of progress
- Lagging indicators: Long-term impact measures

**Best Practices:**
- Make metrics specific, measurable, and time-bound
- Include both quantitative (numbers) and qualitative (user feedback) measures
- Define baseline metrics before starting
- Set ambitious but achievable targets
- Consider both positive metrics (what should increase) and negative metrics (what should decrease)

**Types of Metrics:**
- User engagement: Usage, retention, time spent
- Business impact: Revenue, cost reduction, efficiency gains
- Quality metrics: Error rates, performance, reliability
- Developer experience: Time to implement, code quality, maintainability

### 3. User Stories and Use Cases

**What to Include:**
- Primary user flows: The main paths users will take
- Edge cases: Important but less common scenarios
- User personas: Who are we building for?
- Acceptance criteria: What must be true for this to be considered done?

**Best Practices:**
- Write from the user's perspective (not the system's)
- Focus on value delivered, not implementation details
- Include both happy paths and error scenarios
- Prioritize user stories by impact and effort
- Use the format: "As a [user type], I want [goal] so that [benefit]"

**Example Structure:**
- Primary use case: The main scenario
- Secondary use cases: Important but less common scenarios
- Edge cases: Boundary conditions and error handling
- User personas: Characteristics of target users

### 4. Requirements

**What to Include:**
- Functional requirements: What the system must do
- Non-functional requirements: Performance, reliability, security, scalability
- Constraints: Technical limitations, business rules, compliance requirements
- Dependencies: What must exist or be completed first

**Best Practices:**
- Use "must" for requirements, "should" for nice-to-haves, "could" for future considerations
- Be specific but not prescriptive about implementation
- Separate requirements from solutions
- Prioritize requirements (P0 = critical, P1 = important, P2 = nice-to-have)
- Include both positive requirements (what must happen) and negative requirements (what must not happen)

**Categories of Requirements:**
- Functional: Features and behaviors
- Performance: Speed, throughput, latency
- Reliability: Uptime, error handling, recovery
- Security: Authentication, authorization, data protection
- Usability: User experience, accessibility, learnability
- Scalability: Growth capacity, resource efficiency
- Compliance: Legal, regulatory, policy requirements

### 5. Out of Scope

**What to Include:**
- Explicitly state what is NOT included in this work
- Future considerations that are explicitly deferred
- Related problems that are intentionally not addressed

**Best Practices:**
- Be explicit about boundaries to prevent scope creep
- Explain why things are out of scope (not just what)
- Reference future work or related initiatives
- Acknowledge valid concerns that are being deferred

**Why This Matters:**
- Prevents misunderstandings about what will be delivered
- Helps stakeholders understand trade-offs
- Enables focused execution without constant renegotiation
- Documents decisions for future reference

### 6. Technical Considerations

**What to Include:**
- Architecture constraints: System design boundaries
- Integration points: External systems, APIs, services
- Data considerations: Storage, privacy, retention
- Performance requirements: Response times, throughput, capacity
- Security requirements: Authentication, authorization, encryption

**Best Practices:**
- Provide context without prescribing implementation
- Reference existing systems, patterns, or standards
- Identify risks and mitigation strategies
- Consider scalability and maintainability
- Document assumptions about technical environment

**Key Areas:**
- System architecture: High-level design constraints
- Data model: What data is needed and how it's structured
- APIs and interfaces: How components interact
- Infrastructure: Deployment, monitoring, operations
- Security: Threat model, security requirements

### 7. Design Considerations

**What to Include:**
- User experience principles: How users should feel when using this
- Design constraints: Brand guidelines, accessibility requirements
- Interaction patterns: How users will interact with the system
- Content requirements: Copy, messaging, localization needs

**Best Practices:**
- Focus on user experience goals, not specific UI designs
- Reference design system or style guide if applicable
- Include accessibility requirements explicitly
- Consider different user contexts (mobile, desktop, etc.)
- Leave room for design exploration

### 8. Risks and Mitigations

**What to Include:**
- Technical risks: What could go wrong technically?
- Product risks: What could prevent success?
- Timeline risks: What could cause delays?
- Mitigation strategies: How will we address these risks?

**Best Practices:**
- Be honest about uncertainties
- Prioritize risks by likelihood and impact
- Propose concrete mitigation strategies
- Identify early warning signs
- Plan for failure scenarios

**Risk Categories:**
- Technical: Implementation challenges, performance issues
- Product: User adoption, market fit, competition
- Timeline: Dependencies, complexity, resource availability
- Business: Stakeholder alignment, changing priorities

### 9. Timeline and Milestones

**What to Include:**
- Key milestones: Major checkpoints in the work
- Dependencies: What must happen before this can start or complete
- Critical path: The sequence of work that determines timeline
- Buffer time: Contingency for unknowns

**Best Practices:**
- Be realistic about timelines
- Identify critical dependencies early
- Build in time for iteration and learning
- Define clear milestone criteria
- Communicate uncertainty where it exists

### 10. Open Questions

NEVER UNDER ANY CIRCUMSTANCES HAVE OPEN QUESTIONS, YOU ARE ON YOUR OWN NO HUMAN IS COMING TO ANSWER ANY QUESTION YOU MAY HAVE, DO YOUR BEST.

## Writing Style and Tone

### Clarity Over Cleverness

Use simple, direct language. Avoid jargon unless it's necessary and well-defined. Write for a diverse audience—engineers, designers, product managers, and stakeholders should all be able to understand the document.

### Specific Over Vague

Instead of "make it fast," say "respond to user requests within 200ms for 95% of requests." Instead of "improve user experience," say "reduce the number of steps to complete checkout from 5 to 2."

### Facts Over Opinions

Base statements on data, research, or validated assumptions. When opinions are included, label them as such and explain the reasoning.

### Collaborative Over Prescriptive

Frame requirements as problems to solve, not solutions to implement. Leave room for engineering and design expertise to find the best approach.

## Common Anti-Patterns to Avoid

### Solution-First Thinking

**Bad**: "Add a dropdown menu to filter users by role"
**Good**: "Users need to quickly find specific users by role. Current search is too slow when there are many users."

### Vague Requirements

**Bad**: "Make it user-friendly"
**Good**: "New users should be able to complete their first task without reading documentation, measured by 80% task completion rate."

### Missing Context

**Bad**: "Implement authentication"
**Good**: "Users need secure access to their data. Current system has no authentication, creating security risks. We need to implement authentication that supports SSO for enterprise customers and email/password for individual users."

### Over-Specification

**Bad**: "Use React hooks, implement useState for form state, use useEffect for API calls..."
**Good**: "Form should validate input in real-time and submit via API. Handle network errors gracefully."

### Ignoring Constraints

**Bad**: "Build the perfect solution"
**Good**: "Must work within existing authentication system. Cannot change database schema. Must support 10,000 concurrent users."

## PRD Review Checklist

Before considering a PRD complete, verify:

- [ ] Problem statement is clear and validated
- [ ] Success metrics are specific and measurable
- [ ] User stories cover primary and edge cases
- [ ] Requirements are prioritized and testable
- [ ] Out of scope is explicitly stated
- [ ] Technical considerations are documented
- [ ] Risks are identified with mitigation strategies
- [ ] Open questions are listed with owners
- [ ] Document is readable by non-technical stakeholders
- [ ] Dependencies are identified and tracked

## For Technical Debt Paydown Sprints

When writing a PRD for a tech debt paydown sprint, adapt the structure:

**Problem Statement**: Focus on the technical debt problem—what's broken, what's slow, what's risky?

**Goals and Success Metrics**: Measure improvements in code quality, performance, maintainability, or developer experience.

**User Stories**: Frame from developer perspective—"As a developer, I want..."

**Requirements**: Focus on refactoring goals, code quality improvements, dependency updates, documentation improvements.

**Technical Considerations**: Emphasize architecture improvements, performance optimizations, security fixes.

**Success Metrics**: Code quality scores, test coverage, build times, error rates, developer satisfaction.

Remember: Even tech debt work should have clear outcomes and measurable success criteria. The PRD helps ensure the sprint delivers real value, not just "cleaning up code."

## Conclusion

A great PRD is a tool for alignment, not a specification to be followed blindly. It should enable good decisions, facilitate communication, and serve as a historical record of why decisions were made. The best PRDs are living documents that evolve as understanding deepens, always focused on solving real problems and delivering measurable value.
