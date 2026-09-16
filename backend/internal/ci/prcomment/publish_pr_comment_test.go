package prcomment

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPublishCreatesThenUpdatesOneBotCommentAcrossPages(t *testing.T) {
	run := testCommentRun()
	var saved *githubComment
	creates, updates, pageTwo := 0, 0, 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" || r.Header.Get("X-GitHub-Api-Version") != "2022-11-28" {
			t.Error("missing GitHub request headers")
		}
		switch r.Method + " " + r.URL.Path {
		case "GET /repos/owner/repo/issues/7/comments":
			if r.URL.Query().Get("page") == "1" {
				comments := make([]githubComment, 100)
				// A human copying our marker must not become an update target.
				comments[0] = botComment(11, RenderPRComment(Summary{}, run))
				comments[0].User.Type = "User"
				comments[1] = botComment(12, RenderPRComment(Summary{}, run))
				comments[1].User.Login = "other-bot[bot]"
				writeGitHubFixture(t, w, comments)
			} else {
				pageTwo++
				comments := []githubComment{}
				if saved != nil {
					comments = append(comments, *saved)
				}
				writeGitHubFixture(t, w, comments)
			}
		case "GET /repos/owner/repo/pulls/7":
			writeGitHubFixture(t, w, map[string]any{"state": "open", "head": map[string]string{"sha": run.Head}})
		case "POST /repos/owner/repo/issues/7/comments", "PATCH /repos/owner/repo/issues/comments/42":
			var payload struct{ Body string }
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Error(err)
			}
			comment := botComment(42, payload.Body)
			saved = &comment
			if r.Method == "POST" {
				creates++
			} else {
				updates++
			}
			writeGitHubFixture(t, w, saved)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	api := GitHubAPI{URL: server.URL, Token: "test-token", Client: server.Client()}
	for attempt := int64(1); attempt <= 2; attempt++ {
		run.Attempt = attempt
		body := RenderPRComment(Summary{}, run)
		published, err := PublishPRComment(context.Background(), api, run, body)
		if err != nil || !published || saved == nil || saved.Body != body {
			t.Fatalf("could not publish attempt %d: %v, %v", attempt, published, err)
		}
	}
	if creates != 1 || updates != 1 || pageTwo != 2 {
		t.Fatalf("expected one created comment followed by an update across pages: creates=%d updates=%d pageTwo=%d", creates, updates, pageTwo)
	}
}

func TestPublishSkipsSupersededRunsAndClosedPRs(t *testing.T) {
	for _, tc := range []struct {
		name    string
		head    string
		state   string
		runID   int64
		attempt int64
	}{
		{name: "new commit", head: "new-head", state: "open", runID: 99, attempt: 1},
		{name: "closed PR", state: "closed", runID: 99, attempt: 1},
		{name: "newer run", state: "open", runID: 101, attempt: 1},
		{name: "newer attempt", state: "open", runID: 100, attempt: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run := testCommentRun()
			previous := run
			previous.RunID, previous.Attempt = tc.runID, tc.attempt
			head := tc.head
			if head == "" {
				head = run.Head
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("stale publisher must not write: %s", r.Method)
				}
				if strings.HasSuffix(r.URL.Path, "/comments") {
					writeGitHubFixture(t, w, []githubComment{botComment(42, RenderPRComment(Summary{}, previous))})
				} else {
					writeGitHubFixture(t, w, map[string]any{"state": tc.state, "head": map[string]string{"sha": head}})
				}
			}))
			defer server.Close()
			published, err := PublishPRComment(context.Background(), GitHubAPI{URL: server.URL, Client: server.Client()}, run, "new body")
			if err != nil || published {
				t.Fatalf("expected skipped update: %v, %v", published, err)
			}
		})
	}
}

func TestPublishReportsAPIFailureWithoutCreatingDuplicate(t *testing.T) {
	for _, status := range []int{http.StatusForbidden, http.StatusTooManyRequests, http.StatusInternalServerError} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				if r.Method != "GET" {
					t.Error("failed lookup must not fall back to creating a comment")
				}
				w.WriteHeader(status)
			}))
			defer server.Close()
			published, err := PublishPRComment(context.Background(), GitHubAPI{URL: server.URL, Client: server.Client()}, testCommentRun(), "body")
			if published || err == nil || !strings.Contains(err.Error(), fmt.Sprint(status)) || requests != 1 {
				t.Fatalf("expected an actionable API error: published=%v error=%v requests=%d", published, err, requests)
			}
		})
	}
}

func botComment(id int64, body string) githubComment {
	comment := githubComment{ID: id, Body: body}
	comment.User.Type = "Bot"
	comment.User.Login = "github-actions[bot]"
	return comment
}

func writeGitHubFixture(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Error(err)
	}
}
