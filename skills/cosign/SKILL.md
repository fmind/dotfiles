---
name: cosign
description: Sign, verify, and attest container images keyless with Sigstore cosign, pinning OIDC identity and issuer. Use for image signing, verification, or SBOM attestations.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/cosign
  created: "2026-09-02"
  updated: "2026-09-03"
---

# Cosign

Keyless signing with Sigstore: an OIDC identity (a GitHub Actions workflow or a developer's browser login) signs the image digest and anyone verifies it without managing keys; [containerize](../containerize/SKILL.md) builds the image and [github-actions](../github-actions/SKILL.md) wires the CD job.

## Commands

```bash
cosign sign --yes <registry>/<slug>@<digest>              # --yes is mandatory: cosign prompts otherwise and hangs agents and CI
cosign verify \
  --certificate-identity 'https://github.com/<owner>/<repo>/.github/workflows/cd.yml@refs/tags/<tag>' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
  <registry>/<slug>@<digest>
cosign attest --yes --type cyclonedx --predicate sbom.json <registry>/<slug>@<digest>
cosign verify-attestation --type cyclonedx \
  --certificate-identity-regexp 'https://github.com/<owner>/<repo>/' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
  <registry>/<slug>@<digest>
```

## GitHub Actions

Pin `cosign` in `mise.toml` `[tools]` so `mise-action` installs it with the rest of the toolchain; the signing job needs `permissions: id-token: write` plus `packages: write` (or the registry's equivalent), `cache: false` on `mise-action`, and signs the digest the build step recorded (`ko build --image-refs`). The [github-actions](../github-actions/SKILL.md) `cd.yml` template implements this wiring.

```toml
[tools]
cosign = "latest"
```

## Gotchas

- **Sign digests, never tags**: a tag can be repointed after signing.
- **Verification pins identity and issuer**: a valid signature from an unexpected workflow is not provenance; record the expected identity in the release documentation.
- **Signing writes to the registry**: `sign` and `attest` push signatures next to the image, so a local packaging request does not authorize them.
- **SBOM first, then attest**: generate the SBOM per [trivy](../trivy/SKILL.md) and attach it as an attestation so consumers verify inventory and signature together.

## Documentation

- [cosign](https://docs.sigstore.dev/cosign/)
- Companion skills: [containerize](../containerize/SKILL.md) (builds the image), [github-actions](../github-actions/SKILL.md) (CD job), [secure](../secure/SKILL.md).
