package ci

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"migration-lab/backend/internal/loader"
)

type workflowEvent struct {
	Number      int `json:"number"`
	PullRequest struct {
		Head struct {
			SHA string `json:"sha"`
		} `json:"head"`
		Base struct {
			SHA string `json:"sha"`
		} `json:"base"`
	} `json:"pull_request"`
}

type migrationPlan struct {
	Event           string            `json:"event"`
	Head            string            `json:"head"`
	Base            string            `json:"base,omitempty"`
	AnalyzerSHA     string            `json:"analyzerSha"`
	Directory       string            `json:"directory"`
	Targets         []string          `json:"targets"`
	ExistingChanges []migrationChange `json:"existingChanges"`
}

func planMigrations(repository, directory, eventName string, event workflowEvent) (migrationPlan, error) {
	plan := migrationPlan{Event: eventName, Targets: []string{}, ExistingChanges: []migrationChange{}}
	if eventName != "pull_request" && eventName != "workflow_dispatch" {
		return plan, fmt.Errorf("supported triggers are pull_request and workflow_dispatch")
	}
	if filepath.IsAbs(directory) || strings.TrimSpace(directory) == "" {
		return plan, fmt.Errorf("migrations-directory must be a nonempty repository-relative path")
	}
	repository, err := filepath.Abs(repository)
	if err != nil {
		return plan, err
	}
	repository, err = filepath.EvalSymlinks(repository)
	if err != nil {
		return plan, err
	}
	dir, err := filepath.EvalSymlinks(filepath.Join(repository, directory))
	if err != nil {
		return plan, err
	}
	relative, err := filepath.Rel(repository, dir)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return plan, fmt.Errorf("migrations-directory must stay inside the repository")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return plan, err
	}
	for _, entry := range entries {
		if entry.Type()&os.ModeSymlink != 0 && filepath.Ext(entry.Name()) == ".sql" && !strings.HasPrefix(entry.Name(), ".") {
			return plan, fmt.Errorf("SQL migration symlinks are not supported")
		}
	}
	migrations, err := loader.LoadMigrationInfosFromDir(dir)
	if err != nil {
		return plan, err
	}
	if len(migrations) == 0 {
		return plan, fmt.Errorf("migrations-directory contains no SQL migration files")
	}
	head, err := gitOutput(repository, "rev-parse", "HEAD")
	if err != nil {
		return plan, err
	}
	plan.Head = strings.TrimSpace(string(head))
	plan.Directory = filepath.ToSlash(relative)
	if eventName == "workflow_dispatch" {
		plan.Targets = []string{migrations[len(migrations)-1].ID}
		return plan, nil
	}
	if plan.Head != event.PullRequest.Head.SHA {
		return plan, fmt.Errorf("checked-out commit does not match the PR head")
	}
	plan.Base = event.PullRequest.Base.SHA
	if plan.Base == "" {
		return plan, fmt.Errorf("PR base SHA is missing")
	}
	diff, err := gitOutput(repository, "diff", "--name-status", "-z", "--find-renames", plan.Base+"..."+plan.Head, "--", plan.Directory)
	if err != nil {
		return plan, err
	}
	added, existing, err := changedMigrations(diff, plan.Directory)
	if err != nil {
		return plan, err
	}
	plan.ExistingChanges = existing
	// The shared loader provides the same Flyway ordering as the CLI.
	for _, migration := range migrations {
		if added[migration.ID] {
			plan.Targets = append(plan.Targets, migration.ID)
			delete(added, migration.ID)
		}
	}
	if len(added) > 0 {
		return plan, fmt.Errorf("some selected migrations were not loaded")
	}
	return plan, nil
}
