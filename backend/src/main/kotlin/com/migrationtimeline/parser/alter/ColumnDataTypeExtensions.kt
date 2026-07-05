package com.migrationtimeline.parser.alter

import net.sf.jsqlparser.statement.alter.AlterExpression

fun AlterExpression.ColumnDataType.isSetNotNull(): Boolean {
    val colDataType = this.colDataType?.toString()?.uppercase() ?: return false
    return colDataType == "SET"
}

fun AlterExpression.ColumnDataType.isDropNotNull(): Boolean {
    val specs = this.columnSpecs ?: return false
    if (specs.isEmpty()) return false
    return specs[0].uppercase() == "DROP"
}

fun AlterExpression.ColumnDataType.hasNotNullConstraint(): Boolean {
    val specs = this.columnSpecs ?: return false
    val upper = specs.map { it.uppercase() }
    val notIndex = upper.indexOf("NOT")
    val nullIndex = upper.indexOf("NULL")
    return notIndex != -1 && nullIndex == notIndex + 1
}

fun AlterExpression.ColumnDataType.isSetDefault(): Boolean {
    val colDataType = this.colDataType?.toString()?.uppercase() ?: return false
    return colDataType == "SET"
}

fun AlterExpression.ColumnDataType.isDropDefault(): Boolean {
    val specs = this.columnSpecs ?: return false
    if (specs.isEmpty()) return false
    return specs[0].uppercase() == "DROP"
}

fun AlterExpression.ColumnDataType.hasDefaultConstraint(): Boolean {
    val specs = this.columnSpecs ?: return false
    return specs.any { it.uppercase() == "DEFAULT" }
}

fun AlterExpression.ColumnDataType.extractDefaultValue(): String {
    val specs = this.columnSpecs ?: return ""
    val upper = specs.map { it.uppercase() }
    val defaultIndex = upper.indexOf("DEFAULT")
    if (defaultIndex == -1 || defaultIndex + 1 >= specs.size) return ""
    return specs.drop(defaultIndex + 1).joinToString(" ")
}
