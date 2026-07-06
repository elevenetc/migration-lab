package com.migrationtimeline.contracts

import com.migrationtimeline.analysis.analyse
import com.migrationtimeline.models.AddColumn
import com.migrationtimeline.models.AddConstraint
import com.migrationtimeline.models.AlterColumnType
import com.migrationtimeline.models.Column
import com.migrationtimeline.models.CreateTable
import com.migrationtimeline.models.DropColumn
import com.migrationtimeline.models.DropConstraint
import com.migrationtimeline.models.DropDefault
import com.migrationtimeline.models.DropNotNull
import com.migrationtimeline.models.DropTable
import com.migrationtimeline.models.Migration
import com.migrationtimeline.models.MigrationTimelineResponse
import com.migrationtimeline.models.RenameColumn
import com.migrationtimeline.models.RenameTable
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
                    AddColumn(
                        migrationId = "migration-2",
                        tableName = "users",
                        column = Column("status", "VARCHAR(50)", listOf("DEFAULT 'active'"))
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
            ),
            Migration(
                id = "migration-8",
                version = "V8__rename_users_to_accounts",
                timestamp = 8000L,
                operations = listOf(
                    RenameTable(
                        migrationId = "migration-8",
                        tableName = "users",
                        newTableName = "accounts"
                    )
                )
            ),
            Migration(
                id = "migration-9",
                version = "V9__rename_email_to_email_address",
                timestamp = 9000L,
                operations = listOf(
                    RenameColumn(
                        migrationId = "migration-9",
                        tableName = "accounts",
                        columnName = "email",
                        newColumnName = "email_address"
                    )
                )
            ),
            Migration(
                id = "migration-10",
                version = "V10__add_unique_constraint",
                timestamp = 10000L,
                operations = listOf(
                    AddConstraint(
                        migrationId = "migration-10",
                        tableName = "accounts",
                        constraintName = "uk_accounts_email",
                        constraintType = "UNIQUE"
                    )
                )
            ),
            Migration(
                id = "migration-11",
                version = "V11__drop_unique_constraint",
                timestamp = 11000L,
                operations = listOf(
                    DropConstraint(
                        migrationId = "migration-11",
                        tableName = "accounts",
                        constraintName = "uk_accounts_email"
                    )
                )
            ),
            Migration(
                id = "migration-12",
                version = "V12__drop_accounts",
                timestamp = 12000L,
                operations = listOf(
                    DropTable(
                        migrationId = "migration-12",
                        tableName = "accounts"
                    )
                )
            ),
            Migration(
                id = "migration-13",
                version = "V13__drop_legacy_column",
                timestamp = 13000L,
                operations = listOf(
                    DropColumn(
                        migrationId = "migration-13",
                        tableName = "users",
                        columnName = "legacy_field"
                    )
                )
            ),
            Migration(
                id = "migration-14",
                version = "V14__create_partitioned_table",
                timestamp = 14000L,
                operations = listOf(
                    CreateTable(
                        migrationId = "migration-14",
                        tableName = "measurements",
                        columns = listOf(
                            Column("id", "SERIAL", emptyList()),
                            Column("created_at", "TIMESTAMP", listOf("NOT NULL")),
                            Column("value", "NUMERIC", emptyList())
                        ),
                        isPartitioned = true,
                        partitionOf = null
                    )
                )
            ),
            Migration(
                id = "migration-15",
                version = "V15__create_partition",
                timestamp = 15000L,
                operations = listOf(
                    CreateTable(
                        migrationId = "migration-15",
                        tableName = "measurements_2024",
                        columns = emptyList(),
                        isPartitioned = false,
                        partitionOf = "measurements"
                    )
                )
            ),
            Migration(
                id = "migration-16",
                version = "V16__alter_partitioned",
                timestamp = 16000L,
                operations = listOf(
                    AddColumn(
                        migrationId = "migration-16",
                        tableName = "measurements",
                        column = Column("description", "TEXT", emptyList())
                    )
                )
            )
        )

        val createTableMap = migrations
            .flatMap { it.operations }
            .filterIsInstance<CreateTable>()
            .associateBy { it.tableName }

        val analysis = analyse(migrations)

        val response = MigrationTimelineResponse(
            timeline = migrations,
            map = migrations.associateBy { it.id },
            createTableMap = createTableMap,
            analysis = analysis
        )

        val fixture = json.encodeToString(MigrationTimelineResponse.serializer(), response)
        val outputFile = File("../api-contracts/fixtures/migration-response.json")
        outputFile.parentFile.mkdirs()
        outputFile.writeText(fixture)
    }
}
