---
name: project-license
description: "Write the right LICENSE and manifest field: MIT for public fmind, fmind-ai, and mlops-courses repos, CC-BY-4.0 for written courses, proprietary otherwise. Use when adding a license."
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/project-license
  created: "2026-06-23"
  updated: "2026-09-03"
---

# Project License

Select, write, and declare the LICENSE a repository needs from its namespace, visibility, and content; [github-repository](../github-repository/SKILL.md) owns the remaining repository settings.

## Workflow

1. **Detect the namespace**: `git remote -v`, the manifest (`pyproject.toml`, `go.mod`, `package.json`), or the parent directory gives the owning organization or user.
1. **Read the existing license first**: `ls LICENSE*` and `gh repo view --json nameWithOwner,isPrivate,licenseInfo`; an existing license stays unless the user asked to replace it.
1. **Select the license**:
   - Public code under `fmind`, `fmind-ai`, or `mlops-courses`: MIT, from [MIT](references/MIT).
   - Written course material (lessons, exercises, prose): CC-BY-4.0 as `LICENSE.txt`, from [CC-BY-4.0](references/CC-BY-4.0); `mlops-courses/mlops-coding-course` is the reference example.
   - Every private repository and every other namespace: proprietary, from [PROPRIETARY](references/PROPRIETARY); never an open-source license.
1. **Write the file** at the repository root:
   - Copyright holder from the namespace: `fmind` and `fmind-ai` use `Médéric Hurier (Fmind)`; `mlops-courses` uses `MLOps Courses`. When unsure, copy the holder from a sibling repository.
   - Resolve `<year>` to the current calendar year (`date +%Y`).
1. **Declare it in the manifest**:
   - Python: PEP 639 SPDX `license = "MIT"` or `license = "LicenseRef-Proprietary"` plus `license-files = ["LICENSE"]` in `pyproject.toml`.
   - Node: `"license": "MIT"` or `"UNLICENSED"` in `package.json`; Go has no manifest field.

## Gotchas

- **Namespace is not ownership**: `mlops-courses` repositories are public and MIT although the namespace is neither `fmind` nor `fmind-ai`; a namespace-only rule would relicense them as proprietary.
- **`LICENSE.txt`**: a course repository may carry `LICENSE.txt`; writing `LICENSE` next to it leaves two conflicting licenses.
- **Code and prose differ**: one course organization holds both, MIT for code repositories and CC-BY-4.0 for the written course; check what the repository publishes.
- **SPDX only**: plain `"Proprietary"` is not a valid SPDX expression and modern build tools reject it; use `LicenseRef-Proprietary`.

## Documentation

- [Choose an Open Source License](https://choosealicense.com/) · [SPDX license list](https://spdx.org/licenses/) · [PEP 639](https://peps.python.org/pep-0639/)
- [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/) — the legal code shipped in `references/CC-BY-4.0`.
- Companion skills: [new-project](../new-project/SKILL.md) (calls this skill when bootstrapping), [github-repository](../github-repository/SKILL.md) (repository settings).
