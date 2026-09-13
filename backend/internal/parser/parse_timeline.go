package parser

import (
	"migration-timeline/backend/internal/models"
)

// ParseTimeline parses a whole timeline in order. Some operations cannot be
// described from their own statement alone — an ALTER COLUMN TYPE needs the type
// the column had before — so the migrations are parsed first and then resolved
// against the schema their predecessors built.
func ParseTimeline(infos []models.MigrationInfo) ([]*models.Migration, error) {
	migrations := make([]*models.Migration, 0, len(infos))
	for _, info := range infos {
		migration, err := ParseMigration(info.ID, info.SQL, info.Timestamp)
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, migration)
	}
	return withResolvedPreviousColumnTypes(migrations), nil
}

// columnTypes is table name -> column name -> declared type, as the timeline has
// built it up so far.
type columnTypes map[string]map[string]models.SQLType

// withResolvedPreviousColumnTypes replays the timeline to fill
// AlterColumnType.PreviousType, returning new migrations rather than editing the
// ones it was given. A column no earlier migration declared keeps a zero
// PreviousType, which the classifier reads as unknown.
func withResolvedPreviousColumnTypes(migrations []*models.Migration) []*models.Migration {
	types := columnTypes{}
	resolved := make([]*models.Migration, 0, len(migrations))

	for _, migration := range migrations {
		statements := make([]models.Statement, 0, len(migration.Statements))
		for _, statement := range migration.Statements {
			operations := make([]models.Operation, 0, len(statement.Operations))
			for _, operation := range statement.Operations {
				operations = append(operations, types.applyOperation(operation))
			}
			statement.Operations = operations
			statements = append(statements, statement)
		}
		next := *migration
		next.Statements = statements
		resolved = append(resolved, &next)
	}

	return resolved
}

// applyOperation records what the operation does to the tracked types and returns it,
// filled in from what was tracked before it.
func (types columnTypes) applyOperation(operation models.Operation) models.Operation {
	switch op := operation.(type) {
	case models.CreateTable:
		columns := map[string]models.SQLType{}
		for _, column := range op.Columns {
			columns[column.Name] = column.Type
		}
		types[op.TableName] = columns
	case models.AddColumn:
		types.set(op.TableName, op.Column.Name, op.Column.Type)
	case models.AlterColumnType:
		op.PreviousType = types.of(op.TableName, op.ColumnName)
		types.set(op.TableName, op.ColumnName, op.NewType)
		return op
	case models.RenameColumn:
		types.set(op.TableName, op.NewColumnName, types.of(op.TableName, op.ColumnName))
		delete(types[op.TableName], op.ColumnName)
	case models.DropColumn:
		delete(types[op.TableName], op.ColumnName)
	case models.RenameTable:
		types[op.NewTableName] = types[op.TableName]
		delete(types, op.TableName)
	case models.DropTable:
		delete(types, op.TableName)
	}
	return operation
}

func (types columnTypes) of(table, column string) models.SQLType {
	return types[table][column]
}

func (types columnTypes) set(table, column string, declared models.SQLType) {
	if types[table] == nil {
		types[table] = map[string]models.SQLType{}
	}
	types[table][column] = declared
}
