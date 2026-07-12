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

	// CSS is optional: the canvas frontend inlines its styles and emits no .css asset.
	cssContent := loadOptionalAsset(".css")

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

// loadOptionalAsset returns the first embedded asset with ext, or "" if none exists.
func loadOptionalAsset(ext string) string {
	content, err := loadAsset(ext)
	if err != nil {
		return ""
	}
	return content
}

func buildHTML(jsonData, css, js string) string {
	var sb strings.Builder

	sb.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Migration Timeline Report</title>
`)
	if css != "" {
		sb.WriteString("  <style>\n")
		sb.WriteString(css)
		sb.WriteString("\n  </style>\n")
	}
	sb.WriteString(`</head>
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
