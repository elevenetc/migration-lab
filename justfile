# List available commands
default:
    @just --list

# Run all tests
test: generate-contracts test-backend test-contracts

# Generate API contract fixtures from backend models
generate-contracts:
    cd backend && go test ./internal/contracts/...

# Validate frontend types against API contracts
test-contracts:
    cd frontend && npx tsc --noEmit

# Run all backend tests
test-backend:
    cd backend && go test ./...

# Run backend server
run-backend *args:
    cd backend && go run ./cmd/server {{ args }}

# Install frontend dependencies
install-frontend:
    cd frontend && npm install

# Run frontend dev server
run-frontend:
    cd frontend && npm run dev

# Build backend
build-backend:
    mkdir -p build
    cd backend && go build -o ../build/server ./cmd/server

# Build frontend
build-frontend:
    cd frontend && npm run build

# Clean all
clean:
    rm -rf build
    cd backend/internal/report && rm -rf dist
    cd frontend && rm -rf node_modules dist

# Build and run all services with Docker Compose (detached)
compose-up:
    docker compose up --build -d

# Stop and remove Docker Compose services (use 'clear' to also remove volumes)
compose-down *args:
    docker compose down {{ if args == "clear" { "-v" } else { "" } }}

# Rebuild and reload only changed services
compose-apply:
    #!/usr/bin/env bash
    set -euo pipefail

    # Build backend locally
    mkdir -p build
    cd backend && go build -o ../build/server ./cmd/server

    # Rebuild Docker images and restart
    docker compose up --build -d

# Run Go CLI to analyze migrations (default behavior)
cli-analyze sql:
    @cd backend && go run ./cmd/cli "{{ sql }}"

# Run Go CLI with --run flag to also execute migrations (requires Docker)
cli-run sql:
    @cd backend && go run ./cmd/cli --run "{{ sql }}"

# Generate HTML report from migrations (requires build-cli first)
cli-report path output="report.html":
    @./build/migration-timeline --report="{{ output }}" "{{ path }}"

# Build CLI binary with embedded frontend assets
build-cli:
    #!/usr/bin/env bash
    set -euo pipefail

    echo "Building frontend..."
    (cd frontend && npm run build)

    echo "Copying frontend dist to backend/internal/report..."
    rm -rf backend/internal/report/dist
    cp -r frontend/dist backend/internal/report/dist

    echo "Building CLI binary..."
    mkdir -p build
    (cd backend && go build -o ../build/migration-timeline ./cmd/cli)

    echo "Done! Binary: build/migration-timeline"
