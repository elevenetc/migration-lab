package com.migrationtimeline.routes.dummy

import com.migrationtimeline.models.MigrationTimelineResponse
import com.migrationtimeline.routes.migrationsToResponse

private val createUsers = """
    CREATE TABLE users (
        id SERIAL PRIMARY KEY,
        name VARCHAR(255) NOT NULL
    );
""".trimIndent()

private val addAgeColumn = """
    ALTER TABLE users ADD COLUMN age INTEGER;
""".trimIndent()

fun simpleUsers(): MigrationTimelineResponse {
    return migrationsToResponse(
        mapOf(
            "v1" to createUsers,
            "v2" to addAgeColumn
        )
    )
}
