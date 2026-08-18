# Supported Static Analysis

Warnings detected by analyzing migrations.

## Supported

- `AccessExclusiveLock` - ALTER on partitioned tables; statement-scoped (one warning per
  statement, `opIndex = -1`), since the lock is acquired per statement

## Related

Every operation and statement also carries a `performanceClass` - see
[supported-performance-classes.md](supported-performance-classes.md). It is an attribute of every operation
rather than a sparse warning, and describes cost, not lock scope; the two axes compose.
