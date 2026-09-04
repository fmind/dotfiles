---
name: gitleaks
description: Detect leaked secrets with gitleaks in staged changes, the working tree, or full git history; handle allowlists and rotation. Use for any gitleaks scan or leaked credential.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/gitleaks
  created: "2026-09-02"
  updated: "2026-09-03"
---

# Gitleaks

Find credentials before they reach a remote and the ones that already did; each scope answers a different question, and [secure](../secure/SKILL.md) orders the pass.

## Commands

```bash
gitleaks git --staged --verbose                        # pre-commit: the change about to be committed
gitleaks git --log-opts="--max-count=100" --verbose    # check:leaks: the recent commits (bounded, fast)
gitleaks git --verbose                                 # full history: the scheduled audit
gitleaks dir . --verbose                               # working tree, including untracked files
gitleaks git --redact=100 --report-format sarif --report-path gitleaks.sarif
```

## Mise Task

Expose a scope-aware scan as `check:leaks` per [mise](../mise/SKILL.md): scan the working tree before the repository has a commit, retain the bounded history scan afterward, and forward `--staged` from the pre-commit hook in [lefthook](../lefthook/SKILL.md):

```toml
[tasks."check:leaks"]
description = "Audit staged changes, the working tree, or recent commits for leaked secrets (gitleaks)"
run = """
sh -c '
if [ "$#" -gt 0 ]; then
  gitleaks git --verbose "$@"
elif git rev-parse --verify HEAD >/dev/null 2>&1; then
  gitleaks git --log-opts="--max-count=100" --verbose
else
  gitleaks dir . --verbose
fi
' --
"""
```

## When a Secret Is Found

1. **Rotate first**: a secret in history is compromised even after the commit disappears.
1. **Remove it from the source**: move the value to an environment variable or an encrypted file per [sops-secrets](../sops-secrets/SKILL.md).
1. **Rewrite history only when asked**: rewrites affect every clone; confirm with the user before `git filter-repo`.
1. **Allowlist true false positives** with an inline `gitleaks:allow` comment or a rule in `.gitleaks.toml`, each with a reason.

## Gotchas

- **Shallow CI checkouts**: fetch the depth the task scans (`fetch-depth: 100`) and keep the full-history audit in the scheduled `security.yml` job at `fetch-depth: 0` per [github-actions](../github-actions/SKILL.md).
- **Fresh repository**: the task uses `gitleaks dir` until `HEAD` exists, avoiding a misleading Git error while still scanning untracked files; `--staged` keeps pre-commit scoped to the index.
- **`--redact` in shared logs**: never print a found secret in CI output or an uploaded report.

## Documentation

- [gitleaks](https://github.com/gitleaks/gitleaks)
- Companion skills: [secure](../secure/SKILL.md), [lefthook](../lefthook/SKILL.md), [trivy](../trivy/SKILL.md) (also reports secrets in `fs` scans).
