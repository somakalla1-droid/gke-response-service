# GKE Response Service

Application B for the GKE SRE assessment. It provides an independent downstream response with pod, cluster, region, request correlation, controlled-error, and controlled-latency behavior. Request completion logs are newline-delimited JSON with `severity`, `status_code`, `request_id`, and `latency_ms` fields for Cloud Logging, BigQuery, and Grafana analysis.

## Run locally

```bash
go test ./...
go run ./cmd/server
curl -s http://localhost:8080/ | jq
```

Endpoints: `/`, `/healthz`, `/readyz`, `/metrics`, `/error`, and `/delay?ms=250`.

Configuration: `PORT`, `APP_VERSION`, `CLUSTER_NAME`, `REGION`, `POD_NAME`,
`RESPONSE_MESSAGE`, `OBSERVABILITY_ENABLED`, and `GOOGLE_CLOUD_PROJECT`.

Cloud observability is disabled by default so local development does not need
Google credentials. When enabled in GKE, the service exports sampled server
spans through the standard authenticated OTLP endpoint to Cloud Trace, starts
Cloud Profiler, adds trace correlation fields to structured request logs, and
formats `/error` events for Error Reporting. A telemetry initialization failure
is logged but does not prevent the service from serving traffic. Health,
readiness, and metrics scrapes are excluded from tracing to avoid paying for
low-value operational spans.

## Container and Kubernetes

```bash
docker build -t gke-response-service:local .
docker run --rm -p 8080:8080 gke-response-service:local
helm lint charts/gke-response-service
```

The image is distroless and non-root. The application-owned Helm chart configures two replicas, health probes, constrained resources, HPA, and a PodDisruptionBudget. Its container security context explicitly uses the distroless `nonroot` UID/GID (`65532`), allowing Kubernetes to verify the `runAsNonRoot` policy without relying on the image's named user.

### Kubernetes configuration

The chart always creates a ConfigMap containing `APP_VERSION`, `CLUSTER_NAME`, `REGION`, and `RESPONSE_MESSAGE`. Set the deployment values without editing manifests:

```bash
helm upgrade --install response charts/gke-response-service \
  --set config.clusterName=gke-primary \
  --set config.region=us-central1
```

For sensitive settings, use an externally managed Kubernetes Secret by setting `secret.existingSecret`, or enable chart-managed creation with `secret.create=true` and a non-empty `secret.stringData` map. Never commit real secret values in a values file. The assessment's production design will use Secret Manager with Workload Identity Federation.

### Google Secret Manager on GKE

For the assessment deployment, the secret is delivered as a read-only file by the GKE Secret Manager CSI add-on. The Helm chart creates a `SecretProviderClass` and mounts the file, but does **not** mirror the value into a Kubernetes Secret. The target GKE cluster must have the Secret Manager add-on enabled, and the release ServiceAccount must have the corresponding Workload Identity principal access granted in Google Secret Manager.

Example for the prepared assessment resources:

```bash
helm upgrade --install response charts/gke-response-service \
  --namespace assessment-apps \
  --create-namespace \
  --set serviceAccount.name=response-service-workload \
  --set config.clusterName=gke-primary \
  --set config.region=us-central1 \
  --set observability.enabled=true \
  --set observability.projectId=gke-sre-assesment \
  --set secretManager.enabled=true \
  --set secretManager.projectNumber=150538255871
```

The application must treat `/var/run/secrets/gsm/response-demo-token` as sensitive: do not log, print, commit, or inject its contents into a Kubernetes Secret. To verify the mount later, test only for the presence of the file, not its value.
