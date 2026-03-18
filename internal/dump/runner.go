package dump

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Runner executes pg_dump with progress tracking.
type Runner struct {
	BinaryPath string
}

// NewRunner creates a Runner for the given pg_dump binary path.
func NewRunner(binaryPath string) *Runner {
	return &Runner{BinaryPath: binaryPath}
}

// Run starts pg_dump with the given config and returns a channel of progress events.
// totalObjects is passed to ParseProgress for percentage calculation.
// The caller should range over the returned channel until it is closed.
func (r *Runner) Run(ctx context.Context, cfg Config, totalObjects int) (<-chan ProgressEvent, error) {
	args := BuildArgs(cfg)
	cmd := exec.CommandContext(ctx, r.BinaryPath, args...)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+cfg.Password)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start pg_dump: %w", err)
	}

	// Tee stderr: feed ParseProgress while capturing tail for error reporting.
	var stderrBuf bytes.Buffer
	tee := io.TeeReader(stderr, &stderrBuf)
	progress := ParseProgress(tee, totalObjects)

	// Forward progress events, then wait for process exit.
	out := make(chan ProgressEvent)
	go func() {
		defer close(out)
		for ev := range progress {
			out <- ev
		}
		if err := cmd.Wait(); err != nil {
			out <- ProgressEvent{
				Err: fmt.Errorf("pg_dump failed: %w\n%s", err, stderrBuf.String()),
			}
			return
		}
		// ParseProgress already sent Done on EOF; if process exited 0
		// and progress already emitted Done, nothing more to send.
	}()

	return out, nil
}
