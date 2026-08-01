---
name: release-audit
description: Audit an existing GitHub release without changing it. Use when verifying exact-head CI, local and remote tag identity, public release state, or declared release assets such as checksums, SBOMs, signatures, and attestations; do not use it to publish or repair a release.
license: MIT
---

# Release Audit

Prove each release boundary independently and report gaps without creating, moving, deleting, or publishing tags or releases.

## Contract

Require a repository-owned `.release-audit.json` so an audit never invents an artifact requirement or assumes binaries exist:

```json
{
  "schema": "fmind.dev/release-audit-contract/v1",
  "workflows": ["CI"],
  "artifacts": []
}
```

An empty `artifacts` list explicitly declares a source-only release. Artifact entries require `name` and `kind`; supported kinds are `asset`, `checksum` with `covers`, `sbom`, `signature` with `subject`, `certificate_identity`, and `certificate_oidc_issuer`, and `attestation` with `subject`.

## Audit

Run from the repository checkout:

```bash
python3 ~/.agents/skills/release-audit/scripts/audit.py <tag> [--contract .release-audit.json]
python3 ~/.agents/skills/release-audit/scripts/audit.py <tag> --json
```

The collector is read-only. It compares explicit SHAs for `HEAD`, its upstream, the local tag, the remote tag, each required workflow run, and the public release target. Declared assets are downloaded only to a temporary directory; GitHub release digests, checksum coverage, SBOM structure, Sigstore signatures, and GitHub attestations are validated only when the contract requests them.

Treat the final state precisely:

- `missing`: required local, remote, release, workflow, or contract evidence does not exist.
- `draft`: the matching GitHub release exists but is not public.
- `stale`: SHA identities diverge, a tag moved, or required CI belongs to another commit.
- `artifact-incomplete`: the published release is missing or fails a declared artifact check.
- `published`: every declared boundary matches the exact commit and every promised artifact is valid.

Keep the JSON report as the automation interface; the default human report is a concise rendering of the same checks. A timeout or collection error is `missing` evidence, never success.

## Publication Boundary

Stop after reporting. Invoke the separate [release skill](../release/SKILL.md) only when the user explicitly requests publication or repair; never infer that authority from an audit request.
