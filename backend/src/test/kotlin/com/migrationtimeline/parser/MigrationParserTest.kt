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
    fun `parse ALTER COLUMN DROP NOT NULL migration`() {
        dropNotNullSql.toMigration().isEqualTo("dropNotNull(users(name))")
    }

    @Test
    fun `parse ALTER COLUMN SET DEFAULT migration`() {
        setDefaultSql.toMigration().isEqualTo("setDefault(users(status='active'))")
    }

    @Test
    fun `parse ALTER COLUMN DROP DEFAULT migration`() {
        dropDefaultSql.toMigration().isEqualTo("dropDefault(users(status))")
    }

    @Test
    fun `parse RENAME TABLE migration`() {
        renameTableSql.toMigration().isEqualTo("renameTable(users->accounts)")
    }

    @Test
    fun `parse RENAME COLUMN migration`() {
        renameColumnSql.toMigration().isEqualTo("renameColumn(users(name->full_name))")
    }

    @Test
    fun `parse ADD CONSTRAINT PRIMARY KEY migration`() {
        addPrimaryKeyConstraintSql.toMigration().isEqualTo("addConstraint(users(pk_users:PRIMARY KEY))")
    }

    @Test
    fun `parse ADD CONSTRAINT UNIQUE migration`() {
        addUniqueConstraintSql.toMigration().isEqualTo("addConstraint(users(uk_users_email:UNIQUE))")
    }

    @Test
    fun `parse ADD CONSTRAINT FOREIGN KEY migration`() {
        addForeignKeyConstraintSql.toMigration().isEqualTo("addConstraint(orders(fk_orders_user:FOREIGN KEY))")
    }

    @Test
    fun `parse ADD CONSTRAINT CHECK migration`() {
        addCheckConstraintSql.toMigration().isEqualTo("addConstraint(users(ck_users_age:CHECK))")
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

val dropNotNullSql = """
            ALTER TABLE users ALTER COLUMN name DROP NOT NULL;
        """.trimIndent()

val setDefaultSql = """
            ALTER TABLE users ALTER COLUMN status SET DEFAULT 'active';
        """.trimIndent()

val dropDefaultSql = """
            ALTER TABLE users ALTER COLUMN status DROP DEFAULT;
        """.trimIndent()

val renameTableSql = """
            ALTER TABLE users RENAME TO accounts;
        """.trimIndent()

val renameColumnSql = """
            ALTER TABLE users RENAME COLUMN name TO full_name;
        """.trimIndent()

val addPrimaryKeyConstraintSql = """
            ALTER TABLE users ADD CONSTRAINT pk_users PRIMARY KEY (id);
        """.trimIndent()

val addUniqueConstraintSql = """
            ALTER TABLE users ADD CONSTRAINT uk_users_email UNIQUE (email);
        """.trimIndent()

val addForeignKeyConstraintSql = """
            ALTER TABLE orders ADD CONSTRAINT fk_orders_user FOREIGN KEY (user_id) REFERENCES users(id);
        """.trimIndent()

val addCheckConstraintSql = """
            ALTER TABLE users ADD CONSTRAINT ck_users_age CHECK (age >= 0);
        """.trimIndent()