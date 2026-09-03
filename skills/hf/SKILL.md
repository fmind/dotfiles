---
name: hf
description: Use the Hugging Face hf CLI to download, upload, cache, and run jobs for models, datasets, and spaces with token auth and bounded transfers. Use for any Hugging Face Hub task.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/hf
  created: "2026-09-02"
  updated: "2026-09-03"
---

# Hugging Face CLI

Use `hf` for Hub operations from the shell. The CLI generates its own command skill from the installed version; this skill owns authentication, cache hygiene, and the authority boundary around uploads and paid jobs.

## Workflow

1. **Resolve the account**: `hf auth whoami`; log in with `hf auth login` or pass `HF_TOKEN` through the environment. Never run `hf auth token` in a transcript or log.
1. **Read before writing**: `hf models info <repo-id>`, `hf models card <repo-id>`, `hf models ls --sort downloads --limit 10`, and the `datasets` and `spaces` twins answer most questions without a download.
1. **Download into an ignored directory**: gated repositories need the license accepted on the website first.

   ```bash
   hf download <repo-id> --local-dir models/<name>
   hf download <owner>/<dataset> --repo-type dataset --local-dir data/<name>
   ```

1. **Keep the cache under control**: `hf cache ls`, `hf cache prune`, `hf cache rm <repo-id>`; `HF_HOME` relocates it (see [reclaim-disk](../reclaim-disk/SKILL.md)).
1. **Upload with authority**: `hf repos create <repo-id> --private` then `hf upload <repo-id> <local-path>`; confirm repository, visibility, and license before the first push, then verify with `hf models info` or `hf repos ls`.
1. **Remote compute with authority**: `hf jobs run` and `hf jobs uv run` bill by hardware flavor; confirm the flavor and timeout, then watch `hf jobs logs` and `hf jobs ps`.

## Gotchas

- **Name**: `hf` replaced `huggingface-cli`; the mise tool is `pipx:huggingface_hub`.
- **Large transfers**: `hf download` and `hf upload` resume; `hf cache verify` checks checksums after an interrupted run.
- **Pin a revision**: pass `--revision <sha>` for reproducible downloads; anything loaded with `trust_remote_code` is third-party code to review first.

## Official Skills

Upstream: `huggingface/skills`, the same packages the CLI marketplace serves. `hf skills add` writes the CLI's own skill and `hf skills list` shows the marketplace; install at project scope (never `-g`) after reviewing the snapshot (see [agent-skills](../agent-skills/SKILL.md)):

```bash
hf skills list
hf skills add                  # the CLI skill into .agents/skills
hf skills add <name>
hf skills update
```

## Documentation

- [hf CLI guide](https://huggingface.co/docs/huggingface_hub/en/guides/cli) · [huggingface/skills](https://github.com/huggingface/skills)
- Companion skills: [kaggle](../kaggle/SKILL.md), [colab](../colab/SKILL.md), [python-stack](../python-stack/SKILL.md), [reclaim-disk](../reclaim-disk/SKILL.md).
