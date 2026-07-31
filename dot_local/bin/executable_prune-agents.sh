#!/usr/bin/env bash
# prune-agents.sh — reclaim disk from agent session transcripts and build caches.
#
# Deliberately narrower than prune-system.sh. Two things it must never do, because both
# make the machine slower at the only thing it is optimized for (running Fgentic's gates):
#
#   * `docker container prune` / `docker volume prune` — stopped k3d containers ARE the
#     clusters. Pruning them destroys every local cluster and its node state.
#   * `go clean -cache -modcache` — the Go build cache is what keeps `mise run test` at
#     ~77 s instead of several minutes, and ~/go/pkg/mod is where agents read pinned
#     upstream source.

set -euo pipefail

readonly KEEP_DAYS="${PRUNE_KEEP_DAYS:-7}"

# POSIX df output keeps this portable across GNU/Linux and macOS.
free_gb() { df -Pk / | awk 'NR == 2 { print int($4 / 1024 / 1024) }'; }

before=$(free_gb)
echo "Free before: ${before}G (keeping the last ${KEEP_DAYS} days of transcripts)"

# 1. Agent session transcripts — the fastest-growing consumer, ~1.5 GiB/day.
for dir in "${HOME}/.codex/sessions" "${HOME}/.claude/projects"; do
  if [ -d "${dir}" ]; then
    find "${dir}" -type f -mtime "+${KEEP_DAYS}" -delete 2>/dev/null || true
    find "${dir}" -type d -empty -delete 2>/dev/null || true
    echo "✓ pruned ${dir}"
  fi
done

# 2. Docker build cache only. Never containers, volumes, or images.
if command -v docker &>/dev/null && docker info >/dev/null 2>&1; then
  docker builder prune -af >/dev/null 2>&1 || true
  echo "✓ docker build cache pruned (containers, volumes and images untouched)"
fi

# 3. Regenerable package caches.
[ -d "${HOME}/.npm/_npx" ] && rm -rf "${HOME}/.npm/_npx" && echo "✓ npx cache cleared"
[ -d "${HOME}/.cache/trivy" ] && rm -rf "${HOME}/.cache/trivy" && echo "✓ trivy cache cleared"
command -v uv &>/dev/null && uv cache prune >/dev/null 2>&1 && echo "✓ uv cache pruned"

# 4. mise: unused tool versions and download tarballs, not the installs in use.
if command -v mise &>/dev/null; then
  mise prune -y >/dev/null 2>&1 || true
  mise cache clear >/dev/null 2>&1 || true
  rm -rf "${HOME}/.local/share/mise/http-tarballs"/* "${HOME}/.local/share/mise/downloads"/* 2>/dev/null || true
  echo "✓ mise unused versions and tarballs pruned"
fi

after=$(free_gb)
echo "Free after:  ${after}G  (reclaimed $((after - before))G)"
