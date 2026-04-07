#!/usr/bin/env bash
set -euo pipefail

echo "Starting dotfiles installation..."

mkdir -p "$HOME/.local/bin"
export PATH="$HOME/.local/bin:$HOME/.local/share/mise/bin:$HOME/.local/share/mise/shims:$PATH"

install_linux_deps() {
    local sudo_cmd=""

    if ! command -v apt-get >/dev/null 2>&1; then
        return
    fi

    if [ "$(id -u)" -ne 0 ]; then
        if ! command -v sudo >/dev/null 2>&1; then
            echo "Skipping apt packages: sudo is not available."
            return
        fi
        sudo_cmd="sudo"
    fi

    $sudo_cmd apt-get update -qq
    $sudo_cmd apt-get install -yq git curl ca-certificates libatomic1
}

install_macos_deps() {
    if [ "$(uname -s)" != "Darwin" ]; then
        return
    fi

    if ! command -v brew >/dev/null 2>&1; then
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
        [ -x /opt/homebrew/bin/brew ] && eval "$(/opt/homebrew/bin/brew shellenv)"
        [ -x /usr/local/bin/brew ] && eval "$(/usr/local/bin/brew shellenv)"
    fi

    brew install git curl
}

bootstrap_core_tools() {
    command -v chezmoi >/dev/null 2>&1 || curl -sS https://get.chezmoi.io | bash -s -- -b "$HOME/.local/bin"
    command -v mise >/dev/null 2>&1 || curl -sS https://mise.run | bash
}

apply_dotfiles() {
    if [ -d "$HOME/dotfiles" ]; then
        chezmoi init --apply --promptDefaults --source "$HOME/dotfiles"
    else
        chezmoi init --apply --promptDefaults fmind
    fi
}

install_mise_tools() {
    local source_dir

    if ! command -v mise >/dev/null 2>&1; then
        return
    fi

    source_dir="$(chezmoi source-path)"
    if [ -f "$source_dir/mise.toml" ]; then
        mise trust --yes "$source_dir/mise.toml"
    fi

    mise install -y
}

install_linux_deps
install_macos_deps
bootstrap_core_tools
apply_dotfiles
install_mise_tools

echo "Installation complete. Start fish when you want the managed shell."
