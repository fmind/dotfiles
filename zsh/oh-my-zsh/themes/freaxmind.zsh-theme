function zle-line-init zle-keymap-select {
    VIM_PROMPT="%{$fg_bold[yellow]%} ℕ %{$reset_color%}"
    RPS1="${${KEYMAP/vicmd/$VIM_PROMPT}/(main|viins)/} $(dirinfos) $EPS1"
    zle reset-prompt
}

function dirinfos() {
    NDIR=`find . -maxdepth 1 -type d | wc -l`
    NFILE=`find . -maxdepth 1 -type f | wc -l`
    PERM=`stat -c %a /etc`
    ISGIT=`git branch >/dev/null 2>/dev/null && echo '∢'`

    echo "%{$fg_bold[magenta]%}${ISGIT}% %{$reset_color%} %{%b%}%{$fg[white]%}δ:${NDIR} φ:${NFILE} π:${PERM}%{$reset_color%}"
}


PROMPT='%{$fg[green]%}∑%n%{$fg[red]%}∀%m%{$fg[blue]%}%{%B%}∧%2~%{$fg[white]%}%{%b%}%(!.√.∫)%{$reset_color%} '
RPROMPT='$(dirinfos)'

zle -N zle-line-init
zle -N zle-keymap-select
