# gw-exchanger

Сервис курсов валют. Предоставляет gRPC API для получения курсов USD/RUB/EUR.
Используется сервисом `gw-currency-wallet` для обмена валют.

## Что делает

- Хранит курсы валют в PostgreSQL (таблица `rates`)
- Отдаёт курсы по gRPC:
    - `GetExchangeRates` — все курсы к RUB
    - `GetExchangeRateForCurrency` — кросс-курс from→to
- Поддерживаемые валюты: **USD, RUB, EUR**

## Стек

- Go 1.22
- gRPC + Protocol Buffers
- PostgreSQL 16 + pgx
- `log/slog` (structured logging)
