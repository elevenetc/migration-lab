package com.migrationtimeline.models

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class MigrationTimelineResponse(
    val timeline: List<Migration>,
    val map: Map<String, Migration>, // migrationId -> Migration
    val createTableMap: Map<String, CreateTable> // tableName -> CreateTable
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
    val columns: List<Column> = emptyList()
) : Operation()

@Serializable
@SerialName("ALTER_TABLE")
data class AlterTable(
    override val migrationId: String,
    val tableName: String,
    val addedColumns: List<Column> = emptyList(),
    val droppedColumns: List<String> = emptyList()
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
data class Column(
    val name: String,
    val type: String,
    val constraints: List<String> = emptyList()
)

