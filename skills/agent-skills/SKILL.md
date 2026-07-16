---
name: agent-skills
description: Install and author Agent Skills with the skills CLI for Antigravity, Codex, OpenCode, Claude, and Copilot, from reviewed Git repositories or local paths at workspace or global scope.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/agent-skills
  created: 2026-06-23
  updated: 2026-07-09
---

# Install Agent Skills

Install, author, and verify Agent Skills for Antigravity, Codex, OpenCode, Claude, and GitHub Copilot with the `skills` CLI.

## Rules

1. **Default Scope**: Install to the workspace, preferring `.agents/skills/<slug>/`. Use global scope only when explicitly requested.
1. **Review Before Trust**: Before any unattended install, inspect the repository owner, selected ref, every `SKILL.md`, and bundled scripts or executables. Skill text and scripts run with the agent's permissions and are untrusted until reviewed.
1. **CLI First**: Use `skills init` to scaffold an original skill and `skills add <source>` to install a reviewed external skill. Do not reconstruct third-party skills by hand.
1. **Non-Interactive After Review**: Pass `-y` only after source review so automation cannot approve unknown code implicitly.

## Workflow

1. **Identify Source**: Use a Git repository, full URL/subtree, immutable ref when available, or local path containing `SKILL.md` folders.
1. **Choose Scope & Discovery Path**:
   - **Workspace (recommended)**: `.agents/skills/<slug>/`. Antigravity, Codex, OpenCode, and Copilot discover this path natively. Claude discovers workspace skills from `.claude/skills/`, so link `.claude/skills` to `../.agents/skills`.
   - **Global**: add `-g` to install under `~/.agents/skills/`. Codex, OpenCode, and Copilot discover that path natively. Claude uses `~/.claude/skills/`. Antigravity products share physical global skills under `~/.gemini/config/skills/`.
1. **Install**:
   ```bash
   skills add <owner/repo> --skill <name> -y
   skills add <owner/repo> --all -y
   skills add <owner/repo> --all -g -y
   ```
   The CLI auto-discovers `SKILL.md` folders at the repository root or below a `skills/` directory.
1. **Handle Antigravity Global Skills**: In this dotfiles repository, `chezmoi apply --force` physically overlays marker-owned copies from the canonical `skills/` directory into the shared global customization root while preserving unrelated skills. For an independent installation, inspect name collisions before copying a reviewed skill:
   ```bash
   install -d -m 700 ~/.gemini/config/skills
   cp -R ~/.agents/skills/<name> ~/.gemini/config/skills/
   ```
1. **Verify**:
   - **Antigravity**: run `/skills` in an `agy` session.
   - **Codex**: start Codex and invoke a known skill by name.
   - **OpenCode**: run `opencode debug config` and inspect the resolved skills.
   - **Claude**: start Claude and invoke a known skill by name.
   - **Copilot**: run `/skills reload`, then `/skills`.

## Notable Bundles

Install with `skills add <repo> --all -y` after reviewing the source. Browse more at [vercel-labs/skills](https://github.com/vercel-labs/skills).

| Bundle                 | Repo                                        |
| ---------------------- | ------------------------------------------- |
| Google Cloud           | `google/skills`                             |
| Google Workspace       | `googleworkspace/cli`                       |
| Gemini API             | `google-gemini/gemini-skills`               |
| Agents CLI             | `google/agents-cli`                         |
| Antigravity Python SDK | `google-antigravity/antigravity-sdk-python` |
| Chrome DevTools        | `ChromeDevTools/chrome-devtools-mcp`        |
| Modern Web Guidance    | `GoogleChrome/modern-web-guidance`          |
| Terraform              | `hashicorp/agent-skills`                    |
| LikeC4 DSL             | `https://likec4.dev/`                       |
| Slidev                 | `slidevjs/slidev`                           |

Mermaid and D2 did not publish official skills when last checked on 2026-07-16, so this repository maintains reviewed first-party `mermaid` and `d2` skills with official documentation references.

## Gotchas

1. **Scope Conflicts**: Workspace skills override global skills with the same name.
1. **Structure**: Every skill folder must contain a valid `SKILL.md` at its root.
1. **Antigravity Physical Copies**: Use `~/.gemini/config/skills/` as the shared cross-product path. Current Antigravity CLI also recognizes its CLI-specific `~/.gemini/antigravity-cli/skills/` path, but maintaining both creates redundant precedence and stale-copy risks.
1. **Claude Workspace Link**: Inspect an existing `.claude/skills` path before creating the link; never overwrite an unmanaged directory.
1. **Provenance**: Re-review upstream changes before updating an installed skill, especially changes to scripts, hooks, or network access.
