#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
readonly SCRIPT_DIR
REPO_ROOT=$(cd -- "${SCRIPT_DIR}/../.." && pwd)
readonly REPO_ROOT

fixture_root=$(mktemp -d /tmp/mise-trust-fixture.XXXXXX)
cleanup() {
  [[ ${fixture_root} == /tmp/mise-trust-fixture.* && -d ${fixture_root} && ! -L ${fixture_root} ]] || {
    printf 'refusing to remove invalid fixture path: %s\n' "${fixture_root}" >&2
    return 1
  }
  rm -rf -- "${fixture_root}"
}
trap cleanup EXIT
mkdir -p "${fixture_root}/.github/scripts" "${fixture_root}/dot"
cp "${SCRIPT_DIR}"/*.sh "${fixture_root}/.github/scripts/"
cp "${REPO_ROOT}/mise.toml" "${fixture_root}/mise.toml"
cp "${REPO_ROOT}/dot/mise.toml" "${fixture_root}/dot/mise.toml"
git -C "${fixture_root}" init --quiet
git -C "${fixture_root}" add .github/scripts/trust-mise.sh mise.toml dot/mise.toml

export MISE_DATA_DIR="${fixture_root}/mise-data"
export MISE_CONFIG_DIR="${fixture_root}/mise-config"
export MISE_CACHE_DIR="${fixture_root}/mise-cache"
mise -C "${fixture_root}" trust -y mise.toml
mise -C "${fixture_root}" run trust:repo

for config in mise.toml dot/mise.toml; do
  trust_state=$(mise trust --show "${fixture_root}/${config}")
  [[ ${trust_state} == *': trusted' ]] || {
    printf 'expected trusted config: %s\n' "${config}" >&2
    exit 1
  }
done

mkdir "${fixture_root}/nested"
printf '[tools]\n' >"${fixture_root}/nested/mise.toml"
git -C "${fixture_root}" add nested/mise.toml
if "${fixture_root}/.github/scripts/trust-mise.sh" >/dev/null 2>&1; then
  printf 'unexpected nested mise config was trusted without inventory review\n' >&2
  exit 1
fi

printf 'PASS: clean-clone mise trust\n'
