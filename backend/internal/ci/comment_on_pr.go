package ci

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"migration-lab/backend/internal/ci/prcomment"
)

func commentOnPR(ctx context.Context, config Config) error {
	if config.EventName != "pull_request" {
		return nil
	}
	var event workflowEvent
	if err := readJSON(config.EventPath, &event); err != nil {
		return err
	}
	runID, runErr := strconv.ParseInt(config.GitHubRunID, 10, 64)
	attempt, attemptErr := strconv.ParseInt(config.GitHubRunAttempt, 10, 64)
	run := prcomment.Run{
		Repository: config.GitHubRepository, Number: event.Number,
		Head:  event.PullRequest.Head.SHA,
		RunID: runID, Attempt: attempt, Status: config.AnalysisStatus,
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
	token := config.GitHubToken
	if token == "" {
		return fmt.Errorf("GH_TOKEN with pull-requests: write permission is required")
	}
	body := prcomment.RenderPRComment(loadCommentSummary(config.Output, run.Head), run)
	if err := os.WriteFile(filepath.Join(config.Output, "comment.md"), []byte(body), 0o644); err != nil {
		return err
	}
	api := prcomment.GitHubAPI{
		URL: "https://api.github.com", Token: token,
		Client: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
	_, err := prcomment.PublishPRComment(ctx, api, run, body)
	return err
}
