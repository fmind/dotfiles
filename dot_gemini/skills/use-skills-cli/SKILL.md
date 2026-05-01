---
name: use-skills-cli
description: Guide for using the `npx skills` CLI (vercel-labs/skills) — discover, install, and update Agent Skills from public bundles, in project (`.agents/skills/`) or global (`~/.gemini/skills/`) scope.
---

# Use skills CLI

[`vercel-labs/skills`](https://github.com/vercel-labs/skills) ships `npx skills` — a small, stack-agnostic CLI that discovers and installs Agent Skills (SKILL.md folders) into a project or globally for an agent (Gemini CLI, Claude Code, …). Once installed, skills are plain folders and have no runtime dependency on the CLI.

> Despite living under the `vercel-labs` org, the CLI is **not** Vercel-specific. It can pull any GitHub-published `*/skills` or `*/agent-skills` repo (e.g. `vercel-labs/agent-skills`, `addyosmani/web-quality-skills`, or your own).

## Install

`npx` runs ephemerally — no install step. Node.js is already provided by mise on this dotfile setup.

```bash
# Sanity check.
npx skills --version
npx skills --help
```

## Discover

```bash
# Interactive picker for skills matching a query (consults the public leaderboard).
npx skills find "view transitions"

# JSON output for scripting / agent use.
npx skills find "react performance" --json

# List the skills inside a published bundle without installing.
npx skills add <owner>/<repo> --list
```

## Install Skills

```bash
# Install a whole bundle (interactive scope prompt: project vs global).
npx skills add <owner>/<repo>

# Install a specific skill from a bundle.
npx skills add <owner>/<repo> --skill <skill-slug>

# Install several skills at once.
npx skills add <owner>/<repo> \
  --skill <skill-a> \
  --skill <skill-b>
```

Examples:

```bash
# UI/UX review pass (stack-agnostic).
npx skills add vercel-labs/agent-skills --skill web-design-guidelines

# Discovery meta-skill so the agent suggests further skills proactively.
npx skills add vercel-labs/skills --skill find-skills

# Lighthouse-style web-quality audits.
npx skills add addyosmani/web-quality-skills
```

## Scope

`npx skills` writes into one of two locations (it asks interactively):

- **Project scope (default)** → `.agents/skills/<slug>/SKILL.md` — repo-pinned, commits with the codebase, available to anyone who clones it.
- **Global scope** → `~/.gemini/skills/<slug>/SKILL.md` (or `~/.claude/skills/<slug>/`, depending on the active agent) — machine-local. Run `chezmoi add ~/.gemini/skills/<slug>` if you want it dotfile-managed.

## Maintain

```bash
# Update installed skills to their latest published version.
npx skills update
```

Run `npx skills --help` to confirm the full subcommand surface in your installed version — the CLI is young and ships new verbs frequently.

## After Install

1. Restart the agent (Gemini CLI or Claude Code) so progressive disclosure picks up new SKILL descriptions.
2. Project-scope installs commit naturally with the repo. Global-scope installs are machine-local — track with `chezmoi add` if relevant.
3. Re-run `npx skills update` periodically to pull upstream improvements.

## Important Notes

1. The CLI is **stack-agnostic** — works for any Agent Skill regardless of framework or vendor.
2. Skills are **plain folders + SKILL.md** — once installed they have no runtime dependency on `npx skills`.
3. `find-skills` (the meta-skill in `vercel-labs/skills`) is what teaches the agent to suggest installs proactively; the CLI alone is just a fetcher. See `install-find-skills` for the wrapper.
4. For authoring **new, project-local** skills (not installing published ones), see `create-agent-skill` instead.

## Documentation

- [`vercel-labs/skills` repo](https://github.com/vercel-labs/skills)
- [`vercel-labs/agent-skills` bundle](https://github.com/vercel-labs/agent-skills)
- [Vercel changelog: skills v1.1.1 (interactive discovery)](https://vercel.com/changelog/skills-v1-1-1-interactive-discovery-open-source-release-and-agent-support)
- [Vercel KB: Agent Skills — creating, installing, sharing](https://vercel.com/kb/guide/agent-skills-creating-installing-and-sharing-reusable-agent-context)
