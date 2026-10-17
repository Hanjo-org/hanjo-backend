.PHONY: help dev run build test tidy lint seed docker-up docker-down migrate-create migrate-up migrate-down migrate-fresh migrate-version migrate-force

# Load .env file if present
ifneq (,$(wildcard .env))
    include .env
    export
endif

# Database connection URL for migrations (defaults to postgres localhost)
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= postgres
DB_PASSWORD ?=
DB_NAME ?= hanjo_db
DB_SSLMODE ?= disable
DB_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

help: ## Display available make commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

dev: ## Run application with live reloading (Air)
	@air

run: ## Run application locally
	@go run ./cmd/api

seed: ## Run database seeder (initial organization, brands, outlets, users)
	@go run ./cmd/seed

build: ## Build binary executable
	@mkdir -p bin
	@go build -o bin/server ./cmd/api

test: ## Run unit tests
	@go test -v ./...

tidy: ## Clean up and download dependencies
	@go mod tidy

docker-up: ## Start docker containers (app & postgres)
	@docker-compose up -d

docker-down: ## Stop docker containers
	@docker-compose down

migrate-create: ## Create new SQL migration (usage: make migrate-create NAME=create_orders_table)
	@if [ -z "$(NAME)" ]; then echo "❌ Please specify NAME. Example: make migrate-create NAME=create_orders_table"; exit 1; fi
	@migrate create -ext sql -dir migrations -seq $(NAME)

migrate-up: ## Run pending up migrations (usage: make migrate-up [N=1])
	@if [ -z "$(N)" ]; then \
		migrate -path migrations -database "$(DB_URL)" up; \
	else \
		migrate -path migrations -database "$(DB_URL)" up $(N); \
	fi

migrate-down: ## Rollback down migrations (usage: make migrate-down [N=1])
	@if [ -z "$(N)" ]; then \
		migrate -path migrations -database "$(DB_URL)" down 1; \
	else \
		migrate -path migrations -database "$(DB_URL)" down $(N); \
	fi

migrate-fresh: ## Drop everything and re-run all migrations from scratch
	@migrate -path migrations -database "$(DB_URL)" drop -f
	@migrate -path migrations -database "$(DB_URL)" up

migrate-version: ## Print current database migration version
	@migrate -path migrations -database "$(DB_URL)" version

migrate-force: ## Force set migration version (usage: make migrate-force VERSION=1)
	@if [ -z "$(VERSION)" ]; then echo "❌ Please specify VERSION. Example: make migrate-force VERSION=1"; exit 1; fi
	@migrate -path migrations -database "$(DB_URL)" force $(VERSION)
