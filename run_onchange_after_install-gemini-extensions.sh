#!/usr/bin/env bash
set -euo pipefail

command -v gemini >/dev/null || {
  echo "gemini CLI not found, skipping extension install"
  exit 0
}

extensions=(
  "https://github.com/fmind/fgate"
  "https://github.com/googleworkspace/cli"
  "https://github.com/ChromeDevTools/chrome-devtools-mcp"
)

installed=$(gemini extensions list 2>/dev/null || true)

for url in "${extensions[@]}"; do
  name="${url##*/}"
  if printf '%s\n' "$installed" | grep -qiE "(^|[[:space:]/])${name}([[:space:]]|$)"; then
    echo "=> $name already installed"
    continue
  fi
  echo "=> Installing $name (auto-update enabled)..."
  gemini extensions install "$url" --auto-update --consent
done
