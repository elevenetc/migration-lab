package models

type MigrationTimelineResponse struct {
	Timeline       []*Migration           `json:"timeline"`
	Map            map[string]*Migration  `json:"map"`
	CreateTableMap map[string]CreateTable `json:"createTableMap"`
	Analysis       *AnalysisResult        `json:"analysis"`
}
