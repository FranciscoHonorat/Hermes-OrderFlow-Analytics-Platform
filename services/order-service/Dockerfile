FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /order-service ./cmd/server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /order-service /app/order-service

EXPOSE 8080

ENV APP_ENV=development \
    HTTP_PORT=8080 \
    HTTP_MODE=release \
    DB_HOST=postgres \
    DB_PORT=5432 \
    DB_USER=postgres \
    DB_PASSWORD=postgres \
    DB_NAME=orderflow \
    DB_SSLMODE=disable \
    MESSAGING_BROKER=kafka:9092 \
    MESSAGING_CLIENT_ID=order-service \
    MESSAGING_WRITE_TIMEOUT=5s

ENTRYPOINT ["/app/order-service"]
