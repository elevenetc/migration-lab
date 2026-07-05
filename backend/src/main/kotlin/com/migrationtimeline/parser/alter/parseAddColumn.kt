package com.migrationtimeline.parser.alter

import com.migrationtimeline.models.AddColumn
import com.migrationtimeline.models.Column
import net.sf.jsqlparser.statement.alter.Alter
import net.sf.jsqlparser.statement.alter.AlterOperation

fun parseAddColumn(migrationId: String, tableName: String, alter: Alter): List<AddColumn> {
    return alter.alterExpressions
        ?.filter { it.operation == AlterOperation.ADD && it.colDataTypeList != null }
        ?.flatMap { expr ->
            expr.colDataTypeList.map { colDataType ->
                AddColumn(
                    migrationId = migrationId,
                    tableName = tableName,
                    column = Column(
                        name = colDataType.columnName,
                        type = colDataType.colDataType.toString().replace(" ", ""),
                        constraints = emptyList()
                    )
                )
            }
        } ?: emptyList()
}
