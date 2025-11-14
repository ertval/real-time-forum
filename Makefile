# -----------------------------------------------------
# 🗨️ Forum Project - Makefile
# -----------------------------------------------------

APP_NAME       = forum
BINARY_NAME    = $(APP_NAME)
CMD_DIR        = ./cmd
MAIN_FILE      = $(CMD_DIR)/main.go
DB_FILE        = internal/db/forum.db
PORT           = 8080
DOCKER_IMAGE   = forum-app
CONTAINER_NAME = forum_app

# Default target
.PHONY: all
all: run

# -----------------------------------------------------
# 🧱 Build & Run (Local)
# -----------------------------------------------------

build:
	@echo "🔧 Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) $(MAIN_FILE)
	@echo "✅ Build complete: ./$(BINARY_NAME)"

run:
	@echo "🚀 Starting $(APP_NAME) server..."
	@go run $(MAIN_FILE)

# -----------------------------------------------------
# 🧹 Clean & Reset
# -----------------------------------------------------

clean:
	@echo "🧹 Cleaning build files..."
	@rm -f $(BINARY_NAME)
	@echo "🧽 Done."

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

# -----------------------------------------------------
# 📦 Dependencies
# -----------------------------------------------------

deps:
	@echo "📦 Installing dependencies..."
	@go mod tidy
	@echo "✅ Dependencies ready."

# -----------------------------------------------------
# 🐳 Docker Targets
# -----------------------------------------------------

# Default Dockerfile (dev version for now)
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
	@echo "🔥 Removing container and image..."
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
	@echo "🔍 Inspecting image metadata for $(DOCKER_IMAGE)..."
	@docker inspect $(DOCKER_IMAGE) | less

docker-prune:
	@echo "🧹 Cleaning up unused Docker resources..."
	@docker system prune -f

# -----------------------------------------------------
# 🧠 Helpers
# -----------------------------------------------------

help:
	@echo ""
	@echo "📘 Available commands:"
	@echo "  make run              - Run Go server locally"
	@echo "  make build            - Build Go binary"
	@echo "  make clean            - Remove binary"
	@echo "  make reset-db         - Delete SQLite DB files"
	@echo "  make test             - Run tests"
	@echo "  make fmt              - Format code"
	@echo "  make vet              - Static code analysis"
	@echo "  make deps             - Install dependencies"
	@echo ""
	@echo "🐳 Docker commands:"
	@echo "  make docker-build     - Build Docker image"
	@echo "  make docker-run       - Run container with local DB volume"
	@echo "  make docker-up        - Run docker-compose up"
	@echo "  make docker-down      - Stop docker-compose"
	@echo "  make docker-clean     - Remove Docker image/container"
	@echo "  make docker-images    - List local images"
	@echo "  make docker-ps        - Show active containers"
	@echo "  make docker-logs      - Follow logs of container"
	@echo "  make docker-shell     - Open shell in running container"
	@echo "  make docker-inspect   - Inspect Docker image details"
	@echo "  make docker-prune     - Remove unused Docker data"
	@echo ""
