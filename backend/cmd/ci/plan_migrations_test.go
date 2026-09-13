package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestPlanWholePRAndNumericOrder(t *testing.T) {
	repo, base := testRepository(t)
	writeTestFile(t, repo, "db/V2__with space.sql", "ALTER TABLE users ADD email TEXT;")
	commitTestRepo(t, repo)
	writeTestFile(t, repo, "db/V2__with space.sql", "ALTER TABLE users ADD email VARCHAR(100);")
	writeTestFile(t, repo, "db/V10__last.sql", "SELECT 10;")
	writeTestFile(t, repo, "db/V1__create.sql", "CREATE TABLE users (id BIGINT);")
	writeTestFile(t, repo, "db/nested/V3__ignored.sql", "SELECT 1;")
	writeTestFile(t, repo, "db/.hidden.sql", "SELECT 1;")
	writeTestFile(t, repo, "db/notes.txt", "notes")
	commitTestRepo(t, repo)
	plan := testPRPlan(t, repo, base, "db")
	if !slices.Equal(plan.Targets, []string{"V2__with space.sql", "V10__last.sql"}) {
		t.Fatalf("unexpected targets: %v", plan.Targets)
	}
	if !reflect.DeepEqual(plan.ExistingChanges, []migrationChange{{Status: "M", Paths: []string{"db/V1__create.sql"}}}) {
		t.Fatalf("unexpected existing changes: %+v", plan.ExistingChanges)
	}
}

func TestPlanExcludesChangesOnBaseBranch(t *testing.T) {
	repo, _ := testRepository(t)
	testGit(t, repo, "checkout", "-b", "feature")
	writeTestFile(t, repo, "db/V2__pr.sql", "SELECT 2;")
	commitTestRepo(t, repo)
	testGit(t, repo, "checkout", "main")
	writeTestFile(t, repo, "db/V3__base.sql", "SELECT 3;")
	base := commitTestRepo(t, repo)
	testGit(t, repo, "checkout", "feature")
	if plan := testPRPlan(t, repo, base, "db"); !slices.Equal(plan.Targets, []string{"V2__pr.sql"}) {
		t.Fatalf("unexpected targets: %v", plan.Targets)
	}
}

func TestPlanRecordsRenamesAndDeletions(t *testing.T) {
	repo, _ := testRepository(t)
	writeTestFile(t, repo, "db/V2__delete.sql", "SELECT 2;")
	base := commitTestRepo(t, repo)
	testGit(t, repo, "mv", "db/V1__create.sql", "db/V1__renamed.sql")
	testGit(t, repo, "rm", "db/V2__delete.sql")
	commitTestRepo(t, repo)
	plan := testPRPlan(t, repo, base, "db")
	if len(plan.Targets) != 0 || len(plan.ExistingChanges) != 2 {
		t.Fatalf("unexpected plan: %+v", plan)
	}
	var statuses []byte
	for _, change := range plan.ExistingChanges {
		statuses = append(statuses, change.Status[0])
	}
	if !slices.Contains(statuses, byte('R')) || !slices.Contains(statuses, byte('D')) {
		t.Fatalf("expected rename and deletion, got %q", statuses)
	}
}

func TestPlanManualSelectsLatest(t *testing.T) {
	repo, _ := testRepository(t)
	writeTestFile(t, repo, "db/V20260410_2__early.sql", "SELECT 2;")
	writeTestFile(t, repo, "db/V20260410_10__late.sql", "SELECT 10;")
	commitTestRepo(t, repo)
	plan, err := planMigrations(repo, "db", "workflow_dispatch", workflowEvent{})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(plan.Targets, []string{"V20260410_10__late.sql"}) {
		t.Fatalf("unexpected targets: %v", plan.Targets)
	}
}

