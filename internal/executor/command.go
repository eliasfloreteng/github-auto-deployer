package executor

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Executor handles command execution
type Executor struct {
	workDir string
}

// NewExecutor creates a new command executor
func NewExecutor(workDir string) *Executor {
	return &Executor{
		workDir: workDir,
	}
}

// Execute runs a command in the working directory
func (e *Executor) Execute(command string) error {
	// Split command into parts for exec
	// Handle shell commands properly
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = e.workDir

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return &CommandError{
			Command: command,
			Stdout:  stdout.String(),
			Stderr:  stderr.String(),
			Err:     err,
		}
	}

	return nil
}

// CommandError represents a command execution error
type CommandError struct {
	Command string
	Stdout  string
	Stderr  string
	Err     error
}

func (e *CommandError) Error() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("command failed: %s\n", e.Command))
	sb.WriteString(fmt.Sprintf("error: %v\n", e.Err))
	
	if e.Stdout != "" {
		sb.WriteString(fmt.Sprintf("stdout:\n%s\n", e.Stdout))
	}
	
	if e.Stderr != "" {
		sb.WriteString(fmt.Sprintf("stderr:\n%s\n", e.Stderr))
	}
	
	return sb.String()
}

// IsCommandError checks if an error is a command error
func IsCommandError(err error) bool {
	_, ok := err.(*CommandError)
	return ok
}
