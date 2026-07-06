set -g fish_greeting ''

if test -e ~/.private.fish
    source ~/.private.fish
end

if status is-interactive; and command -q fastfetch
    if set -q ZELLIJ; or set -q SSH_TTY
        fastfetch
    end
end
