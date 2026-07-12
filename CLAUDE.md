# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Migration Timeline is a tool for analyzing and visualizing database migrations. It parses Flyway PostgreSQL migrations,
builds AST representations, and renders interactive timelines.

## Architecture

Two-module monorepo:

- **backend/**: Go service for migration parsing, AST generation, and static analysis
- **frontend/**: React + TypeScript + Zustand web client for visualization

## Tech Stack

### Backend

- Go with Echo framework
- pg_query_go for SQL parsing
- Cobra for CLI
- Testcontainers for integration tests

### Frontend

- React with TypeScript
- Zustand for state management
- HTML5 Canvas for timeline rendering

#### Structure

- `src/api/` - API types and fetch functions (mirrors backend models)
- `src/store/` - Zustand store with async actions
- `src/components/` - React components
- `src/canvas/` - Canvas rendering helpers (layout, drawing, colors)
- `src/contracts/` - API contract validation

#### Data Flow

1. App mounts -> calls `loadMigrations()`
2. Store fetches `/api/migrations`
3. Timeline reads store, groups by table, renders to canvas

#### Visualization

- Tables as rows, operations as nodes positioned by timestamp
- X-axis: migrations with same timestamp align vertically in same column
- Ribbons connect operations on same table chronologically
- CREATE_TABLE (blue) vs ALTER_TABLE (darker blue)

## CLI

The CLI (`backend/cmd/cli`) outputs static analysis JSON by default:

- Default - Parse migrations and output static analysis as JSON
- `--run` - Also run migrations against a PostgreSQL container (requires Docker)
- `--report` - Generate self-contained HTML report instead of JSON

### Building CLI with Report Support

```bash
just build-cli
```

This builds the frontend, embeds assets into `backend/internal/report/dist/`, and produces `build/migration-timeline`.

### Analyzing Migrations (Default)

```bash
# From directory
./build/migration-timeline /path/to/migrations

# From inline SQL
./build/migration-timeline "CREATE TABLE users (id INT);"

# With migration runner (requires Docker)
./build/migration-timeline --run /path/to/migrations
```

### Generating Reports

```bash
# From directory
./build/migration-timeline --report /path/to/migrations

# With custom output path
./build/migration-timeline --report=output.html /path/to/migrations

# From inline SQL
./build/migration-timeline --report "CREATE TABLE users (id INT);"
```

The generated HTML is fully self-contained (inlined CSS, JS, and data) and opens directly in a browser.

## Verification

After making changes, run tests:

```
just test
```

`just test` runs, in order: `generate-contracts` (backend fixtures), `test-backend` (Go),
`typecheck-frontend` (`tsc --noEmit`, compile-time type-drift guard), and `test-frontend`
(Vitest unit tests). Frontend-only: `just test-frontend` (or `cd frontend && npm test`).

## Frontend tests

Vitest (`*.test.ts` under `frontend/src`, run via `just test-frontend`) covers:

- `src/contracts/validate-api-contracts.test.ts` - validates generated fixtures against TS types at runtime
- `src/canvas/*.test.ts` - pure layout/color/title logic (`computeLayout`, `getOperationColor`, etc.)

Canvas *drawing* (pixel output) has no automated coverage - verify it in the browser (see Debugging).

## API Contract Testing

Backend and frontend types are kept in sync two ways: `tsc` fails to compile if types drift,
and the Vitest contract test fails if the generated fixtures violate the TS types at runtime.

1. `generate_test.go` generates JSON fixtures from backend models to `api-contracts/fixtures/`
2. `validate-api-contracts.test.ts` imports fixtures, asserts type-assignability (compile-time)
   and runs the validators (runtime)
3. Exhaustive switch ensures all Operation variants are handled

When adding new Operation types:

1. Add to `internal/models/operation.go`
2. Add to `migrationApi.ts` and update `Operation` union
3. Add case to `validate-api-contracts.test.ts` switch
4. Update `internal/contracts/generate_test.go` to include example

`just test` fails if types drift.

## Constraints

- PostgreSQL and Flyway migrations only
- Frontend simplicity prioritized over polish for MVP
- See [docs/supported-operations.md](docs/supported-operations.md) for supported SQL statements

## Operations implementation process

1. Identify SQL statement or operation
2. Implement backend parser in `internal/parser/`
3. Add backend test(s) in `internal/parser/parse_test.go`
4. Implement frontend rendering
5. Add example dataset in `internal/datasets/` (registered in `Datasets()`, served via `internal/database`)
6. Run tests
7. Update [docs/supported-operations.md](docs/supported-operations.md)

## Static analysis implementation process

1. Create detection function in `internal/analysis/` (e.g., `detect_something.go`)
2. Add to `analyzeOperation()` in `analyse.go`
3. Add warning type to `internal/models/analysis.go`
4. Add test in `internal/analysis/analysis_test.go`
5. Update [docs/supported-static-analysis.md](docs/supported-static-analysis.md)

## Debugging

- Use `playwright mcp` and `localhost:3000` to verify frontend implementation
- `localhost:3000` makes single request which returns content of `/api/migrations?migrationId=<id>`;
  the id comes from the page URL (`?migrationId=`). Dataset ids are the keys of `datasets.Datasets()`
  (e.g. `ecommerce`, `simple-partition`). A missing/unknown id returns 404.
- `just compose-up` runs the frontend as a Vite dev server with HMR (see `docker-compose.override.yml`), so
  frontend source edits reload live — no `compose-apply` needed for frontend changes
- `just compose-apply` is only needed to pick up backend changes (rebuilds and restarts the backend)
- After editing `internal/datasets/`, run `just compose-apply` to see the updated data at `localhost:3000`
- Pass `/.playwright-mcp` to `playwright`, so it stores logs and screenshots there instead of root

## Backward compatibility

The project is in MVP stage, so breaking changes are expected, backward compatibility is not required.

## Code style

- Prefer functional style over OOP
- Prefer having single file - single function
- Prioritize making functions as pure as possible. For example, instead of passing a mutable list to a function, prefer
  returning a new immutable list with result.
