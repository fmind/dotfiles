---
name: use-jules-cli
description: Guide for using the jules CLI (Jules Tools) to dispatch and inspect async coding sessions on Google Jules.
---

# Use Jules CLI (Jules Tools)

`jules` (the `@google/jules` npm package, also called **Jules Tools**) is the CLI for Google's autonomous async coding agent. The agent runs on a cloud VM, plans + applies changes, runs tests, and pushes results to a remote branch you can pull locally.

The CLI surface is intentionally small — `version`, `completion`, `help`, and `remote` (with `list`, `new`, `pull` subcommands). Run `jules help` to verify the live command list before relying on remembered syntax. Earlier docs that referenced `jules mcp`, `jules session ...`, or `jules auth login` are stale — Jules MCP integration is configured through the **Jules web Settings UI**, not the CLI.

## Install / Auth

```bash
# Install (already provided via mise as `npm:@google/jules`).
npm install -g @google/jules

# First invocation triggers an OAuth browser flow; or set an API key.
export JULES_API_KEY=...

# Confirm the binary works.
jules version
jules help
```

## Dispatch a Remote Session

```bash
# Kick off an async session against the current repo (cwd is auto-inferred).
jules remote new --session "Add tests for the parser module"

# Pin a specific repo and run multiple parallel attempts of the same task.
jules remote new \
  --repo torvalds/linux \
  --session "Migrate API client to v2" \
  --parallel 3
```

Real flags on `jules remote new`: `--session "<prompt>"`, `--repo <owner/name|.>`, `--parallel <N>`. The CLI does not accept `--prompt`, `--base`, or `--target` — branch targeting is decided by Jules itself when the session pushes its result.

## Inspect & Pull

```bash
# List active and recent sessions.
jules remote list
jules remote list --json | jq '.[] | {id, status, branch}'

# Pull the resulting branch into your local checkout.
jules remote pull <session-id>
```

## Shell Completion

```bash
jules completion bash > ~/.local/share/bash-completion/completions/jules
jules completion zsh  > "${fpath[1]}/_jules"
```

## Companion Skill Bundle

The official [`google-labs-code/jules-skills`](https://github.com/google-labs-code/jules-skills) bundle teaches the *local* coding agent how to dispatch Jules effectively. Install via `install-jules-skills`; see the bundle for the current set of skills.

## In-CLI Integration

Jules also ships a [Gemini CLI extension](https://developers.googleblog.com/en/introducing-the-jules-extension-for-gemini-cli/):

```bash
gemini extensions install jules
```

This gives you Jules dispatch from inside a Gemini CLI session without needing the bare `jules` command directly.

## Common Workflows

**Dispatch many issues in parallel.**

```bash
gh issue list --label triaged --json number,title --jq '.[]' | while read -r issue; do
  num=$(echo "$issue" | jq -r .number)
  title=$(echo "$issue" | jq -r .title)
  jules remote new --session "Fix issue #$num: $title"
done
```

**Pull and review.**

```bash
jules remote list --json | jq -r '.[] | select(.status=="completed") | .id' | \
  xargs -I {} jules remote pull {}
```

## Important Notes

1. **`jules mcp` does NOT exist** — Jules' MCP-style integrations (Linear, Stitch, Neon, Tinybird, Context7, Supabase) are configured via the Jules web Settings page, not the CLI.
2. **Jules sessions consume quota** on Google's side; avoid runaway dispatches in loops without a quota check.
3. **The CLI surface is small and may evolve** — run `jules help` to confirm the active commands; defer to the [CLI reference](https://jules.google/docs/cli/reference/) over remembered syntax.
4. **OAuth tokens are stored in `~/.config/jules/`** — protect this directory; use `JULES_API_KEY` for headless/CI use.

## Documentation

- [Jules home](https://jules.google)
- [Jules Tools CLI reference](https://jules.google/docs/cli/reference/)
- [Jules extension for Gemini CLI](https://developers.googleblog.com/en/introducing-the-jules-extension-for-gemini-cli/)
- [Jules MCP integrations (web Settings)](https://jules.google/docs/changelog/2026-02-02/)
- [`google-labs-code/jules-skills`](https://github.com/google-labs-code/jules-skills)
