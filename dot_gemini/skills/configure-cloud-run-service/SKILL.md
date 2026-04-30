---
name: configure-cloud-run-service
description: Guide for declarative Cloud Run service.yaml — image, env, secrets, scaling, concurrency, traffic, VPC, service account, and replace/deploy commands.
---

# Configure Cloud Run Service (`service.yaml`)

Cloud Run services can be defined declaratively in a Knative-style `service.yaml`. The same surface is exposed by `gcloud run deploy`, but the YAML form is reproducible, reviewable, and lives next to the code.

This skill covers the **Cloud Run for Services** YAML (long-running HTTP). For batch jobs, see `gcloud run jobs deploy`.

## Minimal `service.yaml`

```yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: api
  annotations:
    run.googleapis.com/launch-stage: GA
spec:
  template:
    spec:
      containers:
        - image: us-central1-docker.pkg.dev/PROJECT/REPO/api:1.0
          ports:
            - containerPort: 8080
```

Apply / replace:

```bash
gcloud run services replace service.yaml --region=us-central1 --project=$PROJECT
```

## Production-grade `service.yaml`

```yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: api
  labels:
    team: platform
    env: prod
  annotations:
    run.googleapis.com/launch-stage: GA
    run.googleapis.com/ingress: all                    # all | internal | internal-and-cloud-load-balancing
    run.googleapis.com/description: "Customer-facing API"

spec:
  template:
    metadata:
      annotations:
        # Scaling.
        autoscaling.knative.dev/minScale: "1"
        autoscaling.knative.dev/maxScale: "20"

        # Concurrency & resources.
        run.googleapis.com/cpu-throttling: "false"     # always-on CPU outside requests
        run.googleapis.com/startup-cpu-boost: "true"

        # Networking.
        run.googleapis.com/vpc-access-connector: "projects/PROJECT/locations/us-central1/connectors/my-connector"
        run.googleapis.com/vpc-access-egress: "private-ranges-only"

        # Cloud SQL connection (Unix socket).
        run.googleapis.com/cloudsql-instances: "PROJECT:us-central1:db-1"

    spec:
      serviceAccountName: api-runtime@PROJECT.iam.gserviceaccount.com
      timeoutSeconds: 60
      containerConcurrency: 80                          # max concurrent requests per instance

      containers:
        - image: us-central1-docker.pkg.dev/PROJECT/REPO/api:1.0
          name: app
          ports:
            - containerPort: 8080

          resources:
            limits:
              cpu: "2"
              memory: "1Gi"

          env:
            - name: NODE_ENV
              value: production
            - name: DB_URL
              valueFrom:
                secretKeyRef:
                  name: db-url            # Secret Manager secret
                  key: latest

          startupProbe:
            httpGet:
              path: /healthz/ready
            initialDelaySeconds: 0
            periodSeconds: 1
            failureThreshold: 30

          livenessProbe:
            httpGet:
              path: /healthz/live
            periodSeconds: 30

  traffic:
    - percent: 90
      latestRevision: true
    - percent: 10
      revisionName: api-00012-stable
```

## Traffic Splitting & Revisions

```yaml
traffic:
  - revisionName: api-00012-stable
    percent: 100
  - tag: canary
    revisionName: api-00013-canary
    percent: 0                     # accessible at the tagged URL only
```

Equivalent imperative form:

```bash
gcloud run services update-traffic api \
  --to-revisions=api-00012-stable=90,api-00013-canary=10 \
  --region=us-central1
```

## Secrets via Secret Manager

```yaml
env:
  - name: STRIPE_KEY
    valueFrom:
      secretKeyRef:
        name: stripe-key
        key: latest

# Or mount as a file:
volumes:
  - name: stripe
    secret:
      secretName: stripe-key
volumeMounts:
  - name: stripe
    mountPath: /var/secrets/stripe
    readOnly: true
```

The runtime service account needs `roles/secretmanager.secretAccessor` on each referenced secret.

## Networking

