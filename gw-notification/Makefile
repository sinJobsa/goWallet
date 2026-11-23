.PHONY: help build run test test-race lint clean

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

## Собрать бинарник под Linux
build:
	GOOS=linux GOARCH=amd64 go build -o main ./cmd

## Запустить локально с config.env
run:
	go run ./cmd -c config.env

## Линтер
lint:
	golangci-lint run

## Очистить бинарник
clean:
	rm -f main