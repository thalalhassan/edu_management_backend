# --- Variables ---
BINARY_NAME=app
MAIN_PATH=./cmd/api/main.go
MIGRATION_PATH=./cmd/migrate/main.go
SEEDER_PATH=./cmd/seed/main.go

# Load environment variables for local DB migrations
ifneq ("$(wildcard .env)","")
    include .env
    export
endif


# ## help: Show available commands
# help:
# 	@echo "Usage: make [target]"
# 	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

# ## build: Compile the binary for the current platform
# build:
# 	@echo "Building..."
# 	go build -ldflags="-s -w" -o bin/${BINARY_NAME} ${MAIN_PATH}

# --- Commands ---

.PHONY: help
help:
	@echo "\nAvailable commands:\n"
	@echo "  make build           - Build the application"
	@echo "  make run             - Run the application"
	@echo "  make dev             - Run in development mode with hot reload"
	@echo "  make test            - Run tests"
	@echo "  make migrate-up      - Run pending migrations"
	@echo "  make migrate-down    - Rollback last migration"
# 	@echo "  make migrate-status  - Show migration status"
	@echo "  make seed            - Seed database"
# 	@echo "  make seed-clear      - Clear and seed database"
	@echo "  make seed-reset      - Reset database and seed"
	@echo "  make swagger         - Generate swagger docs"
	@echo "  make sqlc            - Generate sqlc go file"
	@echo "  make deps            - Install golang dependencies"
	@echo "  make clean           - Clean build artifacts"
	@echo "  make convert-postman - Convert Postman collection to OpenAPI spec"
	@echo "\nUse 'make [target]' to execute a command.\n"

## Tidy and download dependencies
.PHONY: deps
deps:
	go mod tidy
	go mod download


.PHONY: air
air:
	@air

.PHONY: build
build:
	@go build -o bin/api ./cmd/api

.PHONY: run
run: build
	@./bin/api

.PHONY: dev
dev:
	@air

.PHONY: test
test:
	@go test -v ./...

.PHONY: migrate-up
migrate-up:
	@go run ${MIGRATION_PATH} up

.PHONY: migrate-down
migrate-down:
	@read -p "Are you sure you want to run migration DOWN? [y/N] " confirm && \
	if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
		go run ${MIGRATION_PATH} down; \
	else \
		echo "Aborted."; \
	fi

.PHONY: migrate-status
migrate-status:
	@go run ${MIGRATION_PATH} status

.PHONY: migrate-rollback
migrate-rollback: migrate-down

.PHONY: seed
seed:
	@go run ${SEEDER_PATH}

.PHONY: seed-clear
seed-clear:
	@go run ${SEEDER_PATH} clear


.PHONY: seed-reset
seed-reset:
	@go run ${SEEDER_PATH} reset

.PHONY: db-reset
db-reset: migrate-rollback migrate-up seed
	@echo "✅ Database reset and seeded"

.PHONY: clean
clean:
	@rm -rf bin/
	@go clean

.PHONY: swagger
swagger:
	@swag init -g ${MAIN_PATH} -o ./docs --parseDependency --parseInternal

.PHONY: sql
sql:
	@sqlc generate

.PHONY: convert-postman
convert-postman:
	@npx postman-to-openapi@latest ./docs/postman_collection.json -f ./docs/openapi.yaml