---
name: setup-google-workspace
description: Configure Google Workspace and Apps Script tooling — clasp, gws, and Workspace MCP agents.
---

# Setup Google Workspace

This skill provisions a complete Workspace developer environment from a fresh shell.

## Tools to verify / install

The repo's `~/.config/mise/config.toml` already pins:

- `npm:@google/clasp` — Apps Script local dev (`clasp`).
- `npm:@googleworkspace/cli` — Workspace developer CLI (`gws`).
- `npm:@google/jules` — async coding agent.
- `npm:@google/gemini-cli` — Gemini CLI itself.

Run `mise -C ~ install -y` (or `mr t` from this repo) to install them.

## Authentication checklist

| Tool / agent          | Auth mechanism                                   | Env var                       |
| --------------------- | ------------------------------------------------ | ----------------------------- |
| `clasp`               | OAuth (`clasp login`)                            | —                             |
| `gws`                 | OAuth (`gws auth login`)                         | —                             |
| `gcloud`              | `gcloud auth application-default login`          | sets ADC for `google_credentials` |
| `gemini` (default)    | OAuth-personal (`gemini` then sign in)           | `GEMINI_API_KEY` (optional)   |
| `jules`               | OAuth or API key                                 | `JULES_API_KEY`               |
| `gh` (GitHub)         | OAuth or PAT                                     | `GITHUB_MCP_PAT` for MCP      |
| `firebase`            | OAuth (`firebase login`)                         | —                             |

## Subagents to enable

Drop these subagent files (already present in this repo) into `~/.gemini/agents/`:

- `gws.md`, `clasp.md` — developer ops.
- `google-calendar.md`, `google-chat.md`, `google-drive.md`, `gmail.md`,
  `google-people.md` — Workspace MCP servers.
- `google-stitch.md`, `design-mcp.md` — design workflows.
- `google-pay-wallet.md` — payments and passes.
- `firebase.md`, `genkit.md` — serverless + AI.
- `gemini-code-assist.md`, `gemini-enterprise.md` — Gemini platform.

## Workflow

1. **Install tooling:** `mise -C ~/.local/share/chezmoi run tools`.
1. **Run ADC bootstrap:** `gcloud auth application-default login` so any
   Workspace / GCP MCP using `google_credentials` works without further
   prompts.
1. **Apps Script:** `clasp login` and `clasp create` for new projects.
1. **Workspace CLI:** `gws auth login`; `gws --help` for the command tree.
1. **Verify in Gemini CLI:** start `gemini`, run `/agents list`, and pick
   `@google-drive list my files` as a smoke test.

## Guidelines

- Workspace MCP endpoints (`*.googleapis.com/mcp/v1`) are **Developer Preview**;
  expect breaking changes and prefer read-only operations until GA.
- Never ship `.clasprc.json` or `~/.config/gws/credentials*` in chezmoi —
  they hold OAuth tokens.
- For shared scripts/projects, prefer service accounts with domain-wide
  delegation rather than personal OAuth.
