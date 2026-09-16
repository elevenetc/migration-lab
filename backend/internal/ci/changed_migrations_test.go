package ci

import (
	"slices"
	"testing"
)

func TestChangedMigrationsPreservesNULDelimitedPaths(t *testing.T) {
	diff := []byte("A\x00db/V2__line\nbreak.sql\x00R100\x00db/V1__old.sql\x00db/V1__new.sql\x00M\x00db/nested/ignored.sql\x00")
	added, existing, err := changedMigrations(diff, "db")
	if err != nil || !added["V2__line\nbreak.sql"] || len(added) != 1 {
		t.Fatalf("unexpected additions: %v, %v", added, err)
	}
	if len(existing) != 1 || !slices.Equal(existing[0].Paths, []string{"db/V1__old.sql", "db/V1__new.sql"}) {
		t.Fatalf("unexpected existing changes: %+v", existing)
	}
}

func TestChangedMigrationsRejectsTruncatedRename(t *testing.T) {
	if _, _, err := changedMigrations([]byte("R100\x00db/old.sql\x00"), "db"); err == nil {
		t.Fatal("expected incomplete rename to fail")
	}
}
