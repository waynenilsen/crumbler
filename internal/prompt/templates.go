package prompt

// Preamble templates explain the crumbler system to the AI agent.

const preambleFull = `# Crumbler Agent Instructions

You are working on a task managed by crumbler. Read the README below and decide what to do.

## Your Decision

**1. EXECUTE** - If the task is clear and small enough to do now:
   - Make the necessary code changes
   - When done, run: ` + "`crumbler delete`" + ` then exit

**2. DECOMPOSE** - If the task needs to be broken down:
   - Run: ` + "`crumbler create \"Task A\" \"Task B\" \"Task C\"`" + `
   - Fill in each sub-crumb's README.md with task details
   - Exit (the loop will continue with the first sub-crumb)

## Commands

- ` + "`crumbler create \"Name\" [\"Name2\"...]`" + ` - Create sub-crumbs under current crumb
- ` + "`crumbler delete`" + ` - Delete current crumb (when work is done)
- ` + "`crumbler status`" + ` - Show the crumb tree
- ` + "`crumbler prompt`" + ` - Get the next prompt

## Rules

- Each crumb can have at most 10 children
- Always exit after completing your action so context resets
- Keep READMEs concise but clear

`

const preambleMinimal = `# Crumbler

Read the README. Either EXECUTE (do work, then ` + "`crumbler delete`" + `) or DECOMPOSE (` + "`crumbler create`" + ` sub-tasks).
Exit when done.

`

// Postamble templates remind the agent what to do next.

const postambleFull = `## Next Steps

After you act:
- If you **executed** the work: ` + "`crumbler delete`" + ` then exit
- If you **decomposed** into sub-crumbs: exit (loop continues)

Exit when done so context resets and the loop continues.
`

const postambleMinimal = `Exit after: ` + "`crumbler delete`" + ` (if done) or decomposing.
`

// formatPreamble returns the appropriate preamble.
func formatPreamble(minimal bool) string {
	if minimal {
		return preambleMinimal
	}
	return preambleFull
}

// formatPostamble returns the appropriate postamble.
func formatPostamble(minimal bool) string {
	if minimal {
		return postambleMinimal
	}
	return postambleFull
}
