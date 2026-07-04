package com.migrationtimeline.contracts

import com.migrationtimeline.models.*
import kotlinx.serialization.json.Json
import org.junit.jupiter.api.Test
import java.io.File

class ApiContractFixtureGenerator {

    private val json = Json {
        prettyPrint = true
        encodeDefaults = true
    }

    @Test
    fun `generate migration response fixture`() {
        val migrations = listOf(
            Migration(
                id = "migration-1",
                version = "V1__create_users",
                timestamp = 1000L,
                operations = listOf(
                    CreateTable(
                        migrationId = "migration-1",
                        tableName = "users",
                        columns = listOf(
                            Column("id", "SERIAL", listOf("PRIMARY KEY")),
                            Column("email", "VARCHAR(255)", listOf("NOT NULL", "UNIQUE")),
                            Column("created_at", "TIMESTAMP", listOf("DEFAULT NOW()"))
                        )
                    )
                )
            ),
            Migration(
                id = "migration-2",
                version = "V2__add_user_status",
                timestamp = 2000L,
                operations = listOf(
                    AlterTable(
                        migrationId = "migration-2",
                        tableName = "users",
                        addedColumns = listOf(
                            Column("status", "VARCHAR(50)", listOf("DEFAULT 'active'"))
                        )
                    )
                )
            )
        )

        val createTableMap = migrations
            .flatMap { it.operations }
            .filterIsInstance<CreateTable>()
            .associateBy { it.tableName }

        val response = MigrationTimelineResponse(
            timeline = migrations,
            map = migrations.associateBy { it.id },
            createTableMap = createTableMap
        )

        val fixture = json.encodeToString(MigrationTimelineResponse.serializer(), response)
        val outputFile = File("../api-contracts/fixtures/migration-response.json")
        outputFile.parentFile.mkdirs()
        outputFile.writeText(fixture)
    }
}
