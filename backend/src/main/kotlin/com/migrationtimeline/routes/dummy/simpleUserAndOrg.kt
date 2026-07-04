package com.migrationtimeline.routes.dummy

import com.migrationtimeline.models.MigrationTimelineResponse
import com.migrationtimeline.routes.migrationsToResponse

private val createUser = """
    CREATE TABLE users (
        id SERIAL PRIMARY KEY,
        name VARCHAR(255) NOT NULL
    );
""".trimIndent()

private val addUserLastName = """
    ALTER TABLE users ADD COLUMN lastname VARCHAR(255) NOT NULL;
""".trimIndent()

private val createOrg = """
    CREATE TABLE orgs (
        id SERIAL PRIMARY KEY,
        address VARCHAR(255) NOT NULL
    );
""".trimIndent()

private val changeNameType = """
    ALTER TABLE users ALTER COLUMN name TYPE TEXT;
""".trimIndent()

fun simpleUserAndOrg(): MigrationTimelineResponse {
    return migrationsToResponse(
        mapOf(
            "V1__create_users" to createUser,
            "V2__add_lastname" to addUserLastName,
            "V3__create_orgs" to createOrg,
            "V4__change_name_type" to changeNameType
        )
    )
}