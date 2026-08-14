SERVICE=services/order-service
BINARY=bin/order-service
DOCKER_COMPOSE=docker compose

.PHONY: help build run test fmt tidy docker-build docker-up docker-down

help:
	@echo "Targets:"
	@echo "  build               Build the order-service binary (linux/amd64)"
	@echo "  run                 Run the order-service locally"
	@echo "  test                Run all go tests"
	@echo "  fmt                 Run go fmt"
	@echo "  tidy                Run go mod tidy"
	@echo "  docker-build        Build the order-service docker image"
	@echo "  docker-up           Docker compose up (builds images)"
	@echo "  docker-down         Docker compose down"

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BINARY) ./$(SERVICE)/cmd/server

run:
	go run ./$(SERVICE)/cmd/server

test:
	go test ./...

fmt:
	go fmt ./...

tidy:
	go mod tidy

docker-build:
	docker build -t hermes-orderflow/order-service -f $(SERVICE)/Dockerfile $(SERVICE)

docker-up:
	$(DOCKER_COMPOSE) up --build -d

docker-down:
	$(DOCKER_COMPOSE) down