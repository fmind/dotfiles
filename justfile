# https://just.systems/man/en/

image := "fmind/shell:latest"
sudo := "false"

shell_path := if os() == "macos" { "/bin/zsh" } else { "/usr/bin/zsh" }
apply_become := if os() == "macos" {
    "--become-user " + env_var("USER")
} else if sudo == "true" {
    "--ask-become"
} else {
    ""
}

# Run default task
default: apply

# Apply configuration
apply:
    ansible-playbook {{apply_become}} -i inventory.ini site.yml

# Check configuration
check:
    ansible-playbook -i inventory.ini site.yml --syntax-check

# Build and run docker image
docker:
    docker build -t {{image}} .
    docker run --rm -it {{image}}

# Install repository dependencies
install:
    python3 -m pip install --no-cache-dir --user pipx
    pipx install ansible --include-deps
    pipx inject ansible pipx

# Link agent skills to AI tools
skills:
    mkdir -p .gemini/ .github/
    ln -sf ../.agent/skills/ .gemini/skills
    ln -sf ../.agent/skills/ .github/skills

# Change the default shell after apply
shell:
    chsh -s {{shell_path}}
