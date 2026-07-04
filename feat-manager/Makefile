SHELL := /bin/sh
GOCACHE ?= $(CURDIR)/.cache/go-build

.PHONY: setup deps-up deps-down test run build migrate-up migrate-down fmt seed-phase2 cache-dir

cache-dir:
	mkdir -p "$(GOCACHE)"

setup:
	@if [ ! -f .env ]; then cp .env.example .env; fi
	@echo ".env ready"

deps-up:
	docker compose up -d mysql

deps-down:
	docker compose down

test: cache-dir
	GOCACHE="$(GOCACHE)" go test ./...

run: cache-dir
	set -a && . ./.env && set +a && GOCACHE="$(GOCACHE)" go run ./cmd/api

build: cache-dir
	mkdir -p ./bin
	GOCACHE="$(GOCACHE)" go build -o ./bin/feat-manager ./cmd/api

migrate-up:
	docker compose --profile tools run --rm migrate up

migrate-down:
	docker compose --profile tools run --rm migrate down 1

fmt:
	gofmt -w ./cmd ./internal

seed-phase2:
	docker compose exec -T mysql mysql -ufeat_manager -pfeat_manager feat_manager < db/seeds/phase2_tenants.sql
