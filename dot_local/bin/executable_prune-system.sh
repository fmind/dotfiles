#!/usr/bin/env bash
# prune-system.sh — Clean up system caches, temporary scripts, and Docker resources

set -euo pipefail

echo -e "\033[1mStarting system pruning...\033[0m"

# 1. Docker resources
if command -v docker &>/dev/null; then
  echo "Pruning Docker resources..."
  docker system prune -f
  echo "✓ Docker resources pruned."
else
  echo "Docker not installed, skipping."
fi

# 2. Go caches
if command -v go &>/dev/null; then
  echo "Cleaning Go build and module cache..."
  go clean -cache -testcache -modcache || true
  echo "✓ Go caches cleaned."
fi

# 3. uv cache
if command -v uv &>/dev/null; then
  echo "Cleaning uv package cache..."
  uv cache clean
  echo "✓ uv cache cleaned."
fi

# 4. Python pip cache
if command -v pip &>/dev/null; then
  echo "Cleaning pip cache..."
  pip cache purge
  echo "✓ pip cache cleaned."
fi

# 5. Node.js & npm cache
if command -v npm &>/dev/null; then
  echo "Cleaning npm cache..."
  npm cache clean --force
  echo "✓ npm cache cleaned."
fi

# 6. mise caches and downloads
if command -v mise &>/dev/null; then
  echo "Pruning unused tool versions from mise..."
  mise prune -y
  echo "Cleaning mise cache..."
  mise cache clear
  if [ -d "${HOME}/.local/share/mise/downloads" ]; then
    echo "Cleaning mise downloads..."
    rm -rf "${HOME}/.local/share/mise/downloads"/*
    echo "✓ mise downloads cleaned."
  fi
  echo "✓ mise caches cleaned."
fi

# 7. dprint cache
if command -v dprint &>/dev/null; then
  echo "Cleaning dprint cache..."
  dprint clear-cache
  echo "✓ dprint cache cleaned."
fi

# 8. golangci-lint cache
if command -v golangci-lint &>/dev/null; then
  echo "Cleaning golangci-lint cache..."
  golangci-lint cache clean
  echo "✓ golangci-lint cache cleaned."
fi

# 9. Helm cache
if command -v helm &>/dev/null && [ -d "${HOME}/.cache/helm" ]; then
  echo "Cleaning helm cache..."
  rm -rf "${HOME}/.cache/helm"/*
  echo "✓ helm cache cleaned."
fi

echo -e "\033[32;1m✓ System pruning complete.\033[0m"
