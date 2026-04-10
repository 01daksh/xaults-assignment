# Xaults Assignment

This project is a Go API backed by PostgreSQL and deployed locally with Minikube, Terraform, and Kubernetes.

The repository also includes a monitoring stack:

- Prometheus for scraping `/metrics`
- Grafana for dashboards
- automatic Grafana provisioning for the Prometheus datasource and dashboard
- a Grafana dashboard export for the application's Golden Signals

Terraform applies the Kubernetes resources defined in [k8s/manifests.yaml](/Users/daksh/GolandProjects/xaults-assignment/k8s/manifests.yaml#L1).

## Prerequisites

- Docker Desktop
- Minikube
- `kubectl`
- `terraform`
- `make`

## Quick Start

Start Docker Desktop first, then run:

```sh
make deploy
```

After deployment, open separate terminals for local access:

```sh
make port-forward-api
make port-forward-prometheus
make port-forward-grafana
```

Local URLs:

- API: `http://127.0.0.1:1323`
- Prometheus: `http://127.0.0.1:9090`
- Grafana: `http://127.0.0.1:3000`

Grafana default credentials:

- username: `admin`
- password: `admin`

After Grafana starts, the Prometheus datasource and the `Xaults API Observability` dashboard are provisioned automatically.

## What `make deploy` Does

`make deploy` runs the local assignment flow end to end:

1. builds the API Docker image as `xaults-assignment:latest`
2. starts Minikube with the Docker driver
3. waits for the Minikube node to become Ready
4. loads the image into Minikube
5. initializes Terraform in [terraform/](/Users/daksh/GolandProjects/xaults-assignment/terraform)
6. applies the Kubernetes resources through Terraform with auto-approval

## Manual Terraform Flow

If you want to run the steps yourself:

```sh
minikube start --driver=docker
docker build -t xaults-assignment:latest .
minikube image load xaults-assignment:latest
terraform -chdir=terraform init
terraform -chdir=terraform apply
```

## Verify The Deployment

Check pods and services:

```sh
make status
```

You should see running workloads for:

- `xaults-api`
- `xaults-postgres`
- `prometheus`
- `grafana`

## API Usage

Expose the API:

```sh
make port-forward-api
```

Health endpoint:

```sh
curl http://127.0.0.1:1323/healthz
```

Expected response:

```json
{"status":"healthy"}
```

For Postman:

```text
GET http://127.0.0.1:1323/healthz
```

## Monitoring Stack

### Prometheus Scrape Configuration

The monitoring configuration is included in [k8s/manifests.yaml](/Users/daksh/GolandProjects/xaults-assignment/k8s/manifests.yaml#L196) through the `prometheus-config` `ConfigMap`.

Prometheus scrapes:

- path: `/metrics`
- target: `xaults-api.xaults.svc.cluster.local:80`
- interval: `5s`

This covers the requirement to provide the configuration needed for the monitoring stack to scrape the application's metrics.

### Access Prometheus

Run:

```sh
make port-forward-prometheus
```

Then open:

```text
http://127.0.0.1:9090
```

### Access Grafana

Run:

```sh
make port-forward-grafana
```

Then open:

```text
http://127.0.0.1:3000
```

Login:

- username: `admin`
- password: `admin`

### Dashboard Provisioning

Grafana is provisioned automatically from Kubernetes config in [k8s/manifests.yaml](/Users/daksh/GolandProjects/xaults-assignment/k8s/manifests.yaml#L254):

- datasource: Prometheus
- dashboard provider: `Xaults`
- dashboard: `Xaults API Observability`

The exported dashboard JSON is also stored in the repo at [monitoring/xaults-api-observability-dashboard.json](/Users/daksh/GolandProjects/xaults-assignment/monitoring/xaults-api-observability-dashboard.json#L1).

It visualizes the Golden Signals using these panels:

- Traffic: request rate
- Errors: 5xx error rate and error percentage
- Latency: p95 request latency
- Saturation: process CPU usage and resident memory usage
- Business context: active services and open incidents by severity

Your original dashboard JSON was close, but I adjusted it because `active_services_total` and `open_incidents_total` are business metrics, not true saturation signals. The provisioned/exported version keeps those panels, but uses Go process metrics for saturation so it matches the Golden Signals more accurately for this API.

This satisfies both:

- monitoring stack scrape configuration
- dashboard export for the Golden Signals

## Useful Commands

Show logs:

```sh
make logs-api
make logs-db
make logs-prom
make logs-grafana
```

Rebuild and restart the API after code changes:

```sh
make redeploy
```

## Tear Down

Destroy the Terraform-managed infrastructure:

```sh
make destroy
```

Stop Minikube if needed:

```sh
minikube stop
```

## Notes

- The old Docker Compose workflow is no longer the primary local run path for this assignment.
- PostgreSQL is configured with ephemeral storage for reliable local Minikube startup.
- The root `Makefile` is intended to make review and demo setup faster.
