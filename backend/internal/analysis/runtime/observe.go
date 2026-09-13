package runtime

import (
	"sort"

	"migration-lab/backend/internal/models"
)

// observation is what the database was seen doing while one statement ran.
type observation struct {
	// Class is empty when the signals do not settle the question.
	Class     models.PerformanceClass
	Rewritten []string
	// TuplesRead is summed over every relation of the schema.
	TuplesRead int64
}

// observe derives the performance class of a statement from two snapshots of the
// schema. Every signal is a count, so the answer does not move with the hardware
// the run happened on:
//
//  1. a relation's relfilenode changed if and only if it was rewritten
//  2. otherwise, rows read means the statement scanned data
//  3. otherwise it only touched the catalog
//
// countersUsable is false when the migration could not be wrapped in a
// transaction: the tuple counters are transaction-local, so outside one they read
// back as zero and cannot tell a scan from a catalog change. A rewrite is still
// decidable there, since relfilenode is committed state.
func observe(before, after map[uint32]relationSnapshot, countersUsable bool) observation {
	result := observation{Rewritten: []string{}}

	for oid, was := range before {
		is, present := after[oid]
		if !present {
			continue
		}
		if was.FileNode != is.FileNode {
			result.Rewritten = append(result.Rewritten, is.Name)
		}
		result.TuplesRead += is.TuplesRead - was.TuplesRead
	}
	sort.Strings(result.Rewritten)

	switch {
	case len(result.Rewritten) > 0:
		result.Class = models.TableRewrite
	case !countersUsable:
		result.Class = ""
	case result.TuplesRead > 0:
		result.Class = models.DataScanning
	default:
		result.Class = models.MetadataOnly
	}
	return result
}
