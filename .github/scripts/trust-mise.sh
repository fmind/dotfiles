#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
readonly SCRIPT_DIR
REPO_ROOT=$(cd -- "${SCRIPT_DIR}/../.." && pwd)
readonly REPO_ROOT
readonly EXPECTED_CONFIGS=$'dot/mise.toml\nmise.toml'

tracked_configs=$(git -C "${REPO_ROOT}" ls-files -- ':(glob)**/mise.toml')
discovered_configs=$(
  while IFS= read -r config; do
    case ${config} in
    skills/*/references/mise.toml) continue ;;
    *) printf '%s\n' "${config}" ;;
    esac
  done <<<"${tracked_configs}"
)
if [[ ${discovered_configs} != "${EXPECTED_CONFIGS}" ]]; then
  printf 'mise config inventory changed; update %s after reviewing the new trust boundary.\n' "${BASH_SOURCE[0]}" >&2
  diff -u <(printf '%s\n' "${EXPECTED_CONFIGS}") <(printf '%s\n' "${discovered_configs}") >&2 || true
  exit 1
fi

while IFS= read -r config; do
  mise trust -y "${REPO_ROOT}/${config}"
done <<<"${EXPECTED_CONFIGS}"
