---
name: dependabot
description: Automated dependency updates using GitHub Dependabot — configuration, ecosystem mapping, and validation.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/dependabot
  created: 2026-07-14
  updated: 2026-07-14
---

# Dependabot Dependency Management Standard

Canonical setup for **GitHub Dependabot**, an automated dependency update engine designed to keep dependencies and GitHub Actions pinned, secure, and current.

## 1. Principles

1. **Pin-Everything Strategy**: Pin GitHub Action SHAs, Go modules, npm modules, and Python packages explicitly.
1. **Reduce PR Noise**: Group minor, patch, and digest updates into single, consolidated pull requests using the Dependabot `groups` configuration (e.g., grouping GitHub Action updates or Go module updates) while leaving major updates separate.
1. **Local Validation**: Never merge automated updates blindly. Always run validation pipelines locally (`mise run check` and `mise run test`) before pushing/merging to verify compatibility and catch regressions.
1. **No Auto-Merge**: Do not configure auto-merge for dependencies. Automated systems cannot anticipate protocol, type-checking, or model drift.

## 2. Configuration Setup (`.github/dependabot.yml`)

Place a `dependabot.yml` at `.github/dependabot.yml` at the root of target repositories to manage updates, commit styling, and dependency groupings.

Example configuration for a repository with GitHub Actions and Go modules:

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
        patterns:
          - "*"
  - package-ecosystem: gomod
    directory: /dot
    schedule:
      interval: weekly
      day: monday
    commit-message:
      prefix: "chore(deps)"
    groups:
      go-modules:
        patterns:
          - "*"
```

## 3. Workflow & Commands

1. **Verify Configuration**: Dependabot validation is handled automatically by GitHub upon receiving pushes to `.github/dependabot.yml`. Any syntax errors will be reported in the repository's "Insights" -> "Dependency graph" -> "Dependabot" tab.
1. **Trigger Manual Checks**: To force Dependabot to check for updates immediately, go to your repository on GitHub, navigate to **Insights** -> **Dependency graph** -> **Dependabot**, click on the status/last-check details for a package ecosystem, and click **Check for updates**.
1. **Local Verification**: When a Dependabot PR is opened, fetch the branch locally and run static checkers and tests to verify correctness:
   ```bash
   git fetch origin
   git checkout check-dependabot-branch
   mise run check
   mise run test
   ```

## 4. Licensing & Requirements

1. **100% Free**: Dependabot is natively integrated into GitHub and free for all repositories (both public and private).
1. **GitHub Native**: No tokens or external marketplace applications need to be configured. Enabling the bot is as simple as placing the `dependabot.yml` configuration file in the `.github` directory.

## 5. Documentation

- [GitHub Dependabot Documentation](https://docs.github.com/en/code-security/dependabot)
- [Configuration options for the dependabot.yml file](https://docs.github.com/en/code-security/dependabot/dependabot-version-updates/configuration-options-for-the-dependabot.yml-file)
- [Grouping Dependabot updates](https://docs.github.com/en/code-security/dependabot/dependabot-version-updates/grouping-dependabot-updates-into-a-single-pull-request)
