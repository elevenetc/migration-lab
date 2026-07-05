package com.migrationtimeline.parser

import org.junit.jupiter.api.Test

class TimeOrderTest {

    @Test
    fun `single migration has no prefix`() {
        listOf(
            1 to createTableA
        ).toTimedMigrations().isEqualToTimed("create(a(id))")
    }

    @Test
    fun `sequential timestamps show increasing depth`() {
        listOf(
            1 to createTableA,
            2 to createTableB,
            3 to createTableC
        ).toTimedMigrations().isEqualToTimed(
            """
            create(a(id))
            >create(b(id))
            >>create(c(id))
            """.trimIndent()
        )
    }

    @Test
    fun `same timestamp shows same depth`() {
        listOf(
            1 to createTableA,
            2 to createTableB,
            2 to createTableC
        ).toTimedMigrations().isEqualToTimed(
            """
            create(a(id))
            >create(b(id))
            >create(c(id))
            """.trimIndent()
        )
    }

    @Test
    fun `mixed timestamps`() {
        listOf(
            1 to createTableA,
            1 to createTableB,
            3 to createTableC,
            3 to createTableD
        ).toTimedMigrations().isEqualToTimed(
            """
            create(a(id))
            create(b(id))
            >>create(c(id))
            >>create(d(id))
            """.trimIndent()
        )
    }

    @Test
    fun `gap in timestamps reflects actual time`() {
        listOf(
            0 to createTableA,
            5 to createTableB
        ).toTimedMigrations().isEqualToTimed(
            """
            create(a(id))
            >>>>>create(b(id))
            """.trimIndent()
        )
    }

    @Test
    fun `create and alter grouped by table`() {
        listOf(
            1 to createTableA,
            2 to createTableB,
            3 to alterTableA,
            4 to createTableC
        ).toTimedMigrations().isEqualToTimed(
            """
            create(a(id)) > addColumn(a(name))
            >create(b(id))
            >>>create(c(id))
            """.trimIndent()
        )
    }

    @Test
    fun `multiple alters grouped on same line`() {
        listOf(
            1 to createTableA,
            2 to alterTableA,
            2 to alterTableAAge,
            3 to createTableB
        ).toTimedMigrations().isEqualToTimed(
            """
            create(a(id)) > addColumn(a(name)) > addColumn(a(age))
            >>create(b(id))
            """.trimIndent()
        )
    }
}

private val createTableA = "CREATE TABLE a (id SERIAL);"
private val createTableB = "CREATE TABLE b (id SERIAL);"
private val createTableC = "CREATE TABLE c (id SERIAL);"
private val createTableD = "CREATE TABLE d (id SERIAL);"
private val alterTableA = "ALTER TABLE a ADD COLUMN name VARCHAR(255);"
private val alterTableAAge = "ALTER TABLE a ADD COLUMN age INT;"