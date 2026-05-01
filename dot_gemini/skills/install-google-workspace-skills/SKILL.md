---
name: install-google-workspace-skills
description: Install Google Workspace (`gws` CLI) skills bundle on a per-project basis — Drive, Gmail, Calendar, Sheets, Docs, Chat, Tasks, plus helpers, personas, and recipes. Use for projects that integrate Workspace APIs.
---

# Install Google Workspace Skills

[`googleworkspace/cli`](https://github.com/googleworkspace/cli) ships the `gws` CLI **and** an in-repo bundle of 100+ Agent Skills covering every supported Workspace API plus higher-level helpers, personas, and recipes. This skill explains when and how to install only the slice a project actually needs.

## When to Trigger

- The project integrates one or more Google Workspace APIs (Drive, Gmail, Calendar, Sheets, Docs, Chat, Tasks, People, Slides, Forms, Classroom, Keep, Meet, Admin SDK, Apps Script).
- The user mentions automating a Workspace task, scripting `gws`, or wiring a workflow that touches mail / calendar / drive / sheets.
- Verify first: `ls .agents/skills/ ~/.gemini/skills/ 2>/dev/null | grep -E '^gws-|^persona-|^recipe-'`. If the relevant skills are already installed, skip.

## Install — Pick Skills, Don't Install the Whole Bundle

The bundle has **100+ skills**. Installing all of them floods the agent's progressive-disclosure surface — the prior `googleworkspace/cli` Gemini CLI extension was removed for exactly this reason. Always pick the slice that matches what the project actually does.

```bash
# 1. List available skills (group by category to choose what fits).
npx skills add googleworkspace/cli --list

# 2. Install one or a few skills, project scope (default — writes to .agents/skills/).
npx skills add googleworkspace/cli --skill gws-shared --skill gws-drive --project

# 3. Or install a single skill from its tree URL.
npx skills add https://github.com/googleworkspace/cli/tree/main/skills/gws-gmail --project
```

Use `--global` (writes to `~/.gemini/skills/`) only if the same skill is genuinely needed across many projects; in that case, `chezmoi add ~/.gemini/skills/<slug>` to track it.

`gws-shared` documents authentication and global flags — **install it alongside any other `gws-*` skill**, since the service skills assume its conventions.

## What's in the Bundle

Categories at the time of writing (verify with `npx skills add googleworkspace/cli --list` and the [official Skills Index](https://github.com/googleworkspace/cli/blob/main/docs/skills.md)):

### Services (one per API surface)

`gws-shared`, `gws-drive`, `gws-gmail`, `gws-calendar`, `gws-sheets`, `gws-docs`, `gws-slides`, `gws-tasks`, `gws-people`, `gws-chat`, `gws-classroom`, `gws-forms`, `gws-keep`, `gws-meet`, `gws-events`, `gws-admin-reports`, `gws-modelarmor`, `gws-script`, `gws-workflow`.

### Helpers (shortcut commands prefixed `+` in the CLI)

`gws-drive-upload`, `gws-sheets-append`, `gws-sheets-read`, `gws-gmail-send`, `gws-gmail-reply`, `gws-gmail-reply-all`, `gws-gmail-forward`, `gws-gmail-read`, `gws-gmail-triage`, `gws-gmail-watch`, `gws-calendar-insert`, `gws-calendar-agenda`, `gws-docs-write`, `gws-chat-send`, `gws-events-subscribe`, `gws-events-renew`, `gws-modelarmor-sanitize-prompt`, `gws-modelarmor-sanitize-response`, `gws-modelarmor-create-template`, `gws-script-push`, `gws-workflow-standup-report`, `gws-workflow-meeting-prep`, `gws-workflow-email-to-task`, `gws-workflow-weekly-digest`, `gws-workflow-file-announce`.

### Personas (role-based bundles — pull in the relevant services on demand)

`persona-exec-assistant`, `persona-project-manager`, `persona-team-lead`, `persona-hr-coordinator`, `persona-sales-ops`, `persona-it-admin`, `persona-customer-support`, `persona-event-coordinator`, `persona-content-creator`, `persona-researcher`.

### Recipes (~50 multi-step task sequences)

Cross-service playbooks like `recipe-draft-email-from-doc`, `recipe-organize-drive-folder`, `recipe-block-focus-time`, `recipe-create-doc-from-template`, `recipe-find-free-time`, `recipe-post-mortem-setup`, `recipe-watch-drive-changes`, `recipe-create-events-from-sheet`. Browse the index for the full list.

## Suggested Picks by Project Shape

- **Email automation** → `gws-shared`, `gws-gmail`, plus the relevant `gws-gmail-*` helpers and recipes (`recipe-label-and-archive-emails`, `recipe-create-vacation-responder`).
- **Calendar / scheduling** → `gws-shared`, `gws-calendar`, `gws-calendar-agenda`, `gws-calendar-insert`, recipes like `recipe-find-free-time`, `recipe-block-focus-time`.
- **Drive / Docs ops** → `gws-shared`, `gws-drive`, `gws-docs`, `gws-drive-upload`, recipes like `recipe-organize-drive-folder`, `recipe-create-doc-from-template`.
- **Spreadsheet ETL** → `gws-shared`, `gws-sheets`, `gws-sheets-read`, `gws-sheets-append`.
- **Standups / digests** → `gws-shared`, `gws-workflow`, `gws-workflow-standup-report`, `gws-workflow-weekly-digest`.
- **Apps Script projects** → `gws-shared`, `gws-script`, `gws-script-push` (pairs with the `clasp` workflow — see `use-clasp-cli`).
- **Workspace event subscriptions** → `gws-shared`, `gws-events`, `gws-events-subscribe`, `gws-events-renew`.

## Related: `gws` CLI

The skills assume `gws` is on `$PATH` and authenticated. See `use-google-workspace-cli` for the full CLI reference (auth flows, `--params` / `--json`, `--page-all`, schemas, exit codes). Quick sanity check:

```bash
gws auth setup       # one-time: configures Cloud project, enables APIs, logs in
gws auth login       # subsequent OAuth login
gws auth status
```

## After Install

1. Restart the agent so the new SKILL descriptions are picked up by progressive disclosure.
2. Project-scope installs (`.agents/skills/`) commit naturally with the repo. Global-scope installs are machine-local — `chezmoi add ~/.gemini/skills/<slug>` to track them.
3. Re-run `npx skills update` periodically to pull upstream improvements.

## Important Notes

1. **Do not install the full bundle.** That was the failure mode of the old `googleworkspace/cli` Gemini CLI extension. Always go through `--list` and `--skill` so the agent context stays focused.
2. `gws-shared` is a prerequisite — install it whenever any other `gws-*` skill is selected.
3. This skill installs SKILLs, **not** the CLI binary. Install `gws` separately (npm, Homebrew, Nix, or a release binary — see `use-google-workspace-cli`).
4. The `gws` extension for Gemini CLI (`gemini extensions install https://github.com/googleworkspace/cli`) bundles the same skills *and* registers `gws` as an MCP tool. Skip it on this dotfile setup — it pulls in too much. Prefer skills + CLI.
5. `npx skills` requires Node.js (already provided by mise).

## Documentation

- [`googleworkspace/cli` repo](https://github.com/googleworkspace/cli)
- [Official Skills Index](https://github.com/googleworkspace/cli/blob/main/docs/skills.md)
- [`npx skills` CLI](https://github.com/vercel-labs/skills) — see `use-skills-cli`
- [Google Workspace Developers home](https://developers.google.com/workspace)
