# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy workspace files
COPY go.work go.work.sum ./

# Copy the shared modules
COPY shared ./shared

# Copy order-service
COPY services/order-service ./services/order-service

WORKDIR /app/services/order-service

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /order-service ./cmd/server

# Runtime
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /order-service /app/order-service

EXPOSE 8080

ENV APP_ENV=production \
    HTTP_PORT=8080 \
    HTTP_MODE=release

ENTRYPOINT ["/app/order-service"]