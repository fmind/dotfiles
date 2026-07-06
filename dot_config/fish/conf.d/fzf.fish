# Base commands
set -gx FZF_DEFAULT_COMMAND 'fd --type f --hidden --strip-cwd-prefix --exclude .git'
set -gx FZF_ALT_C_COMMAND 'fd --type d --hidden --strip-cwd-prefix --exclude .git'

# Default options
set -gx FZF_DEFAULT_OPTS \
    --border \
    "--height=50%" \
    "--info=inline" \
    "--layout=reverse" \
    "--bind=ctrl-/:toggle-preview" \
    "--color=fg:#c8d3f5,bg:#222436,hl:#ff757f" \
    "--color=fg+:#c8d3f5,bg+:#2f334d,hl+:#ff757f" \
    "--color=info:#c099ff,prompt:#c099ff,pointer:#86e1fc" \
    "--color=marker:#82aaff,spinner:#86e1fc,header:#ff757f,border:#636da6"

# Preview options
set -gx FZF_ALT_C_OPTS "--preview 'eza --tree --color=always {} | head -200'"
set -gx FZF_CTRL_T_OPTS "--preview 'bat --number --color=always --line-range=:500 {}'"
