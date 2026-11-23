.PHONY: help build run test test-race lint migrate swagger clean

# Показать список целей
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

## Собрать бинарник под Linux
build:
	GOOS=linux GOARCH=amd64 go build -o main ./cmd

## Запустить локально с config.env
run:
	go run ./cmd -c config.env

## Запустить тесты
test:
	go test ./...

## Запустить тесты с race detector
test-race:
	go test -race ./...

## Линтер
lint:
	golangci-lint run

## Применить миграции (нужен запущенный Postgres)
migrate:
	docker exec -i postgres-wallet psql -U postgres -d wallet < migrations/001_init.sql

## Перегенерировать Swagger
swagger:
	swag init -g cmd/main.go -o docs

## Очистить бинарник
clean:
	rm -f main