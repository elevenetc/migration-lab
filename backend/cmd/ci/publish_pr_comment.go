package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type githubComment struct {
	ID   int64  `json:"id"`
	Body string `json:"body"`
	User struct {
		Login string `json:"login"`
		Type  string `json:"type"`
	} `json:"user"`
}

// The workflow serializes publishers for a PR. The head and run checks also
// prevent superseded runs and reruns from replacing a newer result.
func publishPRComment(ctx context.Context, api githubAPI, run commentRun, body string) (bool, error) {
	root := "/repos/" + run.Repository
	commentsPath := fmt.Sprintf("%s/issues/%d/comments", root, run.Number)
	var existing *githubComment
	for page := 1; ; page++ {
		var comments []githubComment
		if err := githubRequest(ctx, api, http.MethodGet, fmt.Sprintf("%s?per_page=100&page=%d", commentsPath, page), nil, &comments); err != nil {
			return false, err
		}
		for _, comment := range comments {
			if comment.User.Type == "Bot" && comment.User.Login == "github-actions[bot]" && strings.HasPrefix(comment.Body, commentMarker+"\n") {
				existing = &comment
				break
			}
		}
		if existing != nil || len(comments) < 100 {
			break
		}
	}
	if existing != nil {
		var previousRun, previousAttempt int64
		metadata := strings.SplitN(existing.Body, "\n", 3)[1]
		if _, err := fmt.Sscanf(metadata, "<!-- migration-lab-run: %d %d -->", &previousRun, &previousAttempt); err == nil {
			if previousRun > run.RunID || (previousRun == run.RunID && previousAttempt > run.Attempt) {
				return false, nil
			}
		}
	}
	var current struct {
		State string `json:"state"`
		Head  struct {
			SHA string `json:"sha"`
		} `json:"head"`
	}
	if err := githubRequest(ctx, api, http.MethodGet, fmt.Sprintf("%s/pulls/%d", root, run.Number), nil, &current); err != nil {
		return false, err
	}
	if current.State != "open" || current.Head.SHA != run.Head {
		return false, nil
	}
	method, path := http.MethodPost, commentsPath
	if existing != nil {
		method = http.MethodPatch
		path = fmt.Sprintf("%s/issues/comments/%d", root, existing.ID)
	}
	err := githubRequest(ctx, api, method, path, struct {
		Body string `json:"body"`
	}{Body: body}, nil)
	return err == nil, err
}
