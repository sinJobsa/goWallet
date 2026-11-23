# --- Stage 1: builder ---
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.work* ./
COPY proto-exchange/ ./proto-exchange/
COPY gw-exchanger/ ./gw-exchanger/
COPY gw-currency-wallet/ ./gw-currency-wallet/
COPY gw-notification/ ./gw-notification/
COPY gw-analytics/ ./gw-analytics/

WORKDIR /app/gw-analytics
RUN apk add --no-cache ca-certificates
RUN for i in 1 2 3 4 5; do \
      GOPROXY=https://goproxy.cn,direct GOSUMDB=off go mod download && break || sleep 5; \
    done
RUN CGO_ENABLED=0 GOOS=linux go build -o /analytics ./cmd

# --- Stage 2: runtime ---
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /analytics /analytics
COPY --from=builder /app/gw-analytics/config.env /config.env

ENTRYPOINT ["/analytics", "-c", "/config.env"]