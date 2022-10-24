# dotfiles

Configuration of my favorite command-line editors, shells, and tools.

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
chsh -s /usr/bin/zsh
# On MacOS systems
chsh -s /bin/zsh
```
