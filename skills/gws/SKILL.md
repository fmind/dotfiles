---
name: gws
description: "Automate Google Workspace with gws: schema-first calls, profile isolation, bounded reads, and verified Drive, Docs, Sheets, Gmail, Calendar, or Chat writes. Use for Workspace from the terminal."
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/gws
  created: "2026-08-30"
  updated: "2026-09-03"
---

# Google Workspace CLI

Use `gws` for Google Workspace automation from the shell: authentication, API discovery, bounded reads, and verified writes across Drive, Docs, Sheets, Gmail, Calendar, and Chat. A connected Workspace app (host connector or MCP server) owns in-document editing when one is available.

## Workflow

1. **Resolve the profile**: `gws auth status` confirms the account and scopes without printing credentials; set `GOOGLE_WORKSPACE_CLI_CONFIG_DIR` to isolate profiles.
1. **Inspect the live schema**: never guess resource names, parameters, or request bodies.

   ```bash
   gws schema <service.resource.method> --resolve-refs
   ```

1. **Start with a bounded read**: pass identifiers and filters through `--params`; use `--page-all --page-limit <n>` only when every page is needed.

   ```bash
   gws drive files list --params '{"pageSize": 10}' --format json
   ```

1. **Prepare writes exactly**: resolve stable file, spreadsheet, document, calendar, space, and message identifiers first; build the body as JSON and validate it against the schema before passing `--json`.
1. **Write with authority**: every write (create, edit, share, move, delete, send mail or Chat, change calendars, bulk update) needs explicit approval of the exact service, IDs or ranges, content, recipients, and sharing effects.
1. **Apply the smallest call**: no broad search followed by an unreviewed batch mutation; keep retries idempotent with stable request identifiers or a read-before-write guard where the API supports them.
1. **Verify by reading back**: fetch the changed resource and compare the requested fields; for messages and events, separate accepted API state from recipient-visible delivery.

## Gotchas

- **`gws auth export` prints decrypted credentials**: never run it during ordinary work.
- **`gws auth status` is a live call**: it may refresh OAuth credentials and query scope or API state; it is not an offline config inspection.
- **Exit code 2 is authentication**: repair it explicitly; on validation or discovery errors, refresh the schema before changing the request.
- **Apps Script**: `gws script` manages projects; `clasp` stays only for local push and pull of script sources.

## Official Skills

Upstream: `googleworkspace/cli` generates its skills from the CLI itself (a shared base skill that every service skill requires, plus one per API). List the current release, then install what the task needs at project scope after reviewing the snapshot (see [agent-skills](../agent-skills/SKILL.md)):

```bash
skills add googleworkspace/cli --list
skills add googleworkspace/cli --skill <name> -y
```

## Documentation

- [Google Workspace CLI](https://github.com/googleworkspace/cli) · [Workspace API reference](https://developers.google.com/workspace)
- Companion skills: [acli](../acli/SKILL.md) (same authority rules for Atlassian), [gcloud](../gcloud/SKILL.md) (Google Cloud), [agent-mcp](../agent-mcp/SKILL.md) (connected apps).
