SERVICE=services/order-service
BINARY=bin/order-service
DOCKER_COMPOSE=docker compose

# go.work is a multi-module workspace: `go <cmd> ./...` only works from
# inside one of these module directories, never from the repo root. Every
# target below loops over them explicitly instead.
MODULES=shared services/order-service services/inventory-service services/cdc-connector services/user-service

.PHONY: help build run test test-integration fmt tidy docker-build docker-build-inventory docker-build-cdc docker-build-user docker-up docker-down

help:
	@echo "Targets:"
	@echo "  build               Build the order-service binary (linux/amd64)"
	@echo "  run                 Run the order-service locally"
	@echo "  test                Run all go unit tests (every workspace module)"
	@echo "  test-integration    Run integration tests (requires Docker; every workspace module)"
	@echo "  fmt                 Run go fmt (every workspace module)"
	@echo "  tidy                Run go mod tidy (every workspace module)"
	@echo "  docker-build        Build the order-service docker image"
	@echo "  docker-build-inventory  Build the inventory-service docker image"
	@echo "  docker-build-cdc    Build the cdc-connector docker image"
	@echo "  docker-build-user   Build the user-service docker image"
	@echo "  docker-up           Docker compose up (builds images)"
	@echo "  docker-down         Docker compose down"

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BINARY) ./$(SERVICE)/cmd/server

run:
	go run ./$(SERVICE)/cmd/server

test:
	@for m in $(MODULES); do \
		echo "==> go test ./... ($$m)"; \
		(cd $$m && go test ./...) || exit 1; \
	done

test-integration:
	@for m in $(MODULES); do \
		echo "==> go test -tags=integration ./... ($$m)"; \
		(cd $$m && go test -tags=integration ./...) || exit 1; \
	done

fmt:
	@for m in $(MODULES); do \
		(cd $$m && go fmt ./...); \
	done

tidy:
	@for m in $(MODULES); do \
		echo "==> go mod tidy ($$m)"; \
		(cd $$m && go mod tidy) || exit 1; \
	done

docker-build:
	docker build -t hermes-orderflow/order-service -f $(SERVICE)/Dockerfile .

docker-build-inventory:
	docker build -t hermes-orderflow/inventory-service -f services/inventory-service/Dockerfile .

docker-build-cdc:
	docker build -t hermes-orderflow/cdc-connector -f services/cdc-connector/Dockerfile .

docker-build-user:
	docker build -t hermes-orderflow/user-service -f services/user-service/Dockerfile .

docker-up:
	$(DOCKER_COMPOSE) up --build -d

docker-down:
	$(DOCKER_COMPOSE) down