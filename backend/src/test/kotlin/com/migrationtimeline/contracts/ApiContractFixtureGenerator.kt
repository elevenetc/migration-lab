package com.migrationtimeline.contracts

import com.migrationtimeline.models.AlterColumnType
import com.migrationtimeline.models.AlterTable
import com.migrationtimeline.models.Column
import com.migrationtimeline.models.CreateTable
import com.migrationtimeline.models.DropDefault
import com.migrationtimeline.models.DropNotNull
import com.migrationtimeline.models.Migration
import com.migrationtimeline.models.MigrationTimelineResponse
import com.migrationtimeline.models.SetDefault
import com.migrationtimeline.models.SetNotNull
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
                        ),
                        droppedColumns = listOf("legacy_field")
                    )
                )
            ),
            Migration(
                id = "migration-3",
                version = "V3__change_email_type",
                timestamp = 3000L,
                operations = listOf(
                    AlterColumnType(
                        migrationId = "migration-3",
                        tableName = "users",
                        columnName = "email",
                        newType = "TEXT"
                    )
                )
            ),
            Migration(
                id = "migration-4",
                version = "V4__set_email_not_null",
                timestamp = 4000L,
                operations = listOf(
                    SetNotNull(
                        migrationId = "migration-4",
                        tableName = "users",
                        columnName = "email"
                    )
                )
            ),
            Migration(
                id = "migration-5",
                version = "V5__drop_status_not_null",
                timestamp = 5000L,
                operations = listOf(
                    DropNotNull(
                        migrationId = "migration-5",
                        tableName = "users",
                        columnName = "status"
                    )
                )
            ),
            Migration(
                id = "migration-6",
                version = "V6__set_status_default",
                timestamp = 6000L,
                operations = listOf(
                    SetDefault(
                        migrationId = "migration-6",
                        tableName = "users",
                        columnName = "status",
                        defaultValue = "'active'"
                    )
                )
            ),
            Migration(
                id = "migration-7",
                version = "V7__drop_status_default",
                timestamp = 7000L,
                operations = listOf(
                    DropDefault(
                        migrationId = "migration-7",
                        tableName = "users",
                        columnName = "status"
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
