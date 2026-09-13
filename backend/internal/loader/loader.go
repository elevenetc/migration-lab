package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"migration-lab/backend/internal/models"
	"migration-lab/backend/internal/parser"
)

// LoadMigrationInfosFromDir reads all *.sql files from dir into MigrationInfos,
// sorted by Flyway version, with a sequential Timestamp assigned per sorted
// position. It returns an error if dir doesn't exist or isn't a directory.
func LoadMigrationInfosFromDir(dir string) ([]models.MigrationInfo, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", dir)
	}

	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return nil, err
	}

	var infos []models.MigrationInfo
	for _, file := range files {
		if strings.HasPrefix(filepath.Base(file), ".") {
			continue
		}
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		infos = append(infos, models.MigrationInfo{
			ID:  filepath.Base(file),
			SQL: string(content),
		})
	}

	sort.SliceStable(infos, func(i, j int) bool {
		return parser.CompareVersions(infos[i].ID, infos[j].ID)
	})
	for i := range infos {
		infos[i].Timestamp = int64(i + 1)
	}

	return infos, nil
}
