# Suppress Java 25 native access warnings for Gradle
export GRADLE_OPTS := "--enable-native-access=ALL-UNNAMED"

# List available commands
default:
    @just --list

# Run all tests
test: generate-contracts test-backend test-contracts

# Generate API contract fixtures from backend models
generate-contracts:
    ./gradlew :backend:test --tests "*.ApiContractFixtureGenerator"

# Validate frontend types against API contracts
test-contracts:
    cd frontend && npx tsc --noEmit

# Run all backend tests
test-backend:
    ./gradlew :backend:test

# Run backend server
run-backend:
    ./gradlew :backend:run

# Install frontend dependencies
install-frontend:
    cd frontend && npm install

# Run frontend dev server
run-frontend:
    cd frontend && npm run dev

# Build backend
build-backend:
    ./gradlew :backend:build

# Build frontend
build-frontend:
    cd frontend && npm run build

# Clean all
clean:
    ./gradlew :backend:clean
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
    services=""
    if ! git diff --quiet HEAD -- backend/ 2>/dev/null || \
       ! git diff --quiet --staged -- backend/ 2>/dev/null || \
       [ -n "$(git ls-files --others --exclude-standard backend/)" ]; then
        services="$services backend"
    fi
    if ! git diff --quiet HEAD -- frontend/ 2>/dev/null || \
       ! git diff --quiet --staged -- frontend/ 2>/dev/null || \
       [ -n "$(git ls-files --others --exclude-standard frontend/)" ]; then
        services="$services frontend"
    fi
    if [ -z "$services" ]; then
        echo "No changes detected in backend/ or frontend/"
    else
        echo "Rebuilding:$services"
        docker compose up --build -d $services
    fi
