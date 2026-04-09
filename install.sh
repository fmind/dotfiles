#!/usr/bin/env bash
set -euo pipefail

# Bootstrap dotfiles with chezmoi and mise.
# Prerequisites: git, curl (and ca-certificates on Linux).
# Targets: Linux (Debian/Ubuntu), macOS (Apple Silicon), Docker.

export PATH="$HOME/.local/bin:$HOME/.local/share/mise/bin:$HOME/.local/share/mise/shims:$PATH"
SOURCE_DIR="$HOME/.local/share/chezmoi"

# Load Homebrew on macOS
if [ "$(uname -s)" = "Darwin" ]; then
    eval "$(/opt/homebrew/bin/brew shellenv 2>/dev/null || true)"
fi

# Check prerequisites
for cmd in git curl; do
    command -v "$cmd" >/dev/null || { echo "Error: $cmd is required."; exit 1; }
done

# Install chezmoi
command -v chezmoi >/dev/null || {
    echo "=> Installing chezmoi..."
    mkdir -p "$HOME/.local/bin"
    curl -fsSL https://get.chezmoi.io | bash -s -- -b "$HOME/.local/bin"
}

# Install mise
command -v mise >/dev/null || {
    echo "=> Installing mise..."
    curl -fsSL https://mise.run | bash
}

# Apply dotfiles
echo "=> Applying dotfiles..."
if [ -d "$SOURCE_DIR" ]; then
    chezmoi init --apply --source "$SOURCE_DIR"
else
    chezmoi init --apply fmind
fi

# Trust mise configs
for config in "$SOURCE_DIR/mise.toml" "$HOME/.config/mise/config.toml"; do
    [ -f "$config" ] && mise trust -y "$config"
done

echo "=> Done! Run 'mise run tools' to install the full toolchain."
