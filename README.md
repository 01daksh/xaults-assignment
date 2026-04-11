# Xaults Assignment

## 1. ♨️ Architectural Choices 

This project is designed as a small service-health and incident-management API written in Go, with PostgreSQL as the primary datastore.

The main architectural choices are:

- The API is built with Echo.
- The codebase is split by domain, mainly `services` and `incidents`, so request handling, business logic, and persistence stay separated.
- Dependency wiring is handled through Wire-generated constructors using `Google Wire`, which keeps controller/service/repository composition.
- PostgreSQL(SQL) is used as the system of record because the domain is relational: incidents belong to services, service health depends on incident state, and consistency matters more than flexible schema designs like MongoDB.
- GORM is used for persistence to speed up CRUD development and model migration in a small project.
- The local deployment target is Kubernetes on Minikube, managed through Terraform, so the project can be reviewed with infrastructure-as-code.

The runtime also includes monitoring by default:

- Prometheus access  metrics from `/metrics` endpoint.
- Grafana is provisioned with a Prometheus datasource, along with the Golden Signals and some more metrices.


## 2. Data Modeling Rationale And Connection Management Strategy

### Data Modeling Rationale

The data model is centered around two entities: `Services` & `Incident`

The relation:

- A service is the primary managed object in the system.
- Incidents are attached to services.
- A service can have many incidents.
- A service's effective health can be derived from the current incident state.

This relational model fits PostgreSQL well because:

- querying incidents by service is common
- the domain is operational data, not document-oriented data

The project also emits business metrics from this model:

- `active_services_total`
- `open_incidents_total{severity=...}`

Those are useful for understanding the application state alongside standard traffic/error/latency/saturation signals.

### Connection Management Strategy

Database connection setup lives in [internal/database/postgres.go](/Users/daksh/GolandProjects/xaults-assignment/internal/database/postgres.go), with configuration sourced from [config/config.go](/Users/daksh/GolandProjects/xaults-assignment/config/config.go).

The current connection strategy is:

- open one shared GORM-backed database handle during application startup
- fail fast if the database cannot be reached
- configure the underlying `sql.DB` pool using:
  - `DB_MAX_OPEN_CONNS(default = 25)`
  - `DB_MAX_IDLE_CONNS(default = 5)`
  - `DB_CONN_MAX_LIFETIME(default = 5 mins)`
- reuse that pool across requests rather than opening per-request connections, has used it very frequently in many projects


## 3. 🪄 Project Startup

### Prerequisites

- Go (version >= 1.25.0)
- Docker Desktop
- Minikube
--https://minikube.sigs.k8s.io/docs/start/

(MacOS)
```sh
brew install minikube
```

- `kubectl`
-- Install from: https://kubernetes.io/docs/tasks/tools/

(MacOS)
```sh
brew install kubectl
```
- `terraform`
-- Install from: https://developer.hashicorp.com/terraform/downloads

(MacOS)
```sh
brew tap hashicorp/tap
brew install hashicorp/tap/terraform
```
- `make`

(MacOS)
```sh
brew install make
```


### Clone the Repository

First, clone the project and move into the directory:

```sh
git clone https://github.com/01daksh/xaults-assignment.git
cd xaults-assignment
```

```sh
go mod tidy
```

### Quick Start

Start Docker Desktop first, then run:

```sh
make deploy
```

After deployment, port forward for accessing the apis and analytics application through the host system:

```sh
make port-forward-all
```

Local URLs:

- API: http://127.0.0.1:1323
- Grafana: http://127.0.0.1:3000 (username and password: `admin` for both)
- Prometheus: http://127.0.0.1:9090

Grafana default credentials:

- username: `admin`
- password: `admin`

After Grafana starts, the Prometheus datasource and the `Xaults API Observability` dashboard are provisioned automatically.

### What `make deploy` Does

`make deploy` runs the local assignment flow end to end:

1. builds the API Docker image as `xaults-assignment:latest`
2. starts Minikube with the Docker driver
3. waits for the Minikube node to become Ready
4. enables the Minikube ingress addon and waits for its admission webhook to become ready
5. loads the image into Minikube
6. initializes Terraform in [terraform/](/Users/daksh/GolandProjects/xaults-assignment/terraform)
7. applies the Kubernetes resources through Terraform with auto-approval, with retries for transient local-cluster races
8. waits for the main deployments to become available before returning

### Manual Terraform Flow

If you want to run the steps yourself:

```sh
minikube start --driver=docker
kubectl wait --for=condition=Ready node/minikube --timeout=120s
docker build -t xaults-assignment:latest .
minikube image load xaults-assignment:latest
terraform -chdir=terraform init
terraform -chdir=terraform apply
```

### Verify The Deployment

Check pods and services:

```sh
make status
```

You should see running workloads for:

- `xaults-api`
- `xaults-postgres`
- `prometheus`
- `grafana`

### API Usage

📡 Health endpoint:

```sh
curl http://127.0.0.1:1323/healthz
```

Expected response:

```json
{"status":"healthy"}
```

API:

```text
GET http://127.0.0.1:1323/healthz
```

🛠️ Create a Service:

```sh
curl --location 'http://localhost:1323/services' \
--header 'Content-Type: application/json' \
--data '{
  "name": "payment-services",
  "description": "Handles payment",
  "health_status": "degraded"
}'
```

API:

```
POST http://localhost:1323/services
```

Get All Services:

```sh
curl --location 'http://localhost:1323/services'
```

API:
```
GET http://localhost:1323/services
```

🚨Incident API:

Create Incident
```sh
curl --location 'http://localhost:1323/services/5/incidents' \
--header 'Content-Type: application/json' \
--data '{
  "title": "low Load",
  "description": "lorem ipsum",
  "severity": "low"
}'
```

API:
```
POST http://localhost:1323/services/{service_id}/incidents
```

Body:
```
{
  "title": "low Load",
  "description": "lorem ipsum",
  "severity": "low"
}
```

Get All Incidents of a specific service:

```sh
curl --location 'http://localhost:1323/services/5/incidents'
```

API:
```
GET http://localhost:1323/services/{service_id}/incidents
```


### Monitoring Stack

Prometheus scrape configuration is included in [k8s/manifests.yaml](/Users/daksh/GolandProjects/xaults-assignment/k8s/manifests.yaml#L236) through the `prometheus-config` `ConfigMap`.

Prometheus scrapes:

- path: `/metrics`
- target: `xaults-api.xaults.svc.cluster.local:80`
- interval: `5s`

The exported dashboard JSON is stored at [monitoring/xaults-api-observability-dashboard.json](/Users/daksh/GolandProjects/xaults-assignment/monitoring/xaults-api-observability-dashboard.json).

It includes:

- Traffic: request rate
- Errors: 5xx error rate and error percentage
- Latency: p95 request latency
- Saturation: CPU usage and memory usage
- Business context: active services and open incidents by severity

### Useful Commands

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

### Tear Down

Destroy the Terraform-managed infrastructure:

```sh
make destroy
```

Stop Minikube if needed:

```sh
minikube stop
```

### Tests

Make sure you are in the `xaults-assignment` folder

```sh
go test ./tests/... -v
```

Even Added in the Gihub Workflow


### Grafana DashBoard

<img width="1792" height="973" alt="image" src="https://github.com/user-attachments/assets/7b8fb353-6e08-438b-b6d3-1a73392a0a44" />

### Postman (example)

<img width="1360" height="949" alt="image" src="https://github.com/user-attachments/assets/ab9f7c20-bffb-468c-aebc-b9c87953a122" />

