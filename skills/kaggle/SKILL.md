---
name: kaggle
description: Use the kaggle CLI for competitions, datasets, kernels, and models with token auth, bounded downloads, and explicit submission authority. Use for any Kaggle CLI task.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/kaggle
  created: "2026-09-02"
  updated: "2026-09-03"
---

# Kaggle CLI

Use `kaggle` for competition, dataset, kernel, and model operations from the shell. The official Kaggle skills document every command and metadata file; this skill owns authentication, download scope, and the authority boundary around submissions and publications.

## Workflow

1. **Resolve the account**: `kaggle auth login` (OAuth) or `KAGGLE_API_TOKEN` in the environment; the legacy `~/.kaggle/kaggle.json` still works. Never run `kaggle auth print-access-token` during ordinary work.
1. **Read before writing**: `kaggle competitions list`, `kaggle competitions files <slug>`, `kaggle datasets files <owner>/<name>`, and `kaggle competitions submission-limits <slug> --json` cost nothing and reveal the rules in force.
1. **Download into an ignored directory**: accept the competition rules on the website first (the CLI returns 403 otherwise).

   ```bash
   kaggle competitions download <slug> -p data/
   kaggle datasets download <owner>/<name> -p data/ --unzip
   ```

1. **Kernels as code**: `kaggle kernels init -p <dir>` writes `kernel-metadata.json`; `kaggle kernels push -p <dir>` publishes it; `kaggle kernels status <owner>/<slug>` and `kaggle kernels output <owner>/<slug> -p out/` retrieve the run.
1. **Submit with authority**: a submission counts against the daily limit and shows on the leaderboard, so confirm the competition, file, and message first, then verify.

   ```bash
   kaggle competitions submit <slug> -f submission.csv -m "<message>"
   kaggle competitions submissions <slug>
   ```

1. **Publish with authority**: `kaggle datasets create -p <dir>` and `kaggle datasets version -p <dir> -m "<message>"` are private by default (`--public` flips it); confirm the license and visibility in the metadata before the first push.

## Gotchas

- **Pinned version**: in a project that pins `kaggle`, call `uv run kaggle` so the pinned version runs instead of the global shim.
- **Quota**: `kaggle quota` shows the accelerator budget before a kernel push with `--accelerator`.
- **Scripts**: pass `-W` to silence the out-of-date warning so JSON output stays parseable.

## Official Skills

Upstream: `Kaggle/kaggle-cli` (command guide) and `Kaggle/kaggle-skills` (competition formats). List the current releases, then install what the task needs at project scope after reviewing the snapshot (see [agent-skills](../agent-skills/SKILL.md)):

```bash
skills add Kaggle/kaggle-cli --list
skills add Kaggle/kaggle-skills --list
skills add <owner/repo> --skill <name> -y
```

## Documentation

- [Kaggle CLI](https://github.com/Kaggle/kaggle-cli) · [Kaggle API](https://www.kaggle.com/docs/api)
- Companion skills: [python-stack](../python-stack/SKILL.md) (project layout), [duckdb](../duckdb/SKILL.md) (inspect downloads), [hf](../hf/SKILL.md) (Hub models and datasets), [colab](../colab/SKILL.md) (rented accelerators).
