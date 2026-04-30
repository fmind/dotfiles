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

for url in "${extensions[@]}"; do
  name="${url##*/}"
  if output=$(gemini extensions install "$url" --auto-update --consent 2>&1); then
    echo "Installed $name"
  elif ! printf '%s' "$output" | grep -q 'already installed'; then
    echo "  install failed for $name"
  fi
done
