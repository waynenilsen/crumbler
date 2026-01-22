# Ticket Decomposition Guide: Breaking ERDs into Implementation Tickets

## Critical: ERD Must Come First

**IMPORTANT**: The ERD (Engineering Requirements Document) MUST be completed before breaking down into tickets. The ERD provides the technical architecture, data models, API designs, and implementation approach that tickets will execute. You cannot effectively decompose work into tickets without understanding the technical solution design.

**Workflow**: PRD → ERD → Ticket Decomposition → Implementation

## Philosophy and Core Principles

Ticket decomposition transforms the technical design (ERD) into actionable, implementable work units. Great ticket decomposition enables parallel work, clear progress tracking, and manageable implementation complexity. Each ticket should represent a single, focused unit of work that can be completed independently (or with clear dependencies).

### Key Principles

**Implementation-Focused**: Tickets translate ERD technical requirements into concrete code changes, tests, and deliverables. They move from design space (ERD) to implementation space (code).

**Independently Actionable**: Each ticket should be completable by a developer without constant coordination, once dependencies are met.

**Testable and Verifiable**: Every ticket must have clear acceptance criteria that can be verified through testing or review.

**Dependency-Aware**: Tickets explicitly document what must be completed before this work can start, enabling parallel work where possible.

**Right-Sized**: Tickets should be small enough to complete in a sprint but large enough to represent meaningful value.

## Essential Ticket Structure

### 1. Title and Description

**What to Include:**
- Clear, concise title that describes what will be implemented
- Detailed description explaining the context and what needs to be done
- Reference to relevant ERD sections (architecture, data models, APIs, etc.)
- Link to sprint goals this ticket supports

**Best Practices:**
- Use action verbs: "Implement", "Create", "Add", "Refactor", "Update"
- Be specific: "Create User model with email validation" not "User stuff"
- Include ERD references: "Implements User entity from ERD section 4.1"
- Explain the "why" briefly: What problem does this solve?

**Example Structure:**
- Title: "Create User model and database migration"
- Description: "Implement the User entity as specified in ERD section 4.1 (Data Model). This includes the database schema, model class, and basic validation rules. This is foundational work for the authentication system."

### 2. Acceptance Criteria

**What to Include:**
- Specific, testable conditions that must be met
- Both positive cases (happy path) and edge cases
- Error handling requirements
- Performance requirements if applicable
- Integration points with other components

**Best Practices:**
- Use "Given-When-Then" format for clarity
- Make criteria binary (pass/fail, not subjective)
- Include edge cases: empty inputs, null values, boundary conditions
- Reference ERD requirements explicitly
- Include both functional and non-functional requirements

**Example Structure:**
- "Given a valid email and password, when a user registers, then a User record is created in the database"
- "Given an invalid email format, when registration is attempted, then validation error is returned"
- "Given a duplicate email, when registration is attempted, then unique constraint error is returned"
- "Given valid input, when User is created, then password is hashed using bcrypt (as per ERD security section)"

### 3. Technical Details

**What to Include:**
- Specific files/components to be created or modified
- API endpoints, data models, or services involved
- Dependencies on other tickets or external systems
- Technical constraints or considerations
- References to ERD sections (architecture, data model, API design, etc.)

**Best Practices:**
- List specific files/modules that will be changed
- Reference ERD architecture diagrams or component descriptions
- Document data model changes (tables, fields, relationships)
- Specify API contracts if applicable (endpoints, request/response formats)
- Note any breaking changes or migration requirements

**Example Structure:**
- Files to create: `models/user.go`, `migrations/001_create_users.sql`
- Files to modify: `config/database.go` (add User model)
- ERD references: Section 4.1 (User entity), Section 5.2 (Authentication API)
- Dependencies: Ticket #2 (Database setup) must be completed first

### 4. Dependencies

**What to Include:**
- Blocking dependencies: What must be completed before this ticket can start?
- Enabling dependencies: What would make this easier but isn't strictly required?
- Related tickets: What other tickets are related or should be considered together?
- External dependencies: Third-party services, APIs, or infrastructure needed

