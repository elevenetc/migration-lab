package runtime

import (
	"reflect"
	"testing"

	"migration-timeline/backend/internal/models"
)

func snapshot(fileNode uint32, tuplesRead int64) relationSnapshot {
	return relationSnapshot{Name: "accounts", FileNode: fileNode, Rows: 1000, TuplesRead: tuplesRead}
}

// relfilenode changes if and only if the table was rewritten, so it settles the
// question whatever the tuple counters say.
func TestObserveReportsARewriteFromTheChangedFileNode(t *testing.T) {
	before := map[uint32]relationSnapshot{1: snapshot(500, 0)}
	after := map[uint32]relationSnapshot{1: snapshot(700, 2000)}

	seen := observe(before, after, true)

	if seen.Class != models.TableRewrite {
		t.Errorf("class = %q, want %q", seen.Class, models.TableRewrite)
	}
	if !reflect.DeepEqual(seen.Rewritten, []string{"accounts"}) {
		t.Errorf("rewritten = %v, want [accounts]", seen.Rewritten)
	}
	if seen.TuplesRead != 2000 {
		t.Errorf("tuples read = %d, want 2000", seen.TuplesRead)
	}
}

func TestObserveReportsAScanFromRowsRead(t *testing.T) {
	before := map[uint32]relationSnapshot{1: snapshot(500, 0)}
	after := map[uint32]relationSnapshot{1: snapshot(500, 1000)}

	seen := observe(before, after, true)

	if seen.Class != models.DataScanning {
		t.Errorf("class = %q, want %q", seen.Class, models.DataScanning)
	}
	if len(seen.Rewritten) != 0 {
		t.Errorf("expected nothing rewritten, got %v", seen.Rewritten)
	}
}

func TestObserveReportsMetadataOnlyWhenNothingWasTouched(t *testing.T) {
	before := map[uint32]relationSnapshot{1: snapshot(500, 0)}
	after := map[uint32]relationSnapshot{1: snapshot(500, 0)}

	if seen := observe(before, after, true); seen.Class != models.MetadataOnly {
		t.Errorf("class = %q, want %q", seen.Class, models.MetadataOnly)
	}
}

// Outside a transaction the tuple counters read back as zero, which is
// indistinguishable from a statement that read nothing — so only a rewrite is
// decidable there.
func TestObserveWithoutATransactionOnlyDecidesARewrite(t *testing.T) {
	unchanged := observe(
		map[uint32]relationSnapshot{1: snapshot(500, 0)},
		map[uint32]relationSnapshot{1: snapshot(500, 0)},
		false)
	if unchanged.Class != "" {
		t.Errorf("class = %q, want it left unknown", unchanged.Class)
	}

	rewritten := observe(
		map[uint32]relationSnapshot{1: snapshot(500, 0)},
		map[uint32]relationSnapshot{1: snapshot(700, 0)},
		false)
	if rewritten.Class != models.TableRewrite {
		t.Errorf("class = %q, want %q", rewritten.Class, models.TableRewrite)
	}
}

// A rewrite builds a transient heap and drops the old relation; only relations
// present in both snapshots can be compared.
func TestObserveIgnoresRelationsOnlyOneSnapshotHas(t *testing.T) {
	before := map[uint32]relationSnapshot{1: snapshot(500, 0)}
	after := map[uint32]relationSnapshot{
		1: snapshot(500, 0),
		2: {Name: "accounts_new", FileNode: 900},
	}

	if seen := observe(before, after, true); seen.Class != models.MetadataOnly {
		t.Errorf("class = %q, want %q", seen.Class, models.MetadataOnly)
	}
}
