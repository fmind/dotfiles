# Fmind Visual Theme

The canonical live sources are `~/internals/publications/BRANDING.md`, `~/internals/publications/IDENTITY.md`, and [www.fmind.dev](https://www.fmind.dev/). Use this bundled reference when those local files are unavailable.

## Tokens

| Token | Value | Use |
| --- | --- | --- |
| Heading | Outfit Variable | Titles and major labels |
| Body | Inter Variable | Body copy and diagram labels |
| Background | `#0F172A` | Canvas |
| Panel | `#1E293B` | Cards, nodes, code surfaces |
| Foreground | `#F8FAFC` | Primary text |
| Muted | `#CBD5E1` | Secondary text |
| Primary | `#646CFF` | Focus, active edges, key nodes |
| Border | `#334155` | Dividers and secondary edges |

Use generous space, crisp geometry, restrained indigo, and evidence-led labels. Avoid gradients, decorative illustration, generic AI imagery, and dense dashboards unless the content requires them.

## Portable Mermaid Frontmatter

Use Mermaid's `base` theme because it is the customizable theme. This form stays inside the Mermaid source and can be used in `.mmd`, GitHub fences, and Slidev fences when the target Mermaid version supports frontmatter.

```mermaid
---
config:
  fontFamily: "ui-sans-serif, system-ui, sans-serif"
  flowchart:
    diagramPadding: 16
  theme: base
  themeVariables:
    darkMode: true
    background: "#0F172A"
    primaryColor: "#1E293B"
    primaryTextColor: "#F8FAFC"
    primaryBorderColor: "#646CFF"
    secondaryColor: "#334155"
    secondaryTextColor: "#F8FAFC"
    secondaryBorderColor: "#646CFF"
    tertiaryColor: "#0F172A"
    tertiaryTextColor: "#CBD5E1"
    tertiaryBorderColor: "#334155"
    lineColor: "#CBD5E1"
    textColor: "#F8FAFC"
    noteBkgColor: "#1E293B"
    noteTextColor: "#F8FAFC"
    noteBorderColor: "#646CFF"
---
flowchart LR
  Evidence --> Decision --> Outcome
```

For a renderer that does not support Mermaid frontmatter, move the same configuration into its supported site-level Mermaid configuration instead of falling back to unthemed output.

Keep Mermaid on the root-level system sans stack even when the surrounding deck uses Inter. Diagram-specific `flowchart.htmlLabels` is deprecated; more importantly, setting the root `fontFamily` before layout keeps label measurements stable without changing Slidev's HTML-label layout mode.

## LikeC4

Define Fmind colors as named tokens inside the LikeC4 `specification` block, then use those tokens in styles. Do not scatter raw hex values across views.

## D2

Start from a built-in D2 theme and use `theme-overrides` or `dark-theme-overrides` under `vars.d2-config`. Use the four `--font-*` flags together with copied Inter TTF files when exact font rendering matters.
