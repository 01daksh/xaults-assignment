SHELL := /bin/zsh

IMAGE_NAME := xaults-assignment:latest
NAMESPACE := xaults
TERRAFORM_DIR := terraform

.PHONY: help minikube-start wait-cluster build image-load tf-init deploy redeploy destroy status logs-api logs-db logs-prom logs-grafana port-forward-api port-forward-prometheus port-forward-grafana

help:
	@echo "Available targets:"
	@echo "  make minikube-start       Start Minikube with the Docker driver"
	@echo "  make wait-cluster         Wait for Minikube and the Kubernetes node to be ready"
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

build:
	docker build -t $(IMAGE_NAME) .

image-load:
	minikube image load $(IMAGE_NAME)

tf-init:
	terraform -chdir=$(TERRAFORM_DIR) init

deploy: minikube-start wait-cluster build image-load tf-init
	terraform -chdir=$(TERRAFORM_DIR) apply -auto-approve

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
	kubectl port-forward -n $(NAMESPACE) svc/xaults-api 1323:80

port-forward-prometheus:
	kubectl port-forward -n $(NAMESPACE) svc/prometheus 9090:9090

port-forward-grafana:
	kubectl port-forward -n $(NAMESPACE) svc/grafana 3000:3000