**Best Practices:**
- Be explicit about blocking vs. enabling dependencies
- Order tickets by dependency chain
- Identify opportunities for parallel work (no dependencies)
- Document why dependencies exist (technical reason)
- Consider creating a dependency graph or ordering

**Example Structure:**
- Blocks: None (can start immediately)
- Blocked by: Ticket #1 (Database schema setup)
- Related: Ticket #3 (User API endpoints), Ticket #4 (User authentication)
- External: None

### 5. Testing Requirements

**What to Include:**
- Unit tests: What should be unit tested?
- Integration tests: What integration scenarios need testing?
- Edge cases: What edge cases must be covered?
- Performance tests: Are there performance requirements to validate?

**Best Practices:**
- Specify test coverage expectations (e.g., "80% coverage on new code")
- List specific test scenarios to implement
- Reference ERD testing strategy section
- Include both positive and negative test cases
- Consider integration with existing test infrastructure

**Example Structure:**
- Unit tests: User model validation, password hashing, email format validation
- Integration tests: User creation flow, database constraints
- Edge cases: Empty strings, null values, SQL injection attempts, very long strings
- Performance: User creation should complete in <100ms for 95% of requests

## Decomposition Strategies

### Strategy 1: By Data Model (ERD Entities)

Break down tickets based on ERD entities and their relationships.

**Approach:**
- One ticket per major entity (User, Order, Product, etc.)
- Separate tickets for entity creation, relationships, and queries
- Follow entity dependencies (create User before UserProfile)

**When to Use:**
- ERD has clear, distinct entities
- Entities have minimal coupling
- Data model is the primary focus

**Example:**
- ERD defines: User, Order, Product, OrderItem entities
- Tickets: 
  - Ticket 1: Create User entity and migration
  - Ticket 2: Create Product entity and migration
  - Ticket 3: Create Order entity and migration
  - Ticket 4: Create OrderItem entity and relationships
  - Ticket 5: Create queries for User orders

### Strategy 2: By API Endpoints

Break down tickets based on API endpoints defined in the ERD.

**Approach:**
- One ticket per API endpoint or resource
- Group related endpoints (CRUD operations) when appropriate
- Separate tickets for authentication, validation, error handling

**When to Use:**
- ERD has well-defined API specifications
- API-first architecture
- Clear endpoint boundaries

**Example:**
- ERD defines: POST /users, GET /users/:id, PUT /users/:id, DELETE /users/:id
- Tickets:
  - Ticket 1: POST /users endpoint (create user)
  - Ticket 2: GET /users/:id endpoint (get user)
  - Ticket 3: PUT /users/:id endpoint (update user)
  - Ticket 4: DELETE /users/:id endpoint (delete user)
  - Ticket 5: Input validation middleware
  - Ticket 6: Error handling and response formatting

### Strategy 3: By Feature/Workflow

Break down tickets based on user-facing features or workflows.

**Approach:**
- One ticket per feature or user flow
- Include all components needed for the feature (API, data, UI if applicable)
- Follow user journey dependencies

**When to Use:**
- Feature-driven development
- Clear user workflows in PRD
- End-to-end feature delivery is priority

**Example:**
- Feature: User Registration
- Tickets:
  - Ticket 1: User registration API endpoint
  - Ticket 2: Email validation and verification
  - Ticket 3: Password strength validation
  - Ticket 4: Registration confirmation email
  - Ticket 5: Error handling and user feedback

### Strategy 4: By Technical Layer

Break down tickets based on architectural layers (data, service, API, etc.).

**Approach:**
- Separate tickets for each architectural layer
- Data layer → Service layer → API layer → Integration
- Follow layer dependencies (data before service, service before API)

**When to Use:**
- Layered architecture in ERD
- Clear separation of concerns
- Team organized by technical expertise

