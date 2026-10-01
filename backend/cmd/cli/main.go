package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"migration-lab/backend/internal/analysis/runtime"
	"migration-lab/backend/internal/analysis/static"
	"migration-lab/backend/internal/loader"
	"migration-lab/backend/internal/models"
	"migration-lab/backend/internal/parser"
	"migration-lab/backend/internal/report"
	"migration-lab/backend/internal/runner"

	"github.com/spf13/cobra"
)

var runFlag bool
var reportFlag string
var runtimeFlag bool
var targetMigrationIDFlag string
var rowsFlag int64
var deadlineFlag int64

var rootCmd = &cobra.Command{
	Use:   "migration-lab <sql-or-dir>",
	Short: "Analyze PostgreSQL migration performance and impact",
	Long: `Migration Lab analyzes PostgreSQL migration performance and impact.

By default, outputs static analysis as JSON.
Use --runtime to measure execution, locking, and retry behavior of every migration against a seeded
PostgreSQL container, or --runtime --migration <filename> to measure a single one.
If argument is a directory, processes all *.sql files in it.
If argument is a SQL string, processes each of its statements as a migration of its own.

Examples:
  migration-lab /path/to/migrations
  migration-lab "CREATE TABLE users (id INT);"
  migration-lab --run /path/to/migrations
  migration-lab --runtime /path/to/migrations
  migration-lab --runtime --migration V42__add_index.sql /path/to/migrations
  migration-lab --runtime --rows 1000000 --deadline-ms 30000 /path/to/migrations
  migration-lab --report /path/to/migrations
  migration-lab --report=output.html /path/to/migrations`,
	Args:          cobra.ExactArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Flags().Changed("migration") && !runtimeFlag {
			return fmt.Errorf("--migration requires --runtime")
		}
		return nil
	},
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
	rootCmd.Flags().BoolVar(&runtimeFlag, "runtime", false, "Measure every migration in turn against a seeded PostgreSQL container (requires Docker)")
	rootCmd.Flags().StringVar(&targetMigrationIDFlag, "migration", "", "Measure only this migration with --runtime, and exit non-zero unless it completes")
	rootCmd.MarkFlagsMutuallyExclusive("runtime", "report")
	rootCmd.Flags().Int64Var(&rowsFlag, "rows", 0, "Rows to top every touched table up to (default 1000000)")
	rootCmd.Flags().Int64Var(&deadlineFlag, "deadline-ms", 0, "Deadline standing for the pod grace period (default 5000)")
}

func processDirectory(dir string) error {
	infos, err := loader.LoadMigrationInfosFromDir(dir)
	if err != nil {
		return err
	}
	return processMigrationInfos(infos)
}

func processSQL(sql string) error {
	statements, err := parser.SplitStatements(sql)
	if err != nil {
		return err
	}
	infos := make([]models.MigrationInfo, len(statements))
	for i, statement := range statements {
		infos[i] = models.MigrationInfo{
			ID:        fmt.Sprintf("statement_%d", i+1),
			SQL:       statement,
			Timestamp: int64(i + 1),
		}
	}
	return processMigrationInfos(infos)
}

func processMigrationInfos(infos []models.MigrationInfo) error {
	if targetMigrationIDFlag != "" {
		var err error
		infos, err = selectRuntimeMigrations(infos, targetMigrationIDFlag)
		if err != nil {
			return err
		}
	}
	if reportFlag != "" {
		return generateReport(infos)
	}

	var result cliResult

	if runFlag {
		runResult := runner.RunMigrations(context.Background(), infos)
		result.RunResult = &runResult
	}

	if runtimeFlag {
		runtimeResult, err := analyseRuntime(infos, targetMigrationIDFlag)
		if err != nil {
			return err
		}
		result.RuntimeResult = runtimeResult
	}

	migrations, err := parseMigrations(infos)
	if err != nil {
		return err
	}
	result.StaticResult = static.Analyse(migrations)

	if err := printJSON(result); err != nil {
		return err
	}

	return runtimeFailure(result.RuntimeResult, targetMigrationIDFlag)
}

// analyseRuntime measures the target migration, or every migration when targetMigrationID
// is empty.
func analyseRuntime(infos []models.MigrationInfo, targetMigrationID string) (*models.RuntimeAnalysisResult, error) {
	if len(infos) == 0 {
		return nil, fmt.Errorf("no migrations to analyze")
	}

	result, err := runtime.Analyse(context.Background(), runtime.Request{
		Migrations:        infos,
		TargetMigrationID: targetMigrationID,
		Rows:              rowsFlag,
		Deadline:          time.Duration(deadlineFlag) * time.Millisecond,
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func parseMigrations(infos []models.MigrationInfo) ([]*models.Migration, error) {
	migrations, err := parser.ParseTimeline(infos)
	if err != nil {
		return nil, fmt.Errorf("failed to parse migrations: %w", err)
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

	_, err = fmt.Printf("Report generated: %s\n", outputPath)
	return err
}

func toTimelineResponse(migrations []*models.Migration) *models.MigrationTimelineResponse {
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

type cliResult struct {
	StaticResult  *models.StaticAnalysisResult  `json:"staticResult,omitempty"`
	RunResult     *models.RunMigrationsResult   `json:"runResult,omitempty"`
	RuntimeResult *models.RuntimeAnalysisResult `json:"runtimeResult,omitempty"`
}

func printJSON(v interface{}) error {
	output, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	_, err = fmt.Println(string(output))
	return err
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.New(os.Stderr, "error: ", 0).Print(err)
		os.Exit(1)
	}
}
