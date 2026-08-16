# Changelog

All notable changes to this project are documented in this file.

## [1.18.3] - 2026-08-16

### 📚 Documentation

- _(fish)_ Drop the GROK_WEB_FETCH rationale comment

## [1.18.2] - 2026-08-16

### 🐛 Bug Fixes

- _(grok)_ Correct the lock note now that mise checksums grok

## [1.18.1] - 2026-08-16

### 🐛 Bug Fixes

- _(chezmoi)_ Prune stale agent config keys and pin compat sessions

## [1.18.0] - 2026-08-16

### 🚀 Features

- _(grok)_ Integrate Grok Build CLI as a first-class agent

## [1.17.1] - 2026-08-14

### 🐛 Bug Fixes

- _(mise)_ Use asdf backend for ollama on macOS (#76)

## [1.17.0] - 2026-08-11

### 🚀 Features

- _(skills)_ Add handover skill and GOTH asset bundling

## [1.16.0] - 2026-08-10

### 🚀 Features

- _(dot)_ Add 1-letter path aliases for agent and query commands

## [1.15.2] - 2026-08-10

### ♻️ Refactor

- _(dot)_ Remove externals pull dirs and update agent clean prefixes

## [1.15.1] - 2026-08-10

### 🧹 Miscellaneous

- _(mise)_ Remove ollama

## [1.15.0] - 2026-08-09

### 🚀 Features

- _(dot,nvim)_ Add completion tools and fix likec4 loading

## [1.14.1] - 2026-08-09

### 🐛 Bug Fixes

- _(k3d)_ Remove custom eviction hard kubelet argument

## [1.14.0] - 2026-08-09

### 🚀 Features

- _(skills)_ Merge the product lifecycle into product-loop and rename code-review to diff-review

### 🐛 Bug Fixes

- _(dot)_ Stop verify from reporting healthy tools as broken
- _(k8s-local)_ Move ingress off 8080/8443 and stop DiskPressure from blocking scheduling

## [1.13.1] - 2026-08-08

### 🐛 Bug Fixes

- _(dot)_ Compare evaluated paths in the skill resource containment check

## [1.13.0] - 2026-08-08

### 🚀 Features

- _(skills)_ Add twenty skills and tighten the repository gate

## [1.12.1] - 2026-08-07

### 🐛 Bug Fixes

- _(dot)_ Remove just from completions

## [1.12.0] - 2026-08-07

### 🚀 Features

- _(skills)_ Add skills for cloud-run, hugo, sops-secrets, terraform-stack, and typst

## [1.11.1] - 2026-08-07

### 📚 Documentation

- _(agents)_ Update global instructions and chezmoiignore

## [1.11.0] - 2026-08-06

### 🚀 Features

- _(tools)_ Add slack-cli

## [1.10.7] - 2026-08-04

### 🐛 Bug Fixes

- _(chezmoi)_ Drop leading slashes rejected by chezmoi 2.72
- _(mise)_ Authenticate GitHub API with the gh CLI token

### 🧹 Miscellaneous

- _(mise)_ Build in the all gate and refresh lockfiles
- _(nvim)_ Refresh LazyVim plugin lockfile

## [1.10.6] - 2026-08-04

### 🐛 Bug Fixes

- _(mise)_ Align task execution directories and binary paths (#74)
- _(mise)_ Pin task directories to config_root instead of cwd

## [1.10.5] - 2026-08-02

### 📚 Documentation

- _(agents)_ Require language identifier for markdown code blocks

## [1.10.4] - 2026-08-02

### 🐛 Bug Fixes

- _(claude)_ Update model identifier to opus[1m]

## [1.10.3] - 2026-08-02

### ♻️ Refactor

- _(skills)_ Align skill declarations and update contract tests

## [1.10.2] - 2026-08-02

### ♻️ Refactor

- _(release)_ Simplify release tag workflow and reapply dot binary

## [1.10.1] - 2026-08-02

### 🐛 Bug Fixes

- _(dot)_ Increase default capability probe timeout to 15s

## [1.10.0] - 2026-08-02

### 🚀 Features

- _(verify)_ Gate install freshness on build inputs and redeploy post-commit

### 🐛 Bug Fixes

- _(cluster)_ Create the local cluster when no k3d cluster exists

### 🧹 Miscellaneous

- _(upgrade)_ Bump global tools and Neovim plugins

## [1.9.0] - 2026-08-02

### 🚀 Features

- _(github)_ Add agent-runnable issue queue
- _(release)_ Gate publication on exact-head CI
- _(dot)_ Emit bounded and redacted project context
- _(agent)_ Add cross-agent discovery and hook doctor
- _(cluster)_ Collect bounded sanitized diagnostics
- _(security)_ Add scheduled full-history scans
- _(stacks)_ Resolve exact local dependency source
- _(skill)_ Add exact-head release audit
- _(skill)_ Add cross-cutting repository review
- _(agent)_ Add versioned session query exports
- _(agent)_ Sync Copilot sessions from personal hook (#69)
- _(skill)_ Add evidence-first project backlog (#70)
- _(skill)_ Add bounded Kubernetes review (#71)
- _(skill)_ Add cooperative issue execution (#72)
- _(config)_ Update root directories for fmind, fmind-ai, and mlops-courses
- _(dot)_ Unify agent sources, add configurable bounds, simplify bootstrap

### 🐛 Bug Fixes

- _(verify)_ Probe CLI capabilities instead of shims
- _(ai)_ Pack diffs with explicit omission evidence
- _(agent)_ Make session ingestion lineage-safe
- _(prune)_ Retain raw sessions until verified
- _(cluster)_ Isolate and verify kubeconfig targets
- _(bootstrap)_ Verify pinned installers
- _(bootstrap)_ Trust nested mise configs hermetically
- _(verify)_ Support Helm 4 capability probe (#73)

### ♻️ Refactor

- _(mise)_ Separate update and convergence phases

### 🧪 Testing

- _(docs)_ Validate agent command contracts
- _(ci)_ Validate workflows and skill contracts

## [1.8.0] - 2026-07-31

### 🚀 Features

- _(dot)_ Consolidate prune command into dot cli and refine skills

## [1.7.3] - 2026-07-31

### 🐛 Bug Fixes

- _(skills)_ Correct inaccurate commands and broken hook ordering

### 🧹 Miscellaneous

- _(prune)_ Include the antigravity brain dir in agent transcript pruning

## [1.7.2] - 2026-07-31

### 🐛 Bug Fixes

- _(claude)_ Keep Claude Code runtime settings across chezmoi applies

### 🧪 Testing

- _(dot)_ Expand unit test coverage across dot cli subcommands

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
