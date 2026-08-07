---
name: hugo
description: Canonical Hugo static-site stack with the Hextra docs theme — Hugo Modules, mise tasks, dprint, lefthook, and GitHub Pages deploy. Use for documentation sites, project docs, and static websites.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/hugo
  created: 2026-08-07
  updated: 2026-08-07
---

# Hugo Site Standard (Hugo 0.164+, Hextra)

Canonical guidelines for static sites with **Hugo extended** and the **Hextra** theme: documentation sites, project docs, and simple websites. Go applications stay on the [go-stack skill](../go-stack/SKILL.md) (GOTH); this skill is for content, not apps.

## 1. Core Stack

- **Hugo Extended**: Required by Hextra. Install via mise's `ubi` backend with `matching = "extended"` ([mise.toml](references/mise.toml)) — the `aqua`/`asdf` backends ship the **standard** build (no `+extended` in `hugo version`), which fails Hextra's `module.hugoVersion` check.
- **Theme via Hugo Modules**: Import `github.com/imfing/hextra` as a Hugo Module — no git submodules, no vendored theme. Modules ride on Go (provisioned by mise): the theme version is pinned in `go.mod`/`go.sum`, fetched by `hugo mod get`, upgraded by `hugo mod get -u`.
- **Hextra Feature Set** (all self-hosted, zero CDN references in the output): FlexSearch client-side search, dark/light/system theme toggle, Mermaid diagrams, **server-side KaTeX** (math renders at build time into fingerprinted, integrity-hashed assets), callout/card/tab shortcodes, per-page edit links, git-based last-modified dates.
- **Task Runner & Hooks**: `mise.toml` ([mise.toml](references/mise.toml)) exposes the canonical vocabulary per the [mise skill](../mise/SKILL.md) — `install`, `format` (dprint), `check` (dprint + gitleaks + strict build), `test` (root-URL build + lychee offline link check), `build` (`--gc --minify`), `watch` (live-reload server). `lefthook.yml` ([lefthook.yml](references/lefthook.yml)) wires pre-commit (format → leaks → check) and pre-push (test) per the [lefthook skill](../lefthook/SKILL.md).
- **Quality Gates**: `check:site` builds with `--panicOnWarning --printPathWarnings` so deprecations and template issues fail fast; `refLinksErrorLevel: ERROR` in [hugo.yaml](references/hugo.yaml) makes broken `ref`/`relref` links build errors; `test` renders a root-URL copy into `tmp/test` and runs `lychee --offline` over it to catch every dead internal link — checking a copy, not `public/`, because subpath deployments (GitHub project pages) emit `/<repo>/`-prefixed links that `--root-dir` cannot resolve.

## 2. Site Scaffolding Workflow

1. **Information**: Define site `Title`, `Base URL`, module `Import Path` (e.g. `github.com/username/repo`), and repository URL.
1. **Bootstrap**: Run `hugo new site . --format=yaml --force` (inside the repo root or a `docs/` subdirectory), then `hugo mod init <import_path>`. Requires Hugo from step 3's `mise.toml` — copy it first and run `mise trust && mise install`.
1. **Config Initialization**:
   - `mise.toml` ([mise.toml](references/mise.toml)) and `lefthook.yml` ([lefthook.yml](references/lefthook.yml)).
   - `hugo.yaml` ([hugo.yaml](references/hugo.yaml)) — replace `<base_url>`, `<title>`, `<repo_url>`; keep `locale` (`languageCode` is deprecated since v0.158) and the Goldmark `passthrough` block (KaTeX).
   - `dprint.json` (setup per the [dprint skill](../dprint/SKILL.md)), `.gitignore` ([gitignore](references/gitignore)), `LICENSE` per the [project-license skill](../project-license/SKILL.md).
1. **Scaffold Content**:
   - Landing page `content/_index.md` from [index.md](references/index.md) (Hextra hero layout).
   - Docs section `content/docs/_index.md` from [docs-index.md](references/docs-index.md) and a first page from [docs-page.md](references/docs-page.md) (shows callouts, Mermaid, and per-page math).
