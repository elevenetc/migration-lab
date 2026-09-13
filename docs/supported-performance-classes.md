# Operation Performance Classes

Every parsed operation carries a `performanceClass`: how expensive it is relative to table size, derived from the AST
alone. Every statement carries one too — the worst class among its operations.

It is a **prediction, not a verdict.** [Runtime analysis](supported-runtime-analysis.md) observes the class the database
actually produced and reports the two disagreeing, so the accuracy of this classifier is a measured number rather than an
assumption. The prediction is what the offline lane has — no Docker, an IDE, a pre-commit hook — and what stands in
wherever the measurement is missing or untrustworthy.

The classes are ordinal, cheapest first:

| Class           | Meaning                                                   | Cost vs table size |
|-----------------|-----------------------------------------------------------|--------------------|
| `METADATA_ONLY` | Catalog change only; the lock is held for microseconds    | Constant           |
| `DATA_SCANNING` | Reads every existing row to validate it or build an index | Scales             |
| `TABLE_REWRITE` | Rewrites the whole table on disk                          | Scales (worst)     |

The class is orthogonal to the `AccessExclusiveLock` warning, which describes lock *scope*. A `METADATA_ONLY`
operation can still take an ACCESS EXCLUSIVE lock on a partition parent and all its partitions; the two compose.

## Classification (PostgreSQL 11+)

| Operation                                         | Class           | Notes                                    |
|---------------------------------------------------|-----------------|------------------------------------------|
| `CREATE TABLE`                                    | `METADATA_ONLY` | Empty table                              |
| `DROP TABLE`                                      | `METADATA_ONLY` | Drops the relation                       |
| `DROP COLUMN`                                     | `METADATA_ONLY` | Marks the attribute dropped              |
| `RENAME TABLE` / `RENAME COLUMN`                  | `METADATA_ONLY` | Catalog rename                           |
| `SET DEFAULT` / `DROP DEFAULT`                    | `METADATA_ONLY` | Affects future rows only                 |
| `DROP NOT NULL`                                   | `METADATA_ONLY` | Removes a constraint                     |
| `DROP CONSTRAINT`                                 | `METADATA_ONLY` | Catalog change                           |
| `ADD COLUMN` (no default)                         | `METADATA_ONLY` | Instant since PG 11                      |
| `ADD COLUMN ... DEFAULT <constant>`               | `METADATA_ONLY` | PG 11+ stores the default in the catalog |
| `ADD CONSTRAINT ... NOT VALID`                    | `METADATA_ONLY` | Skips validation of existing rows        |
| `SET NOT NULL`                                    | `DATA_SCANNING` | Validates every existing row             |
| `ADD CONSTRAINT` FK / CHECK                       | `DATA_SCANNING` | Full scan to validate                    |
| `ADD CONSTRAINT` UNIQUE / PRIMARY KEY / EXCLUSION | `DATA_SCANNING` | Builds an index over existing rows       |
| `ADD COLUMN ... DEFAULT <non-volatile call>`      | `METADATA_ONLY` | `now()` is evaluated once and stored     |
| `ALTER COLUMN TYPE`, coercible and non-tightening | `METADATA_ONLY` | `varchar(255)` to `text`, and widening   |
| `ADD COLUMN ... DEFAULT <volatile call>`          | `TABLE_REWRITE` | `random()`, evaluated per row            |
| `ALTER COLUMN TYPE`, anything else                | `TABLE_REWRITE` | Rewrites and verifies                    |

An operation type the classifier does not know is reported as `METADATA_ONLY`.

## The two context-dependent cases

Neither is decidable from the statement's own text, and both used to be answered with a blanket `TABLE_REWRITE`.

**A default rewrites the table only if it is volatile.** PostgreSQL 11+ evaluates a non-volatile default once and stores
it in the catalog as the column's missing value; a volatile one is evaluated per row. Volatility is read from the parsed
expression **tree**, not from the deparsed string, because a cast hides the call: `random()::int` carries no trailing
`()` and is still volatile.

**A type change rewrites the table only if the cast is not binary-coercible, or the new type imposes a constraint the old
one did not.** The target type alone cannot answer this and the pairs are not symmetric — `varchar(100)` to `text` is
free, `text` to `varchar(100)` rewrites and verifies. The old type is resolved by replaying the timeline, so an
`ALTER COLUMN TYPE` is classified against the schema its predecessors built.

Coercible without a rewrite:

| From                            | To                                       |
|---------------------------------|------------------------------------------|
| any type                        | itself                                   |
| `varchar(n)`                    | `text`, `varchar`, or `varchar(m >= n)`  |
| `numeric(p,s)`                  | `numeric`, or `numeric(p2 >= p, s)`      |

## What stays conservative

The prediction still errs towards expensive where the AST cannot settle it:

- **An unknown function is volatile.** Only a short list of built-ins (`now`, `current_timestamp`, `upper`, `md5`, …) is
  known non-volatile; anything else, including a user-defined function, predicts a rewrite.
- **An expression the walker cannot read through is volatile.** An expression we cannot inspect is not one we can call
  cheap.
- **An unknown old type predicts a rewrite.** A column no earlier migration in the timeline declared — a table created
  outside it, or a single-migration analysis — has no old type to compare.
- **An unmodelled operation is `METADATA_ONLY`**, which is the one optimistic default left. Runtime analysis reports such
  a statement on the strength of the observation alone, precisely because nothing said it was safe.

## Frontend

`DATA_SCANNING` and `TABLE_REWRITE` operations raise a cell warning (`data scan` / `table rewrite`), listed alongside
the static analysis warnings in the focus panel and flagged by the warning icon on the cell.
`METADATA_ONLY` raises nothing.

## Implementation

- `backend/internal/models/classify_performance.go` - `ClassifyPerformance(op)`, pure and total
- `backend/internal/models/classify_alter_column_type.go` - the coercibility rules above
- `backend/internal/parser/extract_default_volatility.go` - volatility from the expression tree
- `backend/internal/parser/parse_timeline.go` - `ParseTimeline`, which replays the timeline to resolve each
  `ALTER COLUMN TYPE`'s old type
- `backend/internal/models/worst_performance_class.go` - statement-level roll-up
- `backend/internal/models/operation.go` - injects `performanceClass` at marshal time, so the field can never go stale
  relative to the operations
- `frontend/src/grid/getPerformanceWarning.ts` - maps a class to a cell warning

The `performance-class` dataset (`backend/internal/datasets/performance_class.go`) exercises every class, including the
two fast paths that look expensive and are not, and the volatile default that hides behind a cast.
