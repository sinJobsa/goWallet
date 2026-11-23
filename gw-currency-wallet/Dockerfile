FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.work* ./
COPY proto-exchange/ ./proto-exchange/
COPY gw-exchanger/ ./gw-exchanger/
COPY gw-currency-wallet/ ./gw-currency-wallet/
COPY gw-notification/ ./gw-notification/
COPY gw-analytics/ ./gw-analytics/

WORKDIR /app/gw-currency-wallet
RUN apk add --no-cache ca-certificates
RUN for i in 1 2 3 4 5; do \
      GOPROXY=https://goproxy.cn,direct GOSUMDB=off go mod download && break || sleep 5; \
    done

COPY gw-currency-wallet/ ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /wallet ./cmd

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /wallet /wallet
COPY --from=builder /app/gw-currency-wallet/config.env /config.env

EXPOSE 8080

ENTRYPOINT ["/wallet", "-c", "/config.env"]