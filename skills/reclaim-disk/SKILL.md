---
name: reclaim-disk
description: Reclaim disk space safely with dot prune and per-tool cache cleanup without disturbing running agents or uncommitted work. Use when the disk is full or low.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/reclaim-disk
  created: "2026-09-02"
  updated: "2026-09-03"
---

# Reclaim Disk

Free disk space in safety order: the big shared caches first through `dot prune`, project directories last; [dot-cli](../dot-cli/SKILL.md) owns the `dot prune` flag reference.

## Workflow

1. **Measure**: `duf` for the filesystem, then `dust -d 2 ~` and `dust -d 1 ~/.cache ~/.local/share` for the largest directories.
1. **Check for running agents**: `pgrep -af "claude|codex|agy|opencode|copilot|grok"`; never delete the working directory, `.agents/tmp`, or session store of a live session.
1. **Preview**: `dot prune --dry-run --all=deep` lists what the configured targets would remove (agent sessions, Go, uv, npm, mise, Docker, tool caches).
1. **Prune**: `dot prune --all=deep`, or one target at a time (`dot prune --go=module`, `dot prune --docker=system`) when a cache still serves an active project.
1. **Tool-specific extras** when still short: `mise prune -y`, `pnpm store prune`, `go clean -modcache`, `uv cache clean`, `nvim --headless '+Lazy! clean' +qa`.
1. **Project directories last**: `.venv`, `node_modules`, `target`, `dist`, `tmp` in inactive projects; they rebuild from lockfiles.
1. **Verify and report**: `duf` again; list the freed space per target and anything skipped because an agent was using it.

## Gotchas

- **Never delete keys or secrets**: `~/.config/chezmoi/key.txt` (the age key), `~/.ssh`, any `secrets.fish` or `*.age` file.
- **Uncommitted work**: check `git status` before removing anything inside a repository.
- **Docker volumes**: `docker volume ls` shows data volumes; remove one only when the user names it.
- **k3d cluster**: `dot prune --docker=system` deletes stopped k3d clusters; stop and keep the cluster only when the user needs its state.

## Documentation

- [dust](https://github.com/bootandy/dust) · [duf](https://github.com/muesli/duf)
- Companion skills: [dot-cli](../dot-cli/SKILL.md) (every `dot prune` target and depth), [k8s-local](../k8s-local/SKILL.md) (the local cluster `--docker=system` can remove).
