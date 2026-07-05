package com.migrationtimeline.parser.alter

import com.migrationtimeline.models.DropNotNull
import net.sf.jsqlparser.statement.alter.Alter
import net.sf.jsqlparser.statement.alter.AlterOperation

fun parseDropNotNull(migrationId: String, tableName: String, alter: Alter): List<DropNotNull> {
    return alter.alterExpressions
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
}
