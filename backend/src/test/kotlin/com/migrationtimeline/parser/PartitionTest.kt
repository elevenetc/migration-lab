package com.migrationtimeline.parser

import org.junit.jupiter.api.Test

class PartitionTest {

    @Test
    fun `parse CREATE TABLE with PARTITION BY`() {
        createPartitionedTableSql.toMigration().isEqualTo("createPartitioned(events(id,name,year))")
    }

    @Test
    fun `parse CREATE TABLE PARTITION OF`() {
        createPartitionOf2025Sql.toMigration().isEqualTo("createPartition(events_2025:events)")
    }

    @Test
    fun `parse partitioned table with multiple partitions`() {
        listOf(
            1 to createPartitionedTableSql,
            2 to createPartitionOf2025Sql,
            2 to createPartitionOf2026Sql
        ).toTimedMigrations().isEqualToTimed(
            """
            createPartitioned(events(id,name,year))
            >createPartition(events_2025:events)
            >createPartition(events_2026:events)
            """.trimIndent()
        )
    }

    @Test
    fun `parse PARTITION BY RANGE`() {
        createOrdersHistoryPartitionedSql.toMigration().isEqualTo("createPartitioned(orders_history(id,order_id,status,changed_at))")
    }

    @Test
    fun `parse PARTITION OF with RANGE values`() {
        createOrdersHistory2024Sql.toMigration().isEqualTo("createPartition(orders_history_2024:orders_history)")
    }
}

private val createOrdersHistoryPartitionedSql = """
    CREATE TABLE orders_history (
        id SERIAL,
        order_id INTEGER NOT NULL,
        status VARCHAR(50) NOT NULL,
        changed_at TIMESTAMP NOT NULL
    ) PARTITION BY RANGE (changed_at);
""".trimIndent()

private val createOrdersHistory2024Sql = """
    CREATE TABLE orders_history_2024 PARTITION OF orders_history
        FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');
""".trimIndent()

private val createPartitionedTableSql = """
    CREATE TABLE events (
        id SERIAL,
        name VARCHAR(255) NOT NULL,
        year INT NOT NULL
    ) PARTITION BY LIST (year);
""".trimIndent()

private val createPartitionOf2025Sql = """
    CREATE TABLE events_2025 PARTITION OF events
        FOR VALUES IN (2025);
""".trimIndent()

private val createPartitionOf2026Sql = """
    CREATE TABLE events_2026 PARTITION OF events
        FOR VALUES IN (2026);
""".trimIndent()