---
name: create-claude-command
description: Guide for creating reusable Claude Code slash commands (Markdown).
---

# Create Claude Code Command

Claude Code slash commands are short, reusable prompts stored as Markdown files with YAML frontmatter. They surface in the CLI as `/<filename>`. For more details, refer to the [official Claude Code slash commands documentation](https://docs.claude.com/en/docs/claude-code/slash-commands).

## Location

- **Local (Workspace):** `.claude/commands/<name>.md` (project-specific commands).
  - If the user does not specify where to create the command, assume it should be local.
- **Global (chezmoi source of truth):** `~/.local/share/chezmoi/dot_claude/commands/<name>.md` (personal commands available across all workspaces after deployment).

Ask the user to run `mise run apply` to deploy global commands to `~/.claude/commands/`.

## File Structure

```markdown
---
description: One-line summary that appears in autocomplete.
argument-hint: [optional placeholder shown when typing the command]
allowed-tools: Bash(git diff:*)
---

The full prompt body. Supports:

- `` !`shell command` `` to inline the command's stdout (requires the bash tool to be allowed).
- `$ARGUMENTS` placeholder for whatever the user types after the command.

Return only the requested output, no quotes, no markdown blocks.
```

### Frontmatter keys

- `description` (required): one-line summary surfaced in autocomplete.
- `argument-hint` (optional): hint shown after the command name when the command takes args.
- `allowed-tools` (optional): comma-separated list of tools the command may invoke. Restricting tools is the safest default for commands that rely on shell substitution.
- `model` (optional): override the default model for this specific command.

### Example: conventional commit message

```markdown
---
description: Write a conventional commit message for staged files.
allowed-tools: Bash(git diff:*)
---

Write a conventional commit message for the following staged changes:
!`git diff --cached`

Return only the message, no quotes, no markdown blocks.
```

## Step-by-Step Creation

1. **Pick a verb-first name.** `commit`, `review`, `explain`, `pr-summary`.
2. **Create the file** at the correct scope. For global commands in this dotfiles repo, write the source file under `~/.local/share/chezmoi/dot_claude/commands/`.
3. **Write a tight `description`** — this is what the user sees in autocomplete.
4. **Write the prompt body** — be explicit about the output format. Inline shell output with `` !`...` `` instead of asking the user to paste.

## Guidelines

- One responsibility per command. Compose, don't bloat.
- Always specify the desired output format ("only the message", "JSON only", "single sentence"). Saves the user a round-trip.
- Constrain `allowed-tools` whenever the command uses shell substitution — it reduces accidental tool use.
- Quote shell snippets carefully when they may contain backticks.
