# Supported SQL

SQL statements that can be parsed (backend) and rendered (frontend).

## Supported

- `CREATE TABLE`
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
- `ADD CONSTRAINT` (PRIMARY KEY, UNIQUE, FOREIGN KEY, CHECK, EXCLUDE)

## Not Supported

- Indicate lock: `ACCESS SHARE`, `ROW EXCLUSIVE`, `ACCESS EXCLUSIVE`