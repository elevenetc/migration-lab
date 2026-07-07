package report

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"migration-timeline/backend/internal/models"
)

func GenerateReport(response *models.MigrationTimelineResponse, outputPath string) error {
	jsonData, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	cssContent, err := loadAsset(".css")
	if err != nil {
		return fmt.Errorf("failed to load CSS: %w", err)
	}

	jsContent, err := loadAsset(".js")
	if err != nil {
		return fmt.Errorf("failed to load JS: %w", err)
	}

	html := buildHTML(string(jsonData), cssContent, jsContent)

	if err := os.WriteFile(outputPath, []byte(html), 0644); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	return nil
}

func loadAsset(ext string) (string, error) {
	var content string

	err := fs.WalkDir(FrontendAssets, "dist/assets", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ext {
			data, err := FrontendAssets.ReadFile(path)
			if err != nil {
				return err
			}
			content = string(data)
			return fs.SkipAll
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	if content == "" {
		return "", fmt.Errorf("no %s file found in embedded assets", ext)
	}

	return content, nil
}

func buildHTML(jsonData, css, js string) string {
	var sb strings.Builder

	sb.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Migration Timeline Report</title>
  <style>
`)
	sb.WriteString(css)
	sb.WriteString(`
  </style>
</head>
<body>
  <div id="root"></div>
  <script>
    window.__MIGRATION_DATA__ = `)
	sb.WriteString(jsonData)
	sb.WriteString(`;
  </script>
  <script type="module">
`)
	sb.WriteString(js)
	sb.WriteString(`
  </script>
</body>
</html>`)

	return sb.String()
}
