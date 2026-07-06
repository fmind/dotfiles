---
name: project-license
description: Select and generate the correct LICENSE — MIT for public fmind/fmind-ai repos, otherwise proprietary — and update project manifests. Use when adding or fixing a project's license.
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/project-license
  created: 2026-06-23
  updated: 2026-07-06
---

# Create Project License

Select, generate, and apply the correct LICENSE file for a repository based on its namespace and ownership.

## Workflow

1. **Detect Ownership & Organization**: Inspect git remotes (`git remote -v`), configuration files (`pyproject.toml`, `go.mod`, etc.), or directory naming to determine the repository's target GitHub organization or user namespace.
1. **Select License Standard**:
   - If the project is associated with the GitHub organizations or users **`fmind`** or **`fmind-ai`** (e.g., `github.com/fmind/...`, `github.com/fmind-ai/...`) and is public, use the **MIT License**.
   - Otherwise, the project is **Proprietary** and must NOT use an open-source license.
1. **Generate LICENSE File**:
   - Write the resolved license content to a `LICENSE` file at the root of the project.
   - Use the correct copyright holder name **"Médéric Hurier (Fmind)"** and the current year (e.g., `2026`).
1. **Update Configuration Files**: For Python, set the PEP 639 SPDX `license` field in `pyproject.toml` — `license = "MIT"` for MIT or `license = "LicenseRef-Proprietary"` for proprietary (plain `"Proprietary"` is not a valid SPDX expression and modern build tools reject it) — and add `license-files = ["LICENSE"]`. Go modules have no license field in `go.mod`.

## License Templates

To keep this skill light and maintainable, the license texts are stored as separate files:

- **MIT License Template**: [MIT](templates/MIT) (for `fmind` and `fmind-ai` projects, resolving `<year>`).
- **Proprietary License Reference**: [PROPRIETARY](references/PROPRIETARY) (for any other projects, resolving `<year>`).

## Gotchas

- **Current Year**: Always dynamically determine the current calendar year when writing `<year>`.

## Documentation

- [Choose an Open Source License](https://choosealicense.com/)
