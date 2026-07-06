if test (uname) = "Darwin"
    fish_add_path -g /opt/homebrew/bin /opt/homebrew/sbin
end
fish_add_path -mg ~/.krew/bin ~/go/bin /usr/local/bin /usr/local/sbin
fish_add_path -mg ~/.local/share/mise/shims
fish_add_path -mg ~/.local/bin
