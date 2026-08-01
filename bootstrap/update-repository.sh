#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
readonly repo_root
cd "${repo_root}"

isolated_root="$(mktemp -d)"
readonly isolated_root
trap 'rm -rf "$isolated_root"' EXIT

export HOME="${isolated_root}/home"
export XDG_CACHE_HOME="${isolated_root}/cache"
export XDG_CONFIG_HOME="${isolated_root}/config"
export XDG_DATA_HOME="${isolated_root}/data"
export XDG_STATE_HOME="${isolated_root}/state"
mkdir -p "${HOME}" "${XDG_CACHE_HOME}" "${XDG_CONFIG_HOME}" "${XDG_DATA_HOME}" "${XDG_STATE_HOME}"

echo "Repository update ecosystems: mise tools"
echo "Expected repository changes: mise.toml mise.lock"

# Isolate mise completely from the workstation while still allowing the reviewed
# checkout to install temporary tools needed to resolve exact current pins.
mise trust -y "${repo_root}/mise.toml"
mise upgrade --bump --local --yes
mise lock --bump --yes

changed_files="$(git diff --name-only -- mise.toml mise.lock)"
if [[ -n ${changed_files} ]]; then
  echo "Changed repository files:"
  printf '%s\n' "${changed_files}"
else
  echo "Changed repository files: none"
fi
