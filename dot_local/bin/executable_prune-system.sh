#!/usr/bin/env bash
# prune-system.sh — Clean up system caches, temporary scripts, and Docker resources

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo -e "\033[1mStarting system pruning...\033[0m"

# 1. Run base agent and development cache pruning
if [ -x "${SCRIPT_DIR}/executable_prune-agents.sh" ]; then
  "${SCRIPT_DIR}/executable_prune-agents.sh"
elif [ -x "${SCRIPT_DIR}/prune-agents.sh" ]; then
  "${SCRIPT_DIR}/prune-agents.sh"
elif command -v prune-agents.sh &>/dev/null; then
  prune-agents.sh
elif command -v executable_prune-agents.sh &>/dev/null; then
  executable_prune-agents.sh
fi

# 2. Docker system resources (containers, networks, dangling images)
if command -v docker &>/dev/null && docker info &>/dev/null; then
  echo "Pruning Docker system resources..."
  docker system prune -f
  echo "✓ Docker resources pruned."
fi

# 3. Go caches
if command -v go &>/dev/null; then
  echo "Cleaning Go build and module cache..."
  go clean -cache -testcache -modcache || true
  echo "✓ Go caches cleaned."
fi

# 4. Deep Python package caches
if command -v uv &>/dev/null; then
  echo "Cleaning full uv package cache..."
  uv cache clean
  echo "✓ Full uv cache cleaned."
fi

if command -v pip &>/dev/null; then
  echo "Cleaning pip cache..."
  pip cache purge
  echo "✓ pip cache cleaned."
fi

# 5. Node.js & npm full cache
if command -v npm &>/dev/null; then
  echo "Cleaning full npm cache..."
  npm cache clean --force
  echo "✓ npm cache cleaned."
fi

# 6. Additional mise configuration pruning
if command -v mise &>/dev/null; then
  echo "Pruning untracked mise configuration links..."
  mise prune --configs -y || true
fi

# 7. Linter and build tool caches
if command -v dprint &>/dev/null; then
  echo "Cleaning dprint cache..."
  dprint clear-cache
  echo "✓ dprint cache cleaned."
fi

if command -v golangci-lint &>/dev/null; then
  echo "Cleaning golangci-lint cache..."
  golangci-lint cache clean
  echo "✓ golangci-lint cache cleaned."
fi

# 8. Helm cache
if command -v helm &>/dev/null && [ -d "${HOME}/.cache/helm" ]; then
  echo "Cleaning helm cache..."
  rm -rf "${HOME}/.cache/helm"/*
  echo "✓ helm cache cleaned."
fi

echo -e "\033[32;1m✓ System pruning complete.\033[0m"
