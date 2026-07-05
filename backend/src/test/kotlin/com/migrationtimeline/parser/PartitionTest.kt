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
}

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