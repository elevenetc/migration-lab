package server

import (
	"migration-timeline/backend/internal/analysis/static"
	"migration-timeline/backend/internal/models"
)

func toResponse(migrations []*models.Migration) *models.MigrationTimelineResponse {
	migrationMap := make(map[string]*models.Migration, len(migrations))
	createTableMap := make(map[string]models.CreateTable)

	for _, m := range migrations {
		migrationMap[m.ID] = m
		for _, stmt := range m.Statements {
			for _, op := range stmt.Operations {
				if ct, ok := op.(models.CreateTable); ok {
					createTableMap[ct.TableName] = ct
				}
			}
		}
	}

	analysisResult := static.Analyse(migrations)

	return &models.MigrationTimelineResponse{
		Timeline:       migrations,
		Map:            migrationMap,
		CreateTableMap: createTableMap,
		Analysis:       analysisResult,
	}
}
