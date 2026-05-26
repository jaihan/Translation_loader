.PHONY: up down run test unit integration tidy

up:
	docker compose up -d

down:
	docker compose down -v

tidy:
	go mod tidy

run:
	go run ./cmd/app

test:
	go test ./... -v

unit:
	go test ./tests/unit -v

integration:
	go test ./tests/integration -v
