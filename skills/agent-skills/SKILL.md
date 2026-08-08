---
name: agent-skills
description: Author, validate, publish, or install Agent Skill packages. Use for SKILL.md frontmatter, metadata, scripts, references, discovery paths, CLI checks, and reviewed installs; route third-party trust to skill-security-review.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/agent-skills
  created: 2026-06-23
  updated: 2026-08-08
---

# Agent Skills

Author, validate, publish, and install Agent Skill packages for Antigravity, Codex, OpenCode, Claude, and GitHub Copilot.

## Rules

1. **Default Scope**: Install to the workspace, preferring `.agents/skills/<slug>/`. Use global scope only when explicitly requested.
1. **Review Before Trust**: Before any unattended install, inspect the repository owner, selected immutable ref, every `SKILL.md`, and bundled scripts or executables. Use [skill-security-review](../skill-security-review/SKILL.md) for packages with hooks, MCP or plugin configuration, installers, network access, credential flows, obfuscation, or unclear provenance. Skill text and scripts run with the agent's permissions and are untrusted until reviewed.
1. **Same Snapshot**: Review and install the same immutable local snapshot. The current `skills add` CLI has no commit/ref flag, so installing a mutable `owner/repo` after reviewing another snapshot creates a trust race.
1. **CLI First**: Use `skills init <name>` to scaffold an original skill and `skills add <source>` to install a reviewed external skill. Do not reconstruct third-party skills by hand.
1. **Non-Interactive After Review**: Pass `-y` only after source review so automation cannot approve unknown code implicitly.
1. **Separate Authority**: Authoring, validation, installation, or a dry run does not grant publication authority. Require explicit publication authorization for the exact repository, package root, and version before creating a release.

## Author

1. **Scaffold**: Run `skills init <name>`, then keep the directory name and lowercase hyphenated frontmatter `name` identical. Write a concise `description` that states both capability and trigger; keep workflow detail out of discovery metadata.
1. **Define the Package**: Keep `SKILL.md` focused and move optional scripts, references, templates, and assets into directly linked one-level resource directories. Test helpers in a disposable directory, document required tools and side effects, use relative paths, and keep executable code minimal.
1. **Add Product Metadata**: Add `agents/openai.yaml` only when Codex interface metadata or dependencies are needed. Use the supported `interface`, `dependencies.tools`, and `policy` fields, mention `$skill-name` in `interface.default_prompt`, and link the file directly from `SKILL.md`.
1. **Fit the Catalog**: Check overlap with neighboring skills, route trust review to `skill-security-review`, and update `skills/contracts.json` plus routing fixtures when the package belongs to this first-party catalog.

## Validate

1. **Isolate a Publishable Package**: For standalone distribution, copy only the candidate to `<candidate-root>/<slug>/SKILL.md` with its directly linked resources. The publisher takes the parent candidate root; passing the skill directory itself makes its name appear as `.`. Catalog-relative sibling links can pass the monorepo contract while remaining unusable after independent installation.
1. **Run Publisher Compatibility**: Run `gh skill publish --dry-run <candidate-root>`. This checks upstream Agent Skills compatibility without publishing and does not replace repository-specific tests.
1. **Run Repository Contracts**: `mise run check:skills` owns publisher compatibility. `mise run test` owns the strict repository manifest, frontmatter and Codex interface schema, exact tools in `skills/contracts.json`, CommonMark and HTML5 links, direct one-level resource disclosure, first-party catalog containment, bounded parsed inputs, text and executable safety, cache exclusion, helper tests, progressive disclosure, routing fixtures, and offline safety smoke contracts. `mise run all` is the authoritative combined gate.
1. **Preserve Dirty Worktrees**: Before any full gate, inspect the full gate's task definition and working-tree state. If it runs whole-tree write-formatters and unrelated or user changes are present, validate the exact candidate in an isolated temporary worktree or run equivalent non-mutating checks; never reformat unrelated work.
1. **Interpret Routing Evidence Honestly**: Maintain the positive, neighbor-owned negative, and pressure scenarios in [lifecycle-evaluations.json](tests/lifecycle-evaluations.json), plus the independently seeded regression cases in [routing-boundaries.json](tests/routing-boundaries.json). The deterministic test ranks prompts against skill names and descriptions, enforces rank-one and multi-skill recall floors, detects lexical near-duplicates, and checks selected safety wording. It is a full-catalog lexical smoke test, not proof of host inclusion, truncated-description routing, abstention, semantic selection, pressure compliance, or model behavior; pressure expectations and no-route probes are schema-checked but not executed or semantically graded.
1. **Check Host Discovery**: Before claiming Codex discoverability, run the installed CLI's read-only `codex debug prompt-input` on a fresh session, record the CLI version, selected model and context window, confirm every expected skill and full description survived the host metadata budget, and retain any shortening or omission warning. Validate instruction following separately in a disposable, instrumented agent run. For other hosts, use `/skills` in Antigravity, `opencode debug config`, explicit invocation in Claude, and `/skills reload` then `/skills` in Copilot.

