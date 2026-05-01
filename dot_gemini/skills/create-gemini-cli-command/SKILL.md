---
name: create-gemini-cli-command
description: Guide for creating Gemini CLI slash commands — TOML schema (`prompt`, `description`), `!{shell}` / `@{file}` / `{{args}}` injection, namespaced subdirectories, and scope (workspace vs global via chezmoi).
---

# Create Gemini CLI Command

Gemini CLI slash commands are short, reusable prompts stored as TOML files. They surface in the CLI as `/<filename>`. For more details, refer to the [official Gemini CLI custom commands documentation](https://geminicli.com/docs/cli/custom-commands/).

## Location

- **Local (Workspace):** `.gemini/commands/<name>.toml` (Project-specific commands).
  - If the user does not specify where to create the command, assume it should be local.
- **Global (chezmoi source of truth):** `~/.local/share/chezmoi/dot_gemini/commands/<name>.toml` (Personal commands available across all workspaces after deployment).

Ask the user to run `mise run apply` to deploy global commands to `~/.gemini/commands/`.

## File Structure

```toml
description = "One-line summary that appears in autocomplete and /help."

prompt = """
The full prompt body sent to the model. Supports three forms of injection:

- `!{shell command}` runs the command (with user confirmation) and inlines its stdout.
- `@{path}` embeds the contents of a file or directory (respects .gitignore / .geminiignore; supports images, PDFs, audio, video).
- `{{args}}` is replaced with whatever the user typed after the command name. Inside `!{...}` the substitution is shell-escaped automatically; outside, it is injected raw.

If `{{args}}` is omitted entirely, any user input is appended to the prompt with a blank-line separator.

Return only the requested output, no quotes, no markdown blocks.
"""
```

Only two TOML keys are recognised: `prompt` (required) and `description` (optional but strongly recommended).

### Subdirectories become namespaces

`commands/git/commit.toml` → `/git:commit`. Use this for grouping related commands under a prefix.

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

1. **Pick a verb-first name.** `commit`, `review`, `explain`, `pr-summary`. Use a `<group>/<name>.toml` layout if you want a `:`-namespaced command.
2. **Create the file** at the correct scope. For global commands in this dotfiles repo, write the source under `~/.local/share/chezmoi/dot_gemini/commands/`.
3. **Write a tight `description`** — this is what the user sees in autocomplete and `/help`.
4. **Write the `prompt`** — be explicit about the output format. Inline shell output with `!{...}` and file content with `@{...}` instead of asking the user to paste. Use `{{args}}` for free-form parameters.

## Guidelines

- One responsibility per command. Compose, don't bloat.
- Always specify the desired output format ("only the message", "JSON only", "single sentence"). Saves the user a round-trip.
- Quote shell snippets carefully when they may contain backticks; prefer single-line `!{...}` blocks when possible.
- Keep `{{args}}` substitution outside of `!{...}` shell blocks unless you want raw injection — inside, arguments are auto-escaped, which can break expected quoting.

## Future-proofing

Commands often outlive the tool versions, models, and CLI flags they were written against. Write each prompt so it survives that drift — no rewrite required when the toolchain changes.

- **Detect, don't assume.** Read project manifests (`pyproject.toml`, `package.json`, `go.mod`, `Cargo.toml`, `mise.toml`, `Makefile`, `.pre-commit-config.yaml`, ...) to discover the toolchain, then run the project's own scripts before reaching for raw tools.
- **List alternatives, never one tool by name only.** When you must reference concrete tools, list realistic ecosystem peers (`pnpm | npm | bun`, `mypy | pyright | ty`, `pytest | jest | go test`, `ruff | eslint | biome`) so the model picks what the project actually uses.
- **No model or version pins.** Never name a specific model (`gemini-2.5-pro`, `claude-opus-4-7`) — the harness owns model selection. Never pin tool versions (`pre-commit 3.x`, `Python 3.11`, `Node 20`); state the capability, not the version. Avoid date stamps and knowledge-cutoff phrasing.
- **Prefer stable conventions and official CLIs.** Lean on widely-adopted conventions (Conventional Commits, semver, `AGENTS.md`, `README.md`, `.editorconfig`) and on official CLIs (`gh`, `gcloud`, `git`) — they abstract URL and API churn. Favor long-stable flags over freshly added ones.
- **Don't lock to Gemini CLI internals.** Avoid hardcoding subagent, extension, MCP server, or skill names by name unless the command's purpose is to manage them — renames will silently break the command.
- **Lock the output shape, not the writer.** Specify exact headers, sections, ordering, and field names so downstream tooling that parses output keeps working — and let the model fill them in.

## Documentation

- [Gemini CLI custom commands](https://geminicli.com/docs/cli/custom-commands/)
- [Gemini CLI configuration reference](https://geminicli.com/docs/reference/configuration/)
