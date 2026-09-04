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

# Install dot
echo "=> Installing dot..."
if [ ! -d "${SOURCE_DIR}" ]; then
  chezmoi init --force https://github.com/fmind/dot.git --source "${SOURCE_DIR}" "$@"
else
  echo "=> Updating dot repository..."
  if [ "${SKIP_GIT_PULL:-}" = "true" ] || [ "${CI:-}" = "true" ]; then
    echo "=> Skipping git pull as requested by environment variable."
  else
    git -C "${SOURCE_DIR}" pull --ff-only
  fi
  chezmoi init --force --source "${SOURCE_DIR}" "$@"
fi

# Trust every reviewed config in the checkout before a task can load it. Trust is
# per file, so the nested Go module config needs its own grant: without it the
# first `mise -C dot ...` of the bootstrap stops on an untrusted config.
echo "=> Trusting mise configs..."
mise trust -y "${SOURCE_DIR}/mise.toml"
mise trust -y "${SOURCE_DIR}/dot/mise.toml"

# Complete the ordered bootstrap: apply, trust, tools, hooks, and editor.
echo "=> Completing environment bootstrap..."
mise -C "${SOURCE_DIR}" run install

echo "=> Install complete! You are ready to go."
