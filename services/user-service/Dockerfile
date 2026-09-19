FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY shared/ ./shared/
COPY services/user-service/ ./services/user-service/

WORKDIR /app/services/user-service
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /user-service ./cmd/server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /user-service /app/user-service

EXPOSE 8083

ENV APP_ENV=development \
    HTTP_PORT=8083 \
    HTTP_MODE=release \
    DB_HOST=postgres \
    DB_PORT=5432 \
    DB_USER=postgres \
    DB_PASSWORD=postgres \
    DB_NAME=users \
    DB_SSLMODE=disable \
    JWT_EXPIRY=1h

ENTRYPOINT ["/app/user-service"]
