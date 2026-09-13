package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

func commentOnPR(ctx context.Context, config configuration) error {
	if os.Getenv("GITHUB_EVENT_NAME") != "pull_request" {
		return nil
	}
	var event workflowEvent
	if err := readJSON(os.Getenv("GITHUB_EVENT_PATH"), &event); err != nil {
		return err
	}
	runID, runErr := strconv.ParseInt(os.Getenv("GITHUB_RUN_ID"), 10, 64)
	attempt, attemptErr := strconv.ParseInt(os.Getenv("GITHUB_RUN_ATTEMPT"), 10, 64)
	run := commentRun{
		Repository: os.Getenv("GITHUB_REPOSITORY"), Number: event.Number,
		Head: event.PullRequest.Head.SHA, Directory: config.directory,
		RunID: runID, Attempt: attempt, Status: os.Getenv("ANALYSIS_STATUS"),
	}
	if runErr != nil || attemptErr != nil || runID < 1 || attempt < 1 || run.Number < 1 ||
		!regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`).MatchString(run.Repository) ||
		!regexp.MustCompile(`^[a-f0-9]{40}$`).MatchString(run.Head) {
		return fmt.Errorf("valid GitHub repository, PR number, head SHA, run ID, and attempt are required")
	}
	switch run.Status {
	case "success", "failure", "cancelled", "skipped":
	default:
		return fmt.Errorf("ANALYSIS_STATUS must be a GitHub job result")
	}
	token := os.Getenv("GH_TOKEN")
	if token == "" {
		return fmt.Errorf("GH_TOKEN with pull-requests: write permission is required")
	}
	body := renderPRComment(loadCommentSummary(config.output, run.Head), run)
	if err := os.WriteFile(filepath.Join(config.output, "comment.md"), []byte(body), 0o644); err != nil {
		return err
	}
	api := githubAPI{
		URL: "https://api.github.com", Token: token,
		Client: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
	_, err := publishPRComment(ctx, api, run, body)
	return err
}
