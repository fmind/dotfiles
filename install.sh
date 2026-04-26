#!/usr/bin/env bash
set -euo pipefail

export PATH="$HOME/.local/bin:$HOME/.local/share/mise/bin:$HOME/.local/share/mise/shims:$PATH"
SOURCE_DIR="$HOME/.local/share/chezmoi"

# Install mise
command -v mise >/dev/null || {
  echo "=> Installing mise..."
  curl -fsSL https://mise.run | bash
}

# Install chezmoi
command -v chezmoi >/dev/null || {
  echo "=> Installing chezmoi..."
  mise use --global --yes chezmoi@latest
}

# Install dotfiles
echo "=> Installing dotfiles..."
if [ ! -d "$SOURCE_DIR" ]; then
  chezmoi init git@github.com:fmind/dotfiles.git --source "$SOURCE_DIR"
elif [ ! -d "$SOURCE_DIR/.git" ]; then
  git -C "$SOURCE_DIR" init -b main
  git -C "$SOURCE_DIR" remote add origin git@github.com:fmind/dotfiles.git
fi
  chezmoi init --apply --source "$SOURCE_DIR"

# Trust mise configs
echo "=> Trusting mise config ..."
mise trust -y "$SOURCE_DIR/mise.toml"
mise -C "$SOURCE_DIR" run trust

echo "=> Install complete! You are ready to go."
