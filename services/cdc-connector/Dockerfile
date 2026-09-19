FROM golang:1.26-alpine AS builder

WORKDIR /app

# O build context é a raiz do repositório (ver docker-compose.yml) porque o
# go.mod deste serviço depende de shared/ via `replace ../../shared` — o
# Docker não enxerga nada fora do build context.
COPY shared/ ./shared/
COPY services/cdc-connector/ ./services/cdc-connector/

WORKDIR /app/services/cdc-connector
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /cdc-connector ./cmd/server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /cdc-connector /app/cdc-connector

EXPOSE 8082

ENV APP_ENV=development \
    HTTP_PORT=8082 \
    DB_HOST=postgres \
    DB_PORT=5432 \
    DB_USER=postgres \
    DB_PASSWORD=postgres \
    DB_SSLMODE=disable \
    SOURCES=order-service:orderflow,inventory-service:inventory \
    MESSAGING_BROKER=kafka:9092 \
    MESSAGING_CLIENT_ID=cdc-connector \
    POLL_INTERVAL=2s \
    POLL_BATCH_SIZE=50

ENTRYPOINT ["/app/cdc-connector"]
