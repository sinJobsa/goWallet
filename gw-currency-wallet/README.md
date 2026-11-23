# gw-currency-wallet

Кошелёк-обменник с JWT-аутентификацией. Единственный публичный сервис системы —
принимает HTTP/REST от пользователя, ходит по gRPC в `gw-exchanger` за курсами,
публикует события о крупных переводах в Kafka.

## Что делает

- **Регистрация и авторизация** пользователей (JWT, HS256)
- **Баланс** в трёх валютах: USD, RUB, EUR
- **Пополнение** и **вывод** средств
- **Курсы валют** — через gRPC к `gw-exchanger`
- **Обмен валют** — с кэшем курсов в Redis
- **Публикация в Kafka** события о переводе ≥ 30 000

## Стек

- Go 1.22
- Gin (HTTP)
- PostgreSQL 16 + pgx (пользователи, балансы, транзакции)
- Redis 7 (кэш курсов)
- Kafka (продюсер событий)
- gRPC + Protocol Buffers (клиент к exchanger)
- JWT HS256 (`golang-jwt/jwt/v5`)
- Swagger (`swaggo/swag`)


## Запуск

### 1. Локально (без Docker)

**Нужны:** Postgres, Redis, Kafka, exchanger (запущены).

```bash
# миграции
docker exec -i <postgres-container> psql -U postgres -d wallet < migrations/001_init.sql
```
# запуск
go run ./cmd -c config.env

### 2. Docker
docker compose up -d wallet
docker compose logs -f wallet