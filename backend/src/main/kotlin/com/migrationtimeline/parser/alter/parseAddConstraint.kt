package com.migrationtimeline.parser.alter

import com.migrationtimeline.models.AddConstraint
import net.sf.jsqlparser.statement.alter.Alter
import net.sf.jsqlparser.statement.alter.AlterOperation

fun parseAddConstraint(migrationId: String, tableName: String, alter: Alter): List<AddConstraint> {
    return alter.alterExpressions
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
}
