-include .env
export

.PHONY: run build tidy migrate-up migrate-down lint test

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

tidy:
	go mod tidy

migrate-up:
	@which migrate > /dev/null || (echo "install golang-migrate: https://github.com/golang-migrate/migrate" && exit 1)
	migrate -path migrations -database "$$DATABASE_URL" up

migrate-down:
	migrate -path migrations -database "$$DATABASE_URL" down 1

lint:
	@which golangci-lint > /dev/null || (echo "install golangci-lint: https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...

test:
	go test ./...
