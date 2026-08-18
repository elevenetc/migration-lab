# Operation Performance Classes

Every parsed operation carries a `performanceClass`: how expensive it is relative to table size, derived from the AST
alone. Every statement carries one too — the worst class among its operations.

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
| `ADD COLUMN ... DEFAULT <function call>`          | `TABLE_REWRITE` | Treated as a volatile default            |
| `ALTER COLUMN TYPE`                               | `TABLE_REWRITE` | Rewrites the table                       |

An operation type the classifier does not know is reported as `METADATA_ONLY`.

## Conservative assumptions

Some cases are context-dependent. They are classified to the worse class rather than parsed in full:

- **`ALTER COLUMN TYPE` is always `TABLE_REWRITE`.** Binary-compatible changes (`varchar(10)` to `varchar(20)`,
  `varchar` to `text`) are actually metadata-only, but the parser does not compare old and new types.
- **Any function-call default is `TABLE_REWRITE`.** PostgreSQL evaluates a non-volatile default such as `now()`
  once and stores it without touching existing rows; only volatile defaults such as `random()` force a rewrite. The
  classifier does not track function volatility.
- **An undeparsable default is `TABLE_REWRITE`.** The column carries the `DEFAULT` token in `constraints` but an empty
  `defaultExpr`, so the expensive path is assumed.

## Frontend

`DATA_SCANNING` and `TABLE_REWRITE` operations raise a cell warning (`data scan` / `table rewrite`), listed alongside
the static analysis warnings in the focus panel and flagged by the warning icon on the cell.
`METADATA_ONLY` raises nothing.

## Implementation

- `backend/internal/models/classify_performance.go` - `ClassifyPerformance(op)`, pure and total
- `backend/internal/models/worst_performance_class.go` - statement-level roll-up
- `backend/internal/models/operation.go` - injects `performanceClass` at marshal time, so the field can never go stale
  relative to the operations
- `frontend/src/grid/getPerformanceWarning.ts` - maps a class to a cell warning

The `performance-class` dataset (`backend/internal/datasets/performance_class.go`) exercises every class.
