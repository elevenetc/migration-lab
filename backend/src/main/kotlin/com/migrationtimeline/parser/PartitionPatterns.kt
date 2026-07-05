package com.migrationtimeline.parser

// JSQLParser doesn't support PostgreSQL PARTITION OF syntax
val PARTITION_OF_REGEX = Regex(
    """CREATE\s+TABLE\s+(\w+)\s+PARTITION\s+OF\s+(\w+)""",
    RegexOption.IGNORE_CASE
)

// JSQLParser doesn't support PARTITION BY RANGE (only LIST works)
val PARTITION_BY_RANGE_REGEX = Regex(
    """CREATE\s+TABLE\s+(\w+)\s*\(([^)]+(?:\([^)]*\)[^)]*)*)\)\s*PARTITION\s+BY\s+RANGE""",
    RegexOption.IGNORE_CASE
)
