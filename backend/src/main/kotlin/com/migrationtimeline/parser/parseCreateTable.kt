package com.migrationtimeline.parser

import com.migrationtimeline.models.Column
import com.migrationtimeline.models.CreateTable
import net.sf.jsqlparser.statement.create.table.CreateTable as JsqlCreateTable

fun parseCreateTable(migrationId: String, createTable: JsqlCreateTable): CreateTable {
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

    // JSQLParser puts PARTITION BY in tableOptionsStrings as [PARTITION, BY, LIST, (col)]
    val isPartitioned = createTable.tableOptionsStrings
        ?.any { it.uppercase() == "PARTITION" } ?: false

    return CreateTable(
        migrationId = migrationId,
        tableName = tableName,
        columns = columns,
        isPartitioned = isPartitioned
    )
}
