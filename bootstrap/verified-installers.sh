#!/usr/bin/env bash

# Update these pins only from `repos/jdx/mise/releases/latest` on api.github.com;
# GitHub exposes each immutable release asset's SHA-256 digest in that response.
readonly MISE_BOOTSTRAP_VERSION="2026.7.18"
# Update this pin only from Google's platform manifests under the official
# antigravity-cli-auto-updater Cloud Run service. The versioned GCS artifacts
# have SHA-512 digests but no detached signature, so both values stay reviewable here.
readonly ANTIGRAVITY_BOOTSTRAP_VERSION="1.1.9"

bootstrap_mise_metadata() {
  local os_name=$1
  local architecture=$2
  local asset digest

  case "${os_name}:${architecture}" in
  Linux:x86_64 | Linux:amd64)
    asset="mise-v${MISE_BOOTSTRAP_VERSION}-linux-x64"
    digest="e1fc2899f2bc7dfe9a3553f4c2d5944cf69d11b5f561545504a20f1e11cd6cc5"
    ;;
  Linux:aarch64 | Linux:arm64)
    asset="mise-v${MISE_BOOTSTRAP_VERSION}-linux-arm64"
    digest="7d96bad004ba706d7ce4e999113f61296ff317ad23754471743be883817e8a75"
    ;;
  Darwin:x86_64 | Darwin:amd64)
    asset="mise-v${MISE_BOOTSTRAP_VERSION}-macos-x64"
    digest="89a8a7d0dd0536da4666d51eb7d9a2a3ffbc0546f00efa9d36041958b642cb48"
    ;;
  Darwin:aarch64 | Darwin:arm64)
    asset="mise-v${MISE_BOOTSTRAP_VERSION}-macos-arm64"
    digest="ad914875be24906afc35bca71c2da0f8ef0f74c450937a9701c03e0e992b6be6"
    ;;
  *)
    echo "Unsupported mise bootstrap platform: ${os_name}/${architecture}" >&2
    return 1
    ;;
  esac

  printf '%s|%s|%s\n' "${asset}" "https://github.com/jdx/mise/releases/download/v${MISE_BOOTSTRAP_VERSION}/${asset}" "${digest}"
}

bootstrap_antigravity_metadata() {
  local os_name=$1
  local architecture=$2
  local asset url digest

  case "${os_name}:${architecture}" in
  Linux:x86_64 | Linux:amd64)
    asset="cli_linux_x64.tar.gz"
    url="https://storage.googleapis.com/antigravity-public/antigravity-cli/1.1.9-6572839516635136/linux-x64/${asset}"
    digest="3bebfd6fdaa43fff77d33e12927f3db2b1449b008e4398dbb986ea5ee73c55fce512de22d9a711855464ec4fcfc37ea85113e47248a610e53c0e6d5e5297ed95"
    ;;
  Linux:aarch64 | Linux:arm64)
    asset="cli_linux_arm64.tar.gz"
    url="https://storage.googleapis.com/antigravity-public/antigravity-cli/1.1.9-6572839516635136/linux-arm/${asset}"
    digest="9d28ab7e767d7625a88ec74d72148747ce7ac32089c1a78f418edbed5a35a1743c05e86495184b640110a06c5d28e8933e352ea15e0836dd2d64c8e1cf68f199"
    ;;
  Darwin:x86_64 | Darwin:amd64)
    asset="cli_mac_x64.tar.gz"
    url="https://storage.googleapis.com/antigravity-public/antigravity-cli/1.1.9-6572839516635136/darwin-x64/${asset}"
    digest="2e61abdf7d627e6ad24bfefeed8bb35a00a538adc00398106270e635015f78969327f776791053d2cb6b92912824912fa2eaf65ac9224fafcd3d0aad7ebd8e8d"
    ;;
  Darwin:aarch64 | Darwin:arm64)
    asset="cli_mac_arm64.tar.gz"
    url="https://storage.googleapis.com/antigravity-public/antigravity-cli/1.1.9-6572839516635136/darwin-arm/${asset}"
    digest="4bb5c759cec7e5aa7738f9d5259bb29bc8899fb616a0979be5b192ddade9f143d493ede30dcc1475298ef4060c013bf75a992adc041be8955762b2c5a3061f1b"
    ;;
  *)
    echo "Unsupported Antigravity bootstrap platform: ${os_name}/${architecture}" >&2
    return 1
    ;;
  esac

  printf '%s|%s|%s\n' "${asset}" "${url}" "${digest}"
}

