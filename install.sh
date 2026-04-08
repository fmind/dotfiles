#!/usr/bin/env bash
set -euo pipefail

echo "=> Starting dotfiles bootstrap..."

export PATH="$HOME/.local/bin:$HOME/.local/share/mise/bin:$HOME/.local/share/mise/shims:$PATH"
export MISE_YES=1

mkdir -p "$HOME/.local/bin"

if ! command -v git >/dev/null || ! command -v curl >/dev/null; then
    echo "=> Installing base dependencies..."
    if [ "$(uname -s)" = "Linux" ]; then
        SUDO=$([ "$(id -u)" -ne 0 ] && command -v sudo >/dev/null && echo "sudo" || echo "")
        $SUDO apt-get update -qq
        $SUDO apt-get install -yq --no-install-recommends git curl ca-certificates
    elif [ "$(uname -s)" = "Darwin" ]; then
        if ! command -v brew >/dev/null; then
            /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
        fi
        eval "$(/opt/homebrew/bin/brew shellenv 2>/dev/null || /usr/local/bin/brew shellenv 2>/dev/null || true)"
    fi
fi

if ! command -v chezmoi >/dev/null; then
    echo "=> Installing chezmoi..."
    curl -sS https://get.chezmoi.io | bash -s -- -b "$HOME/.local/bin"
fi

if ! command -v mise >/dev/null; then
    echo "=> Installing mise..."
    curl -sS https://mise.run | bash
fi

echo "=> Applying dotfiles..."
if [ -d "$HOME/.local/share/chezmoi" ]; then
    chezmoi init --apply --force --source "$HOME/.local/share/chezmoi"
else
    chezmoi init --apply --force fmind
fi

echo "=> Trusting mise configurations..."
mise trust -y "$HOME/.local/share/chezmoi" "$HOME/.config/mise/config.toml" >/dev/null 2>&1 || true

echo "=> Bootstrap complete! Run 'mise run toolchain' to setup the full environment."
