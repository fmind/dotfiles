#!/usr/bin/env bash
# prune-agents.sh — Clean up agent session transcripts and development build caches
# (Deliberately keeps Docker containers/volumes and Go build cache intact)

set -euo pipefail

readonly KEEP_DAYS="${PRUNE_KEEP_DAYS:-7}"

echo -e "\033[1mStarting agent pruning...\033[0m"

# 1. Agent session transcripts
echo "Pruning agent transcripts older than ${KEEP_DAYS} days..."
for dir in "${HOME}/.codex/sessions" "${HOME}/.claude/projects" "${HOME}/.agents/sessions"; do
  if [ -d "${dir}" ]; then
    find "${dir}" -type f -mtime "+${KEEP_DAYS}" -delete 2>/dev/null || true
    find "${dir}" -type d -empty -delete 2>/dev/null || true
  fi
done
echo "✓ Agent transcripts pruned."

# 2. Docker build cache
if command -v docker &>/dev/null && docker info &>/dev/null; then
  echo "Pruning Docker build cache..."
  docker builder prune -af &>/dev/null || true
  echo "✓ Docker build cache pruned."
fi

# 3. Package caches
if [ -d "${HOME}/.npm/_npx" ]; then
  echo "Cleaning npx cache..."
  rm -rf "${HOME}/.npm/_npx"
  echo "✓ npx cache cleaned."
fi

if [ -d "${HOME}/.cache/trivy" ]; then
  echo "Cleaning Trivy cache..."
  rm -rf "${HOME}/.cache/trivy"
  echo "✓ Trivy cache cleaned."
fi

if command -v uv &>/dev/null; then
  echo "Pruning uv cache..."
  uv cache prune &>/dev/null || true
  echo "✓ uv cache pruned."
fi

# 4. mise unused versions and downloads
if command -v mise &>/dev/null; then
  echo "Pruning unused tool versions from mise..."
  mise prune -y &>/dev/null || true
  echo "Cleaning mise cache..."
  mise cache clear &>/dev/null || true
  if [ -d "${HOME}/.local/share/mise/downloads" ] || [ -d "${HOME}/.local/share/mise/http-tarballs" ]; then
    echo "Cleaning mise downloads..."
    rm -rf "${HOME}/.local/share/mise/http-tarballs"/* "${HOME}/.local/share/mise/downloads"/* 2>/dev/null || true
    echo "✓ mise downloads cleaned."
  fi
  echo "✓ mise caches cleaned."
fi

echo -e "\033[32;1m✓ Agent pruning complete.\033[0m"
