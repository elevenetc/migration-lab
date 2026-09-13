# Supported Runtime Analysis

Findings produced by *executing* one migration against a real PostgreSQL container, on tables seeded to production-like
size.

Static analysis reads the AST and reasons about what a statement *is*. Runtime analysis runs it and observes what it
*does*: duration, locks held, who was blocked, what a killed attempt leaves behind.

> Status: [What runs today](#what-runs-today) is implemented. Sections marked **planned** are not.

## Relation to the other analysis axes

| Axis                                                   | Source         | Output                       | Question                       |
|--------------------------------------------------------|----------------|------------------------------|--------------------------------|
| [`performanceClass`](supported-performance-classes.md) | AST            | attribute on every operation | What class of cost *should* this be? |
| [Static analysis](supported-static-analysis.md)        | AST            | sparse warning               | What is structurally risky?    |
| Runtime analysis                                       | real execution | measurement + sparse finding | What happened at real scale?   |

Cost is the axis where the two overlap, and runtime is the authority on it: the AST class is a *prediction* that the
measurement scores, and the observed class is what the findings reason about. The prediction stands in only where there is
nothing to observe — see [Observed performance class](#observed-performance-class).

Runtime findings carry the same fields as static warnings (`type`, `operationId`, `tableName`, `message`), so both
render through one path. Measurements ride alongside as extra payload.

## What runs today

Entry point `Analyse` (`internal/analysis/runtime`), given a timeline and which migration to measure:

1. Start a throwaway `postgres:16-alpine` container — pinned, so version-gated fast paths match what is measured
2. Apply every migration *before* the analyzed one
3. Read the catalog: tables, partitioning, fillable columns
4. [Seed](#seeding) the touched tables, then `ANALYZE` each so the planner sees real statistics
5. Open a [reader probe](#probes) per seeded table
6. Execute the statements under a shared [deadline](#deadline-and-retry-safety), sampling locks throughout and
   snapshotting the schema around each one
7. Derive each statement's [observed performance class](#observed-performance-class) and pair it with the prediction
8. Classify the run, check what a killed attempt left behind, turn measurements into [findings](#findings)

Notes:

- Measured inside one transaction that is rolled back, so the seeded database is reusable
- A statement PostgreSQL refuses to run in a transaction (`CONCURRENTLY`, `VACUUM`, `REINDEX`, `ALTER SYSTEM`, database
  and tablespace DDL) forces the migration to run without one; its effects stay applied
- Statements are split with `pg_query`, so unmodelled ones run too — a backfill `UPDATE` is measured like any other

## Observed performance class

What class of cost a statement really was, from two snapshots of the public schema taken around it. Every signal is a
**count**, so the answer does not move with the hardware the run happened on:

1. a relation's `pg_class.relfilenode` changed — which happens if and only if it was rewritten — then `TABLE_REWRITE`,
   and the relations are named in `rewrittenRelations`
2. otherwise `pg_stat_get_xact_tuples_returned` grew, so rows were read: `DATA_SCANNING`
3. otherwise `METADATA_ONLY`

Both snapshots are read on the connection running the migration: a `relfilenode` swapped by an uncommitted `ALTER TABLE`
is invisible to every other backend, and the tuple counters are transaction-local to the calling backend. They are read
outside the timed window, so `durationMs` stays comparable.

The measurement is reported as `observedClass` beside the AST's `predictedClass`, and three cases leave `observedClass`
empty, so the prediction stands in:

- **the statement did not complete.** A cancelled or failed statement leaves the transaction aborted, where nothing can
  be read. The deadline or failure finding is the answer for those anyway
- **the migration could not run in a transaction** (`CONCURRENTLY`, `VACUUM`, …). The tuple counters are
  transaction-local, so outside one they read back as zero and cannot tell a scan from a catalog change. A rewrite is
  still decidable there, since `relfilenode` is committed state
- **a table could not be seeded.** A table holding no rows reads nothing and rewrites nothing measurable, so the
  observation collapses to `METADATA_ONLY` — the same shape a genuinely cheap statement produces, and blind must not read
  as clean. An observed `TABLE_REWRITE` survives, because a changed `relfilenode` proves a rewrite at any row count.
  Seed trust is a property of the whole run, since a partitioned parent is seeded through leaves whose names do not match
  the statement's table

## Inputs

Every touched table is seeded to **1,000,000 rows**. Production data is never copied. `--rows` (and the `rows` query
parameter) lowers the count; tests and local runs use it to stay quick.

**Planned**, in the order the need would arrive:

- Per-table row counts from `pg_class.reltuples`, so a 42M-row table is not measured as a 1M-row one
- Column statistics (`n_distinct`, `null_frac`) — `generate_series` gives uniform non-null values, which is why
  `ADD CONSTRAINT ... UNIQUE` and `SET NOT NULL` pass here and can fail on real data
- Table bytes and index counts, for index build cost

## Seeding

Which tables:

- The tables the migration's operations name, minus ones it creates or drops
- A partitioned parent resolves to its partitions, each pinned to a key value its own bound admits (range lower bound,
  first list value)
- A hash-partitioned parent is seeded itself, letting tuple routing place the rows
- `DEFAULT` partitions and ranges open at the bottom (`FROM (MINVALUE)`) are skipped
- No expansion for foreign keys: `ADD FOREIGN KEY` scans the child, which the operation already names

How: `generate_series`, skipping columns the database fills itself (identity, generated, `serial`).

| Type                                                         | Value                             |
|--------------------------------------------------------------|-----------------------------------|
| `smallint`, `integer`, `bigint`, `numeric`, `real`, `double` | the series counter                |
| `boolean`                                                    | alternating                       |
| `text`, `character varying`, `character`                     | `row_<i>`, cut to declared length |
| `uuid`                                                       | `gen_random_uuid()`               |
| `date`, `timestamp`, `timestamptz`                           | offset back from now              |
| `json`, `jsonb`                                              | an object holding the counter     |
| `bytea`                                                      | the counter's bytes               |

Anything else (enums, arrays, domains) is left to its default. A `NOT NULL` column with no filler fails the seed: raised
as `SEED_FAILED`, and the run continues against an empty table rather than crashing.

## Probes

- One concurrent reader per seeded table, `SELECT 1 FROM t LIMIT 1` every 5 ms, at most four tables
- The reads take `ACCESS SHARE`, so `ACCESS EXCLUSIVE` stalls them exactly as it stalls production readers
- *Blocked* is not inferred from latency — any threshold is wrong on hardware it was not picked for. PostgreSQL is asked
  directly: `pg_stat_activity.wait_event_type = 'Lock'` plus `pg_blocking_pids`, sampled every 20 ms
- Reported per table: reads, errors, slowest read (raw data, not a verdict), time observed waiting
- Locks come from `pg_locks` on the same sampling connection. Relations gone from `pg_class` — the transient heap of a
  rewrite — are dropped; parent *and* every partition show up, so fan-out is visible

**Planned**: a `writer` probe (blocked writers, serialization failures) and a `longTransaction` probe (DDL queueing
behind a slow query, then head-of-line blocking every later reader).

## Deadline and retry safety

The deadline stands for a pod's termination grace period (default 5,000 ms), enforced as `statement_timeout` and
decremented by what earlier statements spent. A statement still running when it expires is cancelled (SQLSTATE `57014`)
and the run stops, the way a killed pod stops.

| Verdict            | Meaning                                                            |
|--------------------|--------------------------------------------------------------------|
| `COMPLETED`        | every statement finished inside the deadline                       |
| `EXCEEDS_DEADLINE` | a statement was cancelled at the deadline                          |
| `FAILED`           | a statement raised an error, or the run could not be set up at all |

What the cancelled attempt left behind decides whether a restarting pod ever converges:

| Retry verdict          | Meaning                                                                              |
|------------------------|--------------------------------------------------------------------------------------|
| `NOT_APPLICABLE`       | the migration completed, nothing was killed                                          |
| `NEEDS_MANUAL_CLEANUP` | an invalid index is left behind; the next attempt fails until it is dropped by hand  |
| `FAILURE_LOOP`         | nothing left behind, so every attempt does the same work and dies at the same point  |
| `SAFE_TO_RETRY`        | reserved for migrations making durable progress each attempt; nothing reports it yet |

`FAILURE_LOOP` is the failure mode this feature exists to catch.

## Findings

| Type                  | Raised when                                                                   |
|-----------------------|-------------------------------------------------------------------------------|
| `EXCEEDS_DEADLINE`    | a statement was cancelled at the deadline                                     |
| `STATEMENT_FAILED`    | a statement raised an error                                                   |
| `EXCLUSIVE_LOCK_HELD` | a reader-blocking lock **and** a cost that scales with table size — see below |
| `BLOCKS_READERS`      | a reader was observed waiting on a lock the migration held                    |
| `INVALID_INDEX_LEFT`  | the cancelled run left an invalid index behind                                |
| `SEED_FAILED`         | a table could not be filled, so its measurements are of an empty table        |
| `CLASS_UNDERSTATED`   | the observed class was worse than the predicted one                           |
| `CLASS_OVERSTATED`    | the observed class was cheaper than the predicted one                         |

The two class findings are what makes the prediction accountable. `CLASS_UNDERSTATED` is the direction that matters: a
statement the offline gate called cheap and the database rewrote — a migration that passes CI and then rewrites the table
in production. `CLASS_OVERSTATED` is noise in the prediction rather than risk in the migration; it is what turns a
warning into background noise. Both are silent unless *both* classes are known, since an unmeasured statement is not
evidence.

Why the lock finding asks what the statement cost:

- Every `ALTER TABLE` takes `ACCESS EXCLUSIVE`, metadata-only ones included, so the mode alone says nothing
- What matters is not milliseconds — CI hardware moves that, and a fast runner hides the problem — but whether the work
  under the lock is constant in table size or grows with it, which is what the performance class answers
- So: raised when the class is worse than `METADATA_ONLY` **and** a reader-blocking lock was observed. No threshold, no
  calibration factor
- The class it uses is `observedClass` where there is one, and `predictedClass` where there is not
- A statement neither measured nor modelled (`VACUUM FULL`, a hand-written index build) is reported on the observation
  alone: nothing said it was safe
- Below that bar the lock is still reported as the measurement's `strongestLock` — an attribute, not a warning

## Interfaces

```
POST /api/migrations/runtime-analysis?migrationId=<dataset>&migration=<migration id>[&rows=&deadlineMs=]
POST /api/migrations/runtime-analysis?migrationsPath=<abs-dir>&migration=<migration id>
```

`migrationsPath` takes precedence over `migrationId`. Omitting `migration` measures the last of the timeline — the one
that just arrived on the branch.

```bash
# Last migration of a directory, every touched table seeded to 1M rows
./build/migration-lab --runtime /path/to/migrations

# Smaller and quicker, with a 30 s grace period
./build/migration-lab --runtime --rows 50000 --deadline-ms 30000 /path/to/migrations

# Select an earlier migration by its filename; later migrations are excluded
./build/migration-lab --runtime --migration V42__add_index.sql /path/to/migrations
```

- CLI prints `runtimeResult` alongside the static `analysisResult`, and exits non-zero unless the verdict is
  `COMPLETED` — which is what fails the CI job
- Web client: each migration cell has a `runtime` button under `focus-in`, result shown in a popup
- [GitHub Actions](github-actions.md): a reusable workflow measures added migrations in pull requests
  and retains JSON and diagnostic logs as artifacts

## Reading the numbers

CI hardware is not production hardware, and a freshly seeded container sits in page cache while production does not.
Absolute durations are optimistic and not portable.

- Prefer ratios over seconds: "exceeds the grace period by 14x" survives a hardware change, "29 minutes" does not
- Lock modes, blocked-reader observations, observed classes and retry verdicts are hardware-independent — trust those.
  The observed class is derived from counts (`relfilenode`, rows read), never from a duration
- No finding is gated on a duration threshold, so a fast runner cannot make a problem disappear. The only threshold is
  the deadline, which the caller sets to its own pod's grace period
- The rollback's own cost is not measured, and a real deploy would pay it

## Planned

- **Strategies** — run the naive form *and* its safer rewrites against the same seeded data, then rank. The existing
  rollback makes this possible without re-seeding

  | Operation                          | Safer form to compare against                                        |
    |------------------------------------|----------------------------------------------------------------------|
  | `ADD COLUMN` with volatile default | nullable column, batched backfill, then set default                  |
  | `ALTER COLUMN TYPE`                | expand/contract via new column and swap; per-partition detach/attach |
  | `SET NOT NULL`                     | `ADD CHECK ... NOT VALID`, `VALIDATE`, then `SET NOT NULL` (PG 12+)  |
  | `ADD CHECK` / `ADD FOREIGN KEY`    | `NOT VALID`, then `VALIDATE CONSTRAINT`                              |
  | `CREATE INDEX`                     | `CONCURRENTLY`, with `INVALID` index cleanup on failure              |
  | DDL on a partitioned parent        | per-partition loop with `lock_timeout` and retry                     |

- **Scaling curves** — measure at two row counts and fit, so scaling is reported instead of one extrapolated point
- **Seed caching** — keyed by schema hash and row count, so repeated CI runs do not re-seed
- **Writer and long-transaction probes** — see [Probes](#probes)
- **Timeline integration** — findings are warning-shaped already, but only the popup and CLI show them; the grid and the
  HTML report do not

## Notes

- Cost: static analysis stays the fast gate on every migration; the runtime pass is the escalation for what it flags
- Out of scope: compatibility between old pods and the new schema during a rolling deploy

## Related

- [supported-operations.md](supported-operations.md)
- [supported-static-analysis.md](supported-static-analysis.md)
- [supported-performance-classes.md](supported-performance-classes.md)
