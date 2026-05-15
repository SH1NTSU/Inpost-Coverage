SHELL := /usr/bin/env bash
BIN := bin
PKG := ./...
DOCKER_COMPOSE := docker-compose

.PHONY: help tidy build build-all run-api run-scraper migrate-up migrate-down test lint fmt docker-up docker-down \
        start stop wait-db check-seed seed seed-fast web-install logs env-init

help:
	@echo "Quick start:"
	@echo "  make start        # everything: db, redis, migrations, api, scraper, web"
	@echo "  make stop         # tear it all down again"
	@echo "  make seed         # one-time data ingest (~30 min, hits OSM)"
	@echo "  make seed-fast    # quick demo seed: just one city (~3 min)"
	@echo ""
	@echo "Lower level:"
	@echo "  make tidy | build | test | lint | fmt"
	@echo "  make run-api | run-scraper"
	@echo "  make migrate-up | migrate-down"
	@echo "  make docker-up | docker-down"
	@echo "  make logs         # tail all service logs"

tidy:
	go mod tidy

build: build-all

build-all:
	@mkdir -p $(BIN)
	@for d in cmd/*; do \
		name=$$(basename $$d); \
		echo "build $$name"; \
		go build -o $(BIN)/$$name ./$$d || exit 1; \
	done

run-api:
	go run ./cmd/api

run-scraper:
	go run ./cmd/scraper

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

test:
	go test -race -count=1 $(PKG)

lint:
	go vet $(PKG)

fmt:
	gofmt -s -w .

docker-up:
	$(DOCKER_COMPOSE) up -d postgres redis

docker-down:
	$(DOCKER_COMPOSE) down

## High-level: one-shot start/stop ---------------------------------------------

start: env-init docker-up wait-db migrate-up web-install check-seed
	@./scripts/start-all.sh

env-init:
	@if [ ! -f .env ]; then \
	  echo "→ Creating .env from .env.example"; \
	  cp .env.example .env; \
	fi
	@if [ ! -f web/.env ]; then \
	  echo "→ Creating web/.env from web/.env.example"; \
	  cp web/.env.example web/.env; \
	fi

stop:
	@./scripts/stop-all.sh
	@$(DOCKER_COMPOSE) stop postgres redis

wait-db:
	@echo "→ Waiting for Postgres..."
	@for i in $$(seq 1 30); do \
	  if $(DOCKER_COMPOSE) exec -T postgres pg_isready -U paczkomat -d paczkomat >/dev/null 2>&1; then \
	    echo "✓ Postgres ready"; exit 0; \
	  fi; sleep 1; \
	done; \
	echo "✗ Postgres did not become ready in 30s"; exit 1

check-seed:
	@count=$$( $(DOCKER_COMPOSE) exec -T postgres psql -U paczkomat -d paczkomat -tAc 'SELECT COUNT(*) FROM points' 2>/dev/null | tr -d '[:space:]' ); \
	if [ -z "$$count" ] || [ "$$count" = "0" ]; then \
	  echo ""; \
	  echo "⚠  No locker data in the database yet."; \
	  echo "   Run \`make seed\` once (~30 min, hits InPost + OSM APIs)"; \
	  echo "   Or \`make seed-fast\` for a quick Kraków-only demo (~3 min)"; \
	  echo "   The UI will still load — it just won't show much without data."; \
	  echo ""; \
	else \
	  echo "✓ Data present ($$count locker rows)"; \
	fi

seed:
	@echo "→ [1/3] Importing InPost lockers across Poland..."
	@go run ./cmd/import-points
	@echo "→ [2/3] Importing competitor lockers from OSM..."
	@go run ./cmd/import-competitors
	@echo "→ [3/3] Importing OSM anchor POIs (shops, fuel, settlements)..."
	@go run ./cmd/import-pois
	@echo "✓ Seed complete"

seed-fast:
	@echo "→ Importing InPost lockers for Kraków only..."
	@go run ./cmd/import-points -city Kraków
	@echo "→ Importing competitor lockers (whole country bbox, ~1 min)..."
	@go run ./cmd/import-competitors -province małopolskie
	@echo "→ Importing OSM anchors for małopolskie only..."
	@go run ./cmd/import-pois -province małopolskie
	@echo "✓ Fast seed complete (małopolskie only; select 'małopolskie' in the UI)"

web-install:
	@cd web && if [ ! -d node_modules ]; then \
	  echo "→ Installing web dependencies..."; \
	  npm install --silent; \
	else \
	  echo "✓ web/node_modules present"; \
	fi

logs:
	@echo "Tailing all service logs (Ctrl-C to stop)..."
	@tail -f /tmp/inpost-logs/api.log /tmp/inpost-logs/scraper.log /tmp/inpost-logs/web.log
