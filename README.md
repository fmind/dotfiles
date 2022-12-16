# dotfiles

Configuration of my favorite languages, editors, shells, and tools.

# Requirements

- ansible
- pipx

# Installation

```bash
# On Linux systems
ansible-playbook -K site.yml
# On MacOS systems
ansible-playbook site.yml --become-user=$USER
```

**On Mac OSX**:
- To enable the unarchive module: `brew install gnu-tar`

# Configuration

```bash
# On Linux systems
chsh -s /usr/bin/fish
# On MacOS systems
echo "/usr/local/bin/fish" | sudo tee -a /etc/shells
chsh -s /usr/local/bin/fish
```
