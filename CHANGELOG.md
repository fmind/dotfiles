# Changelog

All notable changes to this project are documented in this file.

## [1.7.1] - 2026-07-31

### ♻️ Refactor

- _(dot)_ Drop the duplicate agent notify subcommand

### 🧹 Miscellaneous

- _(mise)_ Re-sync the global lockfile with installed platforms

## [1.7.0] - 2026-07-31

### 🚀 Features

- _(dot)_ Notify desktop on agent stop and session end

### 🐛 Bug Fixes

- _(ci)_ Lock trivy and restore a green cross-platform pipeline
- _(check)_ Render chezmoi against a seeded config so CI can validate it
- _(dot)_ Stub the platform browser opener so macOS tests pass

### ♻️ Refactor

- _(dot)_ Resolve command alias collisions

## [1.6.0] - 2026-07-31

### 🚀 Features

- Update dot CLI agent tools, dotfiles tasks, and skills

### 🧹 Miscellaneous

- _(claude)_ Remove DISABLE_TELEMETRY setting

## [1.5.0] - 2026-07-16

### 🚀 Features

- _(skills)_ Add visual and presentation skills

## [1.4.4] - 2026-07-15

### 🧹 Miscellaneous

- _(mise)_ Update lockfile

## [1.4.3] - 2026-07-15

### 🧹 Miscellaneous

- Ignore local skills, remove completions hook, update k8s docs

## [1.4.2] - 2026-07-14

### 🧹 Miscellaneous

- _(gh)_ Remove prefer_editor_prompt

## [1.4.1] - 2026-07-14

### 📚 Documentation

- _(agents)_ Allow direct push to main branch

## [1.4.0] - 2026-07-14

### 🚀 Features

- _(bash)_ Add mise activation and paths

## [1.3.0] - 2026-07-14

### 🚀 Features

- _(skills)_ Add dependabot skill

## [1.2.4] - 2026-07-13

### 🧹 Miscellaneous

- Configure lsd and disable slsa attestations in mise

## [1.2.3] - 2026-07-13

### ♻️ Refactor

- Replace eza with lsd

## [1.2.2] - 2026-07-12

### 🐛 Bug Fixes

- _(mise)_ Pin gws to 0.22.4 for glibc compatibility

## [1.2.1] - 2026-07-12

### 🧹 Miscellaneous

- Add hadolint linter and migrate lazygit config

## [1.2.0] - 2026-07-12

### 🚀 Features

- _(dot)_ Add agent session log management command
- _(dot)_ Improve CLI commands, tests, and configuration files

### ♻️ Refactor

- _(dot)_ Consolidate gcloud login and add python-script skill

## [1.1.1] - 2026-07-08

### ♻️ Refactor

- Trim cliff.toml defaults and dot-release skill

## [1.1.0] - 2026-07-08

### 🚀 Features

- _(dot)_ Add release command and html-slides skill

## [1.0.0] - 2026-07-06

### 🚀 Features

- Initial public release
