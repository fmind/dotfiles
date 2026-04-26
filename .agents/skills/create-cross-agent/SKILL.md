---
name: create-cross-agent
description: Create a subagent in this chezmoi repo that targets Claude Code and/or Gemini CLI, deciding whether the agent can be shared or must be tool-specific.
---

# Create Cross-Tool Subagent

Use this skill when adding a new subagent to this dotfiles repo. Subagents are the **least** portable artifact between Claude Code and Gemini CLI — the file format overlaps but the capability model differs. Read this skill first to decide whether to write one shared agent or two tool-specific ones.

## Format comparison

Both tools use markdown files with YAML frontmatter. The body (everything after the second `---`) is identical in both. The frontmatter overlaps but is **not** identical.

| Field | Claude Code (`dot_claude/agents/<name>.md`) | Gemini CLI (`dot_gemini/agents/<name>.md`) |
|---|---|---|
| `name` | required, must match filename | required, must match filename |
| `description` | required (trigger condition) | required (trigger condition) |
| `tools` | comma-separated string: `Read, Grep, Bash` | YAML list with wildcards: `["*"]`, `["mcp_*"]`, `["mcp_<server>_*"]` |
| `model` | optional: `sonnet` / `opus` / `haiku` | not used (Gemini chooses) |
| `kind` | not used | optional: `local` (default) or `remote` |
| `mcp_servers` | not used (MCP is global in `.mcp.json`) | optional, per-agent inline MCP config (camelCase keys) |
| Body | full markdown system prompt | full markdown system prompt |

**Key takeaway**: the body is portable; the frontmatter is not. Specifically, `tools` is parsed differently in each tool, so it cannot be set to a syntax that satisfies both.

## Decision tree

Before creating the agent, decide which of the three patterns fits.

### 1. Cross-tool agent (single source, symlinked or duplicated)

Use this when **all** of the following hold:

- The agent's purpose is generic (code review, doc writer, refactor planner) — not tied to a specific MCP integration.
- You can omit `tools` (both tools default to inheriting all parent tools).
- You don't need `model` (Claude) or `mcp_servers` (Gemini).

In this case, write **one** file with minimal frontmatter and place a copy (or symlink) in both `dot_claude/agents/` and `dot_gemini/agents/`:

```markdown
---
name: <agent-name>
description: <When to delegate to this agent>
---

# <Agent Title>

You are the specialized <agent-name> agent. Your primary goal is to ...
```

Implementation today: write the file in `dot_claude/agents/<name>.md`, then `git mv` or symlink it to `dot_gemini/agents/<name>.md`. (A future improvement: create `dot_agents/agents/` and symlink both tools' `agents/` to it, mirroring the skills layout — same caveat as below.)

### 2. Two parallel agents (one per tool)

Use this when the agent **must** declare per-tool frontmatter — for example a Gemini agent that bundles an MCP server (`firebase`, `cloud-run`, ...) which Claude can't use the same way, or a Claude agent that pins `model: opus`.

Create:

- `dot_claude/agents/<name>.md` with Claude-style frontmatter (no `mcp_servers`, no `kind`, `tools` as comma-separated string, optional `model`)
- `dot_gemini/agents/<name>.md` with Gemini-style frontmatter (`tools` as YAML list, optional `kind` and `mcp_servers`)

Keep the body identical between the two files. Diff them whenever you change either.

### 3. Single-tool agent

Use this when the agent only makes sense in one tool (e.g. all the MCP-bundled Google Cloud agents under `dot_gemini/agents/`). Don't fake portability — keep the file in just one place.

## Step-by-Step

1. **Decide the pattern** above (cross / parallel / single).
2. **Pick the name**: lowercase, hyphenated, descriptive. Match the filename.
3. **Write the body**: H1 title → persona ("You are the specialized ... agent.") → primary goal → tool-use directives → output format. The body is the part that's portable.
4. **Write the frontmatter** per the comparison table:
   - **Cross**: just `name` + `description`. Place file in both `dot_claude/agents/` and `dot_gemini/agents/`.
   - **Parallel**: tool-specific frontmatter in each file. Body identical.
   - **Single**: full tool-specific frontmatter, single location.
5. **Deploy**: ask the user to run `mise run apply`. Test the agent from each target tool.

## Pitfalls

- **`tools` syntax** is the most common breakage. Comma-separated string crashes Gemini's YAML parser; YAML list crashes Claude's tools parser. If you need to constrain tools, you **must** use the parallel-agents pattern.
- **`mcp_servers`** must be camelCase in Gemini frontmatter, but the *MCP transport keys inside it* (`httpUrl`, `command`, `args`, `env`, `headers`) follow MCP convention. Don't confuse the two layers.
- **No shared agents/ directory yet**. Skills get symlinked via `dot_agents/skills/`, but agents currently live separately because their frontmatter often diverges. Don't symlink `dot_*/agents/` blindly — verify each agent fits the cross-tool pattern first.
- **Don't pin `model`** in a cross-tool agent — Gemini ignores it but the file becomes "Claude-flavored" and tempts future drift.

## Reference: existing agents in this repo

- `dot_claude/agents/` — currently empty.
- `dot_gemini/agents/` — 30 agents, almost all parallel/single-tool (most bundle MCP servers: `firebase`, `cloud-run`, `vertex-ai`, ...). They are good examples of when **not** to attempt cross-tool sharing.
