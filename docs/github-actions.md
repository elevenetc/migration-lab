# GitHub runtime analysis

The reusable workflow at `.github/workflows/runtime-analysis.yml` runs Migration Lab on a
GitHub-hosted Ubuntu 24.04 runner. PostgreSQL is disposable; no server, persistent database,
or Marketplace installation is needed. It posts one summary comment on same-repository PRs and
updates that comment after subsequent runs.

## Set up another repository

Once the workflow is pushed to Migration Lab's GitHub repository, add this file to the repository
whose migrations you want to analyze:

```yaml
# .github/workflows/migrations.yml
name: Migration analysis

on:
  pull_request:
  workflow_dispatch:

permissions:
  contents: read
  pull-requests: write

jobs:
  analyze:
    name: Migration Lab
    uses: elevenetc/migration-lab/.github/workflows/runtime-analysis.yml@main
    with:
      migrations-directory: db/migrations
```

Change `db/migrations` to your repository-relative migration directory. `main` is convenient for
initial testing; replace it with a full commit SHA to pin the workflow. The analyzer source is
checked out at the workflow's own SHA, so pinning also pins the CLI. A tag or GitHub Release is
optional. The source repository must be readable by the caller's token; a public Migration Lab
repository supports this example without extra credentials. Organization Actions policies still apply.
The caller must grant `pull-requests: write`: reusable workflows cannot elevate the caller's
permissions. This also applies when upgrading an existing caller to the version with PR comments.

This workflow targets GitHub.com: it uses `job.workflow_repository` and `job.workflow_sha` to
identify its own source when called from another repository.

## What gets analyzed

- **Pull requests:** all added, non-hidden `*.sql` files directly in the configured directory,
  using the diff from the merge base to the PR head. Editing a file introduced earlier in the same
  PR analyzes it again. Existing migration edits, deletions, and detected renames are recorded in
  `plan.json` but are not runtime targets. Nested directories are not loaded, matching the CLI.
- **Manual runs:** the latest migration in Flyway version order on the selected branch. The
  caller workflow must exist on the default branch for GitHub's **Run workflow** button to appear.
- **No added migrations:** upload the selection plan and succeed without building the CLI or
  starting PostgreSQL. The Go CI command still builds to select the targets. A missing, empty, or
  invalid migration directory fails as a configuration error.

The checkout is the exact PR head, rather than GitHub's synthetic merge commit. Each target uses
the complete history preceding it on that head, including earlier migrations from the PR.
Updates on the base branch alone do not trigger a new run.

Targets are processed sequentially in the CLI's numeric Flyway order. Each runtime invocation
starts a fresh `postgres:16-alpine` container, rebuilds its predecessor schema, seeds eligible tables
to 1,000,000 rows, and measures the target with a 5,000 ms execution deadline. This is a fixed test
budget, not a claim about your deployment's grace period or production performance.

The complete history must first parse successfully. A runtime failure in one target does not stop
the workflow from attempting later targets; those may also fail if the earlier migration cannot be
applied as a predecessor. Each CLI command has a five-minute timeout covering setup, seeding, and
analysis. The entire job has a 30-minute timeout. New runs cancel older runs for the same PR and
directory.

The workflow sets up Go and builds `backend/cmd/ci` before selecting targets. Its `plan` command
uses the shared migration loader for numeric Flyway ordering; its `run` command invokes the CLI
and collects results. Its `comment` command publishes the PR summary from the collected results.
No Python interpreter or packages are required.

The first version builds the CLI from source when targets are found. That also builds the frontend
because the CLI's HTML report package embeds its assets, although this workflow only outputs JSON.
Go and npm dependencies are cached; databases are recreated for each target.

## PR summary

The job maintains one `github-actions[bot]` comment per PR, with a section for each migration:

```markdown
## ✅ Migration runtime analysis - `V2__add_account_status.sql`

- No runtime concerns found.
- **Performance class:** `METADATA_ONLY`
- [View full analysis](https://github.com/<owner>/<repo>/actions/runs/<run-id>/attempts/<attempt>)
```

