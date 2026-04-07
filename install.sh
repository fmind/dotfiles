#!/bin/bash
set -e

echo "Starting dotfiles installation..."

# Configure PATH
mkdir -p "$HOME/.local/bin"
export PATH="$HOME/.local/bin:$HOME/.local/share/mise/bin:$HOME/.local/share/mise/shims:$PATH"

# Install dependencies (macOS)
if [ "$(uname -s)" = "Darwin" ]; then
    if ! command -v brew >/dev/null 2>&1; then
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
        [ -x /opt/homebrew/bin/brew ] && eval "$(/opt/homebrew/bin/brew shellenv)"
        [ -x /usr/local/bin/brew ] && eval "$(/usr/local/bin/brew shellenv)"
    fi
    brew install git curl
fi

# Install dependencies (Debian/Ubuntu)
if command -v apt-get >/dev/null 2>&1; then
    SUDO=$([ "$(id -u)" -ne 0 ] && echo "sudo" || echo "")
    $SUDO apt-get update -qq && $SUDO apt-get install -yq git curl ca-certificates
fi

# Bootstrap core
command -v chezmoi >/dev/null 2>&1 || curl -sS https://get.chezmoi.io | bash -s -- -b "$HOME/.local/bin"
command -v mise >/dev/null 2>&1 || curl -sS https://mise.run | bash

# Initialize dotfiles
if [ -d "$HOME/dotfiles" ]; then
    chezmoi init --apply --promptDefaults --source "$HOME/dotfiles"
else
    chezmoi init --apply --promptDefaults fmind
fi

# Trust the repository
if command -v mise >/dev/null 2>&1 && [ -f "$HOME/dotfiles/mise.toml" ]; then
    mise trust --yes "$HOME/dotfiles/mise.toml"
fi

# Install tools with mise
if command -v mise >/dev/null 2>&1; then
    mise install -y
fi

# Cloud Shell specific config
if [ "$CLOUD_SHELL" = true ]; then
    echo '#!/bin/bash' > "$HOME/.customize_environment"
    echo 'sudo apt-get update -qq && sudo apt-get install -yq fish' >> "$HOME/.customize_environment"
    chmod +x "$HOME/.customize_environment"
fi

echo "Installation complete! Please restart your shell."
