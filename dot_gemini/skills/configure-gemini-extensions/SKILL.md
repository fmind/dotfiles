---
name: configure-gemini-extensions
description: Guide for installing, authoring, and managing Gemini CLI extensions — bundles of MCP servers, slash commands, agents, and skills installed via `gemini extensions`.
---

# Configure Gemini CLI Extensions

A **Gemini CLI extension** is a packaged bundle of MCP servers, slash commands, subagents, and/or skills that ship together. Once installed, the extension's contents are merged into the user's Gemini CLI config as if you'd added them by hand — but updates and uninstallation are handled by `gemini extensions`.

Use extensions for vendor packs (Stitch, Jules, Firebase, GitHub) and team-shared tooling. Use plain `.gemini/settings.json` + `~/.gemini/agents/` for personal setups in this dotfile repo.

## Manage Extensions

```bash
# Discover extensions — opens the extensions catalog in a browser.
gemini extensions explore

# Install from a Git repo or local path (interactive consent).
gemini extensions install <github-org>/<repo>
gemini extensions install ./path/to/extension

# List installed extensions (no remote-search subcommand).
gemini extensions list

# Update / remove.
gemini extensions update <id>
gemini extensions update --all
gemini extensions uninstall <id>

# Disable temporarily without uninstalling.
gemini extensions disable <id>
gemini extensions enable <id>

# Restart all extensions (reload after manual edits).
gemini extensions restart

# Per-extension settings.
gemini extensions config <id>
```

The full subcommand surface is: `config`, `disable`, `enable`, `explore`, `install`, `link`, `list`, `restart`, `uninstall`, `update`. There is no `search`, `info`, `init`, or `--available` flag.

## What an Extension Ships

Each extension is a directory with a `gemini-extension.json` manifest plus optional payload directories:

```
my-extension/
├── gemini-extension.json     # manifest (required)
├── commands/                 # slash commands (.toml)
├── agents/                   # subagent definitions (.md)
├── skills/                   # SKILL.md bundles
└── mcp/                      # MCP server configs (merged into mcpServers)
```

## Manifest (`gemini-extension.json`)

```json
{
  "name": "my-extension",
  "version": "0.1.0",
  "description": "My team's Gemini CLI extension",
  "author": { "name": "Me", "email": "me@example.com" },
  "license": "MIT",

  "mcpServers": {
    "my-mcp": {
      "command": "node",
      "args": ["server.js"],
      "includeTools": []
    }
  },

  "commands": ["commands/review.toml", "commands/git/commit.toml"],
  "agents":   ["agents/reviewer.md"],
  "skills":   ["skills/review-rules"],

  "tools": {
    "exclude": ["bash_destructive"]
  }
}
```

The manifest is **merged** into the active Gemini CLI configuration. Extensions can ship MCP servers without the user editing their `settings.json` directly.

## Authoring an Extension

There is no scaffold subcommand — create the directory layout above by hand (or copy a template from <https://github.com/gemini-cli-extensions>) and write `gemini-extension.json`.

```bash
# Symlink a working copy for live development (re-reads on session start).
gemini extensions link ./my-extension

# Or install a snapshot from the local path.
gemini extensions install ./my-extension

# Publish to GitHub — any repo with a manifest is installable by URL.
git push origin main
# Now anyone can:  gemini extensions install <user>/<repo>
```

`link` is the right choice while iterating: edits to the source directory take effect on the next session without re-installing. `install` copies into `~/.gemini/extensions/<id>/` and won't pick up source changes.

## Common Vendor Extensions

| Extension | What it ships |
|-----------|----------------|
| `stitch`     | Stitch MCP server + design-related slash commands |
| `jules`      | Jules dispatch UI from inside Gemini CLI sessions |
| `firebase`   | Firebase MCP + CLI shortcuts |
| `gemini-cli-extensions/*` | Templates and reference extensions on GitHub |

Browse: <https://github.com/gemini-cli-extensions>

## Scope

Extensions are installed globally to `~/.gemini/extensions/<id>/` — there is no per-repo `extensions.json` manifest. To pin per-project tooling, prefer `.gemini/settings.json` (project MCP servers) plus per-repo skills/agents under `.gemini/`. Document required extensions in the repo's README or contributor guide.

## Versus Skills, Subagents, MCP Servers

| Layer | When to use |
|-------|-------------|
| **MCP server** (in `settings.json`) | Single-product tool surface (e.g. BigQuery) |
| **Subagent** (`agents/foo.md`) | Tool-scoped persona that loads its MCP lazily |
| **Skill** (`skills/foo/SKILL.md`) | Procedural / contextual knowledge with progressive disclosure |
| **Extension** | A *bundle* of any of the above, distributed and updated together |

Use an extension when you need to ship **all four** as a coherent unit (e.g. a vendor pack). For personal customizations in this dotfile repo, prefer the unbundled forms.

## Important Notes

1. **Extensions can run code on install** — only install from sources you trust.
2. **Extension MCP servers are merged**, not overlaid — duplicate server names will clash with what's in `settings.json`.
3. **Updates follow semver** if the manifest declares it; pin to a tag for production.
4. **Linked extensions reload on session start; installed snapshots don't auto-update** — re-run `gemini extensions install ./my-extension` after edits if you used `install` instead of `link`.
5. **Permissions inherit from the active workspace trust** — installing an extension that ships an MCP server doesn't bypass the trust prompt.

## Documentation

- [Gemini CLI extensions overview](https://geminicli.com/docs/extensions/)
- [Extension manifest reference](https://geminicli.com/docs/extensions/reference/)
- [Gemini CLI extensions on GitHub](https://github.com/gemini-cli-extensions)
- [Gemini CLI configuration](https://geminicli.com/docs/reference/configuration/)
