---
name: configure-gemini-cli
description: Guide for configuring Gemini CLI via .gemini/settings.json — model, context, tools, MCP servers, env vars, telemetry, and trust settings.
---

# Configure Gemini CLI

Gemini CLI reads layered configuration from:

1. **System** — `/etc/gemini-cli/settings.json` (admin / managed)
2. **User** — `~/.gemini/settings.json`
3. **Project** — `.gemini/settings.json` (workspace-local; commits with the repo)

Project settings override user settings, which override system settings. The same JSON schema applies to all three.

## Minimal Project Config

```json
{
  "model": "gemini-2.5-pro",
  "tools": {
    "core": ["read", "write", "edit", "shell", "search"]
  }
}
```

## Common Sections

```json
{
  "general": {
    "checkpointing": { "enabled": true },
    "defaultApprovalMode": "auto_edit",
    "plan": { "directory": ".gemini/plans" }
  },

  "experimental": {
    "autoMemory": true,
    "jitContext": true,
    "worktrees": true
  },

  "context": {
    "fileName": ["AGENTS.md", "GEMINI.md"],
    "include": ["src/**/*.ts", "README.md"],
    "exclude": ["node_modules/**", "dist/**"],
    "fileFiltering": {
      "respectGitIgnore": true,
      "respectGeminiIgnore": true
    }
  },

  "tools": {
    "core": ["read", "write", "edit", "shell", "search"],
    "exclude": ["bash_destructive"],
    "shell": {
      "enableInteractiveShell": true
    }
  },

  "mcpServers": {
    "firebase": {
      "command": "npx",
      "args": ["-y", "firebase-tools@latest", "mcp"],
      "includeTools": []
    },
    "bigquery": {
      "httpUrl": "https://bigquery.googleapis.com/mcp",
      "authProviderType": "google_credentials",
      "includeTools": []
    }
  },

  "shell": {
    "trust": "ask",
    "allow": ["git status", "git diff*", "ls", "cat *", "rg *"],
    "deny":  ["rm -rf *", "git push --force*"]
  },

  "telemetry": {
    "enabled": false
  },

  "env": {
    "GOOGLE_CLOUD_PROJECT": "my-project",
    "REGION": "us-central1"
  }
}
```

### `general.*`

| Key | Notes |
|-----|-------|
| `checkpointing.enabled` | Shadow-git snapshots before each tool call; `/restore` rolls back |
| `defaultApprovalMode` | `default` / `auto_edit` / `yolo` / `plan` — overridable per run with `--approval-mode` |
| `plan.directory` | Where `/plan` writes plan-mode artifacts; gitignore this path |
| `maxSessionTurns` | Hard cap; exit code `53` when hit. `-1` = unlimited |

### `experimental.*`

| Key | Notes |
|-----|-------|
| `autoMemory` | Mine recent transcripts, draft skills under `~/.gemini/skills/`; review with `/memory inbox` |
| `jitContext` | Lazy-load subdir `GEMINI.md` files only when a tool touches that path |
| `worktrees` | Enables `gemini -w <branch>` for isolated parallel sessions |

### `tools.shell.*`

`enableInteractiveShell: true` lets shell calls reach a PTY (curses TUIs, prompts) instead of being captured as a flat string. Pair with the `shell.trust` / `allow` / `deny` policy.

## MCP Server Reference

Each entry under `mcpServers` accepts one of these transports:

```jsonc
{
  // STDIO (local subprocess).
  "stdio_server": {
    "command": "npx",
    "args": ["-y", "some-mcp-package"],
    "env": { "FOO": "$FOO" },
    "includeTools": ["read", "list"]
  },

  // HTTP (remote MCP server).
  "http_server": {
    "httpUrl": "https://example.googleapis.com/mcp",
    "authProviderType": "google_credentials",
    "headers": { "X-Foo": "bar" },
    "timeout": 60000,
    "includeTools": []
  }
}
```

`includeTools` is a tool allowlist — pinning it keeps unused tool descriptions out of the context window. Inside a Gemini CLI session, use `/mcp desc <server>` to list a server's tools (and `/mcp schema <server>` for input schemas).

## Shell Permission Model

`shell.trust` is the default policy; `allow` / `deny` are glob lists of commands.

| Value | Behavior |
|-------|----------|
| `"yes"`  | Run any shell command without prompting (dangerous) |
| `"no"`   | Never run shell commands |
| `"ask"`  | Prompt per command, with auto-allow/deny matched first |

## Trust & Project Activation

The first time Gemini CLI sees `.gemini/settings.json` in a repo, it prompts to **trust** the workspace. Trusting a folder enables:

- Project-level config to override user config
- Project MCP servers to start
- Project tools / hooks to run

Accept the prompt at first launch, or pass `--skip-trust` for one-off CI runs. Inside a session, manage workspace permissions interactively with `/permissions`.

## Env Var Substitution

Strings like `"$FOO"` in `env`, `headers`, or `args` are expanded from the shell environment at start-up. For default values, use `${FOO:-default}` syntax.

## Project vs User Scope (chezmoi)

In this dotfile setup:

- **User-level** (`~/.gemini/settings.json`) is rendered from the chezmoi source `dot_gemini/settings.json.tmpl` — global defaults that travel with the dotfiles.
- **Project-level** (`.gemini/settings.json`) lives inside each repo and commits with it. Use the `install-*-mcp` skills to drop product-specific MCP servers here.

## Inspect

From the shell:

```bash
gemini mcp list                    # all MCP servers + connection state
gemini extensions list             # see also: configure-gemini-extensions
gemini skills list                 # skills available in this scope
```

From inside a Gemini CLI session:

```text
/mcp list                          # connection state for each server
/mcp desc <server>                 # tool descriptions (what counts toward context)
/mcp schema <server>               # input schemas for each tool
/permissions                       # workspace trust + shell allow/deny
/tools                             # core tools currently enabled
```

There is no `gemini config show|validate` subcommand — verify settings by reading `~/.gemini/settings.json` and the project `.gemini/settings.json` directly. JSON syntax can be checked with any tool, e.g. `python -m json.tool .gemini/settings.json`.

## Important Notes

1. **Top-level `mcpServers` is camelCase** in `settings.json`. **Subagent frontmatter is the opposite — `mcp_servers` (snake_case);** the camelCase form is silently ignored there. See `create-gemini-subagent`.
2. **`includeTools: []`** means "include none initially" — pin specific tools to prevent context bloat.
3. **Trust is per-folder, per-machine**; on a new machine after `chezmoi apply`, you'll trust each repo once.
4. **Secrets belong in env vars or a secret manager**, never in `settings.json`. Use `"$VAR"` substitution.
5. **Comments are not allowed** in `settings.json` (strict JSON). Move docs to a sibling `.gemini/README.md` if needed.

## Documentation

- [Gemini CLI configuration reference](https://geminicli.com/docs/reference/configuration/)
- [MCP server configuration](https://geminicli.com/docs/tools/mcp-server/)
- [Subagents](https://geminicli.com/docs/core/subagents/)
- [Custom commands](https://geminicli.com/docs/cli/custom-commands/)
- [Skills](https://geminicli.com/docs/cli/skills/)
- [Trust & permissions](https://geminicli.com/docs/cli/trusted-folders/)
