package com.migrationtimeline.parser

import com.migrationtimeline.models.Column

fun parseColumnsFromSql(columnsSql: String): List<Column> {
    return columnsSql.split(",")
        .map { it.trim() }
        .filter { it.isNotBlank() }
        .mapNotNull { colDef ->
            val parts = colDef.split(Regex("\\s+"), limit = 2)
            if (parts.isNotEmpty()) {
                Column(
                    name = parts[0],
                    type = parts.getOrElse(1) { "UNKNOWN" }.split(Regex("\\s+"))[0],
                    constraints = emptyList()
                )
            } else null
        }
}
