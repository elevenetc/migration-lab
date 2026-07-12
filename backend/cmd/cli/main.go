package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"migration-timeline/backend/internal/analysis"
	"migration-timeline/backend/internal/loader"
	"migration-timeline/backend/internal/models"
	"migration-timeline/backend/internal/parser"
	"migration-timeline/backend/internal/report"
	"migration-timeline/backend/internal/runner"

	"github.com/spf13/cobra"
)

var runFlag bool
var reportFlag string

var rootCmd = &cobra.Command{
	Use:   "migration-timeline <sql-or-dir>",
	Short: "Parse SQL migrations and output JSON AST",
	Long: `Parse and analyze SQL migrations or run them against PostgreSQL.

By default, outputs static analysis as JSON.
If argument is a directory, processes all *.sql files in it.
If argument is a SQL string, processes it as a single migration.

Examples:
  migration-timeline /path/to/migrations
  migration-timeline "CREATE TABLE users (id INT);"
  migration-timeline --run /path/to/migrations
  migration-timeline --report /path/to/migrations
  migration-timeline --report=output.html /path/to/migrations`,
	Args:          cobra.ExactArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		arg := args[0]

		info, err := os.Stat(arg)
		if err == nil && info.IsDir() {
			return processDirectory(arg)
		}

		return processSQL(arg)
	},
}

func init() {
	rootCmd.Flags().BoolVar(&runFlag, "run", false, "Run migrations against a PostgreSQL container (requires Docker)")
	rootCmd.Flags().StringVar(&reportFlag, "report", "", "Generate self-contained HTML report (default: report.html)")
}

func processDirectory(dir string) error {
	infos, err := loader.LoadMigrationInfosFromDir(dir)
	if err != nil {
		return err
	}
	return processMigrationInfos(infos)
}

func processSQL(sql string) error {
	statements := splitStatements(sql)
	var infos []models.MigrationInfo
	for i, stmt := range statements {
		infos = append(infos, models.MigrationInfo{
			ID:        fmt.Sprintf("statement_%d", i+1),
			SQL:       stmt,
			Timestamp: int64(i + 1),
		})
	}
	return processMigrationInfos(infos)
}

func splitStatements(sql string) []string {
	var statements []string
	for _, stmt := range strings.Split(sql, ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt != "" {
			statements = append(statements, stmt+";")
		}
	}
	return statements
}

func processMigrationInfos(infos []models.MigrationInfo) error {
	if reportFlag != "" {
		return generateReport(infos)
	}

	var result cliResult

	if runFlag {
		runResult := runner.RunMigrations(context.Background(), infos)
		result.RunResult = &runResult
	}

	migrations, err := parseMigrations(infos)
	if err != nil {
		return err
	}
	result.AnalysisResult = analysis.Analyse(migrations)

	return printJSON(result)
}

func parseMigrations(infos []models.MigrationInfo) ([]*models.Migration, error) {
	var migrations []*models.Migration
	for _, info := range infos {
		m, err := parser.ParseMigration(info.ID, info.SQL, info.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", info.ID, err)
		}
		migrations = append(migrations, m)
	}
	return migrations, nil
}

func generateReport(infos []models.MigrationInfo) error {
	migrations, err := parseMigrations(infos)
	if err != nil {
		return err
	}

	response := toTimelineResponse(migrations)

	outputPath := reportFlag
	if outputPath == "" {
		outputPath = "report.html"
	}

	if err := report.GenerateReport(response, outputPath); err != nil {
		return err
	}

	fmt.Printf("Report generated: %s\n", outputPath)
	return nil
}

func toTimelineResponse(migrations []*models.Migration) *models.MigrationTimelineResponse {
	migrationMap := make(map[string]*models.Migration, len(migrations))
	createTableMap := make(map[string]models.CreateTable)

	for _, m := range migrations {
		migrationMap[m.ID] = m
		for _, op := range m.Operations {
			if ct, ok := op.(models.CreateTable); ok {
				createTableMap[ct.TableName] = ct
			}
		}
	}

	analysisResult := analysis.Analyse(migrations)

	return &models.MigrationTimelineResponse{
		Timeline:       migrations,
		Map:            migrationMap,
		CreateTableMap: createTableMap,
		Analysis:       analysisResult,
	}
}

type cliResult struct {
	AnalysisResult *models.AnalysisResult      `json:"analysisResult,omitempty"`
	RunResult      *models.RunMigrationsResult `json:"runResult,omitempty"`
}

func printJSON(v interface{}) error {
	output, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(output))
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
