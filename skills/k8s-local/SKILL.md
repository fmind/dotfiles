---
name: k8s-local
description: Run an opt-in k3d or kind cluster on the workstation and deploy into it with kubectl, helm, helmfile, or skaffold. Use only when a project needs local Kubernetes.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/k8s-local
  created: "2026-06-23"
  updated: "2026-09-03"
---

# Local Kubernetes

Run an opt-in k3d (default) or kind cluster on the workstation and deploy into it declaratively. [cloud-run](../cloud-run/SKILL.md) is the default deploy target, so the Kubernetes tools are commented out in the global mise config: uncomment them or pin them in the project `mise.toml` first.

## Workflow

1. **Choose the scope**: quick experiments use the global `dot cluster` from [dot-cli](../dot-cli/SKILL.md); a real project stays self-contained (see the project-scope gotcha).
1. **Check Docker**: run `docker info` before starting, stopping, or configuring a cluster.
1. **Start the cluster, OFF by Default**: run it only for a short verification window and stop it as soon as the task ends; k3d containers use `restart=unless-stopped`, so they return after a reboot until stopped.

   ```bash
   dot cluster start   # k3d from ~/.config/k3d/local.yaml: idempotent, owner-only kubeconfig, never touches the default context
   kind create cluster --config ~/.agents/skills/k8s-local/references/kind-config.yaml
   ```

1. **Target the isolated context**: never switch the default kubeconfig; a renamed context needs `dot cluster --context <name> <command>`.

   ```bash
   kubectl --kubeconfig ~/.kube/dot/local.yaml --context k3d-local get nodes
   dot cluster namespace <namespace>
   ```

1. **Validate manifests**: `kubeconform -strict <path>` for schema, `kube-linter lint <path>` for practices and security.
1. **Publish images**: push to the local registry or sideload, always with a specific tag and `imagePullPolicy: IfNotPresent` or `Never` in the manifest.

   ```bash
   docker tag <image>:<tag> registry.localhost:5050/<image>:<tag> && docker push registry.localhost:5050/<image>:<tag>
   k3d image load <image>:<tag> -c local                        # k3d sideload
   kind load docker-image <image>:<tag> --name local-cluster    # kind sideload
   ```

1. **Deploy declaratively**: `helm upgrade --install <release> <chart> -n <namespace> --create-namespace`, `helmfile apply`, `skaffold dev`, or `kubectl apply -k <dir>`; expose with [ingress-template.yaml](references/ingress-template.yaml).
1. **Diagnose with bounded evidence**: `dot cluster diagnose --namespace <namespace>` verifies the isolated target, runs only the read-only allowlist, redacts secrets, and writes an owner-only manifest; review it locally before sharing.
1. **Stop or delete**: `dot cluster stop` is the resting state; `dot cluster delete` is destructive and asks for confirmation; `kind delete cluster --name local-cluster` removes the kind cluster.

## Gotchas

- **Project scope**: pin the tools in `mise.toml`, keep the cluster config in git, use a process-scoped `KUBECONFIG`, and ship `.agents/skills/local-cluster/SKILL.md` so agents use the project's tasks.
- **Engine choice**: k3d for the interactive loop (fast start, Traefik, registry); kind ([kind-config.yaml](references/kind-config.yaml)) for disposable conformance tests, with your own ingress controller.
- **Ports**: `8042`/`8043` (ingress), `6443` (API), and `5050` (registry) must be free; the ingress pair avoids `8080`/`8443` on purpose (see [gotchas](references/gotchas.md)).
- **Registry push name**: the cluster pulls `registry.localhost:5050/...`, but `docker push` to that name fails until the daemon lists it as insecure (see [gotchas](references/gotchas.md)).
- **DiskPressure**: a full workstation disk taints the node `NoSchedule` while it still reports `Ready`; check taints before debugging a `Pending` pod (see [gotchas](references/gotchas.md)).
- **Interactive tools**: `k9s`, `stern`, and `mirrord` are opt-in too; pin them in the project `mise.toml` when the bounded evidence bundle is not enough.

## Documentation

- [k3d](https://k3d.io) · [kind](https://kind.sigs.k8s.io) · [kubectl](https://kubernetes.io/docs/reference/kubectl/) · [Helm](https://helm.sh/docs/) · [Helmfile](https://helmfile.readthedocs.io) · [Skaffold](https://skaffold.dev/docs/)
- Companion skills: [dot-cli](../dot-cli/SKILL.md) (`dot cluster` commands), [containerize](../containerize/SKILL.md) (builds the images), [cloud-run](../cloud-run/SKILL.md) (default deploy target), [sops-secrets](../sops-secrets/SKILL.md) (Flux runtimes).
