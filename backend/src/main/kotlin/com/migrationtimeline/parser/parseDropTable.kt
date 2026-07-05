package com.migrationtimeline.parser

import com.migrationtimeline.models.DropTable
import net.sf.jsqlparser.statement.drop.Drop

fun parseDropTable(migrationId: String, drop: Drop): DropTable? {
    if (drop.type?.uppercase() != "TABLE") return null
    val tableName = drop.name?.name ?: return null
    return DropTable(migrationId = migrationId, tableName = tableName)
}
