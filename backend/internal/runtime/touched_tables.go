package runtime

import "migration-timeline/backend/internal/models"

// TouchedTables returns the tables the migration's operations name, in first
// appearance order. A partitioned parent named here stands for its partitions,
// which the catalog resolves once the prior migrations have been applied.
func TouchedTables(migration *models.Migration) []string {
	var tables []string
	seen := map[string]bool{}

	add := func(table string) {
		if table == "" || seen[table] {
			return
		}
		seen[table] = true
		tables = append(tables, table)
	}

	for _, statement := range migration.Statements {
		for _, operation := range statement.Operations {
			add(operationTable(operation))
		}
	}
	return tables
}

// operationTable is the table an operation reads or rewrites; a table the
// operation itself creates or drops holds nothing worth seeding.
func operationTable(operation models.Operation) string {
	switch op := operation.(type) {
	case models.CreateTable, models.DropTable:
		return ""
	case models.AddColumn:
		return op.TableName
	case models.DropColumn:
		return op.TableName
	case models.AlterColumnType:
		return op.TableName
	case models.SetNotNull:
		return op.TableName
	case models.DropNotNull:
		return op.TableName
	case models.SetDefault:
		return op.TableName
	case models.DropDefault:
		return op.TableName
	case models.RenameTable:
		return op.TableName
	case models.RenameColumn:
		return op.TableName
	case models.AddConstraint:
		return op.TableName
	case models.DropConstraint:
		return op.TableName
	default:
		return ""
	}
}
