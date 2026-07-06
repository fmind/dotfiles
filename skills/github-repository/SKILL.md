---
name: github-repository
description: Configure a GitHub repository's metadata (description, homepage, topics) and solo-developer settings via `gh`, derived from the codebase. Use when setting up or tidying a repo's GitHub settings.
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/github-repository
  created: 2026-06-23
  updated: 2026-07-06
---

# Configure GitHub Repository

Inspect a GitHub repository and configure its metadata (description, homepage, topics) and solo-developer settings via `gh` — clean merge history, secure defaults, and a decluttered UI.

## Workflow

Follow these steps to inspect the codebase and apply the configurations:

### 1. Extract Repository Metadata

Inspect the codebase files to discover or derive the repository metadata:

- **Primary Source (Manifests)**: Check project manifest files:
  - Go: `go.mod` (under `module` path for the package/project name).
  - Python: `pyproject.toml` (under `[project]` fields for name, description, homepage/urls).
- **Secondary Source (README)**: Read the `README.md` first few paragraphs to extract a concise, high-level description (under ~140 characters).
- **Homepage URL**: Derive from hosting setups (e.g., `https://<username>.github.io/<repo>` for GitHub Pages).
- **Topics/Tags**: Select 3 to 6 descriptive tags representing the main programming language, frameworks, tools, or domain (e.g., `agent`, `python`, `cli`).
  - _Constraint_: Topics must be lowercase letters, numbers, or hyphens, 50 characters or less, and start with a letter or number (max 20 per repo).

### 2. Verify GitHub CLI & Inspect Current State

Ensure the GitHub CLI (`gh`) is authenticated, and fetch the repository's current settings to make changes idempotent:

```bash
# Check authentication status
gh auth status

# Get the origin remote URL
git config --get remote.origin.url

# Inspect existing metadata and settings
gh repo view --json description,homepageUrl,repositoryTopics,deleteBranchOnMerge,squashMergeAllowed,mergeCommitAllowed,rebaseMergeAllowed,hasIssuesEnabled,hasProjectsEnabled,hasWikiEnabled,hasDiscussionsEnabled
```

If the repository does not have an active GitHub remote or `gh` is not authenticated, report this and halt.

### 3. Apply Consolidated Configurations

Combine metadata updates and solo developer optimizations into a single consolidated `gh repo edit` command to maximize speed and efficiency:

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

_Note_: If you manage project issues externally or locally in a `task.md` file, you can also append `--enable-issues=false` to fully declutter the sidebar.

## Gotchas

- **Truncation**: Description strings must be single-line and concise (under ~140 characters) to avoid GitHub UI truncation.
- **Advanced Security**: Secret scanning push protection is free for public repositories and personal private repositories, but may require a GitHub Advanced Security subscription for organization-owned private repositories.
- **Visibility Changes**: Changing visibility can have severe consequences. Avoid passing `--visibility` or `--accept-visibility-change-consequences` unless explicitly instructed by the user.

## Documentation

- [GitHub CLI Manual - gh repo edit](https://cli.github.com/manual/gh_repo_edit)
- Companion skills:
  - [github-pull-request](../github-pull-request/SKILL.md) — Manage pull requests with clean title/description standards.
  - [project-license](../project-license/SKILL.md) — Set up the correct LICENSE for the repository.
