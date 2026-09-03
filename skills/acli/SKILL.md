---
name: acli
description: Use the Atlassian acli for Jira work items, sprints, boards, and Confluence pages with OAuth or token auth, JQL-bounded reads, and confirmed writes. Use for Jira or Confluence work.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/acli
  created: "2026-09-02"
  updated: "2026-09-03"
---

# Atlassian CLI

Use `acli` for Jira and Confluence Cloud from the shell. Atlassian publishes no skill for `acli` itself, so this skill owns authentication, bounded reads, and the write boundary.

## Workflow

1. **Resolve the account**: `acli auth status`, then `acli auth login` for OAuth or an API token piped on stdin, never as an argument; `acli auth switch` selects the site when several are configured.

   ```bash
   acli jira auth login --site <site>.atlassian.net --email <email> --token < token.txt
   ```

1. **Bounded reads**: JQL with a limit and explicit fields, JSON for anything a tool parses.

   ```bash
   acli jira workitem search --jql 'project = TEAM AND status != Done' --fields key,summary,status --limit 50 --json
   acli jira workitem view TEAM-123 --fields summary,status,comment --json
   acli confluence page view --id <page-id> --body-format storage --json
   ```

1. **Write with authority**: creation, edits, transitions, comments, assignments, and bulk operations need explicit confirmation of the exact keys and fields; prefer `--generate-json` then `--from-json` for reproducible creations.

   ```bash
   acli jira workitem create --project TEAM --type Task --summary "<summary>" --json
   acli jira workitem transition --key TEAM-123 --status "In Progress"
   acli jira workitem comment create --key TEAM-123 --body "<text>"
   ```

1. **Verify by reading back**: `view --json` after every mutation and compare the requested fields.

## Gotchas

- **Bulk flags**: `--jql` and `--filter` on `transition`, `edit`, or `comment` act on every match; on `transition` and `edit`, `--yes` skips the prompt, not the authority rule.
- **Rovo Dev**: `acli rovodev` is Atlassian's coding agent, not configured here; it reads `.agents/skills` and `~/.agents/skills`, so this catalog is available there without copies.
- **MCP**: `atlassian/atlassian-mcp-server` publishes skills for the Rovo MCP server (`skills add atlassian/atlassian-mcp-server --list`), configured per [agent-mcp](../agent-mcp/SKILL.md); none covers `acli`.

## Documentation

- [acli reference](https://developer.atlassian.com/cloud/acli/reference/commands/)
- Companion skills: [github-issues](../github-issues/SKILL.md) (same authority rules for GitHub), [gws](../gws/SKILL.md) (Google Workspace), [agent-mcp](../agent-mcp/SKILL.md).
