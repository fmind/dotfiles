---
name: d2
description: Create, theme, validate, and export D2 diagrams when an existing D2 source or a bespoke standalone composition is a better fit than the default Mermaid workflow.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/d2
  created: 2026-07-16
  updated: 2026-07-16
---

# D2 Diagram Standard

D2 remains an installed specialist tool. Mermaid is the default for portable GitHub and Slidev diagrams; use D2 when its composition, styling, boards, or layout control materially improves the result.

## Use D2 When

- The repository already owns `.d2` sources.
- The diagram needs bespoke composition, nested containers, multiple boards, scenarios, layers, or steps.
- A standalone SVG, PNG, PDF, PPTX, GIF, or ASCII export is the primary deliverable.
- Mermaid cannot express the visual clearly without renderer-specific workarounds.

Use [Mermaid](../mermaid/SKILL.md) for the ordinary case and [LikeC4](../likec4-dsl/SKILL.md) for a durable multi-view architecture model.

## Workflow

1. State the visual thesis, target medium, and expected reading size.
1. Write the smallest readable `.d2` source. Prefer containers, explicit directions, and short labels over manual coordinates.
1. Start with the default `dagre` layout. Try `elk` only when it produces a materially clearer topology; inspect supported engines with `d2 layout`.
1. Use a built-in theme as the base and override its tokens only for brand compliance. Discover current theme IDs with `d2 themes`.
1. Apply the Fmind palette and font guidance from [fmind-theme](../fmind-visuals/references/fmind-theme.md) when appropriate.
1. Format, validate, and render:

   ```bash
   d2 fmt diagram.d2
   d2 validate diagram.d2
   d2 diagram.d2 diagram.svg
   ```

1. Trust the command exit status, not only the presence of an output file; D2 can leave a partial render after an error.
1. Inspect the real export for clipping, crossings, font fallback, contrast, appendix behavior, and whether links or tooltips survive the target embedding method.

## Output Rules

- Prefer SVG for web pages and documentation. Use PNG only for raster-only destinations.
- Supply the four `--font-*` flags together when using custom TTF fonts so missing weights do not silently fall back.
- PNG and PDF exports may require Playwright/Chromium; keep browser dependencies project-local when reproducibility matters.
- If D2's bundled Playwright download is unavailable, keep the successful SVG export and rasterize it with the project's reviewed browser toolchain instead of weakening validation or pinning an obsolete browser.
- Keep `.d2` beside every generated artifact. Do not treat an exported image as the editable source.
- Use Mermaid instead when the source must render natively in GitHub or be embedded directly inside Slidev Markdown.

## Official References

- [D2 language tour](https://d2lang.com/tour/intro/)
- [D2 CLI manual](https://d2lang.com/tour/man/)
- [D2 themes and overrides](https://d2lang.com/tour/themes/)
- [D2 layout engines](https://d2lang.com/tour/layouts/)
- [D2 fonts](https://d2lang.com/tour/fonts/)
- [D2 export formats](https://d2lang.com/tour/exports/)
