package com.migrationtimeline.analysis

import com.migrationtimeline.models.AccessExclusiveLock
import com.migrationtimeline.parser.parseMigration
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertTrue
import org.junit.jupiter.api.Test

class StaticAnalysisTest {

    @Test
    fun `ALTER on partitioned table produces warning`() {
        val migrations = parseMigration(
            linkedMapOf(
                "V1__create_partitioned" to createPartitionedTable,
                "V2__add_column" to addColumnToPartitioned
            )
        )

        val result = analyse(migrations)

        assertEquals(1, result.warnings.size)
        val warning = result.warnings[0] as AccessExclusiveLock
        assertEquals("events", warning.tableName)
        assertEquals("V2__add_column", warning.operationId.migrationId)
        assertTrue(warning.message.contains("ACCESS EXCLUSIVE"))
    }

    @Test
    fun `ALTER on non-partitioned table produces no warning`() {
        val migrations = parseMigration(
            linkedMapOf(
                "V1__create_table" to createRegularTable,
                "V2__add_column" to addColumnToRegular
            )
        )

        val result = analyse(migrations)

        assertEquals(0, result.warnings.size)
    }

    @Test
    fun `multiple ALTERs on same partitioned table produce multiple warnings`() {
        val migrations = parseMigration(
            linkedMapOf(
                "V1__create_partitioned" to createPartitionedTable,
                "V2__add_column" to addColumnToPartitioned,
                "V3__drop_column" to dropColumnFromPartitioned
            )
        )

        val result = analyse(migrations)

        assertEquals(2, result.warnings.size)
        assertTrue(result.warnings.all { it is AccessExclusiveLock })
        assertTrue(result.warnings.all { (it as AccessExclusiveLock).tableName == "events" })
    }

    @Test
    fun `ALTER on partition child does not produce warning`() {
        val migrations = parseMigration(
            linkedMapOf(
                "V1__create_partitioned" to createPartitionedTable,
                "V2__create_partition" to createPartition,
                "V3__alter_partition" to alterPartitionChild
            )
        )

        val result = analyse(migrations)

        assertEquals(0, result.warnings.size)
    }
}

private val createPartitionedTable = """
    CREATE TABLE events (
        id SERIAL,
        name VARCHAR(255),
        year INT NOT NULL
    ) PARTITION BY LIST (year);
""".trimIndent()

private val createPartition = """
    CREATE TABLE events_2025 PARTITION OF events
        FOR VALUES IN (2025);
""".trimIndent()

private val addColumnToPartitioned = """
    ALTER TABLE events ADD COLUMN description TEXT;
""".trimIndent()

private val dropColumnFromPartitioned = """
    ALTER TABLE events DROP COLUMN name;
""".trimIndent()

private val alterPartitionChild = """
    ALTER TABLE events_2025 ADD COLUMN local_field TEXT;
""".trimIndent()

private val createRegularTable = """
    CREATE TABLE users (
        id SERIAL PRIMARY KEY,
        name VARCHAR(255)
    );
""".trimIndent()

private val addColumnToRegular = """
    ALTER TABLE users ADD COLUMN email VARCHAR(255);
""".trimIndent()
