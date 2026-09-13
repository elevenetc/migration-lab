package main

import (
	"fmt"
	"path"
	"slices"
	"strings"
)

type migrationChange struct {
	Status string   `json:"status"`
	Paths  []string `json:"paths"`
}

// changedMigrations parses Git's NUL-delimited diff, including both rename paths.
// Git paths always use forward slashes, independent of the runner's platform.
func changedMigrations(diff []byte, directory string) (map[string]bool, []migrationChange, error) {
	added := make(map[string]bool)
	var existing []migrationChange
	eligible := func(name string) bool {
		return path.Dir(name) == directory && path.Ext(name) == ".sql" && !strings.HasPrefix(path.Base(name), ".")
	}
	fields := strings.Split(string(diff), "\x00")
	for i := 0; i < len(fields) && fields[i] != ""; {
		status := fields[i]
		count := 1
		if status[0] == 'R' || status[0] == 'C' {
			count = 2
		}
		if i+count >= len(fields) {
			return nil, nil, fmt.Errorf("incomplete Git diff record")
		}
		paths := fields[i+1 : i+count+1]
		if slices.Contains(paths, "") {
			return nil, nil, fmt.Errorf("empty path in Git diff record")
		}
		if status == "A" && eligible(paths[0]) {
			added[path.Base(paths[0])] = true
		} else {
			if slices.ContainsFunc(paths, eligible) {
				existing = append(existing, migrationChange{Status: status, Paths: paths})
			}
		}
		i += count + 1
	}
	return added, existing, nil
}
