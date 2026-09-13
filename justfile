# List available commands
default:
    @just --list

# Run all tests
test: generate-contracts test-backend typecheck-frontend test-frontend

# Generate API contract fixtures from backend models
generate-contracts:
    cd backend && go test ./internal/contracts/...

# Typecheck frontend against API contracts (compile-time type-drift guard)
typecheck-frontend:
    cd frontend && npx tsc --noEmit

# Run frontend unit tests (Vitest: contract + grid mapping tests)
test-frontend:
    cd frontend && npm test

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

# Build and run all services with Docker Compose (detached).
# --wait blocks until every service is healthy and fails if one exits or never
# gets healthy (e.g. the frontend crashing on a stale node_modules volume).
# The frontend runs as a Vite dev server with HMR (docker-compose.override.yml), so
# frontend edits reload live without compose-apply.
compose-up:
    docker compose up --build --wait --wait-timeout 120

# Stop and remove Docker Compose services (use 'clear' to also remove volumes)
compose-down *args:
    docker compose down {{ if args == "clear" { "-v" } else { "" } }}

# Rebuild and restart the backend (frontends hot-reload via HMR, no rebuild needed)
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

# Measure the last migration of a directory against a seeded container (requires Docker)
cli-runtime path rows="1000000" deadline="5000":
    @cd backend && go run ./cmd/cli --runtime --rows {{ rows }} --deadline-ms {{ deadline }} "{{ path }}"

# Generate HTML report from migrations (requires build-cli first)
cli-report path output="report.html":
    @./build/migration-lab --report="{{ output }}" "{{ path }}"

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
    (cd backend && go build -o ../build/migration-lab ./cmd/cli)

    echo "Done! Binary: build/migration-lab"
