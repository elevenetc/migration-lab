package com.migrationtimeline.runner

import org.junit.jupiter.api.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class RunMigrationsTest {

    @Test
    fun `runMigrations executes SQL successfully`() {
        val migrations = mapOf(
            "V1__create_users" to """
                CREATE TABLE users (
                    id SERIAL PRIMARY KEY,
                    name VARCHAR(255) NOT NULL
                );
            """.trimIndent()
        )

        val result = runMigrations(migrations)

        assertTrue(result.success, "Expected success but got: ${result.message}")
        assertEquals(1, result.migrationsApplied)
        assertEquals("All migrations applied successfully", result.message)
    }

    @Test
    fun `runMigrations reports failure on invalid SQL`() {
        val migrations = mapOf(
            "V1__invalid" to "INVALID SQL STATEMENT"
        )

        val result = runMigrations(migrations)

        assertEquals(false, result.success)
        assertEquals(0, result.migrationsApplied)
        assertTrue(result.message.contains("Failed at V1__invalid"))
    }
}
