---
name: fmind-visuals
description: Apply the Fmind visual identity and route slide or diagram work to Slidev, Mermaid, LikeC4, or D2. Use for Fmind talks, decks, article diagrams, architecture visuals, and www.fmind.dev assets.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/fmind-visuals
  created: 2026-07-16
  updated: 2026-07-16
---

# Fmind Visual Communication

Create calm, exact, spacious technical visuals that feel native to `www.fmind.dev`. Read `~/internals/publications/BRANDING.md` and `~/internals/publications/IDENTITY.md` when available; the bundled [Fmind theme reference](references/fmind-theme.md) is the portable fallback.

## Canonical Tool Choice

| Need                                                        | Tool                             | Boundary                                                  |
| ----------------------------------------------------------- | -------------------------------- | --------------------------------------------------------- |
| Slides, talks, workshops, LinkedIn documents                | [Slidev](../slidev/SKILL.md)     | Default for every new deck                                |
| Flow, sequence, state, class, ER, compact technical diagram | [Mermaid](../mermaid/SKILL.md)   | Default for every new diagram                             |
| Durable architecture model with multiple generated views    | [LikeC4](../likec4-dsl/SKILL.md) | Use when the model, not one image, is the source of truth |
| Existing D2 source or bespoke standalone composition        | [D2](../d2/SKILL.md)             | Specialist fallback                                       |

Do not create a custom HTML deck, Typst deck, PowerPoint source, or generated raster diagram unless the user explicitly requests that format or an existing project requires it.

## Brand Contract

- Heading font: Outfit Variable.
- Body font: Inter Variable.
- Background: `#0F172A`.
- Panel: `#1E293B`.
- Foreground: `#F8FAFC`.
- Muted: `#CBD5E1`.
- Primary accent: `#646CFF`.
- Border: `#334155`.
- Voice: calm, exact, pragmatic, technically grounded, and explicit about trade-offs.
- Use the canonical logo and font files from `~/internals/publications/assets/fmind/`; copy them into the deliverable rather than linking to a private local path.
- Use the Bleeding Agent palette only when the user explicitly asks for that sub-brand.

## Slidev Workflow

1. Read the official [Slidev skill](../slidev/SKILL.md), then use the current official docs for any unstable feature.
1. Scaffold with the project's package manager and keep Slidev, Vue, the default theme, and `playwright-chromium` as project-local dependencies for reproducibility.
1. Start from [package.json](references/package.json), [slides.md](references/slides.md), and [style.css](references/style.css), then copy the canonical logo and WOFF2 fonts into `public/brand/`.
1. Keep the starter's DOMPurify override until Monaco no longer pins a vulnerable release; verify any removal with `npm audit`.
1. Keep one claim, mechanism, decision, or artifact per slide. Split dense content instead of shrinking type.
1. Embed Mermaid directly for ordinary diagrams. Use exported LikeC4 or D2 SVGs only when their specialist boundary applies.
1. Run the deck, build it, and export the required review artifact:

   ```bash
   slidev slides.md
   slidev build slides.md
   slidev export slides.md
   ```

1. Prefer Slidev's browser exporter for review PNGs or PPTX; keep CLI PDF export for automation and install `playwright-chromium` locally.
1. Inspect browser, projector-sized, and exported views. Interactive success does not prove PDF or PNG correctness.

## Diagram Workflow

1. Start with [Mermaid](../mermaid/SKILL.md) and its portable Fmind frontmatter.
1. Keep editable source beside exports and near the prose or deck that owns the claim.
1. Prefer native Mermaid embedding in GitHub and Slidev. Export SVG only for destinations that cannot render Mermaid.
1. Use LikeC4 or D2 only when their model or composition advantages outweigh the loss of direct Markdown rendering.

## Definition of Done

- The visual has one clear thesis and no decorative nodes or slides.
- Source terminology, metrics, and boundaries are evidence-backed.
- Fmind colors, fonts, logo, spacing, and domain are consistent.
- Text remains legible on a laptop, projector, mobile preview, and exported page.
- Diagrams have a prose equivalent or alt text for accessibility.
- Editable sources and the requested exports both validate.

## Current References

- [Fmind website](https://www.fmind.dev/)
- [Slidev documentation](https://sli.dev)
- [Mermaid documentation](https://mermaid.js.org/)
- [LikeC4 documentation](https://likec4.dev/)
- [D2 documentation](https://d2lang.com/)
