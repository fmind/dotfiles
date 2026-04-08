#!/usr/bin/env bash
set -euo pipefail

echo "Starting dotfiles installation..."

export PATH="$HOME/.local/bin:$HOME/.local/share/mise/bin:$HOME/.local/share/mise/shims:$PATH"
mkdir -p "$HOME/.local/bin"

install_linux_deps() {
    if command -v curl >/dev/null 2>&1 && command -v git >/dev/null 2>&1; then
        return
    fi
    if ! command -v apt-get >/dev/null 2>&1; then
        return
    fi
    local sudo_cmd=""
    if [ "$(id -u)" -ne 0 ]; then
        if ! command -v sudo >/dev/null 2>&1; then
            echo "Skipping apt packages: sudo is not available."
            return
        fi
        sudo_cmd="sudo"
    fi
    echo "Installing linux dependencies (git, curl)..."
    $sudo_cmd apt-get update -qq
    $sudo_cmd apt-get install -yq git curl ca-certificates
}

install_macos_deps() {
    if [ "$(uname -s)" != "Darwin" ]; then
        return
    fi
    if command -v curl >/dev/null 2>&1 && command -v git >/dev/null 2>&1; then
        return
    fi
    if ! command -v brew >/dev/null 2>&1; then
        echo "Installing Homebrew..."
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
        [ -x /opt/homebrew/bin/brew ] && eval "$(/opt/homebrew/bin/brew shellenv)"
        [ -x /usr/local/bin/brew ] && eval "$(/usr/local/bin/brew shellenv)"
    fi
    echo "Installing macOS dependencies (git, curl)..."
    brew install git curl
}

install_core_tools() {
    if ! command -v chezmoi >/dev/null 2>&1; then
        echo "Installing chezmoi..."
        curl -sS https://get.chezmoi.io | bash -s -- -b "$HOME/.local/bin"
    fi
    if ! command -v mise >/dev/null 2>&1; then
        echo "Installing mise..."
        curl -sS https://mise.run | bash
    fi
}

apply_dotfiles() {
    echo "Applying dotfiles configuration..."
    if [ -d "$HOME/dotfiles" ]; then
        chezmoi init --apply --source "$HOME/dotfiles"
    else
        chezmoi init --apply fmind
    fi
}

install_mise_tools() {
    echo "Installing mise tools..."
    local source_dir
    if ! command -v mise >/dev/null 2>&1; then
        return
    fi
    source_dir="$(chezmoi source-path)"
    if [ -f "$source_dir/mise.toml" ]; then
        mise trust --yes "$source_dir/mise.toml"
    fi

    if [ -z "${GITHUB_TOKEN:-}" ]; then
        echo "WARNING: GITHUB_TOKEN is not set. You might encounter GitHub API rate limits (403 Forbidden) during 'mise install'."
        if [ -t 0 ]; then
            if ! (command -v gh >/dev/null 2>&1 && gh auth status >/dev/null 2>&1); then
                read -r -p "Would you like to login securely with the GitHub CLI (gh) now? [Y/n] " prompt
                if [[ $prompt == "y" || $prompt == "Y" || $prompt == "" ]]; then
                    if ! command -v gh >/dev/null 2>&1; then
                        echo "Installing GitHub CLI via OS package manager to avoid rate limits..."
                        if [ "$(uname -s)" = "Darwin" ]; then
                            brew install gh
                        elif command -v apt-get >/dev/null 2>&1; then
                            local sudo_cmd=""
                            [ "$(id -u)" -ne 0 ] && sudo_cmd="sudo"
                            $sudo_cmd apt-get update -qq && $sudo_cmd apt-get install -yq gh || echo "Could not install gh via apt-get. Proceeding with mise..."
                            if ! command -v gh >/dev/null 2>&1; then
                                mise use -g gh@latest
                            fi
                        else
                            mise use -g gh@latest
                        fi
                    fi
                    echo "Starting GitHub CLI authentication..."
                    gh auth login
                else
                    echo "Skipping interactive login."
                fi
            fi
        fi
    fi

    # Suppress verbose output but keep warnings if necessary
    mise install -y
}

install_linux_deps
install_macos_deps
install_core_tools
apply_dotfiles
install_mise_tools

echo "Installation complete. Start fish when you want the managed shell."
