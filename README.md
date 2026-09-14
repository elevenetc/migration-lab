# Migration Lab

Analyze the performance and impact of PostgreSQL migrations before running them in production.
Migration Lab parses Flyway-style SQL migrations, predicts operation costs, and measures execution,
locking, and retry behavior against PostgreSQL containers seeded with generated data.

- **Static analysis:** inspect SQL without a database, flag risky operations, and predict whether a
  change updates metadata, scans data, or rewrites a table.
- **Runtime analysis:** measure a migration against a seeded database, observe locks,
  compare observed costs with predictions, and report what a cancelled attempt leaves behind.
- **Timeline:** explore tables and migrations in an interactive grid, including renames and partitions.
- **HTML reports:** export the timeline and static analysis as a self-contained file.

Migration Lab is an early-stage project. SQL support is partial, and APIs and output formats may change.
See [supported operations](docs/supported-operations.md) for the current scope.

## Build the CLI

Building locally requires Go 1.25 or later, a C compiler with CGO enabled, Node.js 22 or later with npm,
and [just](https://github.com/casey/just). Docker is needed for execution and runtime analysis.

From the repository root:

```bash
npm ci --prefix frontend
just build-cli
```

This builds the frontend, embeds its assets, and produces `build/migration-lab`. Build the CLI before
running the full test suite on a fresh checkout: the report package requires those generated assets.

### Static analysis

Analyze the included example migrations or a SQL string:

```bash
./build/migration-lab backend/testdata/migrations
./build/migration-lab "CREATE TABLE users (id INT);"
```

The default output is JSON containing `analysisResult`. Static analysis runs without Docker.
Pass your own directory to analyze its `*.sql` files:

```bash
./build/migration-lab /absolute/path/to/migrations
```

Files are ordered by Flyway version, compared numerically by segment. For example,
`V20260410_2__change.sql` comes before `V20260410_10__change.sql`.

### Execution and runtime analysis

With Docker running, execute a migration sequence in a disposable PostgreSQL container:

```bash
./build/migration-lab --run backend/testdata/migrations
```

To measure the last migration, first apply its predecessors and seed the affected tables:

```bash
./build/migration-lab --runtime backend/testdata/rewrite
./build/migration-lab --runtime --rows 50000 --deadline-ms 30000 backend/testdata/rewrite
# Select a migration by filename, applying only its predecessors first
./build/migration-lab --runtime --migration V2__add_email.sql backend/testdata/migrations
```

Runtime analysis uses PostgreSQL 16, defaults to 1,000,000 generated rows per touched table, and gives
the target migration a shared 5,000 ms execution deadline. Setup and seeding happen before that deadline.
The CLI adds `runtimeResult` to the JSON and exits non-zero unless its verdict is `COMPLETED`.

Measurements depend on the generated data, PostgreSQL version, and local resources. Read the
[runtime analysis guide](docs/supported-runtime-analysis.md) for seeding limitations, observed
performance classes, and retry verdicts.

### HTML reports

```bash
./build/migration-lab --report=report.html backend/testdata/migrations
```

Open `report.html` directly in a browser. It includes the frontend assets, migration SQL, and static
analysis, so the timeline works offline. Runtime measurements are available through the CLI and live
web application; they are not included in the HTML report. Review the embedded SQL before sharing a report.

## GitHub Actions

Use the reusable [runtime analysis workflow](.github/workflows/runtime-analysis.yml) from another
repository with one input: `migrations-directory`. It measures new migrations in pull requests, or
the latest migration on a manual run, using a GitHub-hosted Ubuntu runner and disposable PostgreSQL.
Same-repository PRs receive one bot summary comment, updated on subsequent runs. Full results and
logs remain downloadable artifacts. The caller grants `pull-requests: write`; no external server
is needed.

See [GitHub setup and debugging](docs/github-actions.md) for the caller workflow and current limits.

## Web timeline

The Compose setup requires Docker with the Compose plugin and `just`. Go and Node.js run inside the
containers for this setup.

**Use the web application only in a trusted local environment.** The API has no authentication, and
the supplied configuration publishes ports 3000 and 8080 on all host interfaces. The backend can read
mounted migration directories and uses the Docker socket to start containers. Restrict access to those
ports and run only trusted migrations.

Start with the bundled sample files mounted as the migration root:

```bash
MIGRATIONS_ROOT="$PWD/backend/testdata" just compose-up
```

Open [localhost:3000/?migrationId=ecommerce](http://localhost:3000/?migrationId=ecommerce).
Use the dataset selector to explore other examples, including `simple-partition` and `performance-class`.

Tables form the rows and migrations form the columns. Drag to pan, scroll to zoom, and use the left and
right arrow keys to move between cells. Hover a cell for its `focus-in` and `runtime` buttons. `Enter`
focuses the selected cell; `Escape` backs out. The focus panel shows the SQL and warnings, and `runtime`
measures that cell's migration.

### Load your own migrations

Choose a narrow directory to mount read-only:

```bash
MIGRATIONS_ROOT=/absolute/path/to/migrations just compose-up
```

Then open:

```text
http://localhost:3000/?migrationsPath=/absolute/path/to/migrations
```

The path must be accessible inside the backend container. `migrationsPath` takes precedence over
`migrationId`. Without an explicit `MIGRATIONS_ROOT`, Compose mounts `~/dev`; keep the variable set when
working with a narrower directory. It can also be set in a local `.env` file.

Stop the services with:

```bash
just compose-down
```

## Development

The backend is Go with Echo and `pg_query_go`; the frontend is React, TypeScript, Zustand, and
[`active-grid`](https://www.npmjs.com/package/active-grid).

After installing dependencies and building the CLI as described above, run:

```bash
just test
```

This generates API contract fixtures, runs the Go tests, typechecks the frontend, and runs the Vitest
tests. Docker is required for container-backed tests.
The Go tests include the GitHub automation tests, which also require Git. Run those separately with
`just test-automation`.

To run the services directly, use `just run-backend` and `just run-frontend` in separate terminals.
The frontend listens on port 3000 and proxies API requests to port 8080. With Compose, frontend edits
reload automatically; use `just compose-apply` for backend changes, retaining your `MIGRATIONS_ROOT`
setting. That recipe also requires the local Go toolchain.

## Documentation

- [Supported SQL operations](docs/supported-operations.md)
- [Static analysis warnings](docs/supported-static-analysis.md)
- [Predicted performance classes](docs/supported-performance-classes.md)
- [Runtime analysis and findings](docs/supported-runtime-analysis.md)
- [GitHub Actions setup and debugging](docs/github-actions.md)

## License

[Apache License 2.0](LICENSE) © 2026 Eugene Levenetc.
