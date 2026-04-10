SHELL := /bin/zsh

IMAGE_NAME := xaults-assignment:latest
NAMESPACE := xaults
TERRAFORM_DIR := terraform

.PHONY: help minikube-start wait-cluster ingress-ready wait-apps build image-load tf-init deploy redeploy destroy status logs-api logs-db logs-prom logs-grafana port-forward-api port-forward-prometheus port-forward-grafana

help:
	@echo "Available targets:"
	@echo "  make minikube-start       Start Minikube with the Docker driver"
	@echo "  make wait-cluster         Wait for Minikube and the Kubernetes node to be ready"
	@echo "  make ingress-ready        Enable and wait for the Minikube ingress addon"
	@echo "  make wait-apps            Wait for the main deployments to become available"
	@echo "  make build                Build the API Docker image"
	@echo "  make image-load           Load the image into Minikube"
	@echo "  make tf-init              Initialize Terraform"
	@echo "  make deploy               Build, load, and apply infrastructure with Terraform"
	@echo "  make redeploy             Rebuild image, reload it, and restart the API deployment"
	@echo "  make destroy              Destroy Terraform-managed resources"
	@echo "  make status               Show pod and service status"
	@echo "  make logs-api             Show API logs"
	@echo "  make logs-db              Show PostgreSQL logs"
	@echo "  make logs-prom            Show Prometheus logs"
	@echo "  make logs-grafana         Show Grafana logs"
	@echo "  make port-forward-api     Expose API at http://127.0.0.1:1323"
	@echo "  make port-forward-prometheus  Expose Prometheus at http://127.0.0.1:9090"
	@echo "  make port-forward-grafana Expose Grafana at http://127.0.0.1:3000"

minikube-start:
	minikube start --driver=docker

wait-cluster:
	kubectl wait --for=condition=Ready node/minikube --timeout=120s

ingress-ready:
	minikube addons enable ingress
	kubectl wait -n ingress-nginx --for=condition=Available deployment/ingress-nginx-controller --timeout=180s
	kubectl wait -n ingress-nginx --for=condition=Complete job/ingress-nginx-admission-create --timeout=180s
	kubectl wait -n ingress-nginx --for=condition=Complete job/ingress-nginx-admission-patch --timeout=180s
	kubectl wait -n ingress-nginx --for=jsonpath='{.subsets[0].addresses[0].ip}' endpoints/ingress-nginx-controller-admission --timeout=180s

wait-apps:
	kubectl wait -n $(NAMESPACE) --for=condition=Available deployment/xaults-postgres --timeout=180s
	kubectl wait -n $(NAMESPACE) --for=condition=Available deployment/xaults-api --timeout=180s
	kubectl wait -n $(NAMESPACE) --for=condition=Available deployment/prometheus --timeout=180s
	kubectl wait -n $(NAMESPACE) --for=condition=Available deployment/grafana --timeout=180s

build:
	docker build -t $(IMAGE_NAME) .

image-load:
	minikube image load $(IMAGE_NAME)

tf-init:
	terraform -chdir=$(TERRAFORM_DIR) init

deploy: minikube-start wait-cluster ingress-ready build image-load tf-init
	for attempt in 1 2 3; do \
		terraform -chdir=$(TERRAFORM_DIR) apply -auto-approve && break; \
		if [ $$attempt -eq 3 ]; then exit 1; fi; \
		echo "terraform apply failed; waiting for cluster components to settle before retry $$((attempt + 1))..."; \
		sleep 15; \
	done
	$(MAKE) wait-apps

redeploy: build image-load
	kubectl rollout restart deployment/xaults-api -n $(NAMESPACE)

destroy:
	terraform -chdir=$(TERRAFORM_DIR) destroy

status:
	kubectl get pods -n $(NAMESPACE)
	kubectl get svc -n $(NAMESPACE)

logs-api:
	kubectl logs -n $(NAMESPACE) deployment/xaults-api -f

logs-db:
	kubectl logs -n $(NAMESPACE) deployment/xaults-postgres -f

logs-prom:
	kubectl logs -n $(NAMESPACE) deployment/prometheus -f

logs-grafana:
	kubectl logs -n $(NAMESPACE) deployment/grafana -f

port-forward-api:
	@echo "Starting API port-forward..."
	kubectl port-forward -n $(NAMESPACE) svc/xaults-api 1323:80 > api.log 2>&1 &

port-forward-prometheus:
	@echo "Starting Prometheus port-forward..."
	kubectl port-forward -n $(NAMESPACE) svc/prometheus 9090:9090 > prometheus.log 2>&1 &

port-forward-grafana:
	@echo "Starting Grafana port-forward..."
	kubectl port-forward -n $(NAMESPACE) svc/grafana 3000:3000 > grafana.log 2>&1 &

port-forward-all: port-forward-api port-forward-prometheus port-forward-grafana
