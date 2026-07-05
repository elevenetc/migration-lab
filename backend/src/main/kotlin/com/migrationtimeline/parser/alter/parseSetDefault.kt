package com.migrationtimeline.parser.alter

import com.migrationtimeline.models.SetDefault
import net.sf.jsqlparser.statement.alter.Alter
import net.sf.jsqlparser.statement.alter.AlterOperation

fun parseSetDefault(migrationId: String, tableName: String, alter: Alter): List<SetDefault> {
    return alter.alterExpressions
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
}
