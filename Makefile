BUILD_DIR := bin

.PHONY: server
server:
	go build -o $(BUILD_DIR)/server ./cmd/server

.PHONY: fmt
fmt:
	go fmt $(shell go list ./...)
	swag fmt

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)/*

.PHONY: deps
deps:
	go mod tidy

.PHONY: generate
generate:
	go generate ./...

.PHONY: docs
docs:
	@echo "Formatting swag doc comments..."
	swag fmt
	@echo ""
	@echo "Generating docs..."
	swag init -g ./cmd/server/server.go

.PHONY: migrate
migrate:
	@echo "Running migration command..."
	go run ./cmd/migration

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  make server    - Build the server executable"
	@echo "  make fmt       - Format the code"
	@echo "  make clean     - Clean the build artifacts"
	@echo "  make deps      - Install dependencies"
	@echo "  make generate  - Run go generate"
	@echo "  make docs      - Generate Swagger documentation (normal and sandbox)"
	@echo "  make migrate   - Execute database migrations"
	@echo "  make help      - Show this help message"