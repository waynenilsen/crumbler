package prompt

// Preamble templates explain the crumbler system to the AI agent.

const preambleFull = `# Crumbler Agent Instructions

You are working on a project managed by crumbler, a simple task decomposition system.

## How Crumbler Works

- Work is organized into "crumbs" - directories containing a README.md file
- Crumbs can be nested to any depth (parent crumbs contain child crumbs)
- You always work on the "current" crumb (the deepest, first-by-ID leaf)
- When work is complete, delete the crumb with: crumbler delete

## Workflow

1. **DECOMPOSE**: If the README is empty, plan the work:
   - Write a clear description of what needs to be done in the README
   - If the work is too big for one task, create sub-crumbs: crumbler create "Task Name"
   - Sub-crumbs inherit context from their parent's README

2. **EXECUTE**: If the README has content, do the work:
   - Follow the instructions in the README
   - Make the necessary code changes
   - When done, delete the crumb: crumbler delete

3. **DONE**: When no crumbs remain, the project is complete

## Commands

- crumbler create "Name" - Create a sub-crumb under current crumb
- crumbler delete        - Delete the current crumb (when work is done)
- crumbler status        - Show the crumb tree
- crumbler prompt        - Get the next prompt (you're reading one now)

## Important Rules

- Each crumb can have at most 10 children (IDs 01-10)
- Always work depth-first: complete children before parents
- Keep READMEs concise but clear
- Delete crumbs when their work is done

`

const preambleMinimal = `# Crumbler

Work on crumbs (README.md directories). DECOMPOSE if empty, EXECUTE if has content.
Commands: create "Name", delete, status, prompt

`

// Postamble templates remind the agent what to do next.

const postambleFull = `## Next Steps

When you complete this task:
1. If you created sub-crumbs, run: crumbler prompt
2. If you executed the work, run: crumbler delete
3. Then run: crumbler prompt

Always end your turn by running crumbler to continue the loop.
`

const postambleMinimal = `Run: crumbler delete (if done) or crumbler prompt (to continue)
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
