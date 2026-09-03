---
name: sops-secrets
description: "Manage secrets with sops and age: commit encrypted files, decrypt only in memory, and wire Flux or OpenTofu runtimes. Use for any secret stored in git."
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/sops-secrets
  created: "2026-08-07"
  updated: "2026-09-03"
---

# Secrets with sops and age

Encrypted secrets live in git next to the code they configure and plaintext exists only in memory; [gitleaks](../gitleaks/SKILL.md) catches what was committed before encryption and [cloud-run](../cloud-run/SKILL.md) owns runtime secrets in Secret Manager.

## Model

- **age** provides the key pair: one private key per machine or human, public recipients everywhere; prefer it over PGP and cloud KMS for solo use, adding a `gcp_kms` recipient only for team revocation.
- **sops** encrypts the values of YAML, JSON, and ENV files; keys stay readable, so diffs review cleanly and `git log` tells which secret changed, never what it is.
- **Naming**: encrypted files are committed as `*.enc.yaml`, `*.enc.json`, or `*.enc.env`; [sops.yaml](references/sops.yaml) keys its rules off that suffix and plaintext siblings stay gitignored.
- **Policy as file**: `.sops.yaml` at the repo root ([sops.yaml](references/sops.yaml)) declares which paths get encrypted and for which recipients, so no ad-hoc flags are needed.

## Keys

1. **Generate** once per machine: `age-keygen -o ~/.config/sops/age/keys.txt` (sops' default key location); print the public half with `age-keygen -y ~/.config/sops/age/keys.txt`.
1. **Distribute** only the public key, as the `age:` recipient in each repo's `.sops.yaml`.
1. **Back up** the private key in a password manager; never commit it to any dotfiles repo.
1. **Rotate**: add the new recipient to `.sops.yaml`, run `sops updatekeys <file>` on every encrypted file, then remove the old recipient and repeat; `sops rotate -i <file>` re-keys the data key after an exposure.

## Commands

```bash
sops edit secrets.enc.yaml                          # create or edit: decrypts to $EDITOR, re-encrypts on save
sops encrypt -i config.yaml                         # encrypt an existing file in place, then rename it *.enc.yaml
sops decrypt secrets.enc.yaml                       # decrypt to stdout for piping, never redirect to a file
sops exec-env secrets.enc.env 'mise run watch'      # inject as env vars, memory-only (preferred)
sops exec-file secrets.enc.json 'tool --config {}'  # tools that need a file path get a tmpfs one
```

- **Prefer `exec-env` and `exec-file`** over decrypting to disk: nothing to leak, commit, or clean up.
- **CI**: store the private key as the single `SOPS_AGE_KEY` GitHub Actions secret; every other secret rides encrypted in the repo and jobs wrap commands in `sops exec-env`.
- **Integrations**: Flux, OpenTofu, and runtime-manager wiring lives in [integrations.md](references/integrations.md).

## Gotchas

- **Key names still leak**: sops encrypts values, not keys, so `stripe_production_key:` in a public repo is information; name keys neutrally when the repo is public.
- **Never edit ciphertext by hand**: sops stores a MAC over the file and out-of-band edits fail decryption; go through `sops edit` or `sops set`.
- **Rule match is positional**: `sops edit` picks the first `creation_rules` entry whose `path_regex` matches the path relative to `.sops.yaml`, so run sops from the repo root, and run `updatekeys` after any recipient change.
- **gitleaks coexists**: encrypted `ENC[AES256_GCM,…]` values do not trip `check:leaks`; a finding in an `*.enc.*` file means a value was committed before encryption, so rotate it per [gitleaks](../gitleaks/SKILL.md).
- **Staged hook**: the [lefthook](../lefthook/SKILL.md) pre-commit runs gitleaks on staged content, catching a staged plaintext sibling of an `*.enc.*` file.

## Documentation

- [sops](https://getsops.io/docs/) · [age](https://age-encryption.org)
- Companion skills: [gitleaks](../gitleaks/SKILL.md), [lefthook](../lefthook/SKILL.md), [cloud-run](../cloud-run/SKILL.md) (runtime secrets), [terraform-stack](../terraform-stack/SKILL.md), [k8s-local](../k8s-local/SKILL.md), [secure](../secure/SKILL.md).