func TestPlanRejectsInvalidPaths(t *testing.T) {
	repo, _ := testRepository(t)
	if err := os.Mkdir(filepath.Join(repo, "empty"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, directory := range []string{"", filepath.Join(repo, "db"), "..", "missing", "db/V1__create.sql", "empty"} {
		t.Run(directory, func(t *testing.T) {
			if _, err := planMigrations(repo, directory, "workflow_dispatch", workflowEvent{}); err == nil {
				t.Fatal("expected an invalid directory error")
			}
		})
	}
	if err := os.Symlink(filepath.Join(repo, "db/V1__create.sql"), filepath.Join(repo, "db/V2__linked.sql")); err != nil {
		t.Fatal(err)
	}
	if _, err := planMigrations(repo, "db", "workflow_dispatch", workflowEvent{}); err == nil || !strings.Contains(err.Error(), "symlinks") {
		t.Fatalf("expected a symlink error, got %v", err)
	}
}

func TestPlanHandlesRootAndLiteralDirectoryNames(t *testing.T) {
	repo, base := testRepository(t)
	writeTestFile(t, repo, "V2__root.sql", "SELECT 2;")
	writeTestFile(t, repo, "db[1]/V1__create.sql", "SELECT 1;")
	commitTestRepo(t, repo)
	if plan := testPRPlan(t, repo, base, "."); !slices.Equal(plan.Targets, []string{"V2__root.sql"}) {
		t.Fatalf("unexpected root targets: %v", plan.Targets)
	}
	if plan := testPRPlan(t, repo, base, "db[1]"); !slices.Equal(plan.Targets, []string{"V1__create.sql"}) {
		t.Fatalf("unexpected literal-path targets: %v", plan.Targets)
	}
}

func TestPlanRejectsWrongHeadAndUnsupportedEvents(t *testing.T) {
	repo, _ := testRepository(t)
	if _, err := planMigrations(repo, "db", "pull_request", workflowEvent{}); err == nil || !strings.Contains(err.Error(), "head") {
		t.Fatalf("expected head mismatch, got %v", err)
	}
	if _, err := planMigrations(repo, "db", "pull_request_target", workflowEvent{}); err == nil || !strings.Contains(err.Error(), "triggers") {
		t.Fatalf("expected unsupported event, got %v", err)
	}
}

func TestPlanCommandWritesGitHubOutputAndProvenance(t *testing.T) {
	for _, addMigration := range []bool{false, true} {
		t.Run(map[bool]string{false: "no migrations", true: "new migration"}[addMigration], func(t *testing.T) {
			repo, base := testRepository(t)
			writeTestFile(t, repo, "README.md", "Documentation")
			if addMigration {
				writeTestFile(t, repo, "db/V2__new.sql", "SELECT 2;")
			}
			head := commitTestRepo(t, repo)
			var event workflowEvent
			event.PullRequest.Base.SHA = base
			event.PullRequest.Head.SHA = head
			output := t.TempDir()
			eventFile := filepath.Join(output, "event.json")
			if err := writeJSON(eventFile, event); err != nil {
				t.Fatal(err)
			}
			githubOutput := filepath.Join(output, "github-output")
			t.Setenv("GITHUB_EVENT_NAME", "pull_request")
			t.Setenv("GITHUB_EVENT_PATH", eventFile)
			t.Setenv("GITHUB_OUTPUT", githubOutput)
			code, err := runCommand(context.Background(), []string{"plan"}, configuration{repository: repo, source: repo, directory: "db", output: output})
			if err != nil || code != 0 {
				t.Fatalf("plan command failed: code %d, %v", code, err)
			}
			var plan migrationPlan
			if err := readJSON(filepath.Join(output, "plan.json"), &plan); err != nil {
				t.Fatal(err)
			}
			if plan.AnalyzerSHA != head || plan.Base != base || plan.Head != head || (len(plan.Targets) > 0) != addMigration {
				t.Fatalf("unexpected plan: %+v", plan)
			}
			want := "has-targets=false\n"
			if addMigration {
				want = "has-targets=true\n"
			}
			if got, err := os.ReadFile(githubOutput); err != nil || string(got) != want {
				t.Fatalf("unexpected GitHub output: %q, %v", got, err)
			}
		})
	}
}

func testRepository(t *testing.T) (string, string) {
	t.Helper()
	repository := t.TempDir()
	testGit(t, repository, "init", "-b", "main")
	writeTestFile(t, repository, "db/V1__create.sql", "CREATE TABLE users (id INT);\n")
	return repository, commitTestRepo(t, repository)
}

func testGit(t *testing.T, repository string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", repository, "-c", "user.name=Test", "-c", "user.email=test@example.invalid", "-c", "commit.gpgsign=false"}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func commitTestRepo(t *testing.T, repository string) string {
	t.Helper()
	testGit(t, repository, "add", ".")
	testGit(t, repository, "commit", "-m", "test")
	return testGit(t, repository, "rev-parse", "HEAD")
}

func writeTestFile(t *testing.T, repository, name, content string) {
	t.Helper()
	path := filepath.Join(repository, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func testPRPlan(t *testing.T, repo, base, directory string) migrationPlan {
	t.Helper()
	var event workflowEvent
	event.PullRequest.Head.SHA = testGit(t, repo, "rev-parse", "HEAD")
	event.PullRequest.Base.SHA = base
	plan, err := planMigrations(repo, directory, "pull_request", event)
	if err != nil {
		t.Fatal(err)
	}
	return plan
}
