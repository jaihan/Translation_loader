.PHONY: help up down restart logs tidy fmt lint vet test unit integration coverage run clean

help:
	@echo "Available commands:"
	@echo "  make up            Start PostgreSQL"
	@echo "  make down          Stop PostgreSQL"
	@echo "  make restart       Restart PostgreSQL"
	@echo "  make logs          View PostgreSQL logs"
	@echo "  make run           Run application"
	@echo "  make test          Run all tests"
	@echo "  make unit          Run unit tests"
	@echo "  make integration   Run integration tests"
	@echo "  make fmt           Format code"
	@echo "  make vet           Run go vet"
	@echo "  make clean         Clean test cache"

up:
	docker compose up -d

down:
	docker compose down -v

restart: down up

logs:
	docker compose logs -f postgres

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

run:
	go run ./cmd/app

test:
	go test ./... -v

unit:
	go test ./tests/unit -v

integration:
	go test ./tests/integration -v

clean:
	go clean -testcache