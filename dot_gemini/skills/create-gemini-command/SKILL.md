---
name: create-gemini-command
description: Guide for creating reusable Gemini CLI slash commands (TOML).
---

# Create Gemini Command

Gemini CLI slash commands are short, reusable prompts stored as TOML files. They surface in the CLI as `/<filename>`. For more details, refer to the [official Gemini CLI custom commands documentation](https://geminicli.com/docs/cli/custom-commands/).

## Location

- **Local (Workspace):** `.gemini/commands/<name>.toml` (Project-specific commands).
  - If the user does not specify where to create the command, assume it should be local.
- **Global (chezmoi source of truth):** `~/.local/share/chezmoi/dot_gemini/commands/<name>.toml` (Personal commands available across all workspaces after deployment).

Ask the user to run `mise run apply` to deploy global commands to `~/.gemini/commands/`.

## File Structure

```toml
description = "One-line summary that appears in autocomplete."

prompt = """
The full prompt body. Supports:

- `!{shell command}` to inline the command's stdout.
- `{{arg}}` placeholders bound to CLI flags.

Return only the requested output, no quotes, no markdown blocks.
"""
```

### Example: conventional commit message

```toml
description = "Write a conventional commit message for staged files."

prompt = """
Write a conventional commit message for the following staged changes:
!{git diff --cached}

Return only the message, no quotes, no markdown blocks.
"""
```

## Step-by-Step Creation

1. **Pick a verb-first name.** `commit`, `review`, `explain`, `pr-summary`.
2. **Create the file** at the correct scope. For global commands in this dotfiles repo, write the source file under `~/.local/share/chezmoi/dot_gemini/commands/`.
3. **Write a tight `description`** — this is what the user sees in autocomplete.
4. **Write the `prompt`** — be explicit about the output format. Inline shell output with `!{...}` instead of asking the user to paste.

## Guidelines

- One responsibility per command. Compose, don't bloat.
- Always specify the desired output format ("only the message", "JSON only", "single sentence"). Saves the user a round-trip.
- Quote shell snippets carefully when they may contain backticks.
