---
name: configure-gemini-cli-hooks
description: Guide for authoring Gemini CLI hooks — the 11 lifecycle events, settings.json schema, stdin/stdout JSON contract, exit codes, and patterns for blocking dangerous tools, auto-formatting, and session-start context injection.
---

# Configure Gemini CLI Hooks

Hooks are shell commands that Gemini CLI runs automatically at fixed points in a session. They are the canonical way to enforce repo policies (block secrets, auto-format, reject dangerous shell commands) without rewriting prompts.

Hooks live under the **`hooks`** key in `settings.json` (NOT `hooksConfig` — that key does not exist).

## Events

| Event | Fires when | Typical use |
|-------|------------|-------------|
| `SessionStart` | Session boots | Inject project context, load env |
| `SessionEnd`   | Session exits | Persist transcript, post-process logs |
| `BeforeAgent`  | Before each user-prompt turn | Validate / rewrite prompt |
| `AfterAgent`   | After the model finishes a turn | Append a status banner, kick off CI |
| `BeforeModel`  | Before the LLM request | Inject extra system text, short-circuit with synthetic response |
| `AfterModel`   | After each model response chunk | Cheap logging only — fires per chunk |
| `BeforeToolSelection` | Before the model picks a tool | Restrict the tool whitelist for this turn |
| `BeforeTool`   | Before a tool runs | **Block dangerous calls**, mutate args |
| `AfterTool`    | After a tool returns | Auto-format, append `additionalContext` |
| `Notification` | UI notification fires | Forward to Slack, log a metric |
| `PreCompress`  | Context window about to compress | Skip-or-augment compression |

### Lifecycle matchers (exact strings)

- `SessionStart`: `"startup"`, `"resume"`, `"clear"`
- `SessionEnd`:   `"exit"`, `"clear"`, `"logout"`, `"prompt_input_exit"`, `"other"`
- `PreCompress`:  `"auto"`, `"manual"`

### Tool matchers (regex)

`BeforeTool` / `AfterTool` match the regex against the **tool name** (e.g. `"write_file|replace"`, `"^run_shell.*"`, `"mcp_github_.*"`). Wildcard: `"*"` or `""`.

## settings.json Schema

```json
{
  "hooks": {
    "BeforeTool": [
      {
        "matcher": "write_file|replace",
        "sequential": false,
        "hooks": [
          {
            "name": "block-secrets",
            "type": "command",
            "command": "$GEMINI_PROJECT_DIR/.gemini/hooks/block-secrets.sh",
            "timeout": 5000,
            "description": "Reject writes containing secrets"
          }
        ]
      }
    ],
    "SessionStart": [
      {
        "matcher": "startup",
        "hooks": [
          { "type": "command", "command": "$GEMINI_PROJECT_DIR/.gemini/hooks/inject-context.sh" }
        ]
      }
    ]
  }
}
```

Per-entry fields:

| Field | Type | Notes |
|-------|------|-------|
| `type` | string | Only `"command"` is supported today |
| `command` | string | Shell string; env vars expanded |
| `name` | string | Optional friendly id used in logs |
| `timeout` | number | Milliseconds; default `60000` |
| `description` | string | Optional, surfaces in `/permissions` |

Group fields: `matcher` (string, see above) and `sequential` (boolean — run hooks in order vs. concurrent).

## Hook I/O Contract

### stdin (every event)

The CLI sends a single JSON object to the hook's stdin. Common base fields:

```json
{
  "session_id": "...",
  "transcript_path": "/path/to/session.jsonl",
  "cwd": "/home/me/repo",
  "hook_event_name": "BeforeTool",
  "timestamp": "2026-04-27T12:00:00Z"
}
```

Per-event extras (non-exhaustive — see docs for full schemas):

| Event | Extra fields |
|-------|--------------|
| `BeforeTool` | `tool_name`, `tool_input`, `mcp_context`, `original_request_name` |
| `AfterTool`  | + `tool_response { llmContent, returnDisplay, error }` |
| `BeforeAgent` | `prompt` |
| `AfterAgent` | `prompt`, `prompt_response`, `stop_hook_active` |
| `BeforeModel` / `AfterModel` | `llm_request` / `llm_response` |
| `SessionStart` | `source: "startup"|"resume"|"clear"` |
| `SessionEnd`   | `reason: "exit"|"clear"|"logout"|"prompt_input_exit"|"other"` |
| `Notification` | `notification_type`, `message`, `details` |
| `PreCompress`  | `trigger: "auto"|"manual"` |

### stdout (JSON only)

```jsonc
{
  // Optional flow control:
  "decision": "allow" | "deny" | "block",       // BeforeTool/BeforeAgent/BeforeModel
  "reason": "free-text rejection reason",       // shown to the model on deny
  "continue": true,                             // false = abort the whole session
  "stopReason": "...",                          // accompanies continue: false
  "systemMessage": "shown in the chat",         // any event
  "suppressOutput": false,                      // hide stdout from the user

  // Per-event payload:
  "hookSpecificOutput": {
    // BeforeTool — overrides / merges into tool_input:
    "tool_input": { "...": "..." },

    // AfterTool — appended to the tool result before the model sees it:
    "additionalContext": "Linter reformatted file.",
    "tailToolCallRequest": { "name": "write_file", "args": {} },

    // BeforeAgent / SessionStart — appended to the next prompt:
    "additionalContext": "Project status: green",

    // BeforeToolSelection — restricts tool choice:
    "toolConfig": {
      "mode": "AUTO" | "ANY" | "NONE",
      "allowedFunctionNames": ["read_file", "grep_search"]
    },

    // BeforeModel — short-circuit the LLM call:
    "llm_request": { /* override request */ },
    "llm_response": { /* synthetic response */ }
  }
}
```

