---
name: sops-secrets
description: Manage repository secrets with sops + age — encrypted files committed to git, memory-only decryption via exec-env, Flux and OpenTofu integration. Use whenever a project needs secrets in version control or at runtime.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/sops-secrets
  created: 2026-08-07
  updated: 2026-08-07
---

# Secrets Standard (sops 3.13+ / age)

Canonical workflow for **sops** (structured-file encryption) with **age** (modern key pairs) — encrypted secrets live in git next to the code they configure, and plaintext exists only in memory. This operationalizes the global "No Secrets in Output" rule: no plaintext secret ever touches disk, logs, or a commit.

## 1. Core Model

- **age** provides the key pair: one private key per machine/human, public recipients everywhere. Prefer age over PGP (simpler, no keyservers) and over cloud KMS for solo/portable use; add a `gcp_kms` recipient alongside age only when a team needs central revocation.
- **sops** encrypts the **values** of YAML/JSON/ENV files — keys stay readable, so diffs review cleanly and `git log` still tells you _which_ secret changed, never _what_ it is.
- **Naming Convention**: encrypted files are committed as `*.enc.yaml` / `*.enc.json` / `*.enc.env`; the [sops.yaml](references/sops.yaml) rules key off that suffix, and any plaintext siblings stay gitignored.
- **Policy As File**: `.sops.yaml` at the repo root ([sops.yaml](references/sops.yaml)) declares which paths get encrypted and for which recipients — creation is automatic, ad-hoc flags are never needed.

## 2. Key Management

1. **Generate** once per machine: `age-keygen -o ~/.config/sops/age/keys.txt` — sops' default key location; print the public half anytime with `age-keygen -y ~/.config/sops/age/keys.txt`.
1. **Distribute** only the **public** key: paste it as the `age:` recipient in each repo's [sops.yaml](references/sops.yaml).
1. **Back up** the private key in a password manager. Never commit it anywhere — in particular never `chezmoi add` it, since the dotfiles repository is public.
1. **Rotate** by adding the new recipient to `.sops.yaml`, running `sops updatekeys <file>` on every encrypted file, then removing the old recipient and repeating; `sops rotate -i <file>` re-keys the data key after a suspected exposure.

## 3. Daily Workflow

```bash
sops edit secrets.enc.yaml                      # create or edit (decrypts to $EDITOR, re-encrypts on save)
sops -e -i config.yaml                          # encrypt an existing file in place (then rename to *.enc.yaml)
sops -d secrets.enc.yaml                        # decrypt to stdout — for piping, never for redirecting to a file
sops exec-env secrets.enc.env 'mise run watch'  # inject as env vars, memory-only (preferred)
sops exec-file secrets.enc.json 'tool --config {}'  # tools that require a file path get a tmpfs one
```

- **Prefer `exec-env`/`exec-file`** over `sops -d > plain` — decrypted bytes never land in the working tree, so there is nothing to leak, commit, or clean up.
- **CI**: store the private key as the single `SOPS_AGE_KEY` GitHub Actions secret; every other secret rides encrypted in the repo, and jobs wrap commands in `sops exec-env`. One secret to rotate instead of dozens.

## 4. Integrations

- **Kubernetes (Flux)**: commit `*.enc.yaml` manifests encrypted with the `encrypted_regex: ^(data|stringData)$` rule ([sops.yaml](references/sops.yaml)), create the cluster-side key once (`kubectl create secret generic sops-age --namespace=flux-system --from-file=age.agekey`), and point the Kustomization at it (`spec.decryption: {provider: sops, secretRef: {name: sops-age}}`) — Flux decrypts in-cluster, per the [k8s-local skill](../k8s-local/SKILL.md) for local clusters.
- **OpenTofu**: keep secret variables in `secrets.enc.env` as `TF_VAR_<name>=…` lines and run `sops exec-env secrets.enc.env 'tofu plan …'` — see the [terraform-stack skill](../terraform-stack/SKILL.md); state-side encryption is handled there.
- **Runtime Services**: sops secures secrets **in git**; long-running production workloads should read from a runtime manager (GCP Secret Manager via the [cloud-run skill](../cloud-run/SKILL.md), or Vault when dynamic credentials are required) — sops then guards the bootstrap material only.

## 5. Gotchas

- **Key Names Still Leak**: sops encrypts values, not keys — `stripe_production_key:` in a public repo is information. Name keys neutrally when the repo is public.
- **Never Edit Ciphertext By Hand**: sops stores a MAC over the file; out-of-band edits fail decryption. Always go through `sops edit` (or `sops set` in scripts).
- **Rule Match Is Positional**: `sops edit` picks the first `creation_rules` entry whose `path_regex` matches the path **relative to `.sops.yaml`** — run sops from the repo root, and remember `updatekeys` after any recipient change (editing `.sops.yaml` alone re-encrypts nothing).
- **gitleaks Coexists**: encrypted `ENC[AES256_GCM,…]` values do not trip `check:leaks` — a finding in an `*.enc.*` file means a value was committed **before** encryption; rotate that secret, do not just encrypt it.
- **`--staged` Hook Discipline**: the lefthook pre-commit from the [lefthook skill](../lefthook/SKILL.md) runs gitleaks on staged content, catching the classic mistake of staging the plaintext sibling instead of the `*.enc.*` file.

## Documentation

- [sops](https://getsops.io/docs/) · [age](https://age-encryption.org) · [Flux sops guide](https://fluxcd.io/flux/guides/mozilla-sops/)
- Companion skills:
  - [terraform-stack](../terraform-stack/SKILL.md) — encrypted `TF_VAR` injection.
  - [k8s-local](../k8s-local/SKILL.md) — clusters that consume sops-encrypted manifests.
  - [cloud-run](../cloud-run/SKILL.md) — runtime secrets via GCP Secret Manager.
  - [security-scan](../security-scan/SKILL.md) — gitleaks history sweeps.
