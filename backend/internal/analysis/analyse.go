package analysis

import "migration-timeline/backend/internal/models"

type analysisContext struct {
	partitionedTables map[string]bool
}

func Analyse(migrations []*models.Migration) *models.AnalysisResult {
	partitionedTables := make(map[string]bool)

	for _, m := range migrations {
		for _, op := range m.Operations {
			if ct, ok := op.(models.CreateTable); ok && ct.IsPartitioned {
				partitionedTables[ct.TableName] = true
			}
		}
	}

	ctx := &analysisContext{partitionedTables: partitionedTables}
	var warnings []models.Warning

	for _, m := range migrations {
		for _, op := range m.Operations {
			if w := analyzeOperation(op, ctx); w != nil {
				warnings = append(warnings, w)
			}
		}
	}

	return &models.AnalysisResult{
		Migrations: migrations,
		Warnings:   warnings,
	}
}

func analyzeOperation(op models.Operation, ctx *analysisContext) models.Warning {
	return detectAccessExclusiveLock(op, ctx)
}
