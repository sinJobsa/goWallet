# gw-analytics

Сервис аналитики крупных денежных переводов. Читает события из Kafka,
считает метрики и сохраняет агрегаты в ClickHouse. Внутренний сервис.

## Что делает

- **Читает** из Kafka топик `large_transfers` (своя consumer group `gw-analytics`)
- **Вычисляет** latency: `received_at - created_at`
- **Сохраняет** сырые события в `analytics.events` (`ReplacingMergeTree` — дедупликация)
- **Агрегирует** по 1 минуте через materialized view `analytics.aggregates_1m`
- **At-least-once**: commit offset только после успешного сохранения
- **Graceful shutdown** + **structured logging** (JSON)

## Стек

- Go 1.22
- Kafka (`segmentio/kafka-go`) — consumer с manual commit
- ClickHouse (`clickhouse-go/v2`) — нативный протокол
- `log/slog` (structured logging)