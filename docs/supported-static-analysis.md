# Supported Static Analysis

Warnings detected by analyzing migrations.

## Supported

- `AccessExclusiveLock` - ALTER on partitioned tables; statement-scoped (one warning per
  statement, `opIndex = -1`), since the lock is acquired per statement

## Related

Every operation and statement also carries a `performanceClass` - see
[supported-performance-classes.md](supported-performance-classes.md). It is an attribute of every operation
rather than a sparse warning, and describes cost, not lock scope; the two axes compose.

That class is a *prediction*: it is what the offline lane has when there is no Docker to run against, and
[runtime analysis](supported-runtime-analysis.md) observes the class the database actually produced and reports the two
disagreeing. Structure stays static analysis's own — the timeline, the rename and partition edges, which tables to seed,
and the statements a run never reached.
