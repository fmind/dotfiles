---
name: create-claude-hook
description: Guide for adding lifecycle hooks to Claude Code via settings.json
---

# Add a Hook to Claude Code

Hooks are scripts Claude Code runs automatically at specific points in its workflow. Unlike CLAUDE.md instructions, which are advisory, hooks are deterministic — they always fire on their event. For more details, refer to the [official Claude Code hooks documentation](https://docs.claude.com/en/docs/claude-code/hooks-guide) and the [hooks reference](https://docs.claude.com/en/docs/claude-code/hooks).

## When to Use a Hook

Use a hook when an action **must** happen every time without asking — formatting after edits, blocking writes to forbidden paths, logging tool use, posting notifications when a session ends. For anything that needs Claude's reasoning, use a skill instead.

## Configuration Location

Hooks live under the `hooks` key in any `settings.json`:

| Scope | File | Tracked in git? |
|---|---|---|
| User | `~/.claude/settings.json` (this repo: `dot_claude/settings.json`) | Yes (dotfiles) |
| Project | `.claude/settings.json` | Yes |
| Local | `.claude/settings.local.json` | No |

Run `/hooks` inside Claude Code to see what's currently configured.

## Hook Events

| Event | Fires when |
|---|---|
| `SessionStart` | A new or resumed session begins |
| `UserPromptSubmit` | The user submits a prompt, before Claude processes it |
| `PreToolUse` | Before Claude executes a tool call (Bash, Edit, Read, MCP, ...) |
| `PostToolUse` | After a tool call completes (success or failure) |
| `Notification` | Claude shows a notification (e.g., needs input) |
| `Stop` | Claude finishes responding |
| `SubagentStop` | A subagent finishes |
| `PreCompact` | Before automatic context compaction |
| `SessionEnd` | The session terminates |

## Schema

```jsonc
{
  "hooks": {
    "<EventName>": [
      {
        "matcher": "Bash|Edit",     // optional; tool name regex/alternation, "" = all
        "hooks": [
          {
            "type": "command",
            "command": "/absolute/path/to/script.sh",
            "timeout": 30           // seconds, optional
          }
        ]
      }
    ]
  }
}
```

- **`matcher`** — for tool-level events (`PreToolUse`, `PostToolUse`), filters by tool name. Plain alternation (`"Bash|Edit"`) or regex (`"^Notebook"`, `"mcp__github__.*"`). Omit or use `""` to match everything.
- **`hooks[].type`** — currently `"command"` is the universally available form. Plugins may register additional handler types.
- **`hooks[].command`** — shell command. Receives a JSON event payload on stdin.
- **`hooks[].timeout`** — kill the script after N seconds.

## Stdin / Stdout Contract

The hook script reads a JSON object on stdin. The shape depends on the event:

```jsonc
// PreToolUse / PostToolUse
{
  "session_id": "...",
  "transcript_path": "/path/to/transcript.jsonl",
  "cwd": "/working/dir",
  "hook_event_name": "PreToolUse",
  "tool_name": "Bash",
  "tool_input": { "command": "rm -rf /" }
}
```

The script can return JSON on stdout to influence behavior:

```jsonc
{
  "decision": "block",            // "block" cancels the action; omit to continue
  "reason": "explanation shown to Claude",
  "additionalContext": "extra text appended to Claude's context"
}
```

### Exit codes

| Code | Meaning |
|---|---|
| `0` | Success. Stdout JSON (if any) is parsed. |
| `2` | Blocking error. Stderr is fed back to Claude; the action is cancelled. |
| Other | Non-blocking error. First line of stderr surfaces; execution continues. |

## Examples

### Format Python on save

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [
          { "type": "command", "command": "ruff format $CLAUDE_FILE_PATHS" }
        ]
      }
    ]
  }
}
```

### Block edits to `migrations/`

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [
          { "type": "command", "command": "/home/fmind/.claude/hooks/block-migrations.sh" }
        ]
      }
    ]
  }
}
```

`block-migrations.sh`:

```bash
#!/usr/bin/env bash
input=$(cat)
path=$(echo "$input" | jq -r '.tool_input.file_path // empty')
if [[ "$path" == *"/migrations/"* ]]; then
  echo "edits to migrations/ are forbidden — open a separate PR" >&2
  exit 2
fi
exit 0
```

### Persist env vars on session start

```json
{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          { "type": "command", "command": "echo 'export PYTHONDONTWRITEBYTECODE=1' >> \"$CLAUDE_ENV_FILE\"" }
        ]
      }
    ]
  }
}
```

## Step-by-Step Creation

1. **Pick the event.** `PreToolUse` for guardrails, `PostToolUse` for side effects (lint, format), `Stop`/`SessionEnd` for notifications.
2. **Decide the matcher.** Narrow it down — `Edit|Write` is much cheaper than running on every Bash invocation.
3. **Write the script.** Make it idempotent, fast (< 1s ideal), and exit non-zero on real errors only.
4. **Test in isolation.** Pipe a sample event JSON to the script: `echo '{"tool_input":{"file_path":"a.py"}}' | ./script.sh`.
5. **Wire it up** in `settings.json` under the chosen scope.
6. **Verify** with `/hooks` and by triggering the event in a session.

## Guidelines

- **Hooks are enforcement; CLAUDE.md is advice.** If a rule must hold every time, write a hook.
- **Keep them fast.** Hooks block Claude's loop while they run.
- **Use absolute paths** in the `command` field — the working directory may not match.
- **Never embed secrets** in hook scripts that live in this committed repo. Reference env vars instead.
- **Use exit code 2** sparingly — it cancels Claude's action, which is jarring if not warranted.
- **Prefer `PostToolUse` over `PreToolUse`** when possible — running formatters after the edit is cheaper than validating before.
