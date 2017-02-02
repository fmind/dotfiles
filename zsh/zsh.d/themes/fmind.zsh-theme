function dirinfos() {
    NDIR=`find . -maxdepth 1 -type d | wc -l`
    NFILE=`find . -maxdepth 1 -type f | wc -l`
    PERM=`stat -c %a .`
    ISGIT=`git branch >/dev/null 2>/dev/null && echo '±'`

    echo "%{%b%}%{$fg[white]%}${ISGIT} δ${NDIR} φ${NFILE} π${PERM}%{$reset_color%}"
}

PROMPT='%{%B%}%{$fg[red]%}∀%m%{$fg[green]%}∃%n%{$fg[blue]%}∧%1d%{$fg[white]%}∫%{$reset_color%}%{%b%} '
RPROMPT='$(dirinfos)'
