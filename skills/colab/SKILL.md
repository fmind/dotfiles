---
name: colab
description: Use the Google Colab CLI to rent GPU or TPU sessions, run scripts on the remote VM, sync files, and stop sessions to cap spend. Use for Colab compute from the terminal.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/colab
  created: "2026-09-02"
  updated: "2026-09-03"
---

# Google Colab CLI

Use `colab` when a task needs an accelerator the workstation lacks. The official Colab skill documents every command; this skill owns authentication, session hygiene, and the spend boundary.

## Workflow

1. **Authenticate**: OAuth by default (`--auth oauth2`), or `--auth adc` to reuse the Application Default Credentials from [gcloud](../gcloud/SKILL.md); session state lives under `~/.config/colab-cli/`.
1. **Prefer ephemeral runs**: `colab run` rents a VM, runs the script, and releases it; a shebang `#!/usr/bin/env -S colab run --gpu T4` makes a single file self-contained per [python-script](../python-script/SKILL.md).

   ```bash
   colab run --gpu T4 --timeout 3600 train.py
   ```

1. **Keep a session only while iterating**: `colab new -s <name> --gpu L4` (or `--tpu v6e1`), then `colab exec -s <name> -f snippet.py --timeout 600`, `colab upload`, `colab download`, and `colab ls`.
1. **Stop what you started**: `colab sessions` then `colab stop -s <name>`; an idle session keeps consuming compute units. Run `colab status` before claiming a job finished.
1. **Verify**: `colab log` shows the history; download the artifacts before stopping the session.

## Gotchas

- **30-second default**: `colab run` and `colab exec` abort code execution after 30 seconds unless `--timeout <seconds>` covers the whole job.
- **Pinned dependency**: mise installs `google-colab-cli` with `jupyter-kernel-client==0.15.0`; 1.0.0 renamed the client class and breaks every session.
- **Tiers**: accelerator availability depends on the subscription; `colab pay` opens the compute-units page, so treat it as spend.
- **Disposable VM**: keep secrets off the session beyond what the task needs; use `colab drivemount` only when Drive data is required.

## Official Skills

Upstream: `googlecolab/google-colab-cli` (the same text `colab skill` prints). List the current release, then install what the task needs at project scope after reviewing the snapshot (see [agent-skills](../agent-skills/SKILL.md)):

```bash
skills add googlecolab/google-colab-cli --list
skills add googlecolab/google-colab-cli --skill <name> -y
```

## Documentation

- [Colab CLI](https://github.com/googlecolab/google-colab-cli)
- Companion skills: [kaggle](../kaggle/SKILL.md), [hf](../hf/SKILL.md), [python-script](../python-script/SKILL.md), [gcloud](../gcloud/SKILL.md).
