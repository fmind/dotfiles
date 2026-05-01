---
name: use-github-cli
description: Guide for using the GitHub CLI (gh) to manage issues, pull requests, releases, workflows, and agent skills from the terminal.
---

# Use GitHub CLI (gh)

This skill covers the `gh` CLI for day-to-day GitHub work and for managing Agent Skills via the `gh skill` subcommand (added in `gh` v2.90.0+).

## One-time Setup

```bash
# Authenticate (opens browser or accepts a token).
gh auth login
gh auth status

# Pick the default editor / protocol.
gh config set git_protocol ssh
gh config set editor "nvim"
```

## Pull Requests

```bash
# Create a PR from the current branch.
gh pr create --title "fix: handle empty config" --body "..."
gh pr create --fill                 # use commit messages as title/body
gh pr create --draft

# Inspect / iterate.
gh pr list --author "@me" --state open
gh pr view 123 --comments
gh pr diff 123
gh pr checks 123                    # CI status
gh pr review 123 --approve
gh pr merge 123 --squash --delete-branch
```

## Issues

```bash
gh issue list --label bug --assignee "@me"
gh issue create --title "..." --body "..." --label bug
gh issue view 42 --comments
gh issue close 42 --reason completed
gh issue develop 42 --checkout      # create a branch linked to the issue
```

## Workflows & Runs

```bash
gh workflow list
gh workflow run deploy.yml -f env=staging
gh run list --workflow=ci.yml --limit 5
gh run watch                        # tail the latest run
gh run view <run-id> --log-failed   # only failed step logs
gh run rerun <run-id> --failed
```

## Releases

```bash
gh release list
gh release create v1.4.0 --generate-notes ./dist/*.tar.gz
gh release upload v1.4.0 ./dist/extra.zip --clobber
gh release view v1.4.0
```

## Repos, Forks, Search

```bash
gh repo create my-org/new-repo --private --clone
gh repo fork upstream/project --clone
gh search prs "is:open author:@me language:python" --limit 20
gh search code "TODO(auth)" --owner my-org
```

## GraphQL & REST Escape Hatches

```bash
# REST: paginated JSON.
gh api repos/:owner/:repo/issues --paginate --jq '.[].title'

# GraphQL: arbitrary queries.
gh api graphql -f query='query { viewer { login } }'
```

## Agent Skills via `gh skill`

`gh skill` discovers, installs, updates, and publishes Agent Skills directly from GitHub repositories. It writes provenance metadata (repo, ref, tree SHA) into the SKILL.md frontmatter so installs are reproducible.

```bash
# Search the registry / a repo.
gh skill search mcp-apps
gh skill preview github/awesome-copilot/documentation-writer

# Install (whole repo or a single skill, optionally pinned).
gh skill install github/awesome-copilot
gh skill install github/awesome-copilot documentation-writer
gh skill install github/awesome-copilot documentation-writer@v1.2.0

# Target scope (project vs user) explicitly.
gh skill install <repo> --agent gemini --scope user
gh skill install <repo> --agent gemini --scope project

# Maintain installed skills.
gh skill update --all
gh skill update <skill-name> --pin <sha>
gh skill publish ./my-skill        # validate + push
```

Track the [`gh skill` manual](https://cli.github.com/manual/gh_skill) for the canonical subcommand list — the surface is still evolving.

`--scope user` writes to `~/.gemini/skills/`; `--scope project` writes to `.agents/skills/` (the generic, cross-agent location). Globally installed skills are not tracked by chezmoi by default — `chezmoi add ~/.gemini/skills/<slug>` to commit them.

Prefer `npx skills` over `gh skill` for cross-agent compatibility — it installs to `.agents/skills/` by default and works the same regardless of which coding agent is active. Use `gh skill` when you specifically want GitHub-hosted provenance metadata baked into the SKILL.md frontmatter.

## JSON / Templating

`gh` supports JSON output for any list/view command:

```bash
gh pr list --json number,title,author --jq '.[] | "\(.number) \(.title)"'
gh pr view 123 --json title,body,comments
```

## Important Notes

1. Prefer `gh` over `git` for anything involving GitHub APIs (PRs, issues, runs) — output is structured and supports `--json`/`--jq`.
2. `gh auth login` stores credentials in the system keychain; don't paste tokens into the shell.
3. For bulk or CI work, `gh api --paginate` beats hand-rolling REST calls.
4. `gh skill` is distinct from this dotfile setup's chezmoi-managed skills — be intentional about which path a new skill lives in.

## Documentation

- [gh CLI manual](https://cli.github.com/manual/)
- [gh skill manual](https://cli.github.com/manual/gh_skill)
- [gh skill install manual](https://cli.github.com/manual/gh_skill_install)
- [gh skill launch changelog](https://github.blog/changelog/2026-04-16-manage-agent-skills-with-github-cli/)
- [Agent Skills spec](https://agentskills.io)
