.PHONY: up down migrate postgres recreate-db build logs test test-verbose test-coverage proto

DC=docker compose
DB_USER=clicks_user
DB_PASS=clicks_password
DB_NAME=clicks_db
DB_HOST=localhost
DB_PORT=5432
MIGRATIONS_DIR=migrations
SEEDS_DIR=scripts
PROJECT_NAME=clicks-counter
NETWORK=$(PROJECT_NAME)_clicks-network
PROTO_DIR=api/proto

COUNTER_PKG=pkg/counter
STATS_PKG=pkg/stats

up:
	$(DC) up

down:
	$(DC) down -v

build:
	$(DC) build

rebuild:
	$(DC) up --build

logs:
	$(DC) logs -f

postgres:
	$(DC) up -d postgres
	@until docker exec $$(docker ps -q -f name=postgres) pg_isready -U $(DB_USER) -d $(DB_NAME); do \
		echo "Waiting for postgres..."; \
		sleep 1; \
	done

recreate-db: postgres
	docker exec $$(docker ps -q -f name=postgres) dropdb -U $(DB_USER) --if-exists $(DB_NAME)
	docker exec $$(docker ps -q -f name=postgres) createdb -U $(DB_USER) $(DB_NAME)

migrate: recreate-db
	@for file in $(MIGRATIONS_DIR)/*.up.sql; do \
		echo "Applying $$file..."; \
		docker exec -i $$(docker ps -q -f name=postgres) psql -U $(DB_USER) -d $(DB_NAME) < $$file; \
	done

seed:
	@echo "Applying seed files..."
	@for file in $(SEEDS_DIR)/*.sql; do \
		if [ -f "$$file" ] && [ "$${file##*.}" = "sql" ] && ! [ "$${file%.*}" = "down" ]; then \
			echo "Applying seed: $$file"; \
			docker exec -i $$(docker ps -q -f name=postgres) psql -U $(DB_USER) -d $(DB_NAME) < $$file; \
		fi \
	done
	@echo "Seeds applied successfully!"

restart-app:
	$(DC) restart app

app-logs:
	$(DC) logs -f app

db-logs:
	$(DC) logs -f postgres

start: migrate seed build up

reset-db: recreate-db migrate seed

redis-cli:
	docker exec -it $$(docker ps -q -f name=redis) redis-cli

redis-logs:
	$(DC) logs -f redis

proto:
	@echo "Generating proto files..."
	@mkdir -p $(COUNTER_PKG) $(STATS_PKG)
	
	protoc -I=$(PROTO_DIR) \
		--go_out=$(COUNTER_PKG) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(COUNTER_PKG) \
		--go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(COUNTER_PKG) \
		--grpc-gateway_opt=paths=source_relative \
		$(PROTO_DIR)/counter.proto
	
	protoc -I=$(PROTO_DIR) \
		--go_out=$(STATS_PKG) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(STATS_PKG) \
		--go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(STATS_PKG) \
		--grpc-gateway_opt=paths=source_relative \
		$(PROTO_DIR)/stats.proto

.DEFAULT_GOAL := start
