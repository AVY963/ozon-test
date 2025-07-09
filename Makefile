# --- CONFIG ---

LOCAL_BIN     := $(CURDIR)/bin
MIGRATE_VERSION := v4.18.2

define install_tool
	GOBIN=$(LOCAL_BIN) go install $(1)@$(2)
endef

# --- INSTALL TOOLS ---

.PHONY: install
install:
	mkdir -p $(LOCAL_BIN)
	$(call install_tool,github.com/golang-migrate/migrate/v4/cmd/migrate,$(MIGRATE_VERSION))
	GOBIN=$(LOCAL_BIN) go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION)

# --- DATABASE ---

.PHONY: db-up
db-up:
	docker-compose up -d postgres

.PHONY: db-down
db-down:
	docker-compose down

# --- MIGRATIONS ---

.PHONY: migrate-create
migrate-create:
	$(LOCAL_BIN)/migrate create -ext sql -dir migrations -seq $(name)

.PHONY: migrate-up
migrate-up:
	$(LOCAL_BIN)/migrate -path migrations -database "$(DB_URL)" up

.PHONY: migrate-down
migrate-down:
	$(LOCAL_BIN)/migrate -path migrations -database "$(DB_URL)" down

.PHONY: migrate-force
migrate-force:
	$(LOCAL_BIN)/migrate -path migrations -database "$(DB_URL)" force $(version)

# --- RUN ---

.PHONY: run
run:
	go run cmd/main.go

.PHONY: run-postgres
run-postgres:
	DB_TYPE=postgres go run cmd/main.go

.PHONY: run-memory
run-memory:
	DB_TYPE=memory go run cmd/main.go

# --- BUILD ---

.PHONY: build
build:
	go build -o bin/ozon-posts cmd/main.go

# --- DOCKER ---

.PHONY: docker-build
docker-build:
	docker build -t ozon-posts .

.PHONY: docker-run
docker-run:
	docker run -p 8080:8080 ozon-posts

# --- DEVELOPMENT ---

.PHONY: dev-deps
dev-deps: install

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test:
	go test -v ./...

# --- HELP ---

.PHONY: help
help:
	@echo "Available commands:"
	@echo "  install      - Install development tools"
	@echo "  db-up        - Start PostgreSQL in Docker"
	@echo "  db-down      - Stop PostgreSQL in Docker"
	@echo "  migrate-up   - Run database migrations"
	@echo "  migrate-down - Rollback database migrations"
	@echo "  run          - Run application with default settings"
	@echo "  run-postgres - Run application with PostgreSQL"
	@echo "  run-memory   - Run application with in-memory storage"
	@echo "  build        - Build application binary"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  test         - Run tests" 