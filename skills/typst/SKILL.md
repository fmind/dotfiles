---
name: typst
description: Create standalone reports, papers, letters, or CVs with Typst, typstyle, mise tasks, and live PDF preview. Use for any document that would otherwise be LaTeX or Word.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/typst
  created: "2026-08-07"
  updated: "2026-09-03"
---

# Typst Document Standard

Canonical standalone documents with Typst, the modern typesetting system replacing LaTeX: papers, reports, letters, one-pagers, CVs. Slide decks stay on Slidev and diagrams on Mermaid per [fmind-visuals](../fmind-visuals/SKILL.md); websites stay on [hugo](../hugo/SKILL.md).

## 1. Core Stack

- **Compiler**: `typst` via mise — `typst compile` for PDFs, `typst watch` for a live-updating preview.
- **Formatter**: `typstyle` for `.typ` files (checked in `check:format`, applied in `format:typst`) plus dprint for Markdown, TOML, and YAML per [dprint](../dprint/SKILL.md).
- **Tasks and hooks**: [mise.toml](references/mise.toml) exposes the canonical vocabulary per [mise](../mise/SKILL.md); [lefthook.yml](references/lefthook.yml) wires pre-commit and pre-push per [lefthook](../lefthook/SKILL.md).
- **Compile is the test**: `check:doc` renders a scratch PDF so broken references, missing fonts, and bad markup fail before commit; `test` reuses it as the pre-push gate.

## 2. Document Scaffolding Workflow

1. **Bootstrap**: create the project directory, copy [mise.toml](references/mise.toml) and [lefthook.yml](references/lefthook.yml), then `mise trust && mise install`.
1. **Sources**: start `main.typ` from [main.typ](references/main.typ) (metadata, page setup, outline, math, citation hooks).
1. **Config files**: `dprint.json` per [dprint](../dprint/SKILL.md), `.gitignore` from [gitignore](references/gitignore), and `LICENSE` per [project-license](../project-license/SKILL.md) when the document is its own repository.
1. **Validate**: `git init --initial-branch=main`, then `mise run install`, `mise run format`, `mise run check`, `mise run build`; commit sources, never the `dist/` PDF — attach it to a release per [release](../release/SKILL.md).

## 3. Authoring Standard

- **Packages**: import from Typst Universe with a pinned version (`#import "@preview/cetz:0.5.2"`); there is no lockfile, so the import pin is the pin.
- **Templates over copy-paste**: recurring styles (letterhead, report, CV) live in a local `template.typ` imported by each document.
- **Bibliography**: `#bibliography("refs.bib")` reads BibTeX directly, or Hayagriva YAML for new work; cite with `@key`.
- **Fonts**: Typst bundles Libertinus; vendor brand fonts under `fonts/` and compile with `--font-path fonts` (add the flag in [mise.toml](references/mise.toml)); `typst fonts` lists what is visible.
- **Archival PDFs**: publications compile with `--pdf-standard a-2b` (PDF/A) so they render identically decades from now.

## Gotchas

- **Typst is 0.x**: minor releases can change layout or APIs; `mise.lock` pins the working version, so bump deliberately (recompile and eyeball the diff) via [upgrade-tools](../upgrade-tools/SKILL.md).
- **typstyle formats, dprint does not**: `.typ` files are typstyle's job; do not add a dprint plugin for them.
- **`datetime.today()` breaks reproducibility**: fine for drafts (as in [main.typ](references/main.typ)); replace with a literal date when the document is final.
- **Math mode is not LaTeX**: `$x^2$` syntax differs (spaces matter, `dif` for differentials); port formulas by meaning, not by string substitution.

## Documentation

- [Typst](https://typst.app/docs/) · [Typst Universe](https://typst.app/universe/) · [typstyle](https://typstyle-rs.github.io/typstyle/) · [Hayagriva](https://github.com/typst/hayagriva)
- Companion skills: [fmind-visuals](../fmind-visuals/SKILL.md) (slides → Slidev, diagrams → Mermaid), [mise](../mise/SKILL.md) (task vocabulary), [release](../release/SKILL.md) (attach PDFs to releases).
