---
name: create-claude-hook
description: Guide for adding lifecycle hooks to Claude Code via settings.json
---

# Create Claude Code Hook

Hooks are user-defined shell commands that Claude Code executes at specific lifecycle events. They run with the user's full credentials, so treat them as code. For more details, refer to the [official Claude Code hooks documentation](https://docs.claude.com/en/docs/claude-code/hooks).

## Where hooks live

Hooks are configured under the `hooks` key in any of these settings files:

- **User settings (global):** `~/.claude/settings.json` (chezmoi source: `dot_claude/settings.json`).
- **Project settings (shared):** `.claude/settings.json`.
- **Project settings (local, gitignored):** `.claude/settings.local.json`.

## Lifecycle events

| Event | Fires when |
|---|---|
| `PreToolUse` | Before any tool is invoked. Can block the call. |
| `PostToolUse` | After a tool successfully returns. Receives the result via stdin. |
| `UserPromptSubmit` | Before Claude processes a user message. Can rewrite or block it. |
| `Notification` | When the agent surfaces a notification (e.g., needs approval). |
| `Stop` | When the agent finishes its turn. |
| `SubagentStop` | When a subagent finishes its turn. |
| `PreCompact` | Before context compaction runs. |
| `SessionStart` | When a session starts. |
| `SessionEnd` | When a session ends. |

Each event maps to an array of matchers. Each matcher has a `matcher` regex (against tool name or event-specific value) and a `hooks` array of commands to run.

## File Structure

```jsonc
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Edit|Write|MultiEdit",
        "hooks": [
          {
            "type": "command",
            "command": "biome format --write \"$CLAUDE_FILE_PATH\" 2>/dev/null || true"
          }
        ]
      }
    ]
  }
}
```

### Important environment variables

Hooks receive context via env vars and a JSON payload on stdin:

- `CLAUDE_FILE_PATH` — path of the file being edited (PostToolUse with edit tools).
- `CLAUDE_TOOL_NAME` — name of the tool that fired the hook.
- `CLAUDE_PROJECT_DIR` — root of the current workspace.
- stdin — JSON describing the event; parse with `jq` if needed.

A non-zero exit code from a `PreToolUse` hook **blocks** the tool call. Return 0 to allow.

## Step-by-Step Creation

1. **Pick the event** that fires at the right moment. Format-on-edit usually wants `PostToolUse` with matcher `Edit|Write|MultiEdit`.
2. **Decide the scope:** global behavior → user settings; team behavior → project settings; personal-on-this-repo → project local settings.
3. **Author the command** as a one-liner or a script under the project. Keep it idempotent and fast — hooks block the agent until they finish.
4. **Always exit 0** from PostToolUse hooks unless you intend to surface failure. Use `|| true` to swallow tool-not-installed errors.
5. **Test the hook** by triggering its event and inspecting stderr; Claude Code logs hook output to the conversation when it fails.

## Safe Example: format Markdown after an edit

```jsonc
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Edit|Write|MultiEdit",
        "hooks": [
          {
            "type": "command",
            "command": "case \"$CLAUDE_FILE_PATH\" in *.md) markdownlint --fix \"$CLAUDE_FILE_PATH\" 2>/dev/null || true ;; esac"
          }
        ]
      }
    ]
  }
}
```

## Guidelines

- **Treat hooks as code.** They run unattended with full user permissions.
- **Stay fast.** Anything over a second or two will feel slow in the loop.
- **Never block silently.** If a `PreToolUse` hook rejects a call, write a clear reason to stderr.
- **Prefer existing tooling** (`biome`, `ruff format`, `markdownlint`) over inline shell logic.
