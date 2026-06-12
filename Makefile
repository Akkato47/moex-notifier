SHELL := /bin/bash
MODULE := github.com/Akkato47/moex-notifier
BIN_DIR := bin

.PHONY: up down migrate proto lint test build

up:
	docker compose -f deployments/docker-compose.yml up -d

down:
	docker compose -f deployments/docker-compose.yml down

migrate:
	goose -dir migrations postgres "$$DATABASE_URL" up

proto:
	buf generate

lint:
	golangci-lint run ./...

test:
	go test ./...

build:
	go build -o $(BIN_DIR)/ingester  ./cmd/ingester
	go build -o $(BIN_DIR)/processor ./cmd/processor
	go build -o $(BIN_DIR)/notifier  ./cmd/notifier

tidy:
	go mod tidy
