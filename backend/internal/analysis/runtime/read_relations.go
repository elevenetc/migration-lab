package runtime

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// relationSnapshot is what a relation looked like at one point of the run: which
// file backs it, how many rows it is known to hold, and how many rows this
// transaction has read from it.
type relationSnapshot struct {
	Name string
	// FileNode is pg_class.relfilenode, which changes if and only if the relation
	// was rewritten. It is 0 for a partitioned parent, which has no storage.
	FileNode uint32
	// Rows is pg_class.reltuples as ANALYZE last left it.
	Rows int64
	// TuplesRead is pg_stat_get_xact_tuples_returned: rows read by the calling
	// backend in its current transaction, and so zero outside one.
	TuplesRead int64
}

// Every relation of the public schema is read rather than only the seeded ones,
// so a rewrite of a partition the migration never named is still visible.
const relationsQuery = `
SELECT c.oid,
       c.relname,
       c.relfilenode,
       c.reltuples::bigint,
       pg_stat_get_xact_tuples_returned(c.oid)
FROM pg_class c
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = 'public' AND c.relkind IN ('r', 'p')`

// readRelations snapshots the public schema of the database.
//
// It has to run on the connection executing the migration: a relfilenode swapped
// by an uncommitted ALTER TABLE is invisible to every other backend, and the
// transaction-local tuple counters belong to the calling backend alone.
func readRelations(ctx context.Context, conn *pgx.Conn) (map[uint32]relationSnapshot, error) {
	rows, err := conn.Query(ctx, relationsQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to read relations: %w", err)
	}
	defer rows.Close()

	snapshot := map[uint32]relationSnapshot{}
	for rows.Next() {
		var oid uint32
		var relation relationSnapshot
		if err := rows.Scan(&oid, &relation.Name, &relation.FileNode, &relation.Rows, &relation.TuplesRead); err != nil {
			return nil, fmt.Errorf("failed to scan relation: %w", err)
		}
		snapshot[oid] = relation
	}
	return snapshot, rows.Err()
}
