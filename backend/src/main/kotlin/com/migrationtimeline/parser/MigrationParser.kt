package com.migrationtimeline.parser

import com.migrationtimeline.models.AddConstraint
import com.migrationtimeline.models.AlterColumnType
import com.migrationtimeline.models.AlterTable
import com.migrationtimeline.models.Column
import com.migrationtimeline.models.CreateTable
import com.migrationtimeline.models.DropColumn
import com.migrationtimeline.models.DropConstraint
import com.migrationtimeline.models.DropDefault
import com.migrationtimeline.models.DropNotNull
import com.migrationtimeline.models.DropTable
import com.migrationtimeline.models.Migration
import com.migrationtimeline.models.Operation
import com.migrationtimeline.models.RenameColumn
import com.migrationtimeline.models.RenameTable
import com.migrationtimeline.models.SetDefault
import com.migrationtimeline.models.SetNotNull
import net.sf.jsqlparser.parser.CCJSqlParserUtil
import net.sf.jsqlparser.statement.alter.Alter
import net.sf.jsqlparser.statement.alter.AlterExpression
import net.sf.jsqlparser.statement.alter.AlterOperation
import net.sf.jsqlparser.statement.create.table.CreateTable as JsqlCreateTable
import net.sf.jsqlparser.statement.drop.Drop

private fun AlterExpression.ColumnDataType.isSetNotNull(): Boolean {
    val colDataType = this.colDataType?.toString()?.uppercase() ?: return false
    return colDataType == "SET"
}

private fun AlterExpression.ColumnDataType.isDropNotNull(): Boolean {
    val specs = this.columnSpecs ?: return false
    if (specs.isEmpty()) return false
    return specs[0].uppercase() == "DROP"
}

private fun AlterExpression.ColumnDataType.hasNotNullConstraint(): Boolean {
    val specs = this.columnSpecs ?: return false
    val upper = specs.map { it.uppercase() }
    val notIndex = upper.indexOf("NOT")
    val nullIndex = upper.indexOf("NULL")
    return notIndex != -1 && nullIndex == notIndex + 1
}

private fun AlterExpression.ColumnDataType.isSetDefault(): Boolean {
    val colDataType = this.colDataType?.toString()?.uppercase() ?: return false
    return colDataType == "SET"
}

private fun AlterExpression.ColumnDataType.isDropDefault(): Boolean {
    val specs = this.columnSpecs ?: return false
    if (specs.isEmpty()) return false
    return specs[0].uppercase() == "DROP"
}

private fun AlterExpression.ColumnDataType.hasDefaultConstraint(): Boolean {
    val specs = this.columnSpecs ?: return false
    return specs.any { it.uppercase() == "DEFAULT" }
}

private fun AlterExpression.ColumnDataType.extractDefaultValue(): String {
    val specs = this.columnSpecs ?: return ""
    val upper = specs.map { it.uppercase() }
    val defaultIndex = upper.indexOf("DEFAULT")
    if (defaultIndex == -1 || defaultIndex + 1 >= specs.size) return ""
    return specs.drop(defaultIndex + 1).joinToString(" ")
}

object MigrationParser {

    @Suppress("DEPRECATION")
    fun parse(migration: Map<String, String>): List<Migration> {
        return migration.entries.mapIndexed { index, entry ->
            parse(entry.key, entry.value, index.toLong())
        }
    }

    @Suppress("DEPRECATION")
    fun parse(id: String, sql: String, timestamp: Long = 0L): Migration {
        val statements = CCJSqlParserUtil.parseStatements(sql).statements
        val operations = statements.flatMap { statement ->
            when (statement) {
                is JsqlCreateTable -> listOf(parseCreateTable(id, statement))
                is Alter -> parseAlter(id, statement)
                is Drop -> parseDropTable(id, statement)?.let { listOf(it) } ?: emptyList()
                else -> emptyList()
            }
        }
        return Migration(id = id, version = id, timestamp = timestamp, operations = operations)
    }

