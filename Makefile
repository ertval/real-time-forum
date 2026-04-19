# -----------------------------------------------------
# 🗨️ Forum Project - Makefile
# -----------------------------------------------------

APP_NAME = forum

# -----------------------------------------------------
# 📦 Binaries
# -----------------------------------------------------

BACKEND_BIN  = forum-backend
FRONTEND_BIN = forum-frontend

BACKEND_PKG  = ./cmd/backend
FRONTEND_PKG = ./cmd/frontend

PORT = 8080

# -----------------------------------------------------
# 🧱 Build & Run (Local – Go)
# -----------------------------------------------------

build-backend:
	@echo "🔧 Building backend..."
	@go build -o $(BACKEND_BIN) $(BACKEND_PKG)
	@echo "✅ Backend build complete"

build-frontend:
	@echo "🔧 Building frontend..."
	@go build -o $(FRONTEND_BIN) $(FRONTEND_PKG)
	@echo "✅ Frontend build complete"

build-all: build-backend build-frontend

run-backend:
	@echo "🚀 Starting backend server..."
	@go run $(BACKEND_PKG)

run-frontend:
	@echo "🚀 Starting frontend server on http://localhost:3000 ..."
	@go run $(FRONTEND_PKG)

run-all:
	@echo "🔥 Starting backend & frontend..."
	@go run $(BACKEND_PKG) &
	@go run $(FRONTEND_PKG)

# -----------------------------------------------------
# 🛑 Stop Local Processes
# -----------------------------------------------------

stop-backend:
	@pkill -f "$(BACKEND_PKG)" 2>/dev/null || true

stop-frontend:
	@pkill -f "$(FRONTEND_PKG)" 2>/dev/null || true

stop-all: stop-backend stop-frontend

# -----------------------------------------------------
# 🧪 Code Quality
# -----------------------------------------------------

test: test-backend test-frontend policy

lint:
	@./node_modules/.bin/bun run lint

test-backend:
	@go test ./... -v

test-frontend:
	@go test ./internal/tests/... -v
	@./node_modules/.bin/bun run policy

policy:
	@./node_modules/.bin/bun run policy

format: format-backend format-frontend

format-backend:
	@go fmt ./...

format-frontend:
	@./node_modules/.bin/bun run lint:fix

fmt: format

vet:
	@go vet ./...

deps: deps-backend deps-frontend

deps-backend:
	@go mod tidy

deps-frontend:
	@command -v bun >/dev/null 2>&1 && bun install || (npm install && ./node_modules/.bin/bun install)

# -----------------------------------------------------
# 🐳 Docker (Backend Only – Production)
# -----------------------------------------------------

IMAGE      = forum
CONTAINER  = forum_app
PORT       = 8080

# -----------------------------------------------------
# Build & Run
# -----------------------------------------------------

docker-build:
	@echo "🐳 Building Docker image..."
	docker image build -t $(IMAGE) .

docker-run:
	@echo "🚀 Running Docker container..."
	docker container run -d \
		-p $(PORT):8080 \
		-v forum-data:/data \
		--name $(CONTAINER) \
		$(IMAGE)

docker-restart:
	@echo "🔄 Restarting container..."
	docker restart $(CONTAINER)

# -----------------------------------------------------
# Stop & Remove
# -----------------------------------------------------

docker-stop:
	@echo "🛑 Stopping container..."
	-@docker stop $(CONTAINER) 2>/dev/null || true
	-@docker rm $(CONTAINER) 2>/dev/null || true

# -----------------------------------------------------
# Inspection & Logs
# -----------------------------------------------------

docker-ps:
	docker ps -a

docker-images:
	docker images

docker-logs:
	docker logs $(CONTAINER)

docker-logs-follow:
	docker logs -f $(CONTAINER)

docker-inspect:
	docker inspect $(CONTAINER)

# -----------------------------------------------------
# Cleanup (Project-scoped, SAFE)
# -----------------------------------------------------

docker-clean-images:
	@echo "🧹 Removing Forum image..."
	-@docker rmi -f $(IMAGE) 2>/dev/null || true

docker-clean-all: docker-stop docker-clean-images
	@echo "✅ Forum Docker cleanup complete"


# -----------------------------------------------------
# 🌐 Browser Helper (Frontend)
# -----------------------------------------------------

open-browser:
ifeq ($(OS),Windows_NT)
	@start http://localhost:3000
else
	@xdg-open http://localhost:3000 2>/dev/null || open http://localhost:3000
endif
