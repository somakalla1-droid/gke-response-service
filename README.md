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