An observed `TABLE_REWRITE` uses 🚨 in the heading; `DATA_SCANNING` and other results needing
attention use ⚠️. Their first bullet explains the recommendation:

| Recommendation            | When                                                                   |
|---------------------------|------------------------------------------------------------------------|
| Review before merging — … | Scans, rewrites, runtime risks, or incomplete performance observations |
| Do not merge — …          | Execution failure, exceeded deadline, or retry failure                 |
| Analysis incomplete — …   | Missing results, analysis failure, or inadequate seeding               |

Findings and diagnostics appear as additional bullets. **Performance class** is the most expensive
observed class; missing observations show **Unavailable**, and incomplete coverage is marked **partial**.
The final link opens the workflow run with complete results and diagnostics.

Comments update after subsequent runs, including failures or removal of all added migrations.
Use one migration-directory configuration per PR. Manual runs, fork PRs, and Dependabot PRs retain
checks and artifacts without comments; cancellation and job timeouts can prevent comment updates.

## Inspect a run

Open the consuming repository's **Actions** tab, select the workflow run, and inspect its steps.
Download the `migration-lab-<run-id>-<attempt>` artifact for:

| File                                   | Content                                                                                    |
|----------------------------------------|--------------------------------------------------------------------------------------------|
| `plan.json`                            | Caller head/base, analyzer SHA, selected directory and targets, existing migration changes |
| `ci-build.log`                         | Go CI command build diagnostics                                                            |
| `build.log`                            | CLI and frontend build diagnostics, when a build was needed                                |
| `docker-info.txt`                      | Docker environment information                                                             |
| `static-analysis.json` / `.stderr.log` | Static analysis of the migration history and parsing diagnostics                           |
| `results.json`                         | Each target, exit code, and its result/log filenames, in execution order                   |
| `<migration>.sql.json`                 | CLI JSON for each measured target, including `runtimeResult`                               |
| `<migration>.sql.stderr.log`           | Matching execution diagnostics                                                             |
| `automation-error.txt`                 | Selection or orchestration errors, when present                                            |

Per-migration files retain the complete SQL filename: analyzing `V2__add_account_status.sql`
produces `V2__add_account_status.sql.json` and `V2__add_account_status.sql.stderr.log`.
`results.json` lists these filenames in execution order.

Artifacts are retained for seven days. Upload runs after ordinary failures as well as success;
forced cancellation or the overall job timeout can interrupt result collection. Individual command
timeouts retain their partial output and use exit code `124`. SQL and database errors can appear in
these files. Logs are printed after each CLI invocation finishes.

The job fails if parsing, setup, or any runtime target fails or exceeds its deadline. Runtime
findings alone do not fail it: `COMPLETED` may still contain `EXCLUSIVE_LOCK_HELD`, table rewrites, or
`SEED_FAILED`. Inspect the result and its seeding coverage during validation.

## Local checks

```bash
just test-automation
just test
go run github.com/rhysd/actionlint/cmd/actionlint@latest -shellcheck= .github/workflows/runtime-analysis.yml
```

Automation tests live alongside `backend/cmd/ci` and run as part of `go test ./...`. They use
temporary Git repositories and a Go test subprocess as a stub CLI to check whole-PR diffs, numeric
ordering from the shared loader, manual selection, and artifact retention after failures.
Comment tests cover runtime summaries, missing and partial results, bounded and escaped output,
and a local fake GitHub API for create/update, pagination, stale runs, and API failures.
The Go CLI tests cover selecting a real migration and replaying its predecessors against PostgreSQL.
The actionlint configuration suppresses only its currently unknown `job.workflow_repository` and
`job.workflow_sha` properties; both are documented GitHub.com contexts.

## References

- [Reusable workflows](https://docs.github.com/en/actions/how-tos/reuse-automations/reuse-workflows)
- [Job context and reusable workflow source](https://docs.github.com/en/actions/reference/workflows-and-actions/contexts#job-context)
- [Runtime analysis](supported-runtime-analysis.md)