bootstrap_hash() {
  local algorithm=$1
  local path=$2

  case "${algorithm}" in
  sha256)
    if command -v sha256sum >/dev/null 2>&1; then
      sha256sum "${path}" | awk '{print $1}'
    else
      shasum -a 256 "${path}" | awk '{print $1}'
    fi
    ;;
  sha512)
    if command -v sha512sum >/dev/null 2>&1; then
      sha512sum "${path}" | awk '{print $1}'
    else
      shasum -a 512 "${path}" | awk '{print $1}'
    fi
    ;;
  *)
    echo "Unsupported bootstrap checksum algorithm: ${algorithm}" >&2
    return 1
    ;;
  esac
}

bootstrap_verify_artifact() {
  local name=$1
  local path=$2
  local algorithm=$3
  local expected=$4
  local actual

  if [ -L "${path}" ] || [ ! -f "${path}" ]; then
    echo "Integrity boundary failed for ${name}: artifact is not a regular file." >&2
    return 1
  fi
  actual="$(bootstrap_hash "${algorithm}" "${path}")" || {
    echo "Integrity boundary failed for ${name}: ${algorithm} tooling is unavailable." >&2
    return 1
  }
  if [ "${actual}" != "${expected}" ]; then
    echo "Integrity boundary failed for ${name}: ${algorithm} mismatch." >&2
    return 1
  fi
}

bootstrap_download() {
  local url=$1
  local destination=$2

  case "${url}" in
  https://*) ;;
  *)
    echo "Transport boundary failed: bootstrap downloads require HTTPS." >&2
    return 1
    ;;
  esac
  curl --proto '=https' --proto-redir '=https' --tlsv1.2 --fail --location --silent --show-error --output "${destination}" "${url}" || {
    echo "Transport boundary failed: verified bootstrap artifact download failed." >&2
    return 1
  }
}

bootstrap_secure_directory() {
  local name=$1
  local path=$2

  if [ -L "${path}" ]; then
    echo "Integrity boundary failed for ${name}: managed directory is a symbolic link." >&2
    return 1
  fi
  if [ -e "${path}" ] && [ ! -d "${path}" ]; then
    echo "Integrity boundary failed for ${name}: managed directory is not a directory." >&2
    return 1
  fi
  mkdir -p "${path}"
  chmod 700 "${path}"
}

bootstrap_cached_artifact() (
  set -euo pipefail
  local name=$1
  local version=$2
  local asset=$3
  local url=$4
  local algorithm=$5
  local expected=$6
  local cache_root="${HOME}/.cache/dot"
  local cache_dir="${cache_root}/bootstrap/${name}/${version}"
  local cached="${cache_dir}/${asset}"
  local staging_dir=""

  umask 077
  mkdir -p "${HOME}/.cache"
  bootstrap_secure_directory "${name}" "${cache_root}"
  bootstrap_secure_directory "${name}" "${cache_root}/bootstrap"
  bootstrap_secure_directory "${name}" "${cache_root}/bootstrap/${name}"
  bootstrap_secure_directory "${name}" "${cache_dir}"
  if [ -f "${cached}" ] && [ ! -L "${cached}" ] && bootstrap_verify_artifact "${name}" "${cached}" "${algorithm}" "${expected}"; then
    printf '%s\n' "${cached}"
    exit 0
  fi

  staging_dir="$(mktemp -d "${cache_dir}/.download.XXXXXX")"
  trap 'rm -rf "${staging_dir:?}"' EXIT
  if ! bootstrap_download "${url}" "${staging_dir}/${asset}"; then
    exit 1
  fi
  if ! bootstrap_verify_artifact "${name}" "${staging_dir}/${asset}" "${algorithm}" "${expected}"; then
    exit 1
  fi
  mv -f "${staging_dir}/${asset}" "${cached}"
  chmod 600 "${cached}"
  printf '%s\n' "${cached}"
)

