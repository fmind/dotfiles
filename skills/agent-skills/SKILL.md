---
name: agent-skills
description: Install Agent Skills with the `skills` CLI — from a Git repo or local path, into workspace or global scope for Antigravity, OpenCode, and Claude. Use when adding or updating skills.
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/agent-skills
  created: 2026-06-23
  updated: 2026-07-06
---

# Install Agent Skills

Install Agent Skills for Antigravity, OpenCode, and Claude with the `skills` CLI.

## Rules

1. **Default Scope**: Install to the workspace (preferring `.agents/skills/<slug>/` when possible) by default. Use global scope only when explicitly requested.
1. **CLI First**: Use `skills add <owner/repo>` (or `npx skills add ...`) to install. Do not hand-write skills unless explicitly asked.
1. **Non-Interactive**: Always pass `-y` so autonomous runs never block on a prompt.

## Workflow

1. **Identify Source**: A GitHub repo (`<owner/repo>`), a full URL/subtree, or a local path containing `SKILL.md` folders.
1. **Choose Scope & Paths**:
   - **Workspace (default - recommended)**: `.agents/skills/<slug>/`. This is the preferred workspace-relative path. Note that while Antigravity and OpenCode can resolve this path directly, **Claude Code** only discovers workspace-level skills in `.claude/skills/`. Therefore, you must create a symlink `.claude/skills` pointing to `../.agents/skills` for Claude Code to discover them.
   - **Global**: add `-g`; the CLI installs to the canonical `~/.agents/skills/` and automatically links it to each agent's global directory:
     - **Antigravity**: `~/.gemini/antigravity-cli/skills/`
     - **OpenCode**: `~/.config/opencode/skills/`
     - **Claude**: `~/.claude/skills/`
1. **Install**:
   ```bash
   skills add <owner/repo> --skill <name> -y   # one skill
   skills add <owner/repo> --all -y            # whole bundle
   skills add <owner/repo> --all -g -y         # global
   ```
   The CLI auto-discovers `SKILL.md` folders (repo root or a `skills/` subdirectory); there is no `--path` flag.
1. **Verify**:
   - **Antigravity**: Start an `agy` session to auto-discover.
   - **OpenCode**: Prompt _"List available skills"_ in the TUI.
   - **Claude**: Start a `claude` session to auto-discover.

## Notable Bundles

Install with `skills add <repo> --all -y` (browse more at [vercel-labs/skills](https://github.com/vercel-labs/skills)):

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
| LikeC4                 | `https://likec4.dev/`                       |

## Gotchas

1. **Scope Conflicts**: Workspace skills override global skills of the same name.
1. **Structure**: A skill folder must contain `SKILL.md` at its root with valid frontmatter.
1. **Antigravity + Symlinks**: Antigravity does not yet discover symlinked global skills (vercel-labs/skills#633). For global installs targeting Antigravity, add `--copy`, e.g. `skills add <owner/repo> --all -g -a antigravity --copy -y`.
1. **Claude workspace skills**: Because Claude Code only discovers workspace-level skills in `.claude/skills/`, a symlink must be created to link the standard `.agents/skills/` directory to `.claude/skills/` (i.e., `mkdir -p .claude && ln -s ../.agents/skills .claude/skills`).
