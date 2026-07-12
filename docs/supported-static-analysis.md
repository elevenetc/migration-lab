# Supported Static Analysis

Warnings detected by analyzing migrations.

## Supported

- `AccessExclusiveLock` - ALTER on partitioned tables; statement-scoped (one warning per
  statement, `opIndex = -1`), since the lock is acquired per statement
