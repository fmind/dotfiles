---
name: agent-skills
description: Author, validate, publish, or install a SKILL.md package with the skills CLI and gh skill publish. Use for skill layout, frontmatter, resources, discovery.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/agent-skills
  created: "2026-06-23"
  updated: "2026-09-03"
---

# Agent Skills

Author, validate, publish, and install Agent Skill packages for Antigravity, Claude Code, Codex, Copilot, Grok, and OpenCode. [skillify](../skillify/SKILL.md) owns extracting a skill from a session; [skill-security-review](../skill-security-review/SKILL.md) owns the trust review of third-party packages.

## Author

1. **Scaffold**: `skills init <name>`; the directory name and the lowercase hyphenated frontmatter `name` stay identical, and the description is one sentence stating capability then trigger, never a workflow summary.
1. **Package**: keep `SKILL.md` focused; scripts, references, and templates live one level deep and are linked directly from `SKILL.md` per the [package rules](references/package-rules.md).
1. **Fit the catalog**: read neighbors with overlapping descriptions and link them instead of copying; a first-party skill also declares its CLI names in `skills/contracts.json`.

## Validate

1. **Isolate the candidate**: copy only the skill to `<candidate-root>/<slug>/SKILL.md` with its linked resources; the publisher takes the parent root, and catalog-relative sibling links do not survive a standalone install.
1. **Dry run** (proves Agent Skills compatibility, not safety):
   ```bash
   gh skill publish --dry-run <candidate-root>
   ```
1. **Catalog gate**: a first-party skill also passes `mise run check:skills` and `mise run test` (contracts, links, routing probes).
1. **Host discovery**: confirm the skill and its full description survive each host's metadata budget with the read-only listing commands in [host discovery](references/host-discovery.md); a listed skill is not proof it is followed.

## Publish

1. **Authority**: `gh skill publish` adds a topic, a tag, and a GitHub release, so obtain explicit publication authorization for the exact repository, candidate root, and version, then confirm the intended commit and rerun the dry run.
1. **Publish once**, stopping on any partial failure instead of retrying with another tag:
   ```bash
   gh skill publish <candidate-root> --tag <vX.Y.Z>
   ```
1. **Receipt**: record repository, commit, tag, and release URL; without the remote receipt the package is validated, not published.

## Install

1. **Review first**: resolve the source to an immutable ref, list what it ships, and review that snapshot; route packages with scripts, hooks, MCP, or network access to [skill-security-review](../skill-security-review/SKILL.md).
   ```bash
   skills add <owner/repo> --list
   ```
1. **Install at project scope** (`.agents/skills/<slug>/`), passing `-y` only after review; `--all` implies `-y`:
   ```bash
   skills add <owner/repo> --skill <name> -y
   gh skill install <owner/repo> <name> --pin <tag|sha>
   ```
   `skills add` has no ref flag and follows the default branch; `gh skill install --pin` is the pinned alternative when the installed ref must equal the reviewed one.
1. **Global scope**: `-g` (or `gh skill install --scope user`) writes to `~/.agents/skills`, which is this catalog; use it only when asked.
1. **Verify**: `skills list` (`-g` for global), then compare the installed files with the reviewed snapshot and keep source, ref, and destination as the receipt.

## Gotchas

- **Claude link**: Claude reads `.claude/skills`; link it to `../.agents/skills` after checking nothing unmanaged is there (see [agent-project](../agent-project/SKILL.md)).
- **Antigravity paths**: `~/.gemini/config/skills` is the shared global path; the CLI-only `~/.gemini/antigravity-cli/skills` creates stale copies, so keep one.
- **Trust race**: never review one ref and install another; re-review scripts, hooks, and network access before updating an installed skill.

## Documentation

- [Agent Skills specification](https://agentskills.io/specification) · [gh skill manual](https://cli.github.com/manual/gh_skill) · [skills.sh](https://skills.sh) · vendor bundles per tool: [upstream index](references/upstream-index.md)
- Companion skills: [skillify](../skillify/SKILL.md) (extract from a session), [skill-security-review](../skill-security-review/SKILL.md) (trust review), [agent-project](../agent-project/SKILL.md) (repository layout).
