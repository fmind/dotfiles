---
name: create-cross-command
description: Create a slash command in this chezmoi repo that ships to both Claude Code (.md) and Gemini CLI (.toml) from one logical definition.
---

# Create Cross-Tool Slash Command

Use this skill when adding a new slash command to this dotfiles repo that should be available in **both** Claude Code (`/<name>`) and Gemini CLI (`/<name>`). The two tools use different file formats; this skill keeps them in lock-step.

## Why two files (for now)

There is no shared command standard between Claude Code and Gemini CLI. Until one emerges, the canonical approach is:

- `dot_claude/commands/<name>.md`  → Claude Code (Markdown + YAML frontmatter)
- `dot_gemini/commands/<name>.toml` → Gemini CLI (TOML with `description` + `prompt` keys)

Both files **share the same prompt body** but use slightly different syntax for shell substitution and arguments. The mapping is below — keep the two files in sync whenever you change either one.

## Syntax mapping

| Concept | Claude (`.md`) | Gemini (`.toml`) |
|---|---|---|
| Description metadata | YAML frontmatter `description: ...` | TOML key `description = "..."` |
| Shell substitution | `` !`git diff` `` | `!{git diff}` |
| User arguments | `$ARGUMENTS` | `{{args}}` |
| File embed | (paste manually or `!\`cat ...\``) | `@{path/to/file}` |
| Tool restrictions | `allowed-tools: Bash(git:*)` | (no equivalent — handled at session level) |
| Argument hint | `argument-hint: [extra context]` | (no equivalent) |
| Namespacing | (single flat dir) | subdir → `/group:name` |

## File templates

### `dot_claude/commands/<name>.md`

```markdown
---
description: <One-line summary shown in autocomplete.>
argument-hint: [optional placeholder]
allowed-tools: Bash(git:*), Bash(gh:*)
---

<Prompt body. Use !`shell cmd` to inline shell output. Use $ARGUMENTS for free-form input.>
```

### `dot_gemini/commands/<name>.toml`

```toml
description = "<One-line summary shown in autocomplete.>"

prompt = """
<Prompt body. Use !{shell cmd} to inline shell output. Use {{args}} for free-form input.>
"""
```

## Step-by-Step

1. **Pick a verb-first name** (`commit`, `review`, `pr`, `explain`, ...). Keep it identical across both files.
2. **Write the prompt body once** in your head: framing → embedded shell/file context → numbered requirements → output-format constraint at the end.
3. **Create `dot_claude/commands/<name>.md`**:
   - YAML frontmatter with `description` (always) and `allowed-tools` (whenever the body uses `` !`...` `` substitutions).
   - Replace shell snippets with `` !`<cmd>` ``.
   - Replace any free-form placeholder with `$ARGUMENTS`.
4. **Create `dot_gemini/commands/<name>.toml`**:
   - TOML keys `description` and `prompt = """..."""`.
   - Replace shell snippets with `!{<cmd>}`.
   - Replace `$ARGUMENTS` with `{{args}}`.
5. **Verify parity**: skim both files and confirm the prompt body is semantically identical (only the four mappings above should differ).
6. **Deploy**: ask the user to run `mise run apply`. The command then surfaces as `/<name>` in both tools.

## Guidelines

- **One responsibility per command.** If you find yourself adding an `if/else` for the two tools inside the prompt body, split the command instead.
- **Always specify the desired output format** ("only the message", "JSON only", "single sentence"). Saves a round-trip.
- **Restrict `allowed-tools` on the Claude side** whenever the body uses shell substitution — it's the safer default.
- **Keep the body identical**. The only legal divergence is the syntax mapping table above. If you need different *behavior* per tool, that's a sign you should write two separate, explicitly tool-specific commands.
- **Cross-check** by diffing the body of the two files (with the syntax mapping in mind) whenever you edit one.

## Reference: existing commands in this repo

`commit`, `pr`, `review`, `lint`, `fix`, `test`, `explain`, `improve`, `refactor`, `risk`, `spec`, `suggest`, `ask`, `agent`, `readme`, `document`, `diagram` — all exist as `.md` + `.toml` pairs. Use any of them as a template before writing a new one.
