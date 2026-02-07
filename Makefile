.PHONY: all build clean test proto swagger certs up down run-gateway run-users run-orders migrate lint

# Variables
DOCKER_COMPOSE = docker-compose -f deploy/docker-compose.yml
PROTO_DIR = api/proto
GEN_DIR = api/gen

# Build all services
all: build

build:
	go build -o bin/gateway ./cmd/gateway
	go build -o bin/users ./cmd/users
	go build -o bin/orders ./cmd/orders

clean:
	rm -rf bin/
	rm -rf $(GEN_DIR)

# Testing
test:
	go test -v -race -cover ./...

test-unit:
	go test -v -short ./internal/...

# Generate Protocol Buffers
proto:
	@mkdir -p $(GEN_DIR)
	protoc --go_out=$(GEN_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(GEN_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/users/v1/*.proto $(PROTO_DIR)/orders/v1/*.proto

# Generate Swagger documentation
swagger:
	swag init -g cmd/gateway/main.go -o docs/swagger --parseDependency --parseInternal

# Generate TLS certificates
certs:
	@chmod +x scripts/certs/generate.sh
	@bash scripts/certs/generate.sh

# Docker commands
up:
	$(DOCKER_COMPOSE) up -d --build

down:
	$(DOCKER_COMPOSE) down -v

logs:
	$(DOCKER_COMPOSE) logs -f

# Run services locally
run-gateway:
	go run ./cmd/gateway

run-users:
	go run ./cmd/users

run-orders:
	go run ./cmd/orders

# Run all services locally (requires separate terminals)
run-all:
	@echo "Run these commands in separate terminals:"
	@echo "  make run-users"
	@echo "  make run-orders"
	@echo "  make run-gateway"

# Database migrations
migrate:
	go run ./scripts/migrate/main.go

# Linting
lint:
	golangci-lint run ./...

# Install development tools
tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Help
help:
	@echo "Available targets:"
	@echo "  build        - Build all services"
	@echo "  test         - Run all tests"
	@echo "  proto        - Generate gRPC code from proto files"
	@echo "  swagger      - Generate Swagger documentation"
	@echo "  certs        - Generate TLS/mTLS certificates"
	@echo "  up           - Start all services with Docker Compose"
	@echo "  down         - Stop all services"
	@echo "  run-gateway  - Run gateway service locally"
	@echo "  run-users    - Run users service locally"
	@echo "  run-orders   - Run orders service locally"
	@echo "  tools        - Install development tools"
