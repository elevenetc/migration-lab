package com.migrationtimeline.routes.dummy

import com.migrationtimeline.models.MigrationTimelineResponse
import com.migrationtimeline.routes.migrationsToResponse

private val createUsers = """
    CREATE TABLE users (
        id SERIAL PRIMARY KEY,
        name VARCHAR(255) NOT NULL
    );
""".trimIndent()

private val renameUsersToCustomers = """
    ALTER TABLE users RENAME TO customers;
""".trimIndent()

fun simpleTableRename(): MigrationTimelineResponse {
    return migrationsToResponse(
        mapOf(
            "V1__create_users" to createUsers,
            "V2__rename_users_to_customers" to renameUsersToCustomers
        )
    )
}