| Setting | Effect |
|---------|--------|
| `run.googleapis.com/ingress: internal` | Reject external traffic; only VPC + GCLB |
| `run.googleapis.com/vpc-access-connector` | Egress through a VPC connector |
| `run.googleapis.com/vpc-access-egress: all-traffic` | Route ALL egress through the VPC |
| `run.googleapis.com/cloudsql-instances` | Mount the Cloud SQL Unix-socket auth proxy |

## Identity

```yaml
spec:
  template:
    spec:
      serviceAccountName: api-runtime@PROJECT.iam.gserviceaccount.com
```

Always use a dedicated service account per service — never the Compute Engine default. Grant it only the roles its workload needs.

## Deploy

```bash
# Replace the entire spec (declarative).
gcloud run services replace service.yaml \
  --region=us-central1 --project=$PROJECT

# Deploy a single image (imperative — useful for quick rollouts).
gcloud run deploy api \
  --image=us-central1-docker.pkg.dev/$PROJECT/repo/api:$SHA \
  --region=us-central1 --project=$PROJECT

# Read current spec back.
gcloud run services describe api --region=us-central1 --format=yaml > current.yaml
```

## Scaling Knobs

| Annotation | Default | Notes |
|------------|---------|-------|
| `autoscaling.knative.dev/minScale` | 0 | Set ≥1 to keep an instance warm |
| `autoscaling.knative.dev/maxScale` | 100 | Cap concurrency-driven growth |
| `containerConcurrency` (top-level) | 80 | Max concurrent requests per instance |
| `run.googleapis.com/cpu-throttling` | "true" | "false" → always-allocated CPU (more $) |
| `run.googleapis.com/cpu-boost` (startup-cpu-boost) | "false" | Faster cold starts |

## Probes

```yaml
startupProbe:                    # only fires once per cold start
  httpGet: { path: /healthz/ready }
  failureThreshold: 30
  periodSeconds: 1

livenessProbe:                   # restarts container on repeated failure
  httpGet: { path: /healthz/live }
  periodSeconds: 30
```

## Common Workflows

**Bootstrap.**

1. `gcloud iam service-accounts create api-runtime --project=$PROJECT`
2. Grant least-privilege roles (Logs Writer, Secret Accessor, etc.).
3. Drop `service.yaml` next to the code (use the production-grade template).
4. `gcloud run services replace service.yaml ...`.

**Promote a canary.**

1. Edit `traffic:` to give the new revision 10%.
2. `gcloud run services replace service.yaml`.
3. Watch metrics; when stable, set new revision to 100%.

**Local equivalent** (Functions Framework / Express):

```bash
docker build -t api:dev .
docker run --rm -p 8080:8080 -e PORT=8080 api:dev
curl localhost:8080/healthz/ready
```

## Important Notes

1. **`gcloud run services replace` is destructive** — the YAML is the new spec; anything not declared is removed. Read `describe` first.
2. **Pin images by digest (`@sha256:...`)** in production for reproducibility.
3. **Default service account is too broad** — always set `serviceAccountName` to a dedicated one.
4. **`minScale: 0`** means cold starts; set ≥1 for latency-sensitive APIs.
5. **`startupProbe.failureThreshold * periodSeconds`** is the cold-start budget — tune to match real boot time.
6. **Secrets via `secretKeyRef`** require the runtime SA to have `secretmanager.secretAccessor`; the deployer SA needs IAM admin on the secret.

## Documentation

- [Cloud Run YAML reference](https://docs.cloud.google.com/run/docs/reference/yaml/v1)
- [Cloud Run service annotations](https://docs.cloud.google.com/run/docs/configuring)
- [Networking — VPC connectors](https://docs.cloud.google.com/run/docs/configuring/connecting-vpc)
- [Secrets in Cloud Run](https://docs.cloud.google.com/run/docs/configuring/secrets)
- [Cloud SQL connections](https://docs.cloud.google.com/run/docs/configuring/connect-cloudsql)
- [Probes](https://docs.cloud.google.com/run/docs/configuring/healthchecks)
- [Traffic splitting](https://docs.cloud.google.com/run/docs/rollouts-rollbacks-traffic-migration)
