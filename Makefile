.PHONY: deps lint test run docker-up docker-down docker-logs migrate-up migrate-down

GOFILES := $(shell find . -name '*.go' -not -path './vendor/*')

deps:
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go mod tidy

lint:
	golangci-lint run ./...

test:
	go test ./...

run:
	APP_ENV=local go run ./cmd/user-service

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

migrate-up:
	migrate -path ./migrations -database "postgres://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME?sslmode=$$DB_SSLMODE" up

migrate-down:
	migrate -path ./migrations -database "postgres://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME?sslmode=$$DB_SSLMODE" down
