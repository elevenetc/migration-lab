# AGENTS.md

This file provides guidance to coding agents when working with code in this repository.
`CLAUDE.md` is a symlink to this file, so Claude Code reads the same content.

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
- `active-grid` for canvas timeline rendering (grid layout, camera, hit testing, focus overlay)

#### Structure

- `src/api/` - API types and fetch functions (mirrors backend models)
- `src/store/` - Zustand store with async actions
- `src/components/` - React components
- `src/grid/` - active-grid configuration (drawers, focus layout, keyboard, data mapping)
- `src/contracts/` - API contract validation

#### Data Model

`migration > statement > operation`: a migration holds SQL statements (`Statement`: index, kind,
sql), and each statement yields one or more operations (e.g. `ALTER TABLE t ADD a, DROP b` is one
statement with two operations; `DROP TABLE a, b` is one statement with an operation per table).

#### Data Flow

1. App mounts -> calls `loadMigrations()`
2. Store fetches `/api/migrations`
3. `Timeline` reads store, `buildMigrationCells` maps it to grid rows/columns/cells,
   `buildMigrationGrid` turns that into an `ActiveGrid` attached to the canvas

#### Ordering

Directory-loaded migrations (`migrationsPath`) are sorted by Flyway version, compared
segment-wise **numerically** — not lexicographically or by file mtime. `V<date>_<seq>`
names become `<date>.<seq>` (e.g. `V20260410_2` -> `20260410.2`), so `20260410.2` < `20260410.10`.
After sorting, each migration gets a sequential `timestamp` (`1, 2, ...`), matching how the
datasets and CLI assign timestamps. `timestamp` is therefore an ordinal rank, not a real date.

#### Visualization

- Tables as rows, migrations as columns: row 0 holds the migration versions, column 0 the table
  names, so both axes are offset by one
- A cell exists at (table row, migration column) when the migration touches that table; a migration
  touching several tables fills several rows of one column
- A cell shows a kind-colored accent bar and one stacked `action / target` line per operation, in
  statement order
- Kinds: `partition` (`CREATE_TABLE` with `partitionOf`), `create`, `drop`, otherwise `alter`;
  colors come from `MIGRATION_KIND_COLOR`, not from per-table hashes
- `cellPath` connectors route from the parent table's latest earlier cell into a partition cell, and
  from the old table row into a `RENAME_TABLE` cell
- Focusing a cell (footer button, or `Enter` on the selection) opens an HTML panel with the
  migration version, table, its analysis warnings and the parsed `Statement.sql`

#### Rendering

- `buildMigrationCells` is a pure mapping from the API response to rows, columns, cells and
  connectors; `buildMigrationGrid` only turns that model into `setCell` / `setFooter` /
  `setBackground` calls — neither knows about the viewport
- `ActiveGrid` owns the rest: it auto-sizes tracks from each drawer's `layoutWidth`/`layoutHeight`,
  runs its own rAF loop and `ResizeObserver`, and keeps the canvas backing store at
  viewport × `devicePixelRatio`
- Camera: the grid fits on the first frame (small grids are upscaled to fill), wheel zooms, drag
  pans, arrows walk cells, `Escape` backs out of focus, then selection, then to the whole grid
- `Timeline.tsx` only builds the grid and calls `attach` / `detach`; it never repaints by hand

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
- `src/grid/migrationCells.test.ts` - the pure data mapping (rows, columns, grouping, kinds,
  connectors, warnings, SQL collection)
- `src/grid/operationText.test.ts` - `getOperationTitle` / `getOperationTarget` over every
  `Operation` variant

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
2. Add to `analyzeStatement()` in `analyse.go` (statement-scoped warnings use `OpIndex = -1`)
3. Add warning type to `internal/models/analysis.go`
4. Add test in `internal/analysis/analysis_test.go`
5. Update [docs/supported-static-analysis.md](docs/supported-static-analysis.md)

## Debugging

- Use `playwright mcp` and `localhost:3000` to verify frontend implementation
- `localhost:3000` makes single request which returns content of `/api/migrations?migrationId=<id>`;
  the id comes from the page URL (`?migrationId=`). Dataset ids are the keys of `datasets.Datasets()`
  (e.g. `ecommerce`, `simple-partition`). A missing/unknown id returns 404.
- Alternatively, `?migrationsPath=<abs-dir>` loads, parses and analyzes the `*.sql` files from a
  local directory (takes precedence over `migrationId`; a bad path returns 400). With `compose-up`
  the backend runs in a container, so `docker-compose.override.yml` bind-mounts a narrow root
  (`MIGRATIONS_ROOT`, default `~/dev`) read-only at the same path — pass an absolute path under that
  root so it resolves in-container. Set `MIGRATIONS_ROOT` (e.g. in `.env`) if migrations live elsewhere.
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
