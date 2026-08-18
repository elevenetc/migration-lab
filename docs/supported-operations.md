# Supported Operations

SQL statements that can be parsed (backend) and rendered (frontend).

## Supported

- `CREATE TABLE`
- `CREATE TABLE ... PARTITION BY` (partitioned parent table)
- `CREATE TABLE ... PARTITION OF` (child partition)
- `ADD COLUMN`
- `DROP COLUMN`
- `DROP CONSTRAINT`
- `DROP TABLE`
- `DROP NOT NULL`
- `DROP DEFAULT`
- `ALTER COLUMN TYPE`
- `SET NOT NULL`
- `SET DEFAULT`
- `RENAME TABLE`
- `RENAME COLUMN`
- `ADD CONSTRAINT` (PRIMARY KEY, UNIQUE, FOREIGN KEY, CHECK, EXCLUDE), including `NOT VALID`

## Captured modifiers

- `ADD CONSTRAINT ... NOT VALID` sets `AddConstraint.notValid`
- A column's `DEFAULT` expression is deparsed into `Column.defaultExpr`; `constraints` carries the
  `DEFAULT` token whether or not the expression could be deparsed

Both feed [performance classification](supported-performance-classes.md).

