# -----------------------------------------------------
# 🗨️ Forum Project - Makefile
# -----------------------------------------------------

APP_NAME        = forum

# Binaries
BACKEND_BIN     = forum-backend
FRONTEND_BIN    = forum-frontend

# Entry points
BACKEND_MAIN    = ./cmd/backend/main.go
FRONTEND_MAIN   = ./cmd/frontend/main.go

# Legacy compatibility
BINARY_NAME     = $(BACKEND_BIN)
MAIN_FILE       = $(BACKEND_MAIN)

DB_FILE         = internal/db/forum.db
PORT            = 8080

DOCKER_IMAGE    = forum-app
CONTAINER_NAME  = forum_app

# Default target
.PHONY: all
all: run-backend

# -----------------------------------------------------
# 🧱 Build & Run (Local)
# -----------------------------------------------------

build-backend:
	@echo "🔧 Building backend..."
	@go build -o $(BACKEND_BIN) $(BACKEND_MAIN)
	@echo "✅ Backend build complete: ./$(BACKEND_BIN)"

build-frontend:
	@echo "🔧 Building frontend..."
	@go build -o $(FRONTEND_BIN) $(FRONTEND_MAIN)
	@echo "✅ Frontend build complete: ./$(FRONTEND_BIN)"

build-all: build-backend build-frontend

run-backend:
	@echo "🚀 Starting backend server..."
	@go run $(BACKEND_MAIN)

run-frontend:
	@echo "🚀 Starting frontend server on http://localhost:3000 ..."
	@go run $(FRONTEND_MAIN) &
	@sleep 1
	@$(MAKE) open-browser

run-all:
	@echo "🔥 Starting backend & frontend servers..."
	@go run $(BACKEND_MAIN) &
	@go run $(FRONTEND_MAIN) &
	@sleep 1
	@$(MAKE) open-browser

run: run-backend

# -----------------------------------------------------
# 🛑 Stop Processes
# -----------------------------------------------------

stop-backend:
	@echo "🛑 Stopping backend..."
	@pkill -f "$(BACKEND_MAIN)" 2>/dev/null || echo "Backend not running."

stop-frontend:
	@echo "🛑 Stopping frontend..."
	@pkill -f "$(FRONTEND_MAIN)" 2>/dev/null || echo "Frontend not running."

stop-all: stop-backend stop-frontend

# -----------------------------------------------------
# 🧹 Clean & Reset
# -----------------------------------------------------

clean:
	@echo "🧹 Cleaning binaries..."
	@rm -f $(BACKEND_BIN) $(FRONTEND_BIN)

clean-all: clean docker-clean

reset-db:
	@echo "🗑️ Resetting database..."
	@rm -f $(DB_FILE) $(DB_FILE)-wal $(DB_FILE)-shm

# -----------------------------------------------------
# 🧪 Code Quality
# -----------------------------------------------------

test:
	@go test ./... -v

fmt:
	@go fmt ./...

vet:
	@go vet ./...

deps:
	@go mod tidy

# -----------------------------------------------------
# 🐳 Docker (Backend Only)
# -----------------------------------------------------

DOCKERFILE ?= Dockerfile.dev

docker-build:
	@docker build -f $(DOCKERFILE) -t $(DOCKER_IMAGE) .

docker-run:
	@docker run --rm -p $(PORT):8080 \
		--name $(CONTAINER_NAME) \
		-v $(PWD)/internal/db:/app/internal/db \
		$(DOCKER_IMAGE)

docker-up:
	@docker compose up --build

docker-down:
	@docker compose down

docker-clean:
	-@docker rm -f $(CONTAINER_NAME) 2>/dev/null || true
	-@docker rmi -f $(DOCKER_IMAGE) 2>/dev/null || true

# -----------------------------------------------------
# 🌐 Browser Helper
# -----------------------------------------------------

open-browser:
ifeq ($(OS),Windows_NT)
	@start http://localhost:3000
else
	@xdg-open http://localhost:3000 2>/dev/null || open http://localhost:3000
endif
