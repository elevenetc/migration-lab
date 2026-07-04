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

private val setNameNotNull = """
    ALTER TABLE users ALTER COLUMN name SET NOT NULL;
""".trimIndent()

private val dropLastnameNotNull = """
    ALTER TABLE users ALTER COLUMN lastname DROP NOT NULL;
""".trimIndent()

private val setLastnameDefault = """
    ALTER TABLE users ALTER COLUMN lastname SET DEFAULT '';
""".trimIndent()

private val dropLastnameDefault = """
    ALTER TABLE users ALTER COLUMN lastname DROP DEFAULT;
""".trimIndent()

fun simpleUserAndOrg(): MigrationTimelineResponse {
    return migrationsToResponse(
        mapOf(
            "V1__create_users" to createUser,
            "V2__add_lastname" to addUserLastName,
            "V3__create_orgs" to createOrg,
            "V4__change_name_type" to changeNameType,
            "V5__set_name_not_null" to setNameNotNull,
            "V6__drop_lastname_not_null" to dropLastnameNotNull,
            "V7__set_lastname_default" to setLastnameDefault,
            "V8__drop_lastname_default" to dropLastnameDefault
        )
    )
}