---
name: dependabot
description: "Automate dependency updates with GitHub Dependabot: configuration, ecosystem mapping, grouping, validation. Use when enabling or fixing dependabot.yml."
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/dependabot
  created: "2026-07-14"
  updated: "2026-09-03"
---

# Dependabot

Keep GitHub Actions and dependencies current with one `.github/dependabot.yml`; [secure](../secure/SKILL.md) enables it during the security pass and [upgrade-tools](../upgrade-tools/SKILL.md) owns the bumps Dependabot cannot make (mise pins, formatter plugins).

## Workflow

1. **Map the ecosystems**: one `updates` entry per manifest directory, using the value Dependabot expects:

   | Manifest                         | `package-ecosystem` |
   | -------------------------------- | ------------------- |
   | `.github/workflows/*.yml`        | `github-actions`    |
   | `go.mod`                         | `gomod`             |
   | `pyproject.toml` + `uv.lock`     | `uv`                |
   | `package.json` (npm, pnpm, yarn) | `npm`               |
   | `Dockerfile`                     | `docker`            |
   | `*.tf` (Terraform)               | `terraform`         |
   | `*.tf` (OpenTofu)                | `opentofu`          |

1. **Write the config**: weekly schedule, `chore(deps)` commit prefix, and one group per ecosystem for `minor` and `patch` updates so majors arrive alone:

   ```yaml
   version: 2
   updates:
     - package-ecosystem: github-actions
       directory: /
       schedule:
         interval: weekly
         day: monday
       commit-message:
         prefix: "chore(deps)"
       groups:
         actions:
           patterns: ["*"]
           update-types: [minor, patch]
     - package-ecosystem: gomod
       directory: /
       schedule:
         interval: weekly
         day: monday
       commit-message:
         prefix: "chore(deps)"
       groups:
         go-modules:
           patterns: ["*"]
           update-types: [minor, patch]
   ```

1. **Validate**: `zizmor --offline --collect dependabot --strict-collection .` checks the schema and audits the file offline (see [zizmor](../zizmor/SKILL.md)); GitHub re-validates on push and shows errors in the Dependabot tab.
1. **Review each PR locally**: never merge blindly and never enable auto-merge.

   ```bash
   gh pr list --app dependabot    # open Dependabot PRs
   gh pr checkout <number>
   mise run check && mise run test
   ```

## Gotchas

- **Pin to major tags, not SHAs**: `actions/checkout@v7` tracks security patches within the major; SHA pins turn every upstream patch into review noise.
- **Directory is per manifest**: a Go module under `dot/` needs `directory: /dot`; Dependabot does not recurse from `/`.
- **No tokens needed**: Dependabot is native to GitHub and free for public and private repositories; the config file alone enables it.
- **No CLI trigger**: forcing an immediate check happens only in the repository's Dependabot tab (Insights, Dependency graph).

## Documentation

- [Dependabot](https://docs.github.com/en/code-security/dependabot) · [dependabot.yml options](https://docs.github.com/en/code-security/dependabot/dependabot-version-updates/configuration-options-for-the-dependabot.yml-file) · [Grouping updates](https://docs.github.com/en/code-security/tutorials/secure-your-dependencies/optimizing-pr-creation-version-updates)
- Companion skills: [secure](../secure/SKILL.md) (security pass), [upgrade-tools](../upgrade-tools/SKILL.md) (manual bumps), [zizmor](../zizmor/SKILL.md) (offline validation).
