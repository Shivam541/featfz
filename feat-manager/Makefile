SHELL := /bin/sh

.PHONY: setup deps-up deps-down test run build migrate-up migrate-down fmt

setup:
	@if [ ! -f .env ]; then cp .env.example .env; fi
	@echo ".env ready"

deps-up:
	docker compose up -d mysql

deps-down:
	docker compose down

test:
	go test ./...

run:
	set -a && . ./.env && set +a && go run ./cmd/api

build:
	mkdir -p ./bin
	go build -o ./bin/feat-manager ./cmd/api

migrate-up:
	docker compose --profile tools run --rm migrate up

migrate-down:
	docker compose --profile tools run --rm migrate down 1

fmt:
	gofmt -w ./cmd ./internal