    private fun parseCreateTable(migrationId: String, createTable: JsqlCreateTable): CreateTable {
        val tableName = createTable.table.name
        val columns = createTable.columnDefinitions?.map { colDef ->
            val constraints = mutableListOf<String>()

            colDef.columnSpecs?.forEach { spec ->
                when (spec.uppercase()) {
                    "PRIMARY" -> {
                        val idx = colDef.columnSpecs.indexOf(spec)
                        if (idx + 1 < colDef.columnSpecs.size &&
                            colDef.columnSpecs[idx + 1].uppercase() == "KEY"
                        ) {
                            constraints.add("PRIMARY KEY")
                        }
                    }
                    "NOT" -> {
                        val idx = colDef.columnSpecs.indexOf(spec)
                        if (idx + 1 < colDef.columnSpecs.size &&
                            colDef.columnSpecs[idx + 1].uppercase() == "NULL"
                        ) {
                            constraints.add("NOT NULL")
                        }
                    }
                    "UNIQUE" -> constraints.add("UNIQUE")
                }
            }

            Column(
                name = colDef.columnName,
                type = colDef.colDataType.toString().replace(" ", ""),
                constraints = constraints
            )
        } ?: emptyList()

        return CreateTable(
            migrationId = migrationId,
            tableName = tableName,
            columns = columns
        )
    }

