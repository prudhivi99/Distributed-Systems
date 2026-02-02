# Makefile for Go MongoDB API - Phase 4: Docker

.PHONY: help build up down logs ps restart clean prune test shell shell-db health

# Default target
.DEFAULT_GOAL := help

# Variables
APP_NAME := go-mongo-api
DOCKER_COMPOSE := docker compose
DOCKER := docker

help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the Docker images
	@echo "🔨 Building Docker images..."
	$(DOCKER_COMPOSE) build --no-cache

up: ## Start all services (detached mode)
	@echo "🚀 Starting services..."
	$(DOCKER_COMPOSE) up -d
	@echo "✅ Services started!"
	@echo "📍 API available at: http://localhost:8080"
	@echo "📍 MongoDB available at: localhost:27017"

down: ## Stop all services
	@echo "🛑 Stopping services..."
	$(DOCKER_COMPOSE) down
	@echo "✅ Services stopped!"

logs: ## View logs from all services
	$(DOCKER_COMPOSE) logs -f

logs-api: ## View logs from API service only
	$(DOCKER_COMPOSE) logs -f api

logs-db: ## View logs from MongoDB service only
	$(DOCKER_COMPOSE) logs -f mongodb

ps: ## Show running containers
	$(DOCKER_COMPOSE) ps

restart: ## Restart all services
	@echo "🔄 Restarting services..."
	$(DOCKER_COMPOSE) restart
	@echo "✅ Services restarted!"

restart-api: ## Restart only the API service
	@echo "🔄 Restarting API service..."
	$(DOCKER_COMPOSE) restart api
	@echo "✅ API service restarted!"

clean: ## Stop and remove containers, networks, volumes
	@echo "🧹 Cleaning up..."
	$(DOCKER_COMPOSE) down -v
	@echo "✅ Cleanup complete!"

prune: ## Remove all unused Docker resources (CAUTION: system-wide)
	@echo "⚠️  WARNING: This will remove ALL unused Docker resources!"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		$(DOCKER) system prune -af --volumes; \
		echo "✅ Docker system pruned!"; \
	else \
		echo "❌ Aborted."; \
	fi

test: ## Test API endpoints
	@echo "🧪 Testing API endpoints..."
	@echo "\n1. Health Check:"
	@curl -s http://localhost:8080/health | jq '.' || echo "❌ Health check failed"
	@echo "\n2. Create User:"
	@curl -s -X POST http://localhost:8080/api/users \
		-H "Content-Type: application/json" \
		-d '{"name":"Docker Test","email":"docker@example.com","age":25}' | jq '.' || echo "❌ Create user failed"
	@echo "\n3. Get All Users:"
	@curl -s http://localhost:8080/api/users | jq '.' || echo "❌ Get users failed"

shell: ## Open shell in API container
	$(DOCKER_COMPOSE) exec api sh

shell-db: ## Open MongoDB shell
	$(DOCKER_COMPOSE) exec mongodb mongosh -u admin -p password123

health: ## Check health of all services
	@echo "🏥 Checking service health..."
	@echo "\n📍 API Health:"
	@curl -s http://localhost:8080/health | jq '.data' || echo "❌ API not responding"
	@echo "\n📍 MongoDB Health:"
	@$(DOCKER_COMPOSE) exec -T mongodb mongosh -u admin -p password123 --eval "db.adminCommand('ping')" --quiet || echo "❌ MongoDB not responding"

status: ## Show detailed status of services
	@echo "📊 Service Status:"
	@echo ""
	@$(DOCKER_COMPOSE) ps
	@echo ""
	@echo "💾 Volumes:"
	@$(DOCKER) volume ls | grep $(APP_NAME) || echo "No volumes found"
	@echo ""
	@echo "🌐 Networks:"
	@$(DOCKER) network ls | grep $(APP_NAME) || echo "No networks found"

rebuild: ## Rebuild and restart services
	@echo "🔨 Rebuilding and restarting..."
	$(DOCKER_COMPOSE) down
	$(DOCKER_COMPOSE) build --no-cache
	$(DOCKER_COMPOSE) up -d
	@echo "✅ Services rebuilt and restarted!"

dev: ## Start services in development mode (foreground)
	@echo "🔧 Starting in development mode..."
	$(DOCKER_COMPOSE) up

# Quick commands
start: up ## Alias for 'up'
stop: down ## Alias for 'down'