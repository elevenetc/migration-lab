package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func gitOutput(repository string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", append([]string{"--literal-pathspecs", "-C", repository}, args...)...)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}
