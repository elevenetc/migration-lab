package ci

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func runCLI(ctx context.Context, binary string, args []string, output, name string) (code int, err error) {
	stdout, err := os.Create(filepath.Join(output, name+".json"))
	if err != nil {
		return 1, err
	}
	defer func() {
		err = errors.Join(err, stdout.Close())
	}()
	stderr, err := os.Create(filepath.Join(output, name+".stderr.log"))
	if err != nil {
		return 1, err
	}
	defer func() {
		err = errors.Join(err, stderr.Close())
	}()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err = cmd.Run()
	code = 0
	var exitError *exec.ExitError
	switch {
	case ctx.Err() != nil:
		if _, err := fmt.Fprintln(stderr, "\nMigration Lab command timed out or was cancelled."); err != nil {
			return 1, err
		}
		code = 124
	case errors.As(err, &exitError):
		code = exitError.ExitCode()
	case err != nil:
		return 1, err
	}
	if _, err := stderr.Seek(0, io.SeekStart); err != nil {
		return 1, err
	}
	if _, err := io.Copy(os.Stderr, stderr); err != nil {
		return 1, err
	}
	return code, nil
}
