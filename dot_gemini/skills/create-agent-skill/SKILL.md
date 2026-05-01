---
name: create-agent-skill
description: Guide for creating an Agent Skill for this dotfile repo — primarily targeting Gemini CLI, project- vs global-scope, frontmatter format, and progressive disclosure.
---

# Create Agent Skill

Agent Skills are an open standard (originated by Anthropic, adopted by Claude Code, Gemini CLI, Cursor, OpenCode, and others — see [agentskills.io](https://agentskills.io)). A skill is a directory with a `SKILL.md` whose frontmatter is read at startup; the body is loaded only when the agent decides the skill is relevant (progressive disclosure).

This repo treats **Gemini CLI as the primary skill consumer**. Skills load from `~/.gemini/skills/`. Skills are not auto-shared with Claude Code in this layout; if you need cross-tool reach, install the same skill into both surfaces explicitly.

## Discovery Locations

When creating a skill, choose its scope:

| Scope | Path | When to use |
|-------|------|-------------|
| Project | `.agents/skills/<slug>/` (in the repo root) | Project-specific, version-controlled with the repo. Discovered when Gemini CLI is launched from inside the repo. |
| Global | `~/.gemini/skills/<slug>/` — tracked via chezmoi at `dot_gemini/skills/<slug>/` | Personal skill that should travel with the user across every project. After a fresh checkout, `mise run apply` (= `mr a`) deploys it. |

Default to **project scope** unless the user explicitly says "global" or the skill is clearly machine-wide (e.g. a wrapper around a CLI installed everywhere).

## Skill Directory Layout

```text
<slug>/
├── SKILL.md       # (required) frontmatter + instructions
├── scripts/       # (optional) executable helpers
├── references/    # (optional) static docs the agent can grep
└── assets/        # (optional) templates, prompts, fixtures
```

## `SKILL.md` Format

```markdown
---
name: <slug>
description: <one-line trigger that helps the agent decide when to load this skill>
---

# <Skill Title>

Short framing paragraph...

## Instructions

1. ...
2. ...
```

Frontmatter — only two fields are required:

- `name` (required): slug — lowercase letters, digits, hyphens. Must match the folder name.
- `description` (required): single sentence that lets the parent agent decide when this skill is relevant. The body is **not** loaded until the description matches, so make it concrete and trigger-rich.

The body is plain markdown: procedure, conventions, constraints, examples. Stay tool-agnostic where possible. If the skill is genuinely Gemini-specific (e.g. uses a Gemini CLI extension, slash command, or subagent format), name it accordingly (`create-gemini-cli-command`, `install-gemini-skills`, …) so it is obvious from the slug.

## Skill Naming

- Lowercase letters, digits, hyphens only; under 64 characters total.
- Verb-led where possible (`use-…`, `install-…`, `configure-…`, `create-…`, `run-…`).
- Folder name MUST match the `name:` in frontmatter exactly.
- Namespace by tool when it sharpens triggering (`gh-address-comments`, `linear-create-issue`) instead of a generic verb.

## Step-by-Step Creation

1. **Pick a slug.** Apply the rules above (`use-foo-cli`, `install-bar-mcp`, `configure-baz`).
2. **Create the folder.**
   - Project scope (default): `mkdir -p .agents/skills/<slug>/` from the repo root.
   - Global scope: `mkdir -p ~/.local/share/chezmoi/dot_gemini/skills/<slug>/` (then `mr a` to deploy to `~/.gemini/skills/<slug>/`).
3. **Add optional subfolders** (`scripts/`, `references/`, `assets/`) only if the skill bundles resources.
4. **Write `SKILL.md`** with frontmatter (`name`, `description`) and a tight, procedural body.
5. **Verify upstream facts** mentioned in the body — check the tool's current docs before pinning flags, URLs, package names, or endpoints.
6. **Iterate** by trial-running the skill in a new Gemini CLI session; refine the description until progressive disclosure picks it up reliably for the intended trigger.

## Future-Proofing Rules

Skill files outlive the moment they were written — CLI versions, model generations, and product names turn over faster than skill prose gets re-edited. Write each skill so it stays correct without maintenance.

- [ ] **No time-stamped prose.** Avoid "As of 2026…", "announced at Cloud Next 2026", "Q4 2025", "since 2026", "added in April 2026". Keep "currently" / "recently" out of capability descriptions. Write in plain present tense and let linked docs carry the timestamp.
- [ ] **No forward predictions.** Don't write "GA in 2025", "coming Q3", "will ship next release". Predictions become wrong and stay wrong; describe today's behavior and link to the canonical status page.
- [ ] **No rebrand narration.** Don't write "rebranded from X", "formerly known as Y", "the new name for Z". Use the current product name; if disambiguation is genuinely useful (e.g. "Gemini Cloud Assist (NOT Gemini Code Assist)"), keep it to a one-liner aimed at *clarification*, not history.
- [ ] **No specific model versions in prose.** Don't pin "powered by Gemini 3 Pro / Flash" or "uses Claude 4.5 Sonnet". Reference model families abstractly ("Gemini Pro/Flash family") or omit — pinning a generation locks the file to that generation.
- [ ] **Pin executable references; don't assume global installs.** In install commands, prefer `npx -y <pkg>@latest` or `uvx --from <pkg>` over `command: <bare-binary>`. A globally-installed CLI on the user's PATH drifts independently of the skill file. For Docker, pin the image tag explicitly.
- [ ] **Cite canonical product docs, not announcements.** Link to `cloud.google.com/<product>/docs` and similar evergreen paths. Avoid blog posts, release notes, codelabs, and "introducing X" URLs whose paths rot. Where an announcement is genuinely informative, label the link by purpose ("Launch announcement"), not by date.
- [ ] **Defer fast-moving lists to upstream.** Subcommand catalogs, supported-product tables, model IDs — point at the canonical "always-current" reference page rather than enumerating in the skill. Static lists rot the moment the upstream adds an entry.
- [ ] **Describe the skill's *purpose*, not the product's *current state*.** Capabilities and workflows describe what the skill teaches; they don't need to narrate where the product sits in its lifecycle. If a feature is in preview, say "preview" — don't say "newly released in preview".

## Installing Third-Party Skills

For published skills, prefer the `skills` CLI (installed via mise; `npx skills` works on a fresh checkout):

```bash
# Project scope (default — writes to .agents/skills/).
skills add <owner>/<repo> --skill <name>

# Global scope (writes to ~/.gemini/skills/).
skills add --global <owner>/<repo> --skill <name>
```

After a global install, run `chezmoi add ~/.gemini/skills/<slug>` to import the skill into `dot_gemini/skills/<slug>/` so it's tracked.

For wrappers around official bundles (e.g. `install-firebase-skills`, `install-stitch-skills`), hand-author a thin `dot_gemini/skills/install-*-skills/SKILL.md` that documents the exact `skills add ...` invocation — see the existing `install-*-skills` directories for the pattern.

## Documentation

- [Agent Skills standard (agentskills.io)](https://agentskills.io)
- [Gemini CLI skills reference](https://geminicli.com/docs/cli/skills/)
- [`vercel-labs/skills` CLI](https://github.com/vercel-labs/skills) — `skills add`, `skills find`, `skills update`
- [Public skill registry](https://skills.sh)
- Companion skills: `install-find-skills` (discovery meta-skill), `configure-gemini-cli` (skills directories live alongside MCP servers in `~/.gemini/`).
