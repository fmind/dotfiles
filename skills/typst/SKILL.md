---
name: typst
description: Create standalone reports, papers, letters, or CVs with Typst, typstyle, mise tasks, and live PDF preview.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/typst
  created: 2026-08-07
  updated: 2026-08-30
---

# Typst Document Standard (Typst 0.15+)

Canonical guidelines for standalone documents with **Typst** — the modern typesetting system replacing LaTeX: millisecond compiles, readable markup, real error messages. **Scope is documents only** (papers, reports, letters, one-pagers, CVs): slide decks stay on Slidev and diagrams on Mermaid per the [fmind-visuals skill](../fmind-visuals/SKILL.md); websites stay on the [hugo skill](../hugo/SKILL.md).

## 1. Core Stack

- **Compiler**: `typst` 0.15+ via mise — `typst compile` for PDFs, `typst watch` for a live-updating preview.
- **Formatter**: `typstyle` (checked in `check:format`, applied in `format`) plus dprint for the surrounding Markdown/TOML/YAML per the [dprint skill](../dprint/SKILL.md).
- **Task Runner & Hooks**: `mise.toml` ([mise.toml](references/mise.toml)) exposes the canonical vocabulary per the [mise skill](../mise/SKILL.md) — `build` (PDF into `dist/`, cached via `sources`/`outputs`), `check` (`check:doc` compile fail-fast + `check:format` + `check:leaks`), `test` (compile gate for pre-push), `watch` (live preview). `lefthook.yml` ([lefthook.yml](references/lefthook.yml)) wires the hooks per the [lefthook skill](../lefthook/SKILL.md).
- **Compile Is the Test**: a Typst document has one correctness gate — it compiles without errors or warnings; `check:doc` renders to a scratch PDF so broken references, missing fonts, and bad markup fail before commit.

## 2. Document Scaffolding Workflow

1. **Bootstrap**: create the project directory, copy `mise.toml` ([mise.toml](references/mise.toml)) and `lefthook.yml` ([lefthook.yml](references/lefthook.yml)), then run `mise trust && mise install`.
1. **Scaffold Sources**: start `main.typ` from [main.typ](references/main.typ) (metadata, page setup, outline, math, citation hooks); add `dprint.json` per the [dprint skill](../dprint/SKILL.md), `.gitignore` ([gitignore](references/gitignore)), and a `LICENSE` per the [project-license skill](../project-license/SKILL.md) when the document is a repository of its own.
1. **Validation**: `git init --initial-branch=main`, then `mise run install`, `format`, `check`, `build` — commit sources, never the `dist/` PDF (attach it to a release instead, per the [release skill](../release/SKILL.md)).

## 3. Authoring Standard

- **Packages**: import from Typst Universe with a pinned version — `#import "@preview/cetz:0.5.2"` — the compiler downloads and caches them; there is no lockfile, so the pin in the import _is_ the pin.
- **Templates over copy-paste**: recurring document styles (letterhead, report, CV) belong in a local `template.typ` imported by each document — one styling source, per the DRY principle.
- **Bibliography**: `#bibliography("refs.bib")` reads BibTeX directly, or Hayagriva YAML for new work; cite with `@key`.
- **Fonts**: Typst bundles Libertinus; for brand fonts, vendor the files under `fonts/` and compile with `--font-path fonts` (add the flag in [mise.toml](references/mise.toml)) so builds stay reproducible across machines — `typst fonts` lists what is visible.
- **Archival PDFs**: publications and long-lived documents compile with `--pdf-standard a-2b` (PDF/A) so they render identically decades from now.

## 4. Gotchas

- **Typst is 0.x**: minor releases can change layout or APIs; the project `mise.lock` pins the working version, so bump deliberately (recompile and eyeball the diff) via the [upgrade-tools skill](../upgrade-tools/SKILL.md).
- **typstyle formats, dprint does not**: `.typ` files are typstyle's job ([mise.toml](references/mise.toml) wires both); do not add a dprint plugin for them.
- **`datetime.today()` breaks reproducibility**: fine for drafts (as in [main.typ](references/main.typ)); replace with a literal date when the document is final, or the PDF changes on every rebuild.
- **Math mode is not LaTeX**: `$x^2$` syntax differs (spaces matter, `dif` for differentials); port formulas by meaning, not by string substitution.

## Documentation

- [Typst Documentation](https://typst.app/docs/) · [Typst Universe](https://typst.app/universe/) (packages) · [typstyle](https://typstyle-rs.github.io/typstyle/) · [Hayagriva](https://github.com/typst/hayagriva)
- Companion skills:
  - [fmind-visuals](../fmind-visuals/SKILL.md) — identity and routing (slides → Slidev, diagrams → Mermaid).
  - [mise](../mise/SKILL.md) — the task vocabulary this stack implements.
  - [release](../release/SKILL.md) — attach compiled PDFs to tagged releases.
