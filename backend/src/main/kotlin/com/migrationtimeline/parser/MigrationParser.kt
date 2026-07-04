package com.migrationtimeline.parser

import com.migrationtimeline.models.AlterTable
import com.migrationtimeline.models.Column
import com.migrationtimeline.models.CreateTable
import com.migrationtimeline.models.Migration
import net.sf.jsqlparser.parser.CCJSqlParserUtil
import net.sf.jsqlparser.statement.alter.Alter
import net.sf.jsqlparser.statement.create.table.CreateTable as JsqlCreateTable

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
        val operations = statements.mapNotNull { statement ->
            when (statement) {
                is JsqlCreateTable -> parseCreateTable(id, statement)
                is Alter -> parseAlterTable(id, statement)
                else -> null
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

    private fun parseAlterTable(migrationId: String, alter: Alter): AlterTable {
        val tableName = alter.table.name
        val addedColumns = alter.alterExpressions
            ?.filter { it.colDataTypeList != null }
            ?.flatMap { expr ->
                expr.colDataTypeList.map { colDataType ->
                    Column(
                        name = colDataType.columnName,
                        type = colDataType.colDataType.toString().replace(" ", ""),
                        constraints = emptyList()
                    )
                }
            } ?: emptyList()

        return AlterTable(
            migrationId = migrationId,
            tableName = tableName,
            addedColumns = addedColumns
        )
    }
}
