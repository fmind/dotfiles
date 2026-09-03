---
name: improve-docs
description: "Trim, restructure, and verify project documentation: docs sites, wikis, guides, changelogs, long AGENTS.md files. Use when docs are too long, outdated, or noisy."
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/improve-docs
  created: "2026-09-02"
  updated: "2026-09-03"
---

# Improve Documentation

Documentation is read far more than it is written; every stale page costs every reader. This pass keeps what answers a real question and removes the rest. `README.md` and `AGENTS.md` alone are covered by [readme-agents](../readme-agents/SKILL.md); this skill handles the wider set.

## Workflow

1. **Inventory**: list every doc source (`README.md`, `AGENTS.md`, `docs/`, wiki, `CHANGELOG.md`, inline `--help`) and its audience: users, contributors, or agents.
1. **Find drift**: compare each page against the code it describes: commands (`mise tasks`, `--help` output), paths, versions, screenshots. A claim the repo cannot back is removed or fixed, never left.
1. **Cut noise**: delete duplicated sections, marketing filler, and step-by-step narratives that the code already shows; merge pages that answer the same question.
1. **Move workflows out of prose**: repeated agent instructions in `AGENTS.md` become skills per [skillify](../skillify/SKILL.md); `AGENTS.md` keeps rules and layout only.
1. **Verify**: `lychee <files>` for links, `dprint check` for formatting, and the site build (`mise run build`) for a [hugo](../hugo/SKILL.md) site. Read the rendered page, not only the Markdown.
1. **Report**: pages removed, merged, or rewritten, with the reason, plus links or facts that could not be verified.

## Gotchas

- **Do not rewrite voice**: keep the author's tone; change facts and structure, not style.
- **Changelogs are generated**: `CHANGELOG.md` comes from git-cliff per the [release skill](../release/SKILL.md); fix commits, not the file.
- **External links drift**: prefer primary documentation URLs and record the date for facts that age.

## Documentation

- [lychee](https://lychee.cli.rs) · [dprint](https://dprint.dev)
- Companion skills: [readme-agents](../readme-agents/SKILL.md) (the two root files), [skillify](../skillify/SKILL.md) (extract workflows), [hugo](../hugo/SKILL.md) (docs sites), [reduce-complexity](../reduce-complexity/SKILL.md) (the same pass for code).