1. **Git & Validation**:
   - Run `git init --initial-branch=main`, then the verification sequence: `mise run install`, `format`, `check`, `test` (`check:leaks` prints a benign `no commits yet` on fresh repos).
   - Smoke-test locally with `mise run watch`, then commit: `git add . && git commit -m "chore: initial site"`.
1. **Deploy (GitHub Pages)**: Copy [pages.yml](references/pages.yml) to `.github/workflows/pages.yml` and set the repository's Pages source to "GitHub Actions". The workflow reuses `mise run build`, injecting the Pages base URL — local and CI builds stay identical. CI checks come from the [github-actions skill](../github-actions/SKILL.md).

## 3. Content Authoring

- **Front Matter**: `title`, `weight` (sidebar order), `next`/`prev` (page navigation), `math: true` (per-page KaTeX opt-in), `draft: true` (excluded from production builds, included by `watch`).
- **Shortcodes**: Hextra ships `callout`, `cards`/`card`, `tabs`, `steps`, `filetree`, `details`, and the `hextra/hero-*` family. Icon names must exist in Hextra's bundled set — an unknown icon (e.g. `rocket-launch`) is a **build error**, not a fallback.
- **Diagrams**: Fenced `` ```mermaid `` blocks render natively — author them per the [mermaid skill](../mermaid/SKILL.md).
- **Math**: Enable per page (`math: true`); write inline math as `\(...\)` and display math as `$$...$$` or `\[...\]`. Single `$...$` is deliberately not configured — it false-positives on dollar amounts.
- **Links**: Use plain relative Markdown links between pages; `test` (lychee) verifies them against the rendered site, and `refLinksErrorLevel: ERROR` covers `ref`/`relref` shortcode links at build time.

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

Hugo also scaffolds `archetypes/`, `assets/`, `data/`, `i18n/`, `layouts/`, `static/`, `themes/` — keep the empty ones or delete them; Hugo recreates directories on demand. Custom layouts/partials go in `layouts/` (they override the theme module).

## 5. Gotchas & Guidelines

- **Extended vs Standard**: `mise install` must yield `hugo vX.Y.Z…+extended` — if `+extended` is missing, the tool came from the wrong backend; use the `ubi:gohugoio/hugo` entry with `matching = "extended"`.
- **Fresh-Repo Git Info**: `enableGitInfo` warns (and `--panicOnWarning` panics) until the first commit exists. `check:site` therefore disables it via `HUGO_ENABLEGITINFO=false` — git dates are display metadata, not content under check. The production `build` keeps it on; its fresh-repo warning disappears after the initial commit.
- **CI Needs Full History**: `enableGitInfo` reads per-file commit dates — the Pages workflow checks out with `fetch-depth: 0`; a shallow clone silently yields wrong last-modified dates.
- **dprint vs Shortcodes**: dprint's Markdown formatter joins adjacent lines, collapsing multi-line shortcode blocks (`{{< cards >}} … {{< /cards >}}`) onto one line. Harmless — Hugo parses them either way — but expect the reflow on `format`.
- **Deprecations Fail Fast**: `--panicOnWarning` in `check:site` promotes Hugo deprecation warnings (e.g. `languageCode` → `locale`) to build failures — fix the config key, never relax the check.
- **Module Cache**: `hugo mod get` resolves through Go's module proxy and cache (`~/go/pkg/mod`); `hugo mod clean` clears the site's module cache if the theme ever looks stale.
- **Link Check Scope**: `lychee --offline` validates internal links only (fast, deterministic, CI-safe). For an occasional external-link audit, run `lychee 'tmp/test/**/*.html'` without `--offline` manually — do not put network checks in the default gates.

## Documentation

- [Hugo Documentation](https://gohugo.io/documentation/)
- [Hextra Documentation](https://imfing.github.io/hextra/docs/)
- [Hugo Modules](https://gohugo.io/hugo-modules/)
- Companion skills:
  - [mise](../mise/SKILL.md) / [lefthook](../lefthook/SKILL.md) / [dprint](../dprint/SKILL.md) — tasks, hooks, and formatting standards.
  - [mermaid](../mermaid/SKILL.md) — diagram authoring for content pages.
  - [github-actions](../github-actions/SKILL.md) — CI running the same `mise run` gates.
  - [go-stack](../go-stack/SKILL.md) — Go applications and GOTH web apps (not content sites).