**Example:**
- Feature: User Management
- Tickets:
  - Ticket 1: User data model and database schema
  - Ticket 2: User repository/data access layer
  - Ticket 3: User service/business logic layer
  - Ticket 4: User API/controller layer
  - Ticket 5: User API integration tests

### Strategy 5: By Component/Module

Break down tickets based on system components or modules defined in ERD.

**Approach:**
- One ticket per major component or module
- Separate tickets for component interfaces and implementations
- Follow component dependencies

**When to Use:**
- Microservices or modular architecture
- Clear component boundaries in ERD
- Component-based development

**Example:**
- Components: Authentication Service, User Service, Notification Service
- Tickets:
  - Ticket 1: Authentication Service - core implementation
  - Ticket 2: User Service - core implementation
  - Ticket 3: Service-to-service communication
  - Ticket 4: Notification Service integration
  - Ticket 5: Service discovery and configuration

## What Makes a Good Ticket: INVEST Criteria

The INVEST criteria (Independent, Negotiable, Valuable, Estimable, Small, Testable) provides a framework for evaluating ticket quality.

### Independent

**Definition:** The ticket can be worked on independently of other tickets (once dependencies are met).

**How to Achieve:**
- Minimize coupling between tickets
- Make dependencies explicit and clear
- Design tickets so they can be completed in any order (after dependencies)
- Avoid tickets that require constant coordination

**Good Example:**
- "Create User model" - Can be done independently once database is set up

**Bad Example:**
- "Implement user registration (needs discussion with team about validation rules)" - Requires ongoing coordination

### Negotiable

**Definition:** The ticket's details can be discussed and refined, but the core requirement is clear.

**How to Achieve:**
- Focus on "what" and "why" rather than prescribing "how"
- Leave implementation details to the developer
- Allow for technical decisions within constraints
- Reference ERD for constraints, not prescriptive implementation

**Good Example:**
- "Implement user authentication using OAuth2 (as per ERD section 6.1)" - Clear requirement, implementation approach flexible

**Bad Example:**
- "Use library X, function Y, with parameters Z" - Too prescriptive

### Valuable

**Definition:** The ticket delivers value, either to users, the system, or the development process.

**How to Achieve:**
- Each ticket should contribute to sprint goals
- Even infrastructure tickets should enable future value
- Avoid "nice to have" tickets that don't support sprint goals
- Consider incremental value delivery

**Good Example:**
- "Create User model" - Enables user registration feature (sprint goal)

**Bad Example:**
- "Refactor all error messages to use emojis" - Doesn't support sprint goals

### Estimable

**Definition:** The ticket can be estimated with reasonable accuracy.

**How to Achieve:**
- Provide enough detail for estimation
- Break down large tickets into smaller, estimable pieces
- Include technical complexity indicators
- Reference similar past work if applicable

**Good Example:**
- "Create User model with 5 fields, basic validation, and database migration" - Clear scope

**Bad Example:**
- "Implement user management" - Too vague to estimate

### Small

**Definition:** The ticket is small enough to complete in a sprint but large enough to be meaningful.

**How to Achieve:**
- Aim for tickets that can be completed in 1-3 days
- Break down tickets that would take more than a week
- Combine tickets that are too small (< 2 hours) unless they're truly independent
- Consider team velocity and sprint length

**Good Example:**
- "Create User model and migration" - 1-2 days of work

**Bad Example:**
- "Build entire authentication system" - Too large, should be multiple tickets
- "Add semicolon to line 42" - Too small, combine with related work

### Testable

**Definition:** The ticket has clear acceptance criteria that can be verified.

**How to Achieve:**
- Define specific, binary (pass/fail) acceptance criteria
- Include both positive and negative test cases
- Specify how to verify completion (tests, manual testing, code review)
- Make criteria objective, not subjective

**Good Example:**
- Acceptance criteria: "User model validates email format, rejects invalid emails, stores hashed passwords"

**Bad Example:**
- Acceptance criteria: "User model works well" - Not testable

## Decomposition Process

### Step 1: Review ERD Thoroughly

