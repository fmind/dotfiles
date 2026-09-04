---
name: hugo
description: "Canonical Hugo static-site stack with the Hextra docs theme: Hugo Modules, mise tasks, dprint, lefthook, GitHub Pages deploy. Use for docs sites and static websites."
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/hugo
  created: "2026-08-07"
  updated: "2026-09-03"
---

# Hugo Site Standard

Canonical static sites with Hugo extended and the Hextra theme: documentation sites, project docs, and simple websites. Go applications stay on [go-stack](../go-stack/SKILL.md) (GOTH); this skill is for content, not apps.

## 1. Core Stack

- **Hugo extended**: required by Hextra; mise's `hugo-extended` registry entry ([mise.toml](references/mise.toml)) provides it, and `hugo version` must print `+extended`.
- **Theme via Hugo Modules**: `github.com/imfing/hextra` is imported as a module — no submodules, no vendored theme; the version is pinned in `go.mod`/`go.sum`, fetched by `hugo mod get`, upgraded by `hugo mod get -u`.
- **Hextra features** (all self-hosted, zero CDN references): FlexSearch, theme toggle, Mermaid diagrams, server-side KaTeX, callout/card/tab shortcodes, edit links, git-based last-modified dates.
- **Tasks and hooks**: [mise.toml](references/mise.toml) exposes the canonical vocabulary per [mise](../mise/SKILL.md); [lefthook.yml](references/lefthook.yml) wires pre-commit and pre-push per [lefthook](../lefthook/SKILL.md).
- **Quality gates**: `check:site` builds with `--panicOnWarning`; `refLinksErrorLevel: ERROR` in [hugo.yaml](references/hugo.yaml) makes broken refs build errors; `test` renders a root-URL copy and runs `lychee --offline` over it.

## 2. Site Scaffolding Workflow

1. **Information**: define site `Title`, `Base URL`, module `Import Path` (e.g. `github.com/username/repo`), and repository URL.
1. **Bootstrap**: copy [mise.toml](references/mise.toml), run `mise trust && mise install`, then `hugo new site . --format=yaml --force` (repo root or `docs/`) and `hugo mod init <import_path>`.
1. **Config files**:
   - [lefthook.yml](references/lefthook.yml); [hugo.yaml](references/hugo.yaml) with `<base_url>`, `<title>`, `<repo_url>` replaced (keep `locale` and the Goldmark `passthrough` block).
   - `dprint.json` per [dprint](../dprint/SKILL.md), `.gitignore` from [gitignore](references/gitignore), `LICENSE` per [project-license](../project-license/SKILL.md).
1. **Content**: `content/_index.md` from [index.md](references/index.md) (hero layout); `content/docs/_index.md` from [docs-index.md](templates/docs-index.md) and a first page from [docs-page.md](templates/docs-page.md).
1. **Validate**: `git init --initial-branch=main`, then `mise run install`, `mise run format`, `mise run check`, `mise run test`; before the first commit, `check:leaks` scans the working tree.
1. **Finish**: smoke-test with `mise run watch`, then `git add . && git commit -m "chore: initial site"`.
1. **Deploy**: copy [pages.yml](references/pages.yml) to `.github/workflows/pages.yml` and set the Pages source to "GitHub Actions"; the workflow reuses `mise run build` with the Pages base URL, and CI checks come from [github-actions](../github-actions/SKILL.md).

## 3. Content Authoring

- **Front matter**: `title`, `weight` (sidebar order), `next`/`prev`, `math: true` (per-page KaTeX), `draft: true` (excluded from production, included by `watch`).
- **Shortcodes**: `callout`, `cards`/`card`, `tabs`, `steps`, `filetree`, `details`, and `hextra/hero-*`; an unknown icon name is a build error, not a fallback.
- **Diagrams**: fenced `mermaid` blocks render natively — author them per [mermaid](../mermaid/SKILL.md).
- **Math**: `math: true` per page; inline `\(...\)`, display `$$...$$` or `\[...\]`; single `$...$` is deliberately off because it false-positives on dollar amounts.
- **Links**: plain relative Markdown links between pages; `test` (lychee) verifies them, and `refLinksErrorLevel` covers `ref`/`relref`.

## 4. Site Layout

```text
<site>/
├── .github/workflows/pages.yml   // GitHub Pages deploy (reuses mise run build)
├── content/
│   ├── _index.md                 // Landing page (hextra-home layout)
│   └── docs/
│       ├── _index.md             // Docs section root
│       └── getting-started.md    // First page (weight-ordered)
├── .gitignore
├── dprint.json
├── go.mod                        // Hugo module graph — pins the Hextra version
├── go.sum
├── hugo.yaml                     // Single-file site configuration
├── lefthook.yml
├── LICENSE
├── mise.toml
└── README.md
```

Hugo also scaffolds `archetypes/`, `assets/`, `data/`, `i18n/`, `layouts/`, `static/`, `themes/`; keep or delete the empty ones, and put custom layouts in `layouts/` (they override the theme).

## Gotchas

- **Extended vs standard**: if `hugo version` lacks `+extended`, use the `hugo-extended` mise entry, not `hugo`.
- **Fresh-repo git info**: `enableGitInfo` warns until the first commit, so `check:site` sets `HUGO_ENABLEGITINFO=false`; the production `build` keeps it on.
- **CI needs full history**: the Pages workflow checks out with `fetch-depth: 0`, or last-modified dates are silently wrong.
- **dprint reflows shortcodes**: multi-line `{{< cards >}} … {{< /cards >}}` blocks collapse onto one line on `format`; harmless, Hugo parses both.
- **Deprecations fail fast**: `--panicOnWarning` turns Hugo deprecation warnings (e.g. `languageCode` → `locale`) into build failures; fix the key, never relax the check.
- **Module cache**: `hugo mod get` resolves through Go's module proxy (`~/go/pkg/mod`); `hugo mod clean` clears a stale theme.
- **Link check scope**: `lychee --offline` validates internal links only; run `lychee 'tmp/test/**/*.html'` manually for an external audit, never in the default gates.

## Documentation

- [Hugo](https://gohugo.io/documentation/) · [Hextra](https://imfing.github.io/hextra/docs/) · [Hugo Modules](https://gohugo.io/hugo-modules/)
- Companion skills: [mise](../mise/SKILL.md), [lefthook](../lefthook/SKILL.md), [dprint](../dprint/SKILL.md) (tasks, hooks, formatting), [mermaid](../mermaid/SKILL.md) (diagrams), [github-actions](../github-actions/SKILL.md) (CI), [go-stack](../go-stack/SKILL.md) (Go apps).
