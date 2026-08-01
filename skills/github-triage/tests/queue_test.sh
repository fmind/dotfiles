#!/usr/bin/env bash

set -euo pipefail

TEST_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
readonly TEST_DIR
readonly QUEUE="${TEST_DIR}/../scripts/queue.sh"
readonly FIXTURE="${TEST_DIR}/fixtures/issues.json"
readonly REPO_ROOT="${TEST_DIR}/../../.."

fail() {
  printf 'FAIL: %s\n' "$*" >&2
  exit 1
}

runnable=$(${QUEUE} runnable fmind/dotfiles --input "${FIXTURE}")
runnable_numbers=$(jq -c '[.[].number]' <<<"${runnable}")
[[ ${runnable_numbers} == '[11,12,17]' ]] || fail "runnable ordering or exclusion"

stale=$(${QUEUE} stale fmind/dotfiles --input "${FIXTURE}" --now 1785578400 --ttl 7200)
stale_numbers=$(jq -c '[.[].number]' <<<"${stale}")
stale_reason=$(jq -r '.[0].reason' <<<"${stale}")
[[ ${stale_numbers} == '[22]' ]] || fail "stale lease detection"
[[ ${stale_reason} == 'lease heartbeat expired' ]] || fail "stale reason"

fake_bin=$(mktemp -d /tmp/github-queue-test.XXXXXX)
cleanup() {
  [[ ${fake_bin} == /tmp/github-queue-test.* && -d ${fake_bin} && ! -L ${fake_bin} ]] || {
    printf 'refusing to remove invalid fixture path: %s\n' "${fake_bin}" >&2
    return 1
  }
  rm -rf -- "${fake_bin}"
}
trap cleanup EXIT
export QUEUE_TEST_LOG="${fake_bin}/gh.log"
printf "#!/usr/bin/env bash\nprintf '%%s\\\\n' \"\$*\" >> \"\$QUEUE_TEST_LOG\"\n" >"${fake_bin}/gh"
chmod +x "${fake_bin}/gh"
PATH="${fake_bin}:${PATH}" ${QUEUE} setup fmind/dotfiles
PATH="${fake_bin}:${PATH}" ${QUEUE} setup fmind/dotfiles

setup_count=$(wc -l <"${QUEUE_TEST_LOG}")
scoped_count=$(grep -Ec -- '--repo fmind/dotfiles.*--force' "${QUEUE_TEST_LOG}")
[[ ${setup_count} -eq 8 ]] || fail "repeated setup must provision the same four labels"
[[ ${scoped_count} -eq 8 ]] || fail "setup must be repository-scoped and idempotent"

dashboard="${REPO_ROOT}/dot_config/gh-dash/config.yml"
for qualifier in '-is:blocked' '-label:needs-human' '-label:status/in-progress' '-label:kind/epic'; do
  qualifier_count=$(grep -Fc -- "${qualifier}" "${dashboard}")
  [[ ${qualifier_count} -ge 2 ]] || fail "dashboard runnable views must exclude ${qualifier}"
done
for section in 'title: Needs human' 'title: Blocked' 'title: Claimed' 'title: Runnable p0' 'title: Runnable p1'; do
  grep -Fq -- "${section}" "${dashboard}" || fail "missing dashboard section: ${section}"
done

printf 'PASS: github queue fixtures\n'