## Publish

1. **Confirm Authority and Target**: Obtain explicit publication authorization for the exact repository, package root, and semantic version. `gh skill publish` can change repository metadata, create a tag, and publish a GitHub release; validation or a request to author a skill does not authorize those mutations.
1. **Revalidate the Exact Candidate**: Confirm the intended commit and clean candidate, then rerun the isolated package dry run and repository gates. Without this evidence, report the package as unvalidated rather than source-ready.
1. **Publish Once**: With authority, run `gh skill publish <candidate-root> --tag <vX.Y.Z>` and stop on any partial failure; do not retry with a different target or tag implicitly.
1. **Publication receipt**: Record the repository, exact commit, tag, release URL, and command result. Claim `release-published` only when the remote receipt exists; otherwise report the highest proved state, such as `source-ready`, and name the missing external proof.

## Install

1. **Resolve and Review**: Resolve a repository or URL to an immutable commit, materialize it locally, and review that exact snapshot before execution. Use [skill-security-review](../skill-security-review/SKILL.md) when its risk triggers apply.
1. **Choose Scope & Discovery Path**:
   - **Workspace (recommended)**: `.agents/skills/<slug>/`. Antigravity, Codex, OpenCode, and Copilot discover this path natively. Claude discovers workspace skills from `.claude/skills/`, so link `.claude/skills` to `../.agents/skills`.
   - **Global**: add `-g` to install under `~/.agents/skills/`. Codex, OpenCode, and Copilot discover that path natively. Claude reaches it through the `~/.claude/skills/` symlink, Antigravity products through `~/.gemini/config/skills/` — both point back to `~/.agents/skills/`.
1. **Install the Reviewed Snapshot**: Install from the exact reviewed local snapshot; do not review one ref and then install a mutable default branch.
   ```bash
   skills add <exact-reviewed-local-snapshot> --skill <name> -y
   skills add <exact-reviewed-local-snapshot> --all -y
   skills add <exact-reviewed-local-snapshot> --all -g -y
   ```
   The CLI auto-discovers `SKILL.md` folders at the repository root or below a `skills/` directory.
1. **Verify the Install**: Inspect `skills list --json`, the resolved discovery path, and the installed package contents. Compare them with the reviewed snapshot and retain the source URL, commit, destination, and content hash or diff as the install receipt.
1. **Handle Antigravity Global Skills**: In this dotfiles repository nothing is copied — `chezmoi apply --force` renders `dot_gemini/private_config/symlink_skills.tmpl` into a symlink chain, so `~/.gemini/config/skills` → `~/.agents/skills` → the canonical `skills/` directory. Every agent CLI therefore reads one source of truth, and a skill edited in `skills/` is live immediately with no re-apply. For an independent installation outside this repo, inspect name collisions before copying a reviewed skill:
   ```bash
   install -d -m 700 ~/.gemini/config/skills
   cp -R ~/.agents/skills/<name> ~/.gemini/config/skills/
   ```

## Notable Bundles

Candidates, not installed: this repository vendors no external skill. Resolve a listed repository to an immutable commit, review it, then install only the exact local snapshot as described above. Browse more at [vercel-labs/skills](https://github.com/vercel-labs/skills).

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
| LikeC4 DSL             | `likec4/likec4`                             |
| Slidev                 | `slidevjs/slidev`                           |

Mermaid and D2 did not publish official skills when last checked on 2026-08-02, so this repository maintains reviewed first-party `mermaid` and `d2` skills with official documentation references.

## Gotchas

1. **Scope Conflicts**: Workspace skills override global skills with the same name.
1. **Structure**: Every skill folder must contain a valid `SKILL.md` at its root.
1. **Catalog Portability**: This first-party suite is installed as one catalog and may link between packages under `skills/` and `.agents/skills/`. Catalog containment does not prove that one copied skill folder is independently distributable under the [Agent Skills file-reference rule](https://agentskills.io/specification#file-references); before standalone publication, replace sibling links with skill-name routing or bundle the dependency, copy only that folder to a disposable directory, and validate it there.
1. **Direct Disclosure**: Keep progressive-disclosure resources one level deep. Every package resource other than host-owned `agents/openai.yaml` must be named directly by the root `SKILL.md` or by a local path in that OpenAI metadata; a nested reference does not make another resource discoverable. Nested instructional Markdown is validated for missing or unsafe package links. Output artifacts under `templates/` may use destinations that become valid after materialization, but they still reject unsafe local schemes.
1. **Antigravity Shared Path**: Use `~/.gemini/config/skills/` as the shared cross-product path. Current Antigravity CLI also recognizes its CLI-specific `~/.gemini/antigravity-cli/skills/` path, but maintaining both creates redundant precedence and stale-copy risks.
1. **Claude Workspace Link**: Inspect an existing `.claude/skills` path before creating the link; never overwrite an unmanaged directory.
1. **Provenance**: Re-review upstream changes before updating an installed skill, especially changes to scripts, hooks, or network access.
