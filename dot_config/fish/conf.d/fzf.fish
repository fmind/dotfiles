# FZF Integration

# Base command
set -gx FZF_DEFAULT_COMMAND 'fd --type f --hidden --exclude .git'

# Sub-commands
set -gx FZF_CTRL_T_COMMAND "$FZF_DEFAULT_COMMAND"
set -gx FZF_ALT_C_COMMAND 'fd --type d --hidden --exclude .git'

# Sub-options
set -gx FZF_ALT_C_OPTS "--preview 'eza --tree --color=always {} | head -200'"
set -gx FZF_CTRL_T_OPTS "--preview 'bat -n --color=always {}'"

# Colors and Global Options (Catppuccin Mocha)
set -gx FZF_DEFAULT_OPTS "\
--color=bg+:#313244,bg:#1e1e2e,spinner:#f5e0dc,hl:#f38ba8 \
--color=fg:#cdd6f4,header:#f38ba8,info:#cba6f7,pointer:#f5e0dc \
--color=marker:#b4befe,fg+:#cdd6f4,prompt:#cba6f7,hl+:#f38ba8 \
--color=selected-bg:#45475A \
--color=border:#6C7086,label:#CDD6F4 \
--bind ctrl-u:preview-half-page-up,ctrl-d:preview-half-page-down \
--layout=reverse"
