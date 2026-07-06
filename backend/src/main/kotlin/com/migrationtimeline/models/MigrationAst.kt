package com.migrationtimeline.models

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class MigrationTimelineResponse(
    val timeline: List<Migration>,
    val map: Map<String, Migration>, // migrationId -> Migration
    val createTableMap: Map<String, CreateTable>, // tableName -> CreateTable
    val analysis: AnalysisResult
)

@Serializable
data class Migration(
    val id: String,
    val version: String,
    val timestamp: Long,
    val operations: List<Operation>
)

@Serializable
sealed class Operation {
    abstract val migrationId: String
}

@Serializable
@SerialName("CREATE_TABLE")
data class CreateTable(
    override val migrationId: String,
    val tableName: String,
    val columns: List<Column> = emptyList(),
    val isPartitioned: Boolean = false,
    val partitionOf: String? = null
) : Operation()

@Serializable
@SerialName("ADD_COLUMN")
data class AddColumn(
    override val migrationId: String,
    val tableName: String,
    val column: Column
) : Operation()

@Serializable
@SerialName("ALTER_COLUMN_TYPE")
data class AlterColumnType(
    override val migrationId: String,
    val tableName: String,
    val columnName: String,
    val newType: String
) : Operation()

@Serializable
@SerialName("SET_NOT_NULL")
data class SetNotNull(
    override val migrationId: String,
    val tableName: String,
    val columnName: String
) : Operation()

@Serializable
@SerialName("DROP_NOT_NULL")
data class DropNotNull(
    override val migrationId: String,
    val tableName: String,
    val columnName: String
) : Operation()

@Serializable
@SerialName("SET_DEFAULT")
data class SetDefault(
    override val migrationId: String,
    val tableName: String,
    val columnName: String,
    val defaultValue: String
) : Operation()

@Serializable
@SerialName("DROP_DEFAULT")
data class DropDefault(
    override val migrationId: String,
    val tableName: String,
    val columnName: String
) : Operation()

@Serializable
@SerialName("RENAME_TABLE")
data class RenameTable(
    override val migrationId: String,
    val tableName: String,
    val newTableName: String
) : Operation()

@Serializable
@SerialName("RENAME_COLUMN")
data class RenameColumn(
    override val migrationId: String,
    val tableName: String,
    val columnName: String,
    val newColumnName: String
) : Operation()

@Serializable
@SerialName("ADD_CONSTRAINT")
data class AddConstraint(
    override val migrationId: String,
    val tableName: String,
    val constraintName: String,
    val constraintType: String
) : Operation()

@Serializable
@SerialName("DROP_CONSTRAINT")
data class DropConstraint(
    override val migrationId: String,
    val tableName: String,
    val constraintName: String
) : Operation()

@Serializable
@SerialName("DROP_TABLE")
data class DropTable(
    override val migrationId: String,
    val tableName: String
) : Operation()

@Serializable
@SerialName("DROP_COLUMN")
data class DropColumn(
    override val migrationId: String,
    val tableName: String,
    val columnName: String
) : Operation()

@Serializable
data class Column(
    val name: String,
    val type: String,
    val constraints: List<String> = emptyList()
)

