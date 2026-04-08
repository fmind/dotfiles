#!/usr/bin/env bash
set -euo pipefail

echo "Starting dotfiles bootstrap..."

export PATH="$HOME/.local/bin:$HOME/.local/share/mise/bin:$HOME/.local/share/mise/shims:$PATH"
mkdir -p "$HOME/.local/bin"

BOOTSTRAP_SOURCE_DIR=""

install_linux_deps() {
    if [ "$(uname -s)" != "Linux" ]; then
        return
    fi
    if ! command -v apt-get >/dev/null 2>&1; then
        return
    fi
    local packages=()
    command -v update-ca-certificates >/dev/null 2>&1 || packages+=(ca-certificates)
    command -v git >/dev/null 2>&1 || packages+=(git)
    command -v curl >/dev/null 2>&1 || packages+=(curl)
    if [ ${#packages[@]} -eq 0 ]; then
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
    echo "Installing Linux bootstrap dependencies (${packages[*]})..."
    $sudo_cmd apt-get update -qq
    $sudo_cmd apt-get install -yq "${packages[@]}"
}

install_macos_deps() {
    if [ "$(uname -s)" != "Darwin" ]; then
        return
    fi
    if ! command -v brew >/dev/null 2>&1; then
        [ -x /opt/homebrew/bin/brew ] && eval "$(/opt/homebrew/bin/brew shellenv)"
        [ -x /usr/local/bin/brew ] && eval "$(/usr/local/bin/brew shellenv)"
    fi
    if ! command -v brew >/dev/null 2>&1; then
        echo "Installing Homebrew..."
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
        [ -x /opt/homebrew/bin/brew ] && eval "$(/opt/homebrew/bin/brew shellenv)"
        [ -x /usr/local/bin/brew ] && eval "$(/usr/local/bin/brew shellenv)"
    fi
    local packages=()
    command -v git >/dev/null 2>&1 || packages+=(git)
    command -v curl >/dev/null 2>&1 || packages+=(curl)
    if [ ${#packages[@]} -eq 0 ]; then
        return
    fi
    echo "Installing macOS bootstrap dependencies (${packages[*]})..."
    brew install "${packages[@]}"
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
    local source_dir=""
    if source_dir="$(resolve_source_dir)"; then
        echo "Using local dotfiles source: $source_dir"
        BOOTSTRAP_SOURCE_DIR="$source_dir"
        chezmoi init --apply --source "$source_dir"
        return
    fi

    echo "Using remote dotfiles source: fmind"
    chezmoi init --apply fmind
    BOOTSTRAP_SOURCE_DIR="$(chezmoi source-path 2>/dev/null || true)"
}

trust_mise_config() {
    local config_file="$HOME/.config/mise/config.toml"

    if ! command -v mise >/dev/null 2>&1 || [ ! -f "$config_file" ]; then
        return
    fi

    echo "Trusting applied mise config..."
    if ! mise trust "$config_file" >/dev/null 2>&1; then
        echo "Warning: could not trust $config_file automatically. Run 'mise trust \"$config_file\"' manually."
    fi
}

trust_mise_source_manifest() {
    local source_dir="${BOOTSTRAP_SOURCE_DIR:-}"
    local config_file=""

    if ! command -v mise >/dev/null 2>&1; then
        return
    fi

    if [ -z "$source_dir" ] && command -v chezmoi >/dev/null 2>&1; then
        source_dir="$(chezmoi source-path 2>/dev/null || true)"
    fi

    if [ -z "$source_dir" ]; then
        return
    fi

    config_file="$source_dir/mise.toml"
    if [ ! -f "$config_file" ]; then
        return
    fi

    echo "Trusting bootstrap mise manifest..."
    if ! mise trust "$config_file" >/dev/null 2>&1; then
        echo "Warning: could not trust $config_file automatically. Run 'mise trust \"$config_file\"' manually."
    fi
}

resolve_source_dir() {
    if [ -n "${CHEZMOI_SOURCE_DIR:-}" ] && [ -f "$CHEZMOI_SOURCE_DIR/.chezmoi.toml.tmpl" ]; then
        printf '%s\n' "$CHEZMOI_SOURCE_DIR"
        return 0
    fi

    local script_path="${BASH_SOURCE[0]:-}"
    if [ -n "$script_path" ] && [ -f "$script_path" ]; then
        local candidate
        candidate="$(cd "$(dirname "$script_path")" && pwd)"
        if [ -f "$candidate/.chezmoi.toml.tmpl" ]; then
            printf '%s\n' "$candidate"
            return 0
        fi
    fi

    if [ -f "$HOME/dotfiles/.chezmoi.toml.tmpl" ]; then
        printf '%s\n' "$HOME/dotfiles"
        return 0
    fi

    return 1
}

print_next_steps() {
    echo
    echo "Bootstrap complete. Open a new shell, then install the managed toolchain:"
    echo '  mise -C "$HOME" install node python && mise -C "$HOME" install'
    echo "If you already have GitHub CLI, run 'gh auth login' first to reduce GitHub API rate limits."
    echo "If you use a token instead, export GITHUB_TOKEN before running the install."
    echo "Start fish after the install if you want the managed shell."
}

install_linux_deps
install_macos_deps
install_core_tools
apply_dotfiles
trust_mise_source_manifest
trust_mise_config
print_next_steps

echo "Dotfiles bootstrap complete."
