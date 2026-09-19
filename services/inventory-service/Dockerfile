FROM golang:1.25-alpine AS builder

WORKDIR /app

# O build context é a raiz do repositório (ver docker-compose.yml) porque o
# go.mod deste serviço depende de shared/ via `replace ../../shared` — o
# Docker não enxerga nada fora do build context.
COPY shared/ ./shared/
COPY services/inventory-service/ ./services/inventory-service/

WORKDIR /app/services/inventory-service
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /inventory-service ./cmd/server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /inventory-service /app/inventory-service

EXPOSE 8081

ENV APP_ENV=development \
    HTTP_PORT=8081 \
    HTTP_MODE=release \
    DB_HOST=postgres \
    DB_PORT=5432 \
    DB_USER=postgres \
    DB_PASSWORD=postgres \
    DB_NAME=inventory \
    DB_SSLMODE=disable

ENTRYPOINT ["/app/inventory-service"]
