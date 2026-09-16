package ci

import (
	"context"
	"testing"
)

func TestCommentCommandSkipsManualRuns(t *testing.T) {
	code, err := RunCommand(context.Background(), []string{"comment"}, Config{Output: t.TempDir(), EventName: "workflow_dispatch"})
	if err != nil || code != 0 {
		t.Fatalf("manual runs must not need GitHub credentials: %d, %v", code, err)
	}
}
