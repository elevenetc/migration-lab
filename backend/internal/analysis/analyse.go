package analysis

import "migration-timeline/backend/internal/models"

type analysisContext struct {
	partitionedTables map[string]bool
}

func Analyse(migrations []*models.Migration) *models.AnalysisResult {
	partitionedTables := make(map[string]bool)

	for _, m := range migrations {
		for _, stmt := range m.Statements {
			for _, op := range stmt.Operations {
				if ct, ok := op.(models.CreateTable); ok && ct.IsPartitioned {
					partitionedTables[ct.TableName] = true
				}
			}
		}
	}

	ctx := &analysisContext{partitionedTables: partitionedTables}
	var warnings []models.Warning

	for _, m := range migrations {
		for _, stmt := range m.Statements {
			if w := analyzeStatement(stmt, m.ID, ctx); w != nil {
				warnings = append(warnings, w)
			}
		}
	}

	return &models.AnalysisResult{
		Migrations: migrations,
		Warnings:   warnings,
	}
}

func analyzeStatement(stmt models.Statement, migrationID string, ctx *analysisContext) models.Warning {
	return detectAccessExclusiveLock(stmt, migrationID, ctx)
}
