#!/usr/bin/env bash
set -euo pipefail

export PATH="${HOME}/.local/bin:${HOME}/.local/share/mise/bin:${HOME}/.local/share/mise/shims:${PATH}"
SOURCE_DIR="${HOME}/.local/share/chezmoi"

# Error trap handler for clean bootstrapping diagnostics
on_error() {
  local exit_code=$?
  echo "==================================================" >&2
  echo "  ✗ Error: install.sh failed at line $1 with exit code ${exit_code}." >&2
  echo "==================================================" >&2
  echo "  Please check the following bootstrap prerequisites:" >&2
  echo "  1. Ensure you have active internet connectivity." >&2
  echo "  2. Confirm both git and curl are installed on your host." >&2
  echo "  3. On Linux: verify 'build-essential' and 'gnome-keyring' are installed." >&2
  echo "  4. Check that ~/.local/bin is writeable by your current user." >&2
  echo "==================================================" >&2
}
trap 'on_error $LINENO' ERR

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
if [ ! -d "${SOURCE_DIR}" ]; then
  chezmoi init --force https://github.com/fmind/dotfiles.git --source "${SOURCE_DIR}" "$@"
else
  echo "=> Updating dotfiles repository..."
  if [ "${SKIP_GIT_PULL:-}" = "true" ] || [ "${CI:-}" = "true" ]; then
    echo "=> Skipping git pull as requested by environment variable."
  else
    git -C "${SOURCE_DIR}" pull --ff-only
  fi
  chezmoi init --force --source "${SOURCE_DIR}" "$@"
fi

# Trust the repository config so it can drive the remaining bootstrap.
echo "=> Trusting mise config..."
mise trust -y "${SOURCE_DIR}/mise.toml"

# Complete the ordered bootstrap: apply, trust, tools, hooks, editor, and krew.
echo "=> Completing environment bootstrap..."
mise -C "${SOURCE_DIR}" run init

echo "=> Install complete! You are ready to go."