`SessionEnd`, `Notification`, `PreCompress` only honor `systemMessage` — flow-control fields are ignored.

### Exit codes

| Code | Meaning |
|------|---------|
| `0`  | Success — stdout parsed as JSON; **prefer this** with `decision: "deny"` for intentional blocks |
| `2`  | System block — action aborted; **stderr is the rejection reason** shown to the model |
| any other | Warning logged; CLI proceeds with the original parameters (NOT a block) |

**stdout MUST be valid JSON.** A stray `echo`, `set -x`, or progress bar breaks parsing and silently degrades to "allow." Send all logs to stderr.

## Env Vars Passed to Hooks

`GEMINI_PROJECT_DIR`, `GEMINI_PLANS_DIR`, `GEMINI_SESSION_ID`, `GEMINI_CWD`, plus the `CLAUDE_PROJECT_DIR` alias for cross-tool scripts. Hooks inherit the rest of the parent shell env (including any API keys) — set `"environmentVariableRedaction": { "enabled": true }` if that's a concern.

## Working Examples

### 1. Block writes containing secrets (BeforeTool)

```bash
#!/usr/bin/env bash
# .gemini/hooks/block-secrets.sh
set -euo pipefail
input=$(cat)
content=$(echo "$input" | jq -r '.tool_input.content // .tool_input.new_string // ""')
if echo "$content" | grep -qiE 'api[_-]?key|password|secret|bearer [a-z0-9]{20,}'; then
  echo "blocked: potential secret in write" >&2
  cat <<'EOF'
{"decision":"deny","reason":"Potential secret detected. Move it to a secret manager and reference an env var.","systemMessage":"Security hook blocked the write."}
EOF
  exit 0
fi
echo '{"decision":"allow"}'
```

### 2. Auto-format after every code edit (AfterTool)

```bash
#!/usr/bin/env bash
# .gemini/hooks/format.sh
set -euo pipefail
input=$(cat)
path=$(echo "$input" | jq -r '.tool_input.path // .tool_input.file_path // empty')
[[ -z "$path" || ! -f "$path" ]] && { echo '{}'; exit 0; }
case "$path" in
  *.py)            ruff format "$path"       >/dev/null 2>&1 || true ;;
  *.ts|*.tsx|*.js) prettier --write "$path"  >/dev/null 2>&1 || true ;;
  *.go)            gofmt -w "$path"          >/dev/null 2>&1 || true ;;
esac
echo '{"hookSpecificOutput":{"additionalContext":"Formatter ran on '"$path"'"}}'
```

### 3. Inject project status at session start

```bash
#!/usr/bin/env bash
# .gemini/hooks/inject-context.sh
set -euo pipefail
ctx=$(printf 'Branch: %s\nUncommitted: %s files\nLast tag: %s\n' \
  "$(git branch --show-current)" \
  "$(git status --porcelain | wc -l)" \
  "$(git describe --tags --abbrev=0 2>/dev/null || echo 'none')")
jq -n --arg c "$ctx" '{hookSpecificOutput:{additionalContext:$c}, systemMessage:"Loaded git context"}'
```

### 4. Restrict tool choice on a sensitive turn (BeforeToolSelection)

```bash
#!/usr/bin/env bash
# Only let the model read during turns that mention "audit"
input=$(cat)
prompt=$(echo "$input" | jq -r '.prompt // empty')
if [[ "$prompt" == *audit* ]]; then
  cat <<'EOF'
{"hookSpecificOutput":{"toolConfig":{"mode":"AUTO","allowedFunctionNames":["read_file","read_many_files","glob","grep_search","list_directory"]}}}
EOF
else
  echo '{}'
fi
```

## Project Layout

Convention used in this repo (and recommended by `setup-gemini-cli-on-new-project`):

```text
.gemini/
├── settings.json
└── hooks/
    ├── block-secrets.sh
    ├── format.sh
    └── inject-context.sh
```

Mark scripts executable (`chmod +x`) and reference them via `$GEMINI_PROJECT_DIR/.gemini/hooks/<name>.sh` so the same `settings.json` works in any clone.

## Important Notes

1. **stdout must be valid JSON.** Any `echo "starting..."`, `set -x`, or prompt for input breaks the contract → CLI degrades to "allow." Logs go to stderr.
2. **`AfterModel` fires per chunk**, not per response. Keep it cheap or cache aggressively, otherwise streaming throughput drops.
3. **Project hooks are untrusted by default** — Gemini CLI re-prompts whenever the `command` string changes. Pin scripts in-repo so the prompt fires once.
4. **Default timeout is 60 000 ms.** Drop to ~5 000 for `BeforeTool` (fires per call) and bump for hooks that wait on network.
5. **Hooks inherit the full process env** (API keys included). Toggle `"environmentVariableRedaction": { "enabled": true }` if the hook script could leak them.
6. **`continue: false` aborts the whole session**, not just the turn. Use `decision: "deny"` to block a single tool call.
7. **No hook signing yet** — anything in `.gemini/hooks/` runs with user privileges. Review additions like any other script.

## Documentation

- [Hooks overview](https://geminicli.com/docs/hooks/)
- [Hooks reference (full payload schemas)](https://geminicli.com/docs/hooks/reference/)
- [Writing hooks](https://geminicli.com/docs/hooks/writing-hooks/)
- [Hook best practices](https://geminicli.com/docs/hooks/best-practices/)
- Companion skills: `configure-gemini-cli`, `setup-gemini-cli-on-new-project`, `use-gemini-cli`.
