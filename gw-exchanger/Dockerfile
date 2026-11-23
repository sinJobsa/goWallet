# gw-exchanger/Dockerfile
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.work* ./
COPY proto-exchange/ ./proto-exchange/
COPY gw-exchanger/ ./gw-exchanger/
COPY gw-currency-wallet/ ./gw-currency-wallet/
COPY gw-notification/ ./gw-notification/
COPY gw-analytics/ ./gw-analytics/

WORKDIR /app/gw-exchanger
RUN apk add --no-cache ca-certificates
RUN for i in 1 2 3 4 5; do \
      GOPROXY=https://goproxy.cn,direct GOSUMDB=off go mod download && break || sleep 5; \
    done
RUN go build -o /exchanger ./cmd

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /exchanger /exchanger
COPY --from=builder /app/gw-exchanger/config.env /config.env

EXPOSE 50051
CMD ["/exchanger", "-c", "/config.env"]