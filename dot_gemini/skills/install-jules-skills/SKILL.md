---
name: install-jules-skills
description: Install Jules skills bundle — dispatch Jules (Google's async coding agent) for parallel issue triage and code work, plus Jules CLI notes.
---

# Install Jules Skills

[Jules](https://jules.google) is Google's autonomous async coding agent. Its CLI (`@google/jules`, often called *Jules Tools*) lets you launch and inspect coding sessions from the terminal, and the official [`google-labs-code/jules-skills`](https://github.com/google-labs-code/jules-skills) repo packages skills that teach the *current* coding agent how to orchestrate Jules effectively.

## When to Trigger

- The user wants to triage / fix many GitHub issues in parallel and has Jules access.
- The user mentions Jules, async coding agents, or "dispatch a coding agent".
- Verify first: `ls ~/.gemini/skills/ | grep -i jules`. If installed, skip.

## Install

```bash
# Discover available skills.
npx skills add google-labs-code/jules-skills --list

# Project scope (default — installs into .agents/skills/).
npx skills add google-labs-code/jules-skills --skill automate-github-issues

# Global scope (installs into ~/.gemini/skills/).
npx skills add google-labs-code/jules-skills --skill automate-github-issues --global
```

Prefer project scope (`.agents/skills/`) so the skill commits with the repo that uses it; reach for `--global` only when the same skill is genuinely needed across many projects.

## What Gets Installed

2 skills at the time of writing:

- **automate-github-issues** — analyzes open GitHub issues, plans implementation tasks, and dispatches parallel Jules coding agents to fix them. Pairs naturally with `gh` for the GitHub side and the `jules` CLI for the dispatch side.
- **local-action-verification** — runs Jules sessions against a local checkout to verify proposed changes before they reach a remote branch.

The repo grows over time — re-run `--list` periodically to check.

## Jules CLI Quick Reference

Jules Tools is installed via mise (`npm:@google/jules`). The skill above will drive it, but for manual use:

```bash
# Start an async remote session against the current repo.
jules remote new --prompt "Add tests for the parser module"

# List / pull results.
jules remote list
jules remote pull <id>

# Print version and other info.
jules version
jules help
```

Verify the available subcommands with `jules help` — the surface is small (`version`, `remote`, `completion`, `help`) and may evolve.

Jules also has a Gemini CLI extension (`gemini extensions install jules`) for in-IDE workflows.

## After Install

1. Restart the coding agent so the new skill is picked up.
2. The `automate-github-issues` skill will, when triggered, call both `gh` (read issues) and `jules` (dispatch agents) — make sure both are authenticated.
3. Project-scope installs (`.agents/skills/`) commit naturally. For global installs, `chezmoi add ~/.gemini/skills/<slug>` to track them.

## Important Notes

1. This skill installs Jules **skills**, not Jules itself. The CLI comes from mise (`@google/jules`).
2. Jules sessions consume quota on Google's side — use `--list` and `session cancel` to avoid runaway dispatches.
3. The bundled `automate-github-issues` skill works best on repos with well-labeled, well-scoped issues; it is less useful on free-form planning tasks.
4. Don't confuse with `agents-cli` (the Agent Development Kit toolchain) — different product, different skill bundle (`use-google-agents-cli`).

## Documentation

- [Jules home](https://jules.google)
- [Jules Tools CLI reference](https://jules.google/docs/cli/reference/)
- [`google-labs-code/jules-skills`](https://github.com/google-labs-code/jules-skills)
- [Jules extension for Gemini CLI](https://developers.googleblog.com/en/introducing-the-jules-extension-for-gemini-cli/)