bootstrap_version_at_least() {
  local current=${1#v}
  local required=${2#v}
  local -a current_parts required_parts
  local index current_part required_part length

  [[ ${current} =~ ^[0-9]+([.][0-9]+)*$ ]] || return 1
  [[ ${required} =~ ^[0-9]+([.][0-9]+)*$ ]] || return 1
  IFS=. read -r -a current_parts <<<"${current}"
  IFS=. read -r -a required_parts <<<"${required}"
  length=${#current_parts[@]}
  if [ "${#required_parts[@]}" -gt "${length}" ]; then
    length=${#required_parts[@]}
  fi
  for ((index = 0; index < length; index++)); do
    current_part=${current_parts[index]:-0}
    required_part=${required_parts[index]:-0}
    if ((10#${current_part} > 10#${required_part})); then
      return 0
    fi
    if ((10#${current_part} < 10#${required_part})); then
      return 1
    fi
  done
  return 0
}

bootstrap_command_version() {
  local command_path=$1
  "${command_path}" --version 2>/dev/null | sed -n '1s/^[^0-9v]*v\{0,1\}\([0-9][0-9.]*\).*$/\1/p'
}

bootstrap_publish_binary() (
  set -euo pipefail
  local name=$1
  local artifact=$2
  local target=$3
  local expected_version=$4
  local target_dir
  local staging
  local actual_version

  target_dir="$(dirname "${target}")"
  mkdir -p "$(dirname "${target_dir}")"
  if [ -L "${target_dir}" ]; then
    echo "Install boundary failed for ${name}: target directory is a symbolic link." >&2
    return 1
  fi
  mkdir -p "${target_dir}"
  if [ -L "${target}" ]; then
    echo "Install boundary failed for ${name}: target is a symbolic link." >&2
    return 1
  fi
  staging="$(mktemp "${target_dir}/.${name}.XXXXXX")"
  trap 'rm -f "${staging:?}"' EXIT
  install -m 0755 "${artifact}" "${staging}"
  actual_version="$(bootstrap_command_version "${staging}")"
  if [ "${actual_version}" != "${expected_version}" ]; then
    echo "Provenance boundary failed for ${name}: verified artifact reported an unexpected version." >&2
    return 1
  fi
  mv -f "${staging}" "${target}"
)

bootstrap_install_mise() {
  local installed=""
  local metadata asset url digest artifact
  local target="${HOME}/.local/bin/mise"
  local command_path os_name architecture

  if command -v mise >/dev/null 2>&1; then
    command_path="$(command -v mise)"
    installed="$(bootstrap_command_version "${command_path}")"
    if bootstrap_version_at_least "${installed}" "${MISE_BOOTSTRAP_VERSION}"; then
      echo "=> mise ${installed} already satisfies bootstrap pin ${MISE_BOOTSTRAP_VERSION}."
      return 0
    fi
  fi
  os_name="$(uname -s)"
  architecture="$(uname -m)"
  metadata="$(bootstrap_mise_metadata "${os_name}" "${architecture}")"
  IFS='|' read -r asset url digest <<<"${metadata}"
  artifact="$(bootstrap_cached_artifact mise "${MISE_BOOTSTRAP_VERSION}" "${asset}" "${url}" sha256 "${digest}")"
  bootstrap_publish_binary mise "${artifact}" "${target}" "${MISE_BOOTSTRAP_VERSION}"
  hash -r
  echo "=> Installed verified mise ${MISE_BOOTSTRAP_VERSION}."
}

bootstrap_install_antigravity() (
  set -euo pipefail
  local installed=""
  local metadata asset url digest archive extraction_dir binary
  local target="${HOME}/.local/bin/agy"
  local command_path os_name architecture archive_entries

  if command -v agy >/dev/null 2>&1; then
    command_path="$(command -v agy)"
    installed="$(bootstrap_command_version "${command_path}")"
    if bootstrap_version_at_least "${installed}" "${ANTIGRAVITY_BOOTSTRAP_VERSION}"; then
      echo "=> Antigravity CLI ${installed} already satisfies bootstrap pin ${ANTIGRAVITY_BOOTSTRAP_VERSION}."
      return 0
    fi
  fi
  os_name="$(uname -s)"
  architecture="$(uname -m)"
  metadata="$(bootstrap_antigravity_metadata "${os_name}" "${architecture}")"
  IFS='|' read -r asset url digest <<<"${metadata}"
  archive="$(bootstrap_cached_artifact antigravity "${ANTIGRAVITY_BOOTSTRAP_VERSION}" "${asset}" "${url}" sha512 "${digest}")"
  if ! archive_entries="$(tar -tzf "${archive}")"; then
    echo "Provenance boundary failed for Antigravity CLI: archive could not be inspected." >&2
    return 1
  fi
  if [ "${archive_entries}" != "antigravity" ]; then
    echo "Provenance boundary failed for Antigravity CLI: archive layout is not the reviewed single binary." >&2
    return 1
  fi
  extraction_dir="$(mktemp -d "${HOME}/.cache/dot/bootstrap/antigravity/${ANTIGRAVITY_BOOTSTRAP_VERSION}/.extract.XXXXXX")"
  trap 'rm -rf "${extraction_dir:?}"' EXIT
  tar -xzf "${archive}" -C "${extraction_dir}" antigravity
  binary="${extraction_dir}/antigravity"
  if [ -L "${binary}" ] || [ ! -f "${binary}" ]; then
    echo "Provenance boundary failed for Antigravity CLI: extracted payload is not a regular file." >&2
    return 1
  fi
  bootstrap_publish_binary antigravity "${binary}" "${target}" "${ANTIGRAVITY_BOOTSTRAP_VERSION}"
  if [ "${os_name}" = "Darwin" ] && command -v xattr >/dev/null 2>&1; then
    xattr -d com.apple.quarantine "${target}" 2>/dev/null || true
  fi
  echo "=> Installed verified Antigravity CLI ${ANTIGRAVITY_BOOTSTRAP_VERSION}."
)
