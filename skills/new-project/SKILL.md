---
name: new-project
description: Bootstrap a new repository by composing the stack, license, mise, lefthook, dprint, CI, agent files, GitHub settings, and first release skills. Use when creating a project.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/new-project
  created: "2026-09-02"
  updated: "2026-09-03"
---

# New Project

Bootstrap a repository in a fixed order, one stack skill for the code and then the shared layer every project gets, so nothing is forgotten or duplicated; [project-health](../project-health/SKILL.md) owns the recurring refresh afterwards.

## Workflow

1. **Decide the basics**: slug (lowercase, hyphens), owner (`fmind`, `fmind-ai`, `mlops-courses`), visibility, purpose, and the parent directory from the global `AGENTS.md`; ask only when one is missing.
1. **Pick one stack skill** and follow its scaffolding; it already writes `mise.toml`, `lefthook.yml`, `.gitignore`, and a stack `AGENTS.md`:
   - Go library, CLI, web app, or ADK agent: [go-stack](../go-stack/SKILL.md)
   - Python package, CLI, web app: [python-stack](../python-stack/SKILL.md)
   - TypeScript frontend or full-stack website: [typescript-stack](../typescript-stack/SKILL.md), which composes [angular](../angular/SKILL.md) and adds [firebase](../firebase/SKILL.md) or [genkit](../genkit/SKILL.md) only when needed
   - Agent with the agents CLI: [google-adk](../google-adk/SKILL.md)
   - Docs site: [hugo](../hugo/SKILL.md); document: [typst](../typst/SKILL.md); infrastructure: [terraform-stack](../terraform-stack/SKILL.md)
1. **Add the shared layer**, skipping what the stack already produced:
   - `LICENSE` and manifest field: [project-license](../project-license/SKILL.md)
   - `dprint.json`: [dprint](../dprint/SKILL.md); hooks installed: [lefthook](../lefthook/SKILL.md)
   - `trivy.yaml` plus the `check:*` scan tasks: [secure](../secure/SKILL.md)
   - `.github/workflows/ci.yml` and `security.yml`: [github-actions](../github-actions/SKILL.md); `.github/dependabot.yml`: [dependabot](../dependabot/SKILL.md)
   - `AGENTS.md`, `.agents/skills/`, and the `CLAUDE.md` bridge: [agent-project](../agent-project/SKILL.md); `README.md`: [readme-agents](../readme-agents/SKILL.md)
1. **Validate locally**: `mise run install`, `mise run format`, `mise run check`, `mise run test`; before the first commit `check:leaks` scans zero commits and passes.
1. **Create the remote** after the first commit (`chore: initial commit`, see [conventional-commit](../conventional-commit/SKILL.md)), then apply [github-repository](../github-repository/SKILL.md):

   ```bash
   gh repo create <owner>/<slug> --<visibility> --source . --push
   ```

1. **Ship when useful**: a first `v0.1.0` through [release](../release/SKILL.md) once CI is green; a deploy target through [cloud-run](../cloud-run/SKILL.md) when the project serves traffic.
1. **Done when**:
   - `mise run check` and `mise run test` are green locally and CI passed on the first push.
   - `README.md` says what the project is and how to run it; `AGENTS.md` says how agents work in it.
   - No placeholder (`<slug>`, `TODO`) remains in committed files.

## Gotchas

- **Keep the stack `AGENTS.md`**: agent-project's generic template must not overwrite the one the stack skill wrote.
- **Private data**: never scaffold with real secrets; `.env.example` documents names only.

## Documentation

- [gh repo create manual](https://cli.github.com/manual/gh_repo_create) · [Agent Skills](https://agentskills.io)
- Companion skills: [project-health](../project-health/SKILL.md) (recurring refresh), [secure](../secure/SKILL.md) (security checklist), [mise](../mise/SKILL.md) (task vocabulary).
