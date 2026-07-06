package com.migrationtimeline.runner

import com.migrationtimeline.models.RunMigrationsResult
import org.slf4j.LoggerFactory
import org.testcontainers.postgresql.PostgreSQLContainer
import java.sql.DriverManager.getConnection

private val log = LoggerFactory.getLogger("runMigrations")

fun runMigrations(migrations: Map<String, String>): RunMigrationsResult {
    log.info("Starting migrations: ${migrations.size} total")

    return PostgreSQLContainer("postgres:16-alpine").use { postgres ->
        postgres.start()
        log.info("PostgreSQL container started: ${postgres.jdbcUrl}")

        getConnection(
            postgres.jdbcUrl,
            postgres.username,
            postgres.password
        ).use { connection ->
            var applied = 0

            for ((version, sql) in migrations) {
                log.info("Applying migration: $version")
                log.debug("SQL:\n$sql")
                try {
                    connection.createStatement().use { stmt ->
                        stmt.execute(sql)
                    }
                    applied++
                    log.info("Migration $version applied successfully")
                } catch (e: Exception) {
                    log.error("Migration $version failed: ${e.message}")
                    log.error("Failed SQL:\n$sql")
                    return@use RunMigrationsResult(
                        success = false,
                        message = "Failed at $version: ${e.message}",
                        migrationsApplied = applied
                    )
                }
            }

            log.info("All $applied migrations applied successfully")
            RunMigrationsResult(
                success = true,
                message = "All migrations applied successfully",
                migrationsApplied = applied
            )
        }
    }
}
