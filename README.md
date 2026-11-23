# goWallet

Микросервисный кошелёк на Go: регистрация/логин, балансы в нескольких валютах,
депозит/вывод, обмен валют по актуальным курсам, аналитика транзакций и уведомления.

## Архитектура

| Сервис | Назначение | Порт | Хранилище |
|---|---|---|---|
| `wallet` | Основной API: пользователи, балансы, операции | `8080` | PostgreSQL, Redis |
| `exchanger` | Курсы валют (gRPC) | `50051` | PostgreSQL |
| `analytics` | Аналитика транзакций | — | ClickHouse |
| `notification` | Уведомления о событиях | — | MongoDB, Kafka |

Взаимодействие:
- `wallet` → `exchanger` по gRPC (курсы);
- `wallet` → Kafka (события транзакций);
- `analytics` ← Kafka → ClickHouse;
- `notification` ← Kafka → MongoDB.

## Стек

- **Go** (Gin для HTTP, gRPC для exchanger)
- **PostgreSQL** — пользователи, балансы, транзакции, курсы
- **Redis** — кеш курсов
- **Kafka** — шина событий
- **ClickHouse** — аналитическое хранилище
- **MongoDB** — уведомления
- **Docker Compose** — локальный запуск
- **Swagger** (`swaggo`) — документация API

## Быстрый старт

### Требования
- Docker + Docker Compose
- Go 1.22+ (для локальной разработки без Docker)

### Запуск


# собрать и поднять все сервисы
docker compose up -d --build

# проверить, что всё поднялось
docker compose ps

API

Swagger UI: http://localhost:8080/swagger/index.html

Основные эндпоинты (префикс /api/v1):

Метод	Путь	Описание	Auth
POST	/register	Регистрация	—
POST	/login	Логин, возвращает JWT	—
GET	/balance	Балансы пользователя	Bearer
POST	/wallet/deposit	Пополнение	Bearer
POST	/wallet/withdraw	Вывод	Bearer
GET	/exchange/rates	Курсы валют	Bearer
POST	/exchange	Обмен валют	Bearer