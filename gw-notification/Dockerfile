# --- Stage 1: builder ---
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.work* ./
COPY proto-exchange/ ./proto-exchange/
COPY gw-exchanger/ ./gw-exchanger/
COPY gw-currency-wallet/ ./gw-currency-wallet/
COPY gw-notification/ ./gw-notification/
COPY gw-analytics/ ./gw-analytics/

# 2. Зависимости
WORKDIR /app/gw-notification
RUN apk add --no-cache ca-certificates
RUN for i in 1 2 3 4 5; do \
      GOPROXY=https://goproxy.cn,direct GOSUMDB=off go mod download && break || sleep 5; \
    done

# 3. Сборка
RUN CGO_ENABLED=0 GOOS=linux go build -o /notification ./cmd

# --- Stage 2: runtime ---
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /notification /notification
COPY --from=builder /app/gw-notification/config.env /config.env

ENTRYPOINT ["/notification", "-c", "/config.env"]