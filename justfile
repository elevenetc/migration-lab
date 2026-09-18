# --wait blocks until every service is healthy and fails if one exits or never
# gets healthy (e.g. the frontend crashing on a stale node_modules volume).
# The frontend runs as a Vite dev server with HMR (docker-compose.override.yml), so
# frontend edits reload live without dev-apply.

# Build and start the development environment in the background
[group('Development')]
dev-run:
    docker compose up --build --wait --wait-timeout 120

# Stop and remove the development services (use 'clear' to also remove volumes)
[group('Development')]
dev-stop *args:
    docker compose down {{ if args == "clear" { "-v" } else { "" } }}

# Rebuild and restart the backend (frontends hot-reload via HMR, no rebuild needed)
[group('Development')]
dev-apply:
    #!/usr/bin/env bash
    set -euo pipefail

    # Build backend locally
    mkdir -p build
    cd backend && go build -o ../build/server ./cmd/server

    # Rebuild Docker images and restart
    docker compose up --build -d

# Build frontend, backend, and CLI
[group('Build')]
build: build-frontend build-backend build-cli

# Build backend
[group('Build')]
build-backend:
    mkdir -p build
    cd backend && go build -o ../build/server ./cmd/server

# Build frontend
[group('Build')]
build-frontend:
    cd frontend && npm run build

# Build CLI binary with embedded frontend assets
[group('Build')]
build-cli: build-frontend
    #!/usr/bin/env bash
    set -euo pipefail

    echo "Copying frontend dist to backend/internal/report..."
    rm -rf backend/internal/report/dist
    cp -r frontend/dist backend/internal/report/dist

    echo "Building CLI binary..."
    mkdir -p build
    (cd backend && go build -o ../build/migration-lab ./cmd/cli)

    echo "Done! Binary: build/migration-lab"

# Run all tests
[group('Test')]
test: generate-contracts test-backend typecheck-frontend test-frontend

# Run all backend tests
[group('Test')]
test-backend:
    cd backend && go test ./...

# Run frontend unit tests (Vitest: contract + grid mapping tests)
[group('Test')]
test-frontend:
    cd frontend && npm test

# Test GitHub migration selection and result collection without GitHub or Docker
[group('Test')]
test-automation:
    cd backend && go test ./cmd/ci ./internal/ci/...

# Run Go CLI to analyze migrations (default behavior)
[group('CLI')]
cli-analyze sql:
    @cd backend && go run ./cmd/cli "{{ sql }}"

# Run Go CLI with --run flag to also execute migrations (requires Docker)
[group('CLI')]
cli-run sql:
    @cd backend && go run ./cmd/cli --run "{{ sql }}"

# Measure the last migration of a directory against a seeded container (requires Docker)
[group('CLI')]
cli-runtime path rows="1000000" deadline="5000":
    @cd backend && go run ./cmd/cli --runtime --rows {{ rows }} --deadline-ms {{ deadline }} "{{ path }}"

# Generate HTML report from migrations (requires build-cli first)
[group('CLI')]
cli-report path output="report.html":
    @./build/migration-lab --report="{{ output }}" "{{ path }}"

# Clean all
[group('Utilities')]
clean:
    rm -rf build
    cd backend/internal/report && rm -rf dist
    cd frontend && rm -rf node_modules dist

# Generate API contract fixtures from backend models
[group('Utilities')]
generate-contracts:
    cd backend && go test ./internal/contracts/...

# Typecheck frontend against API contracts (compile-time type-drift guard)
[group('Utilities')]
typecheck-frontend:
    cd frontend && npx tsc --noEmit

# Install frontend dependencies
[group('Utilities')]
install-frontend:
    cd frontend && npm install

# List available commands in source order
[default]
[private]
default:
    @just --list --unsorted
