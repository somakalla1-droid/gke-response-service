# GKE Response Service

Application B for the GKE SRE assessment. It provides an independent downstream response with pod, cluster, region, request correlation, controlled-error, and controlled-latency behavior. Request completion logs are newline-delimited JSON with `severity`, `status_code`, `request_id`, and `latency_ms` fields for Cloud Logging, BigQuery, and Grafana analysis.

## Run locally

```bash
go test ./...
go run ./cmd/server
curl -s http://localhost:8080/ | jq
```

Endpoints: `/`, `/healthz`, `/readyz`, `/metrics`, `/error`, and `/delay?ms=250`.

Configuration: `PORT`, `APP_VERSION`, `CLUSTER_NAME`, `REGION`, `POD_NAME`, and `RESPONSE_MESSAGE`.

## Container and Kubernetes

```bash
docker build -t gke-response-service:local .
docker run --rm -p 8080:8080 gke-response-service:local
helm lint charts/gke-response-service
```

The image is distroless and non-root. The application-owned Helm chart configures two replicas, health probes, constrained resources, HPA, and a PodDisruptionBudget.

### Kubernetes configuration

The chart always creates a ConfigMap containing `APP_VERSION`, `CLUSTER_NAME`, `REGION`, and `RESPONSE_MESSAGE`. Set the deployment values without editing manifests:

```bash
helm upgrade --install response charts/gke-response-service \
  --set config.clusterName=gke-primary \
  --set config.region=us-central1
```

For sensitive settings, use an externally managed Kubernetes Secret by setting `secret.existingSecret`, or enable chart-managed creation with `secret.create=true` and a non-empty `secret.stringData` map. Never commit real secret values in a values file. The assessment's production design will use Secret Manager with Workload Identity Federation.
