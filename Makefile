SHELL := /usr/bin/env bash
BIN := bin
PKG := ./...
DOCKER_COMPOSE := docker-compose

.PHONY: help tidy build build-all run-api run-scraper migrate-up migrate-down test lint fmt docker-up docker-down

help:
	@echo "make tidy | build | run-api | run-scraper | migrate-up | migrate-down | test | lint | fmt | docker-up | docker-down"

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
	$(DOCKER_COMPOSE) up -d postgres

docker-down:
	$(DOCKER_COMPOSE) down
