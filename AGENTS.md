# AGENTS.md

This file provides guidance to coding agents when working with code in this repository.
`CLAUDE.md` is a symlink to this file, so Claude Code reads the same content.

## Project Overview

Migration Lab analyzes the performance and impact of PostgreSQL migrations. It parses Flyway migrations,
predicts operation costs with static analysis, and measures execution, locking, and retry behavior against seeded
PostgreSQL containers. Interactive timelines visualize migration history and analysis results.

Use `Migration Lab` for the product name and `migration-lab` for the CLI and package names. `Timeline` remains the
name of the migration-history visualization and the corresponding internal data model.

## Architecture

Two-module monorepo:

- **backend/**: Go service for migration parsing, AST generation, static analysis, and runtime analysis
- **frontend/**: React + TypeScript + Zustand web client for visualization

Analysis comes in two groups, one package each under `internal/analysis`, and the docs are named
after them:

- **static analysis** (`internal/analysis/static`) reads the AST and raises warnings. No database
- **runtime analysis** (`internal/analysis/runtime`) executes one migration against a seeded
  PostgreSQL container and raises findings from what it measured. See
  [docs/supported-runtime-analysis.md](docs/supported-runtime-analysis.md)

Where a rule lives follows from that split: derivable from the migration text -> `static`,
only observable -> `runtime`. Two things sit outside both on purpose:

- `models.OperationID` (`operation_id.go`) anchors a static `Warning` and a `RuntimeFinding` alike,
  so it belongs to neither group
- performance-class prediction (`models/classify_performance.go`) is static analysis by subject but
  lives in `models`, because the class is injected while marshalling an operation and `models`
  cannot import the analysis packages

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

Every operation and statement carries a `performanceClass` (`METADATA_ONLY` / `DATA_SCANNING` /
`TABLE_REWRITE`), injected at marshal time from the AST; a statement's class is the worst among its
operations. See [docs/supported-performance-classes.md](docs/supported-performance-classes.md).

A `RuntimeFinding` carries the same fields as a static `Warning` (`type`, `operationId`, `tableName`,
`message`), so both analysis groups stay renderable through one path. Runtime measurements
(`StatementMeasurement`, `ProbeResult`, `SeededTable`) ride alongside in `RuntimeAnalysisResult`.

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
- A hovered cell shows two stacked footer buttons: `focus-in`, then `runtime` below it, which measures
  the cell's migration and shows the result in `RuntimePopup`. Both are drawn by `cellButton.ts`,
  which owns the geometry and hands each drawer a slot

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
- `--run` - Also run migrations against a PostgreSQL container
- `--runtime` - Also measure the last migration against a seeded container;
  select a target with `--migration <filename>`, tune with `--rows` / `--deadline-ms`;
  exits non-zero unless the verdict is `COMPLETED`
- `--report` - Generate self-contained HTML report instead of JSON

### Building CLI with Report Support

```bash
just build-cli
```

This builds the frontend, embeds assets into `backend/internal/report/dist/`, and produces `build/migration-lab`.

### Analyzing Migrations (Default)

```bash
# From directory
./build/migration-lab /path/to/migrations

# From inline SQL
./build/migration-lab "CREATE TABLE users (id INT);"

# With migration runner
./build/migration-lab --run /path/to/migrations
```

### Runtime Analysis

```bash
# Measure the last migration against tables seeded to 1M rows
./build/migration-lab --runtime /path/to/migrations

# Smaller and quicker, with a 30 s pod grace period
./build/migration-lab --runtime --rows 50000 --deadline-ms 30000 /path/to/migrations
```

See [docs/supported-runtime-analysis.md](docs/supported-runtime-analysis.md) for what the verdicts mean.

### Generating Reports

```bash
# From directory
./build/migration-lab --report /path/to/migrations

# With custom output path
./build/migration-lab --report=output.html /path/to/migrations

# From inline SQL
./build/migration-lab --report "CREATE TABLE users (id INT);"
```

The generated HTML is fully self-contained (inlined CSS, JS, and data) and opens directly in a browser.

## Verification

After making changes, run tests:

```
just test
```

`just test` runs, in order: `generate-contracts` (backend fixtures), `test-backend` (Go),
`typecheck-frontend` (`tsc --noEmit`, compile-time type-drift guard), and `test-frontend`
(Vitest unit tests). The backend tests include the reusable GitHub workflow's Go/Git tests;
run those separately with `just test-automation`.
Frontend-only: `just test-frontend` (or `cd frontend && npm test`).

The reusable workflow is `.github/workflows/runtime-analysis.yml`; its setup and debugging guide is
[docs/github-actions.md](docs/github-actions.md). It runs on GitHub.com Ubuntu runners and builds
the CLI from the workflow's own commit. Its Go orchestration lives in `backend/cmd/ci/`.
It keeps analysis read-only and uses a separate job to create or update one bot comment per
same-repository PR. Callers must grant `pull-requests: write`; fork and Dependabot PRs retain
analysis and artifacts without comments. Comment formatting and GitHub API tests are included
in `just test-automation`.

## Frontend tests

Vitest (`*.test.ts` under `frontend/src`, run via `just test-frontend`) covers:

- `src/contracts/validate-api-contracts.test.ts` - validates generated fixtures against TS types at runtime
- `src/grid/migrationCells.test.ts` - the pure data mapping (rows, columns, grouping, kinds,
  connectors, warnings, SQL collection)
- `src/grid/operationText.test.ts` - `getOperationTitle` / `getOperationTarget` over every
  `Operation` variant
- `src/grid/cellButton.test.ts` - footer button geometry (slots stack, stay inside the decor they size)

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

1. Create detection function in `internal/analysis/static/` (e.g., `detect_something.go`)
2. Add to `analyzeStatement()` in `analyse.go` (statement-scoped warnings use `OpIndex = -1`)
3. Add warning type to `internal/models/warning.go`
4. Add test in `internal/analysis/static/analysis_test.go`
5. Update [docs/supported-static-analysis.md](docs/supported-static-analysis.md)

## Runtime analysis implementation process

1. Decide what has to be *observed* rather than derived — a duration, a lock, a blocked reader, a
   leftover. Anything derivable from the AST belongs in static analysis instead
2. Keep the new logic pure where it can be — parsing, planning and classification should take plain
   values and be unit-testable without Docker
3. Put anything that talks to the database in a file of its own, and wire it into `Analyse`
   (`analyse.go`), which is the package's only entry point
4. Add a finding type to `internal/models/runtime.go` and raise it in `findings.go`
5. Add unit tests to `internal/analysis/runtime/runtime_test.go`; add a container-backed test to
   `internal/analysis/runtime/analyse_test.go` only for behaviour a real PostgreSQL has to confirm
6. Keep every collection in `RuntimeAnalysisResult` an empty slice rather than nil — the frontend types are
   arrays, and Go marshals nil as `null`
7. Extend the fixture in `internal/contracts/generate_test.go` and the validator in
   `validate-api-contracts.test.ts`
8. Update [docs/supported-runtime-analysis.md](docs/supported-runtime-analysis.md), moving what you
   built out of its **Planned** section

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
- The `runtime` button of a cell starts a PostgreSQL container from inside the backend container;
  `docker-compose.yml` already mounts the Docker socket and sets `TESTCONTAINERS_HOST_OVERRIDE`, so it
  works under `compose-up`. Seeding 1M rows takes a few seconds — the popup shows progress
- Footer buttons only draw while the cell is hovered, so drive the pointer with `page.mouse.move`
  before clicking one

## Backward compatibility

The project is in MVP stage, so breaking changes are expected, backward compatibility is not required.

## Code style

- Prefer functional style over OOP
- Prefer having single file - single function
- Name the file after what is inside it: the entry function in snake_case when there is one
  (`measure_statement.go` holds `measureStatement`), the concept when the file holds a family of
  peer functions over one thing (`locks.go`, `partition_bound.go`). Private helpers live beside the
  function they serve
- Prioritize making functions as pure as possible. For example, instead of passing a mutable list to a function, prefer
  returning a new immutable list with result.
