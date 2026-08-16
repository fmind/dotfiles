# Editors
set -gx EDITOR nvim
set -gx K9S_EDITOR nvim
set -gx KUBE_EDITOR nvim
set -gx VISUAL $EDITOR

# Locales
set -gx LANG en_US.UTF-8

# Pagers
set -gx LESS -FRSXMK
set -gx LESSHISTFILE -
set -gx MANPAGER "nvim +Man!"
set -gx PAGER "bat --plain"

# Tools
set -gx CARAPACE_BRIDGES 'zsh,fish,bash'
set -gx COREPACK_ENABLE_AUTO_PIN 0
# Grok ships web_fetch off and exposes no config key for it, so the env var is the only
# switch. Without it Grok can search the web but cannot read a URL it was handed, which
# is the "verify against authoritative docs" half of the workflow.
set -gx GROK_WEB_FETCH 1
set -gx K9S_CONFIG_DIR $HOME/.config/k9s
set -gx KO_DOCKER_REPO registry.localhost:5050
set -gx RIPGREP_CONFIG_PATH $HOME/.config/ripgrep/config
set -gx TRIVY_CONFIG $HOME/.config/trivy/trivy.yaml
