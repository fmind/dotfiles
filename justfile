# https://just.systems/man/en/

# %% CONFIGS

sudo := "false"
image := "fmind/shell:latest"
shell_path := if os() == "macos" { "/bin/zsh" } else { "/usr/bin/zsh" }
apply_become := if os() == "macos" { "--become-user " + env("USER") } else if sudo == "true" { "--ask-become" } else { "" }

# %% TASKS

# List tasks
default:
    @just --list

# Apply all roles
apply:
    ansible-playbook {{ apply_become }} -i inventory.ini site.yml

# Check all roles
check:
    ansible-playbook -i inventory.ini site.yml --syntax-check

# Build and run docker
docker:
    docker build -t {{ image }} .
    docker run --rm -it {{ image }}

# Install repository dependencies
install:
    python3 -m pip install --no-cache-dir --user pipx
    pipx install ansible --include-deps
    pipx inject ansible pipx

# Change the default shell after apply
shell:
    chsh -s {{ shell_path }}
