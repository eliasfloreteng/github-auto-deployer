package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Manager handles git operations
type Manager struct {
	repoPath string
}

// NewManager creates a new git manager
func NewManager(repoPath string) (*Manager, error) {
	// Verify the path exists and is a git repository
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("not a git repository: %s", repoPath)
	}

	return &Manager{
		repoPath: repoPath,
	}, nil
}

// GetCurrentBranch returns the current branch name
func (m *Manager) GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = m.repoPath

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	branch := strings.TrimSpace(string(output))
	return branch, nil
}

// GetRemoteURL returns the remote URL for the repository
func (m *Manager) GetRemoteURL() (string, error) {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	cmd.Dir = m.repoPath

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get remote URL: %w", err)
	}

	url := strings.TrimSpace(string(output))
	return url, nil
}

// ParseRepoInfo extracts owner and repo name from a git URL
func ParseRepoInfo(url string) (owner, repo string, err error) {
	// Handle both HTTPS and SSH URLs
	// HTTPS: https://github.com/owner/repo.git
	// SSH: git@github.com:owner/repo.git

	url = strings.TrimSuffix(url, ".git")

	if strings.HasPrefix(url, "https://github.com/") {
		parts := strings.Split(strings.TrimPrefix(url, "https://github.com/"), "/")
		if len(parts) >= 2 {
			return parts[0], parts[1], nil
		}
	} else if strings.HasPrefix(url, "git@github.com:") {
		parts := strings.Split(strings.TrimPrefix(url, "git@github.com:"), "/")
		if len(parts) >= 2 {
			return parts[0], parts[1], nil
		}
	}

	return "", "", fmt.Errorf("unable to parse repository info from URL: %s", url)
}

// FetchAndPull fetches and pulls the latest changes
func (m *Manager) FetchAndPull(branch string) error {
	// Fetch latest changes
	fetchCmd := exec.Command("git", "fetch", "origin", branch)
	fetchCmd.Dir = m.repoPath

	var fetchErr bytes.Buffer
	fetchCmd.Stderr = &fetchErr

	if err := fetchCmd.Run(); err != nil {
		return fmt.Errorf("git fetch failed: %s", fetchErr.String())
	}

	// Pull changes
	pullCmd := exec.Command("git", "pull", "origin", branch)
	pullCmd.Dir = m.repoPath

	var pullOut bytes.Buffer
	var pullErr bytes.Buffer
	pullCmd.Stdout = &pullOut
	pullCmd.Stderr = &pullErr

	if err := pullCmd.Run(); err != nil {
		// Check if it's a merge conflict
		if strings.Contains(pullErr.String(), "CONFLICT") || 
		   strings.Contains(pullOut.String(), "CONFLICT") {
			return &ConflictError{
				Message: pullErr.String() + pullOut.String(),
			}
		}
		return fmt.Errorf("git pull failed: %s", pullErr.String())
	}

	return nil
}

// HasUncommittedChanges checks if there are uncommitted changes
func (m *Manager) HasUncommittedChanges() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = m.repoPath

	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check git status: %w", err)
	}

	return len(strings.TrimSpace(string(output))) > 0, nil
}

// ConflictError represents a merge conflict error
type ConflictError struct {
	Message string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("merge conflict: %s", e.Message)
}

// IsConflictError checks if an error is a conflict error
func IsConflictError(err error) bool {
	_, ok := err.(*ConflictError)
	return ok
}
