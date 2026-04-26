---
name: use-google-jules
description: Guide for using Jules Tools (the Jules CLI) to interact with Google's autonomous AI coding agent.
---

# Use Jules Tools (jules)

This skill documents how to use Jules Tools (`jules`), a lightweight command-line interface for interacting with Jules, Google's autonomous AI coding agent.

## Command Reference

The `remote` command is the primary way to interact with Jules sessions running in the cloud.

### 1. List Repositories and Sessions

```bash
# List all connected repositories
jules remote list --repo

# List all active and past sessions
jules remote list --session
```

### 2. Create a New Session

Use `remote new` to delegate a task to Jules. Jules can automatically infer the repository from your current directory, so you can often omit the `--repo` flag or use `.`.

```bash
# Start a new session for the current repository
jules remote new --repo . --session "write unit tests for the auth module"

# Start a session for a specific repository
jules remote new --repo torvalds/linux --session "optimize memory allocation"

# Start multiple parallel sessions for the same task
jules remote new --repo . --session "refactor the API client" --parallel 3
```

### 3. Pull Session Results

Once a session completes, pull the code changes or results locally:

```bash
# Pull the results for a specific session ID
jules remote pull --session 123456
```

## Advanced Scripting Examples

Jules Tools is designed to be highly composable with other command-line tools. You can pipe outputs into Jules to automate task delegation.

**1. Create sessions from a TODO.md file:**
Assign each line item from a local `TODO.md` file as a new session in the current repository:

```bash
cat TODO.md | while IFS= read -r line; do
  jules remote new --repo . --session "$line"
done
```

**2. Create a session from a GitHub Issue:**
Pipe the title of the first GitHub issue assigned to you directly into a new Jules session:

```bash
gh issue list --assignee @me --limit 1 --json title | jq -r '.[0].title' | jules remote new --repo .
```

**3. Analyze and assign tasks using the Gemini CLI:**
Use the Gemini CLI to analyze your assigned GitHub issues, identify the most tedious one, and pipe its title to Jules:

```bash
gemini -p "find the most tedious issue, print it verbatim\n$(gh issue list --assignee @me)" | jules remote new --repo .
```

## Global Flags

- `-h, --help`: Displays help information for `jules` or a specific command.
- `--theme <string>`: Sets the theme for the TUI (`dark` or `light`).
