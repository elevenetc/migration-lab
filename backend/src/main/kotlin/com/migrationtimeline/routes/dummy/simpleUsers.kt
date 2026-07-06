package com.migrationtimeline.routes.dummy

private val createUsers = """
    CREATE TABLE users (
        id SERIAL PRIMARY KEY,
        name VARCHAR(255) NOT NULL
    );
""".trimIndent()

private val addAgeColumn = """
    ALTER TABLE users ADD COLUMN age INTEGER;
""".trimIndent()

val simpleUsersMigrations: LinkedHashMap<String, String> = linkedMapOf(
    "v1" to createUsers,
    "v2" to addAgeColumn
)
