// Package developmentsandboxrunner runs only inside an explicitly isolated code
// container. It provides a hard container lifetime, not host-side execution.
package developmentsandboxrunner

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
)

// Run must be PID 1 on Linux: exiting it ends the entire code PID namespace,
// including children that escaped the compiler's original process group.
func Run(seconds int, output io.Writer) int {
	if os.Getpid() != 1 || seconds < 1 || seconds > 120 {
		return 125
	}
	// This timer deliberately exits PID 1 even if an IO copy or child wait blocks.
	timer := time.AfterFunc(time.Duration(seconds)*time.Second, func() { os.Exit(124) })
	defer timer.Stop()
	command := exec.CommandContext(context.Background(), "/usr/local/go/bin/go", "test", "-json", "-count=1", "./...")
	command.Dir = "/workspace"
	command.Env = []string{"PATH=/usr/local/go/bin:/usr/bin:/bin", "HOME=/tmp", "TMPDIR=/tmp", "GOCACHE=/tmp/go-cache", "GOPATH=/tmp/go-path", "GOPROXY=off", "GOSUMDB=off", "GOTOOLCHAIN=local", "GOENV=off", "GOWORK=off", "CGO_ENABLED=0", "GOMAXPROCS=2"}
	bounded := &limitedOutput{writer: output, remaining: 65536}
	command.Stdout, command.Stderr = bounded, bounded
	err := command.Run()
	if bounded.exceeded {
		return 122
	}
	if err == nil {
		return 0
	}
	if exit, ok := err.(*exec.ExitError); ok {
		if exit.ExitCode() >= 0 {
			return exit.ExitCode()
		}
	}
	fmt.Fprintln(output, "sandbox_command_unavailable")
	return 125
}

type limitedOutput struct {
	mu        sync.Mutex
	writer    io.Writer
	remaining int
	exceeded  bool
}

func (w *limitedOutput) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	n := len(p)
	if n > w.remaining {
		w.exceeded = true
		p = p[:w.remaining]
	}
	written, err := w.writer.Write(p)
	w.remaining -= written
	if err != nil {
		return written, err
	}
	return n, nil
}
