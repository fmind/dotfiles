---
name: use-kubectl-cli
description: Guide for the kubectl CLI — context/namespace switching, resource inspection, declarative apply, debugging, port-forwarding, and JSON output for scripting.
---

# Use kubectl CLI

`kubectl` is the canonical Kubernetes CLI. The agent should always know which **context** + **namespace** it is operating in, prefer `kubectl apply -f` (declarative) over imperative `create`, and read with `--output=json` / `kubectl get -o yaml` when scripting.

## One-time Setup

```bash
# Install via mise (or platform package manager).
mise use -g kubectl@1.32

# GKE cluster credentials (registers a context with kubectl).
gcloud container clusters get-credentials <CLUSTER> --region=$REGION --project=$PROJECT

# Sanity check.
kubectl version --short
kubectl cluster-info
```

## Context & Namespace

```bash
kubectl config get-contexts
kubectl config current-context
kubectl config use-context my-cluster

# Namespace switching (prefer kubens if installed; otherwise --namespace).
kubectl config set-context --current --namespace=my-app
kubectl get pods                      # uses the context's default namespace
kubectl get pods -n other             # one-off override
kubectl get pods -A                   # all namespaces
```

## Inspect

```bash
# Common reads.
kubectl get pods                      # short form
kubectl get pods -o wide              # node, IP
kubectl get pods -o yaml              # full spec
kubectl get pods -o json | jq '.items[].metadata.name'

# Drill down.
kubectl describe pod my-pod
kubectl logs my-pod
kubectl logs my-pod -c my-container
kubectl logs -f my-pod                # follow
kubectl logs -l app=api --since=10m   # by label, time-bounded

# Events (most useful single command for "why is this broken").
kubectl get events --sort-by=.lastTimestamp -A
```

## Declarative Apply

```bash
# Apply a manifest (idempotent).
kubectl apply -f deployment.yaml
kubectl apply -k overlays/prod        # kustomize

# Diff before apply (server-side, no mutation).
kubectl diff -f deployment.yaml

# Delete by manifest (matches what apply created).
kubectl delete -f deployment.yaml
```

## Debugging

```bash
# Interactive shell into a running pod.
kubectl exec -it my-pod -- bash
kubectl exec -it my-pod -c sidecar -- sh

# Run a one-off pod.
kubectl run debug --rm -it --image=alpine -- sh

# Ephemeral debug container (preferred for prod pods).
kubectl debug -it my-pod --image=ubuntu --target=app

# Port-forward (local → pod).
kubectl port-forward svc/api 8080:80

# Copy files in/out.
kubectl cp ./local-file my-pod:/tmp/file
kubectl cp my-pod:/var/log/app.log ./
```

## Rollouts

```bash
kubectl rollout status deployment/api
kubectl rollout history deployment/api
kubectl rollout undo deployment/api
kubectl rollout undo deployment/api --to-revision=3
kubectl rollout restart deployment/api
```

## Scaling

```bash
kubectl scale deployment/api --replicas=5
kubectl autoscale deployment/api --min=2 --max=10 --cpu-percent=60
kubectl get hpa
```

## Resource Top (requires metrics-server)

```bash
kubectl top nodes
kubectl top pods -A --sort-by=cpu
```

## RBAC & Secrets

```bash
# Inspect what *you* can do as the current principal.
kubectl auth can-i create deployments
kubectl auth can-i list secrets -n other

# Create a secret (literal values).
kubectl create secret generic db-creds --from-literal=password="$DB_PASS"

# Decode a secret value.
kubectl get secret db-creds -o jsonpath='{.data.password}' | base64 -d
```

## JSON / Scripting Output

```bash
# JSONPath.
kubectl get pods -o jsonpath='{.items[*].metadata.name}'

# Custom columns.
kubectl get pods -o custom-columns=NAME:.metadata.name,NODE:.spec.nodeName

# JSON + jq (most flexible).
kubectl get pods -o json | jq '.items[] | {name: .metadata.name, status: .status.phase}'
```

## Common Workflows

**Diagnose a CrashLoopBackOff.**
```bash
kubectl describe pod my-pod              # events + reason
kubectl logs my-pod --previous           # last terminated container
kubectl get events --field-selector involvedObject.name=my-pod
```

**Roll back after a bad deploy.**
```bash
kubectl rollout history deployment/api
kubectl rollout undo deployment/api
kubectl rollout status deployment/api    # confirm new revision is healthy
```

**Local dev against a cluster.**
```bash
kubectl port-forward svc/api 8080:80
# In another shell:
curl http://localhost:8080/health
```

## Companion Tools

- **kubectx / kubens** — fast context/namespace switching (`brew install kubectx`).
- **stern** — multi-pod log tail by label.
- **k9s** — TUI for cluster browsing.
- **kustomize** — bundled with kubectl as `kubectl kustomize`; prefer overlays over Helm where possible.

## Important Notes

1. **Always know the active context + namespace** — `kubectl config current-context` + `kubectl config view --minify | grep namespace`. Mixing contexts is the most common source of incidents.
2. **Use `kubectl diff` before `kubectl apply`** in production.
3. **`kubectl run` creates Pods, not Deployments** — for anything long-lived, write a manifest.
4. **Logs are bounded** — `--since=10m` / `--tail=100` to avoid blasting the terminal.
5. **`kubectl delete -f` matches by manifest content** — if you've drifted, prefer `kubectl delete <kind>/<name>` explicitly.

## Documentation

- [kubectl reference](https://kubernetes.io/docs/reference/kubectl/)
- [kubectl cheat sheet](https://kubernetes.io/docs/reference/kubectl/cheatsheet/)
- [JSONPath support](https://kubernetes.io/docs/reference/kubectl/jsonpath/)
- [Kustomize](https://kubectl.docs.kubernetes.io/references/kustomize/)
- [GKE cluster credentials](https://docs.cloud.google.com/kubernetes-engine/docs/how-to/cluster-access-for-kubectl)
