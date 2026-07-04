package com.migrationtimeline.parser

import com.migrationtimeline.models.isChainedAsAlter
import com.migrationtimeline.models.isConnectedWithAsAlter
import com.migrationtimeline.routes.migrationsToResponse
import org.junit.jupiter.api.Assertions.assertTrue
import org.junit.jupiter.api.Test

class MigrationParserTest {

    @Test
    fun `parse simple CREATE TABLE migration`() {
        createUsersSql.toMigration().isEqualTo("create(users(id,name))")
    }

    @Test
    fun `parse ALTER TABLE ADD COLUMN migration`() {
        addLastNameSql.toMigration().isEqualTo("alter(users(+last_name))")
    }

    @Test
    fun `parse ALTER TABLE DROP COLUMN migration`() {
        dropLastNameSql.toMigration().isEqualTo("alter(users(-last_name))")
    }

    @Test
    fun `parse ALTER COLUMN TYPE migration`() {
        alterColumnTypeSql.toMigration().isEqualTo("alterType(users(name:TEXT))")
    }

    @Test
    fun `parse ALTER COLUMN SET NOT NULL migration`() {
        setNotNullSql.toMigration().isEqualTo("setNotNull(users(name))")
    }

    @Test
    fun `parse multiple migrations to DSL`() {
        listOf(createUsersSql, addLastNameSql, createOrgSql).toMigrations().isEqualTo(
            """
            create(org(id,name))
            create(users(id,name)) > alter(users(+last_name))
            """.trimIndent()
        )
    }

    @Test
    fun `create and alter table are connected`() {
        val response = migrationsToResponse(
            mapOf(
                "create-users.sql" to createUsersSql,
                "add-last_name.sql" to addLastNameSql,
            )
        )

        val createUsers = response.timeline[0]
        val addLastName = response.timeline[1]

        assertTrue(isConnectedWithAsAlter(createUsers, addLastName))
    }

    @Test
    fun `create and multiple alters are connected`() {
        val response = migrationsToResponse(
            mapOf(
                "create-users.sql" to createUsersSql,
                "add-last_name.sql" to addLastNameSql,
                "add-age.sql" to addAgeSql,
            )
        )

        assertTrue(isChainedAsAlter(response.timeline))
    }
}

val createUsersSql = """
            CREATE TABLE users (
                id SERIAL PRIMARY KEY,
                name VARCHAR(255) NOT NULL
            );
        """.trimIndent()

val addLastNameSql = """
            ALTER TABLE users ADD COLUMN last_name VARCHAR(255);
        """.trimIndent()

val addAgeSql = """
            ALTER TABLE users ADD COLUMN age INT;
        """.trimIndent()

val dropLastNameSql = """
            ALTER TABLE users DROP COLUMN last_name;
        """.trimIndent()

val createOrgSql = """
            CREATE TABLE org (
                id SERIAL PRIMARY KEY,
                name VARCHAR(255) NOT NULL
            );
        """.trimIndent()

val alterColumnTypeSql = """
            ALTER TABLE users ALTER COLUMN name TYPE TEXT;
        """.trimIndent()

val setNotNullSql = """
            ALTER TABLE users ALTER COLUMN name SET NOT NULL;
        """.trimIndent()