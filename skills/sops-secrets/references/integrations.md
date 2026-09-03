# sops integrations

Runtime wiring for consumers of sops-encrypted files; the daily workflow stays in [SKILL.md](SKILL.md).

## Kubernetes (Flux)

1. Commit `*.enc.yaml` manifests encrypted with the `encrypted_regex: ^(data|stringData)$` rule from [sops.yaml](references/sops.yaml), so only the secret payload is ciphertext and the manifest stays diffable.
1. Create the cluster-side key once: `kubectl create secret generic sops-age --namespace=flux-system --from-file=age.agekey`.
1. Point the Kustomization at it with `spec.decryption: {provider: sops, secretRef: {name: sops-age}}`; Flux decrypts in-cluster. Local clusters come from [k8s-local](../k8s-local/SKILL.md).

## OpenTofu

Keep secret variables in `secrets.enc.env` as `TF_VAR_<name>=...` lines and run `sops exec-env secrets.enc.env 'tofu plan'`; state-side encryption is handled by [terraform-stack](../terraform-stack/SKILL.md).

## Runtime services

sops secures secrets in git; long-running workloads read from a runtime manager (Secret Manager via [cloud-run](../cloud-run/SKILL.md), or Vault when dynamic credentials are required), and sops guards only the bootstrap material.

## Documentation

- [Flux sops guide](https://fluxcd.io/flux/guides/mozilla-sops/)
