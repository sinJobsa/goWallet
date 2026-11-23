# gw-notification

Сервис уведомлений о крупных денежных переводах. Читает события из Kafka,
сохраняет в MongoDB. Внутренний сервис — пользователь не общается с ним напрямую.

## Что делает

- **Читает** из Kafka топик `large_transfers` (события о переводах ≥ 30 000)
- **Сохраняет** в MongoDB коллекцию `large_transfers`
- **Идемпотентен**: повторное сообщение не создаёт дубль (`_id = transaction_id`)
- **At-least-once**: commit offset только после успешного сохранения
- **Graceful shutdown**: корректно завершается по `SIGTERM`
- **Structured logging** (JSON) с `transaction_id`

## Стек

- Go 1.22
- Kafka (`segmentio/kafka-go`) — consumer с manual commit
- MongoDB 7 (`mongo-driver`)
- `log/slog` (structured logging)

## Запуск

### 1. Локально

**Нужны:** Kafka, Mongo.

```bash
go run ./cmd -c config.env
```

### 2. Docker
# из корня монорепо
docker compose up -d notification
docker compose logs -f notification