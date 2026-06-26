# RedIntel Sentinel - Makefile
# Common development, build and quality tasks.

APP_NAME      := redintel
PKG           := github.com/Skypieee6/redintel-sentinel
CMD_PATH      := ./cmd/server
BIN_DIR       := bin
BIN           := $(BIN_DIR)/$(APP_NAME)

VERSION       ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT        ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
BUILD_TIME    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -w \
	-X $(PKG)/internal/version.Version=$(VERSION) \
	-X $(PKG)/internal/version.Commit=$(COMMIT) \
	-X $(PKG)/internal/version.BuildTime=$(BUILD_TIME)

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

.PHONY: tidy
tidy: ## Sync go.mod / go.sum
	go mod tidy

.PHONY: build
build: ## Build the server binary
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN) $(CMD_PATH)

.PHONY: run
run: ## Run the server locally (serve)
	go run -ldflags "$(LDFLAGS)" $(CMD_PATH) serve

.PHONY: version
version: ## Print build version info
	go run -ldflags "$(LDFLAGS)" $(CMD_PATH) version --json

.PHONY: migrate-up
migrate-up: ## Apply database migrations
	go run $(CMD_PATH) migrate up

.PHONY: migrate-down
migrate-down: ## Roll back the last migration
	go run $(CMD_PATH) migrate down

.PHONY: test
test: ## Run unit tests with race detector
	go test -race -count=1 ./...

.PHONY: cover
cover: ## Run tests and write coverage profile
	go test -race -covermode=atomic -coverprofile=coverage.out ./...

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: fmt
fmt: ## Format code
	gofmt -s -w .

.PHONY: lint
lint: ## Run golangci-lint (must be installed)
	golangci-lint run ./...

.PHONY: docker-build
docker-build: ## Build the Docker image
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		-t $(APP_NAME):$(VERSION) .

.PHONY: up
up: ## Start the full stack via docker-compose
	docker compose up --build

.PHONY: down
down: ## Stop the docker-compose stack
	docker compose down

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(BIN_DIR) coverage.out
