#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH='' cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
# shellcheck source=bootstrap/verified-installers.sh
. "${ROOT_DIR}/bootstrap/verified-installers.sh"

TEST_ROOT="$(mktemp -d)"
trap 'rm -rf "${TEST_ROOT:?}"' EXIT
export HOME="${TEST_ROOT}/home"
mkdir -p "${HOME}" "${TEST_ROOT}/bin"

fail() {
  echo "bootstrap test failed: $*" >&2
  exit 1
}

file_mode() {
  if stat -c '%a' "$1" >/dev/null 2>&1; then
    stat -c '%a' "$1"
  else
    stat -f '%Lp' "$1"
  fi
}

for platform in "Linux x86_64" "Linux arm64" "Darwin x86_64" "Darwin arm64"; do
  read -r os_name architecture <<<"${platform}"
  mise_metadata="$(bootstrap_mise_metadata "${os_name}" "${architecture}")"
  antigravity_metadata="$(bootstrap_antigravity_metadata "${os_name}" "${architecture}")"
  [[ ${mise_metadata} =~ ^[^|]+\|https://github.com/jdx/mise/releases/download/v${MISE_BOOTSTRAP_VERSION}/[^|]+\|[0-9a-f]{64}$ ]] || fail "invalid mise metadata for ${platform}"
  [[ ${antigravity_metadata} =~ ^[^|]+\|https://storage.googleapis.com/antigravity-public/[^|]+\|[0-9a-f]{128}$ ]] || fail "invalid Antigravity metadata for ${platform}"
done

if grep -E 'curl[^|]*\|[[:space:]]*(ba)?sh' "${ROOT_DIR}/install.sh" "${ROOT_DIR}/run_once_after_install-antigravity-cli.sh.tmpl"; then
  fail "bootstrap still executes a remote response through a shell"
fi

bootstrap_version_at_least 2026.7.18 2026.7.18 || fail "equal version rejected"
bootstrap_version_at_least 2026.8.0 2026.7.18 || fail "newer version rejected"
if bootstrap_version_at_least 2026.7.17 2026.7.18; then
  fail "older version accepted"
fi

fixture="${TEST_ROOT}/verified-tool"
cat >"${fixture}" <<'FIXTURE'
#!/usr/bin/env bash
echo "1.2.3"
FIXTURE
chmod 700 "${fixture}"
fixture_hash="$(bootstrap_hash sha256 "${fixture}")"

cat >"${TEST_ROOT}/bin/curl" <<'FAKE_CURL'
#!/usr/bin/env bash
set -euo pipefail
destination=""
while [ "$#" -gt 0 ]; do
  if [ "$1" = "--output" ]; then
    destination=$2
    shift
  fi
  shift
done
[ -n "$destination" ]
cp "$BOOTSTRAP_TEST_FIXTURE" "$destination"
printf 'download\n' >>"$BOOTSTRAP_TEST_CALLS"
FAKE_CURL
chmod 700 "${TEST_ROOT}/bin/curl"
export PATH="${TEST_ROOT}/bin:/usr/bin:/bin"
export BOOTSTRAP_TEST_FIXTURE="${fixture}"
export BOOTSTRAP_TEST_CALLS="${TEST_ROOT}/curl.calls"

cached="$(bootstrap_cached_artifact smoke 1.2.3 verified-tool https://example.invalid/verified-tool sha256 "${fixture_hash}")"
[ -f "${cached}" ] || fail "verified artifact was not cached"
bootstrap_cached_artifact smoke 1.2.3 verified-tool https://example.invalid/verified-tool sha256 "${fixture_hash}" >/dev/null
download_calls="$(wc -l <"${BOOTSTRAP_TEST_CALLS}")"
[ "${download_calls}" -eq 1 ] || fail "verified cache did not prevent a second download"

target="${HOME}/.local/bin/smoke"
bootstrap_publish_binary smoke "${cached}" "${target}" 1.2.3
published_version="$("${target}" --version)"
published_mode="$(file_mode "${target}")"
[ "${published_version}" = "1.2.3" ] || fail "verified binary was not published"
[ "${published_mode}" = "755" ] || fail "published binary is not executable"

if bootstrap_cached_artifact mismatch 1.0 bad https://example.invalid/bad sha256 "$(printf '0%.0s' {1..64})" >/dev/null 2>"${TEST_ROOT}/mismatch.err"; then
  fail "checksum mismatch was accepted"
fi
grep -q "Integrity boundary failed" "${TEST_ROOT}/mismatch.err" || fail "checksum error omitted the failed boundary"

cat >"${TEST_ROOT}/bin/mise" <<'NEWER_MISE'
#!/usr/bin/env bash
echo "9999.1.0 linux-x64"
NEWER_MISE
chmod 700 "${TEST_ROOT}/bin/mise"
calls_before="$(wc -l <"${BOOTSTRAP_TEST_CALLS}")"
bootstrap_install_mise >/dev/null
calls_after="$(wc -l <"${BOOTSTRAP_TEST_CALLS}")"
[ "${calls_after}" -eq "${calls_before}" ] || fail "newer mise was silently replaced"

echo "✓ Verified installer smoke tests passed (Linux execution; macOS metadata bounded)."
