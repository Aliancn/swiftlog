.PHONY: help start stop restart logs build clean cli cli-all release dev dev-up dev-down version

# Version
VERSION ?= $(shell cat VERSION)

# Default target
help:
	@echo "SwiftLog - Available Commands"
	@echo ""
	@echo "Production:"
	@echo "  make start        - Start all services (production mode)"
	@echo "  make stop         - Stop all services"
	@echo "  make restart      - Restart all services"
	@echo "  make logs         - View logs from all services"
	@echo "  make build        - Build all Docker images"
	@echo "  make clean        - Stop and remove all containers, volumes"
	@echo ""
	@echo "CLI Tools:"
	@echo "  make cli          - Build CLI tool for current platform"
	@echo "  make cli-all      - Build CLI for all platforms"
	@echo "  make release      - Create a new release (requires VERSION)"
	@echo ""
	@echo "Development:"
	@echo "  make dev          - Start services in development mode"
	@echo "  make dev-up       - Start only infrastructure (postgres, loki, redis)"
	@echo "  make dev-down     - Stop infrastructure"
	@echo "  make version      - Show current version"
	@echo ""

# Production deployment
start:
	@echo "🚀 Starting SwiftLog platform..."
	@if [ ! -f .env ]; then \
		echo "⚠️  .env file not found. Copying from .env.example..."; \
		cp .env.example .env; \
		echo "⚠️  Please edit .env file and set OPENAI_API_KEY before continuing!"; \
		exit 1; \
	fi
	docker compose up -d
	@echo "✅ SwiftLog is starting up..."
	@echo "⏳ Waiting for services to be healthy..."
	@sleep 10
	@echo ""
	@echo "📊 Service Status:"
	@docker compose ps
	@echo ""
	@echo "🌐 Access Points:"
	@echo "   Frontend:  http://localhost:3000"
	@echo "   API:       http://localhost:8080"
	@echo "   WebSocket: ws://localhost:8081"
	@echo "   gRPC:      localhost:50051"
	@echo ""
	@echo "📝 Next steps:"
	@echo "   1. Build CLI: make cli"
	@echo "   2. View logs: make logs"

stop:
	@echo "🛑 Stopping SwiftLog platform..."
	docker compose down
	@echo "✅ All services stopped"

restart:
	@echo "🔄 Restarting SwiftLog platform..."
	docker compose restart
	@echo "✅ All services restarted"

logs:
	docker compose logs -f

build:
	@echo "🔨 Building all Docker images..."
	docker compose build --no-cache
	@echo "✅ Build complete"

clean:
	@echo "🧹 Cleaning up SwiftLog platform..."
	@echo "⚠️  This will remove all containers and volumes. Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]
	docker compose down -v
	@echo "✅ Cleanup complete"

# CLI build
cli:
	@echo "🔨 Building SwiftLog CLI..."
	cd cli && go build -o swiftlog -ldflags="-s -w -X main.Version=v$(VERSION)" .
	@echo "✅ CLI built successfully: cli/swiftlog"
	@echo ""
	@echo "📝 To install globally (requires sudo):"
	@echo "   sudo cp cli/swiftlog /usr/local/bin/"
	@echo ""
	@echo "📝 To configure:"
	@echo "   ./cli/swiftlog config set --token YOUR_TOKEN --server localhost:50051"

# Build CLI for all platforms
cli-all:
	@echo "🔨 Building SwiftLog CLI for all platforms..."
	@mkdir -p dist
	@echo "  → Linux amd64..."
	@cd cli && GOOS=linux GOARCH=amd64 go build -o ../dist/swiftlog-linux-amd64 -ldflags="-s -w -X main.Version=v$(VERSION)" .
	@echo "  → Linux arm64..."
	@cd cli && GOOS=linux GOARCH=arm64 go build -o ../dist/swiftlog-linux-arm64 -ldflags="-s -w -X main.Version=v$(VERSION)" .
	@echo "  → macOS amd64..."
	@cd cli && GOOS=darwin GOARCH=amd64 go build -o ../dist/swiftlog-darwin-amd64 -ldflags="-s -w -X main.Version=v$(VERSION)" .
	@echo "  → macOS arm64..."
	@cd cli && GOOS=darwin GOARCH=arm64 go build -o ../dist/swiftlog-darwin-arm64 -ldflags="-s -w -X main.Version=v$(VERSION)" .
	@echo "  → Windows amd64..."
	@cd cli && GOOS=windows GOARCH=amd64 go build -o ../dist/swiftlog-windows-amd64.exe -ldflags="-s -w -X main.Version=v$(VERSION)" .
	@echo ""
	@echo "✅ All builds complete! Binaries in dist/"
	@ls -lh dist/

# Create release
release:
	@echo "🚀 Creating release v$(VERSION)..."
	@echo ""
	@echo "Checklist:"
	@echo "  ✓ VERSION file: $(VERSION)"
	@echo "  ✓ CHANGELOG.md updated?"
	@read -p "Continue? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	@echo ""
	@echo "Building CLI binaries..."
	@make cli-all
	@echo ""
	@echo "Creating git tag..."
	git tag -a v$(VERSION) -m "Release v$(VERSION)"
	@echo ""
	@echo "✅ Release v$(VERSION) ready!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Review: git show v$(VERSION)"
	@echo "  2. Push tag: git push origin v$(VERSION)"
	@echo "  3. GitHub Actions will automatically:"
	@echo "     - Create GitHub Release"
	@echo "     - Build and push Docker images"
	@echo "     - Deploy to configured servers"

# Show version
version:
	@echo "SwiftLog v$(VERSION)"

# Development mode
dev-up:
	@echo "🛠️  Starting infrastructure for development..."
	docker compose up -d postgres loki redis
	@echo "✅ Infrastructure started"
	@echo ""
	@echo "📝 Connection strings:"
	@echo "   PostgreSQL: postgres://swiftlog:changeme@localhost:5432/swiftlog?sslmode=disable"
	@echo "   Loki:       http://localhost:3100"
	@echo "   Redis:      redis://localhost:6379"
	@echo ""
	@echo "💡 Run backend services locally:"
	@echo "   cd backend/cmd/ingestor && go run main.go"
	@echo "   cd backend/cmd/api && go run main.go"
	@echo "   cd backend/cmd/websocket && go run main.go"
	@echo "   cd backend/cmd/ai-worker && go run main.go"

dev-down:
	@echo "🛑 Stopping infrastructure..."
	docker compose down
	@echo "✅ Infrastructure stopped"

dev: dev-up
	@echo ""
	@echo "🛠️  Development environment ready!"
	@echo "   Start your backend services manually as needed."