**Actions:**
- Read the entire ERD, especially:
  - Architecture and system design section
  - Data model and storage section
  - API design and interfaces section
  - Technical requirements section
- Identify major components, entities, and workflows
- Note dependencies between components
- Understand the technical constraints and requirements

**Output:**
- List of major components/entities/workflows
- Dependency graph or notes
- Technical constraints and requirements

### Step 2: Map ERD to Sprint Goals

**Actions:**
- Review sprint goals from PRD
- Identify which ERD components support each sprint goal
- Determine the order of work needed to achieve goals
- Identify any gaps between ERD and sprint goals

**Output:**
- Mapping of ERD components to sprint goals
- Work order/sequence
- Gap analysis (if any)

### Step 3: Identify Natural Boundaries

**Actions:**
- Look for natural breakpoints in the ERD:
  - Entity boundaries (one entity = potential ticket)
  - API endpoint boundaries (one endpoint = potential ticket)
  - Feature boundaries (one feature = potential ticket)
  - Layer boundaries (one layer = potential ticket)
- Consider what can be worked on in parallel
- Identify what must be sequential

**Output:**
- List of potential ticket boundaries
- Parallel work opportunities
- Sequential dependencies

### Step 4: Create Initial Ticket List

**Actions:**
- Create one ticket per natural boundary
- Write clear titles and descriptions
- Add ERD references
- Document dependencies
- Estimate size (small/medium/large or story points)

**Output:**
- Initial list of tickets with basic information

### Step 5: Refine and Validate Tickets

**Actions:**
- Apply INVEST criteria to each ticket
- Split tickets that are too large
- Combine tickets that are too small (unless truly independent)
- Ensure each ticket has clear acceptance criteria
- Verify dependencies are correct
- Check that all tickets together cover the ERD requirements

**Output:**
- Refined ticket list meeting INVEST criteria
- Complete acceptance criteria for each ticket
- Validated dependency chain

### Step 6: Order Tickets

**Actions:**
- Order tickets by dependencies (blocking tickets first)
- Identify tickets that can be worked on in parallel
- Consider risk (do risky tickets early to surface issues)
- Consider value delivery (deliver value early when possible)

**Output:**
- Ordered ticket list
- Parallel work groups
- Risk and value considerations

## Common Patterns

### Pattern 1: Foundation First

**Description:** Create foundational components before building on top of them.

**Example:**
- Ticket 1: Database schema setup
- Ticket 2: Core data models
- Ticket 3: Repository layer
- Ticket 4: Service layer
- Ticket 5: API layer

**When to Use:**
- Clear architectural layers
- Strong dependencies between layers
- Infrastructure-heavy work

### Pattern 2: Vertical Slice

**Description:** Implement a complete feature end-to-end (all layers) before moving to the next feature.

**Example:**
- Ticket 1: User registration (data + service + API)
- Ticket 2: User login (data + service + API)
- Ticket 3: User profile (data + service + API)

**When to Use:**
- Feature-driven development
- Need to deliver working features quickly
- Minimal coupling between features

### Pattern 3: Horizontal Slice

**Description:** Implement one layer across all features before moving to the next layer.

**Example:**
- Ticket 1: All data models
- Ticket 2: All service layers
- Ticket 3: All API endpoints

**When to Use:**
- Team organized by technical expertise
- Need to establish patterns before scaling
- Clear layer boundaries

### Pattern 4: Risk-First

**Description:** Tackle risky or uncertain work early to surface issues.

**Example:**
- Ticket 1: Third-party API integration (risky)
- Ticket 2: Core feature implementation
- Ticket 3: Additional features

**When to Use:**
- High technical uncertainty
- External dependencies
- Unproven technologies

### Pattern 5: Value-First

**Description:** Deliver user-visible value as early as possible.

**Example:**
- Ticket 1: Basic user registration (MVP)
- Ticket 2: Email verification (enhancement)
- Ticket 3: Password reset (enhancement)

**When to Use:**
- Need early user feedback
- Incremental value delivery
- User-facing features

