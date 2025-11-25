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

# Legacy compatibility (kept because your Makefile uses them)
BINARY_NAME     = $(BACKEND_BIN)
MAIN_FILE       = $(BACKEND_MAIN)

DB_FILE         = internal/db/forum.db
PORT            = 8080

DOCKER_IMAGE    = forum-app
CONTAINER_NAME  = forum_app

# Default = run backend
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
	@echo "🎉 All binaries built successfully!"

# Keep original build target just mapped to backend
build: build-backend

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

# Keep original run mapped to backend
run: run-backend

# -----------------------------------------------------
# 🛑 Stop Processes
# -----------------------------------------------------

stop-backend:
	@echo "🛑 Stopping backend server..."
	@pkill -f "$(BACKEND_MAIN)" 2>/dev/null || echo "Backend not running."

stop-frontend:
	@echo "🛑 Stopping frontend server..."
	@pkill -f "$(FRONTEND_MAIN)" 2>/dev/null || echo "Frontend not running."

stop-all: stop-backend stop-frontend
	@echo "🧹 All servers stopped."

# -----------------------------------------------------
# 🧹 Clean & Reset
# -----------------------------------------------------

clean:
	@echo "🧹 Cleaning build files..."
	@rm -f $(BACKEND_BIN) $(FRONTEND_BIN)
	@echo "🧽 Done."

clean-all: clean docker-clean
	@echo "🔥 Everything cleaned."

reset-db:
	@echo "🗑️ Removing SQLite database..."
	@rm -f $(DB_FILE) $(DB_FILE)-wal $(DB_FILE)-shm
	@echo "✅ Database reset complete."

# -----------------------------------------------------
# 🧪 Test & Format
# -----------------------------------------------------

test:
	@echo "🧪 Running tests..."
	@go test ./... -v

fmt:
	@echo "✨ Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatted."

vet:
	@echo "🔍 Running go vet..."
	@go vet ./...

deps:
	@echo "📦 Installing dependencies..."
	@go mod tidy
	@echo "✅ Dependencies ready."

# -----------------------------------------------------
# 🐳 Docker Targets
# -----------------------------------------------------

DOCKERFILE ?= Dockerfile.dev

docker-build:
	@echo "🐳 Building Docker image using $(DOCKERFILE)..."
	@docker build -f $(DOCKERFILE) -t $(DOCKER_IMAGE) .

docker-run:
	@echo "🐳 Running Docker container..."
	@docker run --rm -p $(PORT):$(PORT) \
		--name $(CONTAINER_NAME) \
		-v $(PWD)/internal/db:/app/internal/db \
		$(DOCKER_IMAGE)

docker-up:
	@echo "📦 Starting via docker-compose..."
	@docker compose up --build

docker-down:
	@echo "🧹 Stopping docker-compose..."
	@docker compose down

docker-clean:
	@echo "🔥 Removing Docker image & container..."
	-@docker rm -f $(CONTAINER_NAME) 2>/dev/null || true
	-@docker rmi -f $(DOCKER_IMAGE) 2>/dev/null || true

docker-all: docker-build docker-run

# -----------------------------------------------------
# 🧰 Extra Docker Utilities
# -----------------------------------------------------

docker-images:
	@echo "🖼️  Listing local Docker images..."
	@docker images | grep $(DOCKER_IMAGE) || echo "No image named $(DOCKER_IMAGE) found."

docker-ps:
	@echo "🚢 Active containers..."
	@docker ps -a | grep $(CONTAINER_NAME) || echo "No container named $(CONTAINER_NAME) found."

docker-logs:
	@echo "🪵 Showing logs from container $(CONTAINER_NAME)..."
	@docker logs -f $(CONTAINER_NAME) || echo "Container $(CONTAINER_NAME) not running."

docker-shell:
	@echo "🐚 Opening interactive shell inside $(CONTAINER_NAME)..."
	@docker exec -it $(CONTAINER_NAME) /bin/sh || echo "Container $(CONTAINER_NAME) not running."

docker-inspect:
	@echo "🔍 Inspecting Docker image metadata..."
	@docker inspect $(DOCKER_IMAGE) | less

docker-prune:
	@echo "🧹 Cleaning unused Docker resources..."
	@docker system prune -f

# -----------------------------------------------------
# 🌐 Open Browser Helpers
# -----------------------------------------------------

open-browser:
ifeq ($(OS),Windows_NT)
	@start http://localhost:3000
else
	@xdg-open http://localhost:3000 2>/dev/null || open http://localhost:3000
endif

# -----------------------------------------------------
# 🧠 Helpers
# -----------------------------------------------------

help:
	@echo ""
	@echo "📘 Available commands:"
	@echo "  make build-backend      Build backend only"
	@echo "  make build-frontend     Build frontend only"
	@echo "  make build-all          Build backend + frontend"
	@echo "  make run-backend        Run backend server"
	@echo "  make run-frontend       Run frontend server"
	@echo "  make run-all            Run both servers in parallel"
	@echo "  make stop-backend       Stop backend server"
	@echo "  make stop-frontend      Stop frontend server"
	@echo "  make stop-all           Stop all servers"
	@echo ""
	@echo "🐳 Docker commands:"
	@echo "  make docker-build       Build Docker image"
	@echo "  make docker-run         Run container"
	@echo "  make docker-up          docker-compose up"
	@echo "  make docker-down        docker-compose down"
	@echo "  make docker-clean       Remove Docker image/container"
