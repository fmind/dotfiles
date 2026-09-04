---
name: dprint
description: Canonical dprint setup, the standard formatter for JSON, Markdown, TOML, and YAML. Use when configuring or running formatting for these file types.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/dprint
  created: "2026-06-29"
  updated: "2026-09-03"
---

# dprint

The formatter for configuration and markup files (JSON, Markdown, TOML, YAML); dprint formats only, linting lives in the language stacks ([go-stack](../go-stack/SKILL.md), [python-stack](../python-stack/SKILL.md), [typescript-stack](../typescript-stack/SKILL.md)).

## Configuration

dprint searches the current directory upward for `dprint.json` or `dprint.jsonc` and falls back to the global config (`DPRINT_CONFIG_DIR`) only when nothing is found, so every project needs its own resolvable config or it silently inherits whatever the global one contains.

1. **Copy (default)**: copy a known-good `dprint.json` into the project root; it is self-contained and version-pinned. The first run downloads uncached plugins; later runs use dprint's local cache. Bump plugin versions per repository.
1. **Extends (DRY)**: set `"extends"` to a single source of truth, a local path or a commit-pinned URL such as `"https://raw.githubusercontent.com/fmind/dot/<commit>/dprint.json"`; override rules or add plugins locally.

## Commands

```bash
dprint fmt --allow-no-files   # format:dprint — the hook appends {staged_files}; the flag keeps an empty list from exiting 14
dprint check                  # check:format — non-zero exit on drift
dprint add <plugin>           # add and pin a plugin version
```

## Mise Tasks

```toml
[tasks."format:dprint"]
description = "Format JSON, Markdown, TOML, YAML (dprint)"
run = "dprint fmt --allow-no-files" # mise appends the hook's {staged_files}

[tasks."check:format"]
description = "Check config and markup formatting (dprint)"
run = "dprint check"
```

## Gotchas

- **Plugin references**: prefer the `npm:` form (`npm:@dprint/markdown@0.23.3`, `npm:dprint-plugin-yaml@0.6.0`) over `https://plugins.dprint.dev/...wasm` URLs; both resolve, and the npm form makes the current version one `npm view <plugin> version` away.
- **Plugin order is precedence**: the `plugins` array order decides which plugin claims a file; keep specialized plugins before generic ones.
- **Embedded code blocks**: the Markdown plugin formats fenced JSON, TOML, and YAML only when those plugins are loaded too.
- **Staged vs whole-tree**: `format:dprint` takes `{staged_files}` from the hook and restages fixes; `check:format` always runs on the whole tree.

## Documentation

- [dprint](https://dprint.dev) · [Configuration](https://dprint.dev/config/) · [CLI](https://dprint.dev/cli/)
- Companion skills: [mise](../mise/SKILL.md) (task vocabulary), [lefthook](../lefthook/SKILL.md) (the pre-commit hook that calls `format:dprint`).