## Anti-Patterns to Avoid

### Anti-Pattern 1: God Tickets

**Description:** Tickets that try to do everything at once.

**Bad Example:**
- "Implement user authentication system" (includes registration, login, password reset, email verification, OAuth, etc.)

**Why It's Bad:**
- Too large to complete in a sprint
- Hard to estimate accurately
- Difficult to track progress
- High risk of scope creep

**How to Fix:**
- Break into smaller, focused tickets
- One ticket per major feature or component

### Anti-Pattern 2: Vague Tickets

**Description:** Tickets with unclear requirements or acceptance criteria.

**Bad Example:**
- "Make user stuff work better"
- "Fix bugs in authentication"
- "Improve performance"

**Why It's Bad:**
- Can't estimate accurately
- Unclear what "done" means
- Leads to scope creep
- Difficult to test

**How to Fix:**
- Add specific, testable acceptance criteria
- Reference ERD sections explicitly
- Define what "done" means clearly

### Anti-Pattern 3: Implementation-Prescriptive Tickets

**Description:** Tickets that prescribe specific implementation details rather than requirements.

**Bad Example:**
- "Use React hooks useState and useEffect to implement form validation with library X version Y"

**Why It's Bad:**
- Removes developer autonomy
- Prevents better solutions
- Hard to negotiate or refine
- May not be the best approach

**How to Fix:**
- Focus on requirements and constraints
- Reference ERD for technical constraints
- Let developers choose implementation approach

### Anti-Pattern 4: Missing Dependencies

**Description:** Tickets that don't document dependencies, leading to blockers.

