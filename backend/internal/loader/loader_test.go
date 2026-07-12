package loader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMigrationInfosFromDir_Order(t *testing.T) {
	dir := t.TempDir()

	// Written in deliberately non-sorted order; same-date files differ only by
	// the trailing sequence, which must break the tie correctly.
	names := []string{
		"V20260410_3__c.sql",
		"V20260410_10__d.sql",
		"V20260410_1__a.sql",
		"V20250812_1__init.sql",
		"V20260410_2__b.sql",
	}
	for _, n := range names {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("SELECT 1;"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	infos, err := LoadMigrationInfosFromDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		"V20250812_1__init.sql",
		"V20260410_1__a.sql",
		"V20260410_2__b.sql",
		"V20260410_3__c.sql",
		"V20260410_10__d.sql",
	}
	if len(infos) != len(want) {
		t.Fatalf("got %d infos, want %d", len(infos), len(want))
	}
	for i, w := range want {
		if infos[i].ID != w {
			t.Errorf("position %d: got %s, want %s", i, infos[i].ID, w)
		}
		if infos[i].Timestamp != int64(i+1) {
			t.Errorf("%s: got timestamp %d, want %d", infos[i].ID, infos[i].Timestamp, i+1)
		}
	}
}
