package com.migrationtimeline.parser.alter

import com.migrationtimeline.models.SetNotNull
import net.sf.jsqlparser.statement.alter.Alter
import net.sf.jsqlparser.statement.alter.AlterOperation

fun parseSetNotNull(migrationId: String, tableName: String, alter: Alter): List<SetNotNull> {
    return alter.alterExpressions
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
}
