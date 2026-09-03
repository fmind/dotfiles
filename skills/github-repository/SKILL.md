---
name: github-repository
description: Configure a GitHub repository's description, homepage, topics, and solo-developer settings via gh, derived from the codebase. Use when tidying repo settings.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/github-repository
  created: "2026-06-23"
  updated: "2026-09-03"
---

# GitHub Repository

Derive a repository's description, homepage, and topics from its codebase and apply them with `gh repo edit` together with solo-developer settings: squash-only merges, secure defaults, a decluttered sidebar.

## Workflow

1. **Extract metadata** from the codebase:
   - Manifests: Go `go.mod` (`module` path), Python `pyproject.toml` (`[project]` name, description, urls), TypeScript `package.json` (`name`, `description`, `homepage`).
   - `README.md`: the first paragraphs give a one-line description under ~140 characters.
   - Homepage: derive from hosting, e.g. `https://<owner>.github.io/<repo>` for GitHub Pages.
   - Topics: 3 to 6 lowercase tags for language, frameworks, tools, or domain (`agent`, `python`, `cli`); letters, numbers, and hyphens only, 50 characters max, 20 per repository.
1. **Inspect the current state** so the edit stays idempotent; stop when there is no GitHub remote or `gh` is not authenticated:

   ```bash
   gh auth status
   git config --get remote.origin.url
   gh repo view --json description,homepageUrl,repositoryTopics,deleteBranchOnMerge,squashMergeAllowed,mergeCommitAllowed,rebaseMergeAllowed,hasIssuesEnabled,hasProjectsEnabled,hasWikiEnabled,hasDiscussionsEnabled
   ```

1. **Apply one consolidated edit**; append `--enable-issues=false` only when the project tracks issues elsewhere:

   ```bash
   gh repo edit \
     --description "<description>" \
     --homepage "<homepage-url>" \
     --add-topic "tag1,tag2,tag3" \
     --delete-branch-on-merge \
     --enable-squash-merge \
     --squash-merge-commit-message pr-title-description \
     --enable-merge-commit=false \
     --enable-rebase-merge=false \
     --allow-update-branch \
     --enable-secret-scanning \
     --enable-secret-scanning-push-protection \
     --enable-wiki=false \
     --enable-projects=false \
     --enable-discussions=false
   ```

1. **Verify** with the same `gh repo view --json ...` call and report the fields that changed.

## Gotchas

- **Truncation**: keep the description single-line and under ~140 characters or the GitHub UI truncates it.
- **Advanced Security**: secret-scanning push protection is free for public and personal private repositories; organization-owned private repositories may need GitHub Advanced Security.
- **Visibility**: never pass `--visibility` or `--accept-visibility-change-consequences` unless the user explicitly asks.

## Documentation

- [gh repo edit manual](https://cli.github.com/manual/gh_repo_edit)
- Companion skills: [github-pull-request](../github-pull-request/SKILL.md) (PR titles feed the squash message), [project-license](../project-license/SKILL.md) (LICENSE), [new-project](../new-project/SKILL.md) (bootstrap).