**Bad Example:**
- Ticket: "Create user API endpoint" (doesn't mention it needs User model first)

**Why It's Bad:**
- Work can't start when expected
- Parallel work becomes sequential
- Surprise blockers delay sprint

**How to Fix:**
- Explicitly document all dependencies
- Order tickets by dependency chain
- Identify blocking vs. enabling dependencies

### Anti-Pattern 5: Too Small Tickets

**Description:** Tickets that are trivial or too granular.

**Bad Example:**
- "Add semicolon to line 42"
- "Change variable name from x to y"
- "Add one comment"

**Why It's Bad:**
- Overhead exceeds value
- Hard to track meaningfully
- Creates ticket clutter
- Doesn't represent meaningful work

**How to Fix:**
- Combine small related tasks
- Only create tickets for meaningful units of work
- Use checklists within tickets for small tasks

### Anti-Pattern 6: ERD-Ignorant Tickets

**Description:** Tickets that don't reference or align with the ERD.

**Bad Example:**
- "Create user API" (doesn't reference ERD API design section)

**Why It's Bad:**
- May not align with architecture
- Could introduce inconsistencies
- Misses important requirements
- Doesn't leverage ERD work

**How to Fix:**
- Always reference relevant ERD sections
- Ensure tickets align with ERD architecture
- Use ERD as source of truth for technical requirements

## Ticket Quality Checklist

Before considering a ticket ready for implementation, verify:

- [ ] Title is clear and action-oriented
- [ ] Description explains what and why, references ERD sections
- [ ] Acceptance criteria are specific, testable, and binary (pass/fail)
- [ ] Technical details reference ERD (architecture, data model, APIs, etc.)
- [ ] Dependencies are explicitly documented (blocking vs. enabling)
- [ ] Testing requirements are specified (unit, integration, edge cases)
- [ ] Ticket meets INVEST criteria (Independent, Negotiable, Valuable, Estimable, Small, Testable)
- [ ] Ticket size is appropriate (1-3 days of work, not too small or large)
- [ ] Ticket contributes to sprint goals
- [ ] ERD requirements are fully covered across all tickets

## Examples

### Example 1: Good Ticket (Data Model)

**Title:** Create User entity and database migration

**Description:**
Implement the User entity as specified in ERD section 4.1 (Data Model). This includes the database schema, model class, and basic validation rules. This is foundational work for the authentication system and supports sprint goal "Users can register and log in".

**ERD References:**
- Section 4.1: User entity definition
- Section 4.2: Data relationships (User has many Orders)
- Section 7.1: Security requirements (password hashing)

**Acceptance Criteria:**
- Given valid user data, when User is created, then record is stored in database with hashed password
- Given invalid email format, when User creation is attempted, then validation error is returned
- Given duplicate email, when User creation is attempted, then unique constraint error is returned
- Given password less than 8 characters, when User creation is attempted, then validation error is returned
- Given valid input, when User is created, then password is hashed using bcrypt (as per ERD section 7.1)

**Technical Details:**
- Files to create: `models/user.go`, `migrations/001_create_users.sql`
- Database fields: id, email, password_hash, created_at, updated_at
- Validation: Email format, password length (min 8 chars), email uniqueness
- ERD compliance: Implements User entity from ERD section 4.1

**Dependencies:**
- Blocks: None (foundational work)
- Blocked by: None (can start immediately)
- Related: Ticket #2 (User API endpoints), Ticket #3 (User authentication)

**Testing Requirements:**
- Unit tests: User model validation, password hashing, email format validation
- Integration tests: User creation flow, database constraints
- Edge cases: Empty strings, null values, SQL injection attempts, very long strings
- Coverage: 80%+ on new code

**Size:** Medium (2-3 days)

### Example 2: Good Ticket (API Endpoint)

**Title:** Implement POST /users endpoint for user registration

**Description:**
Create the user registration API endpoint as specified in ERD section 5.1 (API Design). This endpoint accepts user registration data, validates input, creates a User record, and returns appropriate responses. Supports sprint goal "Users can register and log in".

**ERD References:**
- Section 5.1: POST /users endpoint specification
- Section 5.2: Request/response formats
- Section 5.3: Error handling
- Section 4.1: User entity (uses Ticket #1)

**Acceptance Criteria:**
- Given valid registration data, when POST /users is called, then User is created and 201 response is returned
- Given invalid email format, when POST /users is called, then 400 error with validation message is returned
- Given duplicate email, when POST /users is called, then 409 conflict error is returned
- Given missing required fields, when POST /users is called, then 400 error with field-specific messages is returned
- Given server error during creation, when POST /users is called, then 500 error is returned

**Technical Details:**
- Endpoint: POST /api/v1/users
- Request body: { email: string, password: string }
- Response (201): { id: number, email: string, created_at: timestamp }
- Response (400/409/500): { error: string, message: string }
- Uses User model from Ticket #1
- Implements API contract from ERD section 5.1

**Dependencies:**
- Blocks: Ticket #4 (User login endpoint)
- Blocked by: Ticket #1 (User model must exist)
- Related: Ticket #3 (Input validation middleware)

**Testing Requirements:**
- Unit tests: Endpoint handler logic, validation, error handling
- Integration tests: Full registration flow, database interaction, error scenarios
- Edge cases: Malformed JSON, missing fields, very long strings, special characters
- Performance: Response time < 200ms for 95% of requests (per ERD section 7.2)

**Size:** Medium (2-3 days)

### Example 3: Bad Ticket (Too Vague)

**Title:** User stuff

**Description:**
Make users work.

**Why It's Bad:**
- Vague title and description
- No ERD references
- No acceptance criteria
- No technical details
- No dependencies
- Can't estimate or test
- Doesn't meet INVEST criteria

## Conclusion

Effective ticket decomposition transforms ERD technical designs into actionable, implementable work. By following decomposition strategies, applying INVEST criteria, and avoiding common anti-patterns, you can create tickets that enable efficient parallel work, clear progress tracking, and successful sprint delivery.

Remember: The goal is not perfection, but creating tickets that are clear, actionable, and aligned with the ERD. Each ticket should represent a meaningful unit of work that contributes to sprint goals and can be completed independently (once dependencies are met).

The ERD is your source of truth for technical requirements. Always reference it, align tickets with it, and ensure your ticket decomposition fully covers the ERD's technical design.
