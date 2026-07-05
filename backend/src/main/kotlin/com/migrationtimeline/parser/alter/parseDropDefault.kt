package com.migrationtimeline.parser.alter

import com.migrationtimeline.models.DropDefault
import net.sf.jsqlparser.statement.alter.Alter
import net.sf.jsqlparser.statement.alter.AlterOperation

fun parseDropDefault(migrationId: String, tableName: String, alter: Alter): List<DropDefault> {
    return alter.alterExpressions
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
}
