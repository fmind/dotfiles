# Run ls (eza) when the working dir changes.
function __auto_ls --on-variable PWD
    status is-interactive; and ls
end
