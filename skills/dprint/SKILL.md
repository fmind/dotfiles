---
name: dprint
description: Canonical dprint setup — the standard formatter for config and markup files (JSON, Markdown, TOML, YAML). Use when configuring or running formatting for these file types.
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/dprint
  created: 2026-06-29
  updated: 2026-07-06
---

# dprint Formatting Standard

Canonical setup for **dprint**, the primary formatter for configuration and markup languages (JSON, Markdown, TOML, YAML). dprint formats only — linting lives in the language stacks ([go-stack](../go-stack/SKILL.md), [python-stack](../python-stack/SKILL.md)).

## 1. Configuration Strategy

dprint resolves config from the project directory upward and never auto-loads the home config, so every project needs its own resolvable `dprint.json`/`dprint.jsonc`. The home file (`~/.config/dprint/dprint.jsonc`, managed via chezmoi under `dot_config/dprint/dprint.jsonc`) is only a stored template. Seed a project one of two ways:

1. **Copy (default)**: Copy the global config into the project root. Self-contained, version-pinned, and offline; bump plugin versions per repo.
1. **Extends (DRY)**: Set `"extends"` to a single source of truth — a local path or a commit-pinned URL, e.g. `"extends": "https://raw.githubusercontent.com/fmind/dotfiles/<commit>/dot_config/dprint/dprint.jsonc"`. Projects inherit and can still override rules or add plugins.

## 2. Usage

- `dprint fmt` — format in place; the pre-commit hook calls the `format:dprint` task with `{staged_files}`.
- `dprint check` — verify formatting, non-zero exit on diff; wired into `check` as `check:format`.
- `dprint add <plugin>` — add and pin a new plugin version.

## 3. Gotchas & Guidelines

- **Precedence**: The order of plugins in the `plugins` array defines precedence. Ensure generic formatting plugins do not overshadow specialized ones.
- **Embedded Code Blocks**: The Markdown plugin formats fenced code blocks (e.g., JSON, TOML) only when both the Markdown plugin and the respective language plugins are loaded.

## Documentation

- [dprint Documentation](https://dprint.dev)
- [dprint Configuration Guide](https://dprint.dev/config/)
