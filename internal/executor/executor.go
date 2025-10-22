package executor

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Executor handles command execution
type Executor struct {
	workDir string
	timeout time.Duration
}

// NewExecutor creates a new command executor
func NewExecutor(workDir string) *Executor {
	return &Executor{
		workDir: workDir,
		timeout: 10 * time.Minute, // Default 10 minute timeout
	}
}

// SetTimeout sets the command execution timeout
func (e *Executor) SetTimeout(timeout time.Duration) {
	e.timeout = timeout
}

// Execute runs a command in the working directory
func (e *Executor) Execute(command string) (string, error) {
	// Parse command into parts
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}

	// Create command with timeout
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = e.workDir

	// Set up timeout
	done := make(chan error, 1)
	var output []byte
	var err error

	go func() {
		output, err = cmd.CombinedOutput()
		done <- err
	}()

	select {
	case <-time.After(e.timeout):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return "", fmt.Errorf("command timed out after %v", e.timeout)
	case err := <-done:
		if err != nil {
			return string(output), fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
		}
		return string(output), nil
	}
}
