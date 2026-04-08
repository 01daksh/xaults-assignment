.PHONY: up down logs test

up: ## Start all services (builds if needed)
	docker compose up --build -d

down: ## Stop all services and remove volumes
	docker compose down -v

logs: ## Stream application logs
	docker compose logs -f app
