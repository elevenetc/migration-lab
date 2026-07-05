package com.migrationtimeline.parser

import com.migrationtimeline.models.AddConstraint
import com.migrationtimeline.models.AlterColumnType
import com.migrationtimeline.models.AlterTable
import com.migrationtimeline.models.CreateTable
import com.migrationtimeline.models.DropConstraint
import com.migrationtimeline.models.DropDefault
import com.migrationtimeline.models.DropNotNull
import com.migrationtimeline.models.Migration
import com.migrationtimeline.models.Operation
import com.migrationtimeline.models.RenameColumn
import com.migrationtimeline.models.RenameTable
import com.migrationtimeline.models.SetDefault
import com.migrationtimeline.models.SetNotNull
import java.util.UUID
import kotlin.test.assertEquals

fun String.toMigration(): Migration =
    MigrationParser.parse(UUID.randomUUID().toString(), this)

fun List<String>.toMigrations(): List<Migration> =
    map { it.toMigration() }

fun Migration.isEqualTo(expectedDsl: String) {
    assertEquals(expectedDsl, toTestDsl())
}

fun List<Migration>.isEqualTo(expectedDsl: String) {
    assertEquals(expectedDsl, toTestDsl())
}

fun List<Migration>.toTestDsl(): String =
    flatMap { it.operations }
        .groupBy { it.tableName() }
        .toSortedMap()
        .map { (_, ops) -> ops.joinToString(" > ") { it.toTestDsl() } }
        .joinToString("\n")

fun Migration.toTestDsl(): String =
    operations.joinToString(" > ") { it.toTestDsl() }

private fun Operation.tableName(): String = when (this) {
    is CreateTable -> tableName
    is AlterTable -> tableName
    is AlterColumnType -> tableName
    is SetNotNull -> tableName
    is DropNotNull -> tableName
    is SetDefault -> tableName
    is DropDefault -> tableName
    is RenameTable -> tableName
    is RenameColumn -> tableName
    is AddConstraint -> tableName
    is DropConstraint -> tableName
}

private fun Operation.toTestDsl(): String = when (this) {
    is CreateTable -> "create(${tableName}(${columns.joinToString(",") { it.name }}))"
    is AlterTable -> {
        val added = addedColumns.map { "+${it.name}" }
        val dropped = droppedColumns.map { "-$it" }
        val allChanges = (added + dropped).joinToString(",")
        "alter(${tableName}($allChanges))"
    }
    is AlterColumnType -> "alterType(${tableName}(${columnName}:${newType}))"
    is SetNotNull -> "setNotNull(${tableName}(${columnName}))"
    is DropNotNull -> "dropNotNull(${tableName}(${columnName}))"
    is SetDefault -> "setDefault(${tableName}(${columnName}=${defaultValue}))"
    is DropDefault -> "dropDefault(${tableName}(${columnName}))"
    is RenameTable -> "renameTable(${tableName}->${newTableName})"
    is RenameColumn -> "renameColumn(${tableName}(${columnName}->${newColumnName}))"
    is AddConstraint -> "addConstraint(${tableName}(${constraintName}:${constraintType}))"
    is DropConstraint -> "dropConstraint(${tableName}(${constraintName}))"
}