    private fun parseAlter(migrationId: String, alter: Alter): List<Operation> {
        val tableName = alter.table.name
        val operations = mutableListOf<Operation>()

        val addedColumns = alter.alterExpressions
            ?.filter { it.operation == AlterOperation.ADD && it.colDataTypeList != null }
            ?.flatMap { expr ->
                expr.colDataTypeList.map { colDataType ->
                    Column(
                        name = colDataType.columnName,
                        type = colDataType.colDataType.toString().replace(" ", ""),
                        constraints = emptyList()
                    )
                }
            } ?: emptyList()

        val dropColumns = alter.alterExpressions
            ?.filter { it.operation == AlterOperation.DROP && it.columnName != null && it.constraintName == null }
            ?.map { expr ->
                DropColumn(
                    migrationId = migrationId,
                    tableName = tableName,
                    columnName = expr.columnName
                )
            } ?: emptyList()

        operations.addAll(dropColumns)

        if (addedColumns.isNotEmpty()) {
            operations.add(
                AlterTable(
                    migrationId = migrationId,
                    tableName = tableName,
                    addedColumns = addedColumns
                )
            )
        }

        val alterColumnTypes = alter.alterExpressions
            ?.filter { it.operation == AlterOperation.ALTER && it.colDataTypeList != null }
            ?.flatMap { expr ->
                expr.colDataTypeList
                    .filter { !it.isSetNotNull() && !it.isDropNotNull() && !it.isSetDefault() && !it.isDropDefault() }
                    .map { colDataType ->
                        AlterColumnType(
                            migrationId = migrationId,
                            tableName = tableName,
                            columnName = colDataType.columnName,
                            newType = colDataType.colDataType.toString().replace(" ", "")
                        )
                    }
            } ?: emptyList()

        operations.addAll(alterColumnTypes)

        val setNotNulls = alter.alterExpressions
            ?.filter { it.operation == AlterOperation.ALTER && it.colDataTypeList != null }
            ?.flatMap { expr ->
                expr.colDataTypeList
                    .filter { it.isSetNotNull() && it.hasNotNullConstraint() }
                    .map { colDataType ->
                        SetNotNull(
                            migrationId = migrationId,
                            tableName = tableName,
                            columnName = colDataType.columnName
                        )
                    }
            } ?: emptyList()

        operations.addAll(setNotNulls)

        val dropNotNulls = alter.alterExpressions
            ?.filter { it.operation == AlterOperation.ALTER && it.colDataTypeList != null }
            ?.flatMap { expr ->
                expr.colDataTypeList
                    .filter { it.isDropNotNull() && it.hasNotNullConstraint() }
                    .map { colDataType ->
                        DropNotNull(
                            migrationId = migrationId,
                            tableName = tableName,
                            columnName = colDataType.columnName
                        )
                    }
            } ?: emptyList()

        operations.addAll(dropNotNulls)

        val setDefaults = alter.alterExpressions
            ?.filter { it.operation == AlterOperation.ALTER && it.colDataTypeList != null }
            ?.flatMap { expr ->
                expr.colDataTypeList
                    .filter { it.isSetDefault() && it.hasDefaultConstraint() }
                    .map { colDataType ->
                        SetDefault(
                            migrationId = migrationId,
                            tableName = tableName,
                            columnName = colDataType.columnName,
                            defaultValue = colDataType.extractDefaultValue()
                        )
                    }
            } ?: emptyList()

        operations.addAll(setDefaults)

        val dropDefaults = alter.alterExpressions
            ?.filter { it.operation == AlterOperation.ALTER && it.colDataTypeList != null }
            ?.flatMap { expr ->
                expr.colDataTypeList
                    .filter { it.isDropDefault() && it.hasDefaultConstraint() }
                    .map { colDataType ->
                        DropDefault(
                            migrationId = migrationId,
                            tableName = tableName,
                            columnName = colDataType.columnName
                        )
                    }
            } ?: emptyList()

        operations.addAll(dropDefaults)

        val renameTable = alter.alterExpressions
            ?.firstOrNull { it.operation == AlterOperation.RENAME_TABLE }
            ?.let { expr ->
                val newName = expr.newTableName ?: return@let null
                RenameTable(
                    migrationId = migrationId,
                    tableName = tableName,
                    newTableName = newName
                )
            }

        renameTable?.let { operations.add(it) }

        val renameColumns = alter.alterExpressions
            ?.filter { it.operation == AlterOperation.RENAME && it.columnName != null }
            ?.mapNotNull { expr ->
                val oldName = expr.columnName ?: return@mapNotNull null
                val newName = expr.columnOldName ?: return@mapNotNull null
                RenameColumn(
                    migrationId = migrationId,
                    tableName = tableName,
                    columnName = newName,
                    newColumnName = oldName
                )
            } ?: emptyList()

        operations.addAll(renameColumns)

        val addConstraints = alter.alterExpressions
            ?.filter { it.operation == AlterOperation.ADD && it.index != null && it.colDataTypeList == null }
            ?.mapNotNull { expr ->
                val index = expr.index ?: return@mapNotNull null
                val constraintName = index.name ?: return@mapNotNull null
                val constraintType = when {
                    index.type?.uppercase() == "PRIMARY KEY" -> "PRIMARY KEY"
                    index.type?.uppercase() == "UNIQUE" -> "UNIQUE"
                    index.type?.uppercase() == "FOREIGN KEY" -> "FOREIGN KEY"
                    index.type == null && index.toString().uppercase().contains("CHECK") -> "CHECK"
                    index.type == null && index.toString().uppercase().contains("EXCLUDE") -> "EXCLUDE"
                    else -> index.type?.uppercase() ?: "UNKNOWN"
                }
                AddConstraint(
                    migrationId = migrationId,
                    tableName = tableName,
                    constraintName = constraintName,
                    constraintType = constraintType
                )
            } ?: emptyList()

        operations.addAll(addConstraints)

        val dropConstraints = alter.alterExpressions
            ?.filter { it.operation == AlterOperation.DROP && it.constraintName != null }
            ?.map { expr ->
                DropConstraint(
                    migrationId = migrationId,
                    tableName = tableName,
                    constraintName = expr.constraintName
                )
            } ?: emptyList()

        operations.addAll(dropConstraints)

        return operations
    }

    private fun parseDropTable(migrationId: String, drop: Drop): DropTable? {
        if (drop.type?.uppercase() != "TABLE") return null
        val tableName = drop.name?.name ?: return null
        return DropTable(migrationId = migrationId, tableName = tableName)
    }
}
