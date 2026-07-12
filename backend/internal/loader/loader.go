package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"migration-timeline/backend/internal/models"
	"migration-timeline/backend/internal/parser"
)

// LoadMigrationInfosFromDir reads all *.sql files from dir into MigrationInfos,
// sorted by timestamp. It returns an error if dir doesn't exist or isn't a directory.
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
			ID:        filepath.Base(file),
			SQL:       string(content),
			Timestamp: parser.ExtractTimestampFromFilename(file),
		})
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Timestamp < infos[j].Timestamp
	})

	return infos, nil
}
