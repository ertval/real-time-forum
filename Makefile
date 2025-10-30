# -----------------------------------------------------
# 🗨️ Forum Project - Makefile
# -----------------------------------------------------

APP_NAME = forum
BINARY_NAME = $(APP_NAME)
CMD_DIR = ./cmd
MAIN_FILE = $(CMD_DIR)/main.go
DB_FILE = forum.db

# Default target
.PHONY: all
all: run

# -----------------------------------------------------
# 🧱 Build & Run
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
	@rm -f $(DB_FILE)
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
# 🧠 Helpers
# -----------------------------------------------------

help:
	@echo ""
	@echo "📘 Available commands:"
	@echo "  make run          - Run the server"
	@echo "  make build        - Build binary"
	@echo "  make clean        - Remove binary"
	@echo "  make reset-db     - Delete forum.db"
	@echo "  make test         - Run tests"
	@echo "  make fmt          - Format code"
	@echo "  make vet          - Static code analysis"
	@echo "  make deps         - Install dependencies"
	@echo ""

