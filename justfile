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

    BUILD_STATE_DIR=".compose-build-state"
    mkdir -p "$BUILD_STATE_DIR"

    needs_rebuild() {
        local service=$1
        local dir=$2
        local state_file="$BUILD_STATE_DIR/$service.sha"

        # Get current state: latest commit + uncommitted changes hash
        local current_sha
        current_sha=$(git log -1 --format=%H -- "$dir/" 2>/dev/null || echo "none")
        current_sha="$current_sha-$(git diff HEAD -- "$dir/" 2>/dev/null | sha256sum | cut -c1-16)"
        current_sha="$current_sha-$(git ls-files --others --exclude-standard "$dir/" 2>/dev/null | sha256sum | cut -c1-16)"

        # Compare with last build state
        if [ -f "$state_file" ]; then
            local last_sha
            last_sha=$(cat "$state_file")
            if [ "$current_sha" = "$last_sha" ]; then
                return 1  # No rebuild needed
            fi
        fi

        # Save current state for next comparison
        echo "$current_sha" > "$state_file"
        return 0  # Rebuild needed
    }

    services=""
    if needs_rebuild backend backend; then
        services="$services backend"
    fi
    if needs_rebuild frontend frontend; then
        services="$services frontend"
    fi

    if [ -z "$services" ]; then
        echo "No changes detected in backend/ or frontend/"
    else
        echo "Rebuilding:$services"
        docker compose up --build -d $services
    fi
