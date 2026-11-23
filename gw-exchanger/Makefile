.PHONY: help build run test test-race test-integration lint migrate clean

# Показать список целей
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

## Собрать бинарник под Linux
build:
	GOOS=linux GOARCH=amd64 go build -o main ./cmd

## Запустить локально с config.env
run:
	go run ./cmd -c config.env

## Запустить unit-тесты
test:
	go test ./...

## Запустить тесты с race detector
test-race:
	go test -race ./...

## Запустить integration-тесты (нужен запущенный Postgres)
test-integration:
	go test -tags=integration ./internal/storages/postgres/ -v

## Линтер
lint:
	golangci-lint run

## Применить миграции (нужен запущенный Postgres)
migrate:
	docker exec -i postgres-exchanger psql -U postgres -d exchanger < migrations/001_init.sql

## Очистить бинарник
clean:
	rm -f main