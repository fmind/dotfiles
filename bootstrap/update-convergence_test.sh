#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
fixture_root="$(mktemp -d)"
readonly repo_root fixture_root
trap 'rm -rf "$fixture_root"' EXIT

repo_fixture="${fixture_root}/repository"
fake_bin="${fixture_root}/bin"
mkdir -p "${repo_fixture}/bootstrap" "${fake_bin}"
cp "${repo_root}/bootstrap/update-repository.sh" "${repo_fixture}/bootstrap/"
printf '[tools]\nexample = "1"\n' >"${repo_fixture}/mise.toml"
printf '[tools.example]\nversion = "1"\n' >"${repo_fixture}/mise.lock"
git -C "${repo_fixture}" init -q
git -C "${repo_fixture}" config user.email test@example.com
git -C "${repo_fixture}" config user.name Test
git -C "${repo_fixture}" add .
git -C "${repo_fixture}" commit -qm fixture

cat >"${fake_bin}/mise" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "${*}" >>"${TEST_LOG}"
case "${*}" in
  'upgrade --bump --local --yes') printf '# upgraded\n' >>mise.toml ;;
  'lock --bump --yes') printf '# locked\n' >>mise.lock ;;
esac
EOF
chmod +x "${fake_bin}/mise"

repository_log="${fixture_root}/repository.log"
(
  cd "${repo_fixture}"
  TEST_LOG="${repository_log}" PATH="${fake_bin}:${PATH}" bootstrap/update-repository.sh >"${fixture_root}/repository.out"
)
repository_changes="$(git -C "${repo_fixture}" diff --name-only)"
test "${repository_changes}" = $'mise.lock\nmise.toml'
rg -q '^trust -y .*/mise.toml$' "${repository_log}"
rg -q '^upgrade --bump --local --yes$' "${repository_log}"
rg -q '^lock --bump --yes$' "${repository_log}"
rg -q '^Repository update ecosystems: mise tools$' "${fixture_root}/repository.out"
rg -q '^Expected repository changes: mise.toml mise.lock$' "${fixture_root}/repository.out"

cat >"${fake_bin}/mise" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "${*}" >>"${TEST_LOG}"
if [[ "${*}" == 'run apply' && -n "${FAIL_ONCE_FILE:-}" && ! -e "${FAIL_ONCE_FILE}" ]]; then
  : >"${FAIL_ONCE_FILE}"
  exit 42
fi
EOF
chmod +x "${fake_bin}/mise"

convergence_log="${fixture_root}/convergence.log"
fail_once="${fixture_root}/fail-once"
if TEST_LOG="${convergence_log}" FAIL_ONCE_FILE="${fail_once}" PATH="${fake_bin}:${PATH}" \
  "${repo_root}/bootstrap/converge-workstation.sh" >"${fixture_root}/failed.out" 2>"${fixture_root}/failed.err"; then
  echo "expected the first convergence attempt to fail" >&2
  exit 1
fi
rg -q '^Workstation convergence failed at phase: dotfiles$' "${fixture_root}/failed.err"
if rg -q '^run verify$' "${convergence_log}"; then
  echo "convergence continued after a failed phase" >&2
  exit 1
fi

: >"${convergence_log}"
TEST_LOG="${convergence_log}" FAIL_ONCE_FILE="${fail_once}" PATH="${fake_bin}:${PATH}" \
  "${repo_root}/bootstrap/converge-workstation.sh" >"${fixture_root}/retry.out"
for expected in 'install -y' 'run apply' 'trust -y' 'run completions' 'run verify'; do
  rg -q "${expected}" "${convergence_log}"
done
rg -q '^Workstation convergence complete\.$' "${fixture_root}/retry.out"

echo "PASS: repository update and workstation convergence fixtures"
