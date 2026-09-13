package runtime

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// CatalogColumn is one column as the live database describes it.
type CatalogColumn struct {
	Name      string
	DataType  string
	MaxLength int
	// AutoFilled marks a column the database fills itself — an identity column, a
	// generated column, or one defaulting to a sequence — which a seed leaves alone.
	AutoFilled bool
}

// CatalogTable is one table as the live database describes it, after the
// migrations before the analyzed one have been applied.
type CatalogTable struct {
	Name        string
	Partitioned bool
	// Parent is the partitioned table this one is a partition of, "" otherwise.
	Parent string
	// Bound is pg_get_expr(relpartbound): the values the partition admits.
	Bound string
	// PartitionKeyDef is pg_get_partkeydef, set when the table is partitioned.
	PartitionKeyDef string
	Columns         []CatalogColumn
}

const tablesQuery = `
SELECT c.relname,
       c.relkind = 'p'                                    AS partitioned,
       coalesce(p.relname, '')                            AS parent,
       coalesce(pg_get_expr(c.relpartbound, c.oid), '')   AS bound,
       CASE WHEN c.relkind = 'p' THEN pg_get_partkeydef(c.oid) ELSE '' END AS partkeydef
FROM pg_class c
JOIN pg_namespace n ON n.oid = c.relnamespace
LEFT JOIN pg_inherits i ON i.inhrelid = c.oid
LEFT JOIN pg_class p ON p.oid = i.inhparent
WHERE n.nspname = 'public' AND c.relkind IN ('r', 'p')`

// A column the database fills itself is left out of a seed: an identity column,
// a generated one, or one defaulting to a sequence.
const columnsQuery = `
SELECT table_name,
       column_name,
       data_type,
       coalesce(character_maximum_length, 0),
       is_identity = 'YES'
         OR is_generated = 'ALWAYS'
         OR coalesce(column_default, '') LIKE 'nextval%'
FROM information_schema.columns
WHERE table_schema = 'public'
ORDER BY table_name, ordinal_position`

// ReadCatalog reads the public schema of a live database: which tables exist, how
// they are partitioned, and what columns a seed can fill.
func ReadCatalog(ctx context.Context, conn *pgx.Conn) (map[string]CatalogTable, error) {
	catalog, err := readTables(ctx, conn)
	if err != nil {
		return nil, err
	}
	return withColumns(ctx, conn, catalog)
}

func readTables(ctx context.Context, conn *pgx.Conn) (map[string]CatalogTable, error) {
	rows, err := conn.Query(ctx, tablesQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to read tables: %w", err)
	}
	defer rows.Close()

	catalog := map[string]CatalogTable{}
	for rows.Next() {
		var table CatalogTable
		if err := rows.Scan(&table.Name, &table.Partitioned, &table.Parent, &table.Bound, &table.PartitionKeyDef); err != nil {
			return nil, fmt.Errorf("failed to scan table: %w", err)
		}
		catalog[table.Name] = table
	}
	return catalog, rows.Err()
}

func withColumns(ctx context.Context, conn *pgx.Conn, catalog map[string]CatalogTable) (map[string]CatalogTable, error) {
	rows, err := conn.Query(ctx, columnsQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to read columns: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var table string
		var column CatalogColumn
		if err := rows.Scan(&table, &column.Name, &column.DataType, &column.MaxLength, &column.AutoFilled); err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}
		if info, ok := catalog[table]; ok {
			info.Columns = append(info.Columns, column)
			catalog[table] = info
		}
	}
	return catalog, rows.Err()
}
