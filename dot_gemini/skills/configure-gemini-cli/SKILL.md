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

## What you can configure

`settings.json` is a flat object of capability namespaces. Every key has a sensible default — the smallest valid file is `{}`. The exact key names move between releases, so treat the list below as a map of *behaviors* and look up current keys in the [configuration reference](https://geminicli.com/docs/reference/configuration/) when you actually edit the file.

Behaviors you can tune:

- **Model** — which Gemini model answers. Default to the latest generation available; check the [model reference](https://geminicli.com/docs/reference/configuration/) for current IDs. Override per run with `--model`.
- **Approval & safety** — pre-tool checkpointing, default approval mode (`auto_edit` / `plan` / `yolo` / ...), session turn caps, plan-mode artifact location.
- **Memory & context** — which memory files Gemini discovers (`AGENTS.md`, `GEMINI.md`), include/exclude globs, gitignore/geminiignore handling. See `configure-gemini-cli-memory`.
- **Tool surface** — which built-in tools are enabled, which are blocked, whether shell calls run through a PTY for interactive TUIs.
- **MCP servers** — local STDIO subprocesses and remote HTTP endpoints, each with their own tool allowlist. See structure below; use the matching `install-*-mcp` skill rather than hand-rolling entries.
- **Shell policy** — trust mode plus `allow` / `deny` glob lists matched before prompting.
- **Telemetry** — opt-in/out of usage reporting.
- **Environment** — values injected into the session env, available to MCP servers, tools, and hooks via `$VAR` substitution.
- **Experimental** — opt-in features that are still moving (e.g. `autoMemory`, `memoryV2`, `taskTracker`, `gemmaModelRouter`, `jitContext`); check the reference for what's currently behind this flag.

## MCP server shape

The transport shape is stable even as keys around it shift. Every entry under `mcpServers` is one of:

```jsonc
{
  // STDIO — local subprocess.
  "stdio_server": {
    "command": "npx",
    "args": ["-y", "some-mcp-package"],
    "env": { "FOO": "$FOO" },
    "includeTools": ["read", "list"]
  },

  // HTTP — remote MCP server.
  "http_server": {
    "httpUrl": "https://example.googleapis.com/mcp",
    "authProviderType": "google_credentials",
    "headers": { "X-Foo": "bar" },
    "timeout": 60000,
    "includeTools": []
  }
}
```

`includeTools` is a tool allowlist that keeps unused tool descriptions out of the context window — `[]` means "none initially". Inside a session, `/mcp desc <server>` lists a server's tools and `/mcp schema <server>` shows their input schemas.

## Env Var Substitution

Strings like `"$FOO"` in `env`, `headers`, or `args` are expanded from the shell environment at start-up. For default values, use `${FOO:-default}` syntax.

## Trust & Project Activation

The first time Gemini CLI sees `.gemini/settings.json` in a repo, it prompts to **trust** the workspace. Trusting a folder enables:

- Project-level config to override user config
- Project MCP servers to start
- Project tools / hooks to run

Accept the prompt at first launch, or pass `--skip-trust` for one-off CI runs. Inside a session, manage workspace permissions interactively with `/permissions`.

## Project vs User Scope (chezmoi)

In this dotfile setup:

- **User-level** (`~/.gemini/settings.json`) is rendered from the chezmoi source `dot_gemini/settings.json.tmpl` — global defaults that travel with the dotfiles.
- **Project-level** (`.gemini/settings.json`) lives inside each repo and commits with it. Use the `install-*-mcp` skills to drop product-specific MCP servers here.

## Inspect

From the shell:

```bash
gemini mcp list                    # all MCP servers + connection state
gemini extensions list             # see also: configure-gemini-cli-extensions
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

1. **Top-level `mcpServers` is camelCase** in `settings.json`. **Subagent frontmatter is the opposite — `mcp_servers` (snake_case);** the camelCase form is silently ignored there. See `create-gemini-cli-subagent`.
2. **Trust is per-folder, per-machine**; on a new machine after `chezmoi apply`, you'll trust each repo once.
3. **Secrets belong in env vars or a secret manager**, never in `settings.json`. Use `"$VAR"` substitution.
4. **Comments are not allowed** in `settings.json` (strict JSON). Move docs to a sibling `.gemini/README.md` if needed.

## Documentation

- [Gemini CLI configuration reference](https://geminicli.com/docs/reference/configuration/)
- [MCP server configuration](https://geminicli.com/docs/tools/mcp-server/)
- [Subagents](https://geminicli.com/docs/core/subagents/)
- [Custom commands](https://geminicli.com/docs/cli/custom-commands/)
- [Skills](https://geminicli.com/docs/cli/skills/)
- [Trust & permissions](https://geminicli.com/docs/cli/trusted-folders/)
