package git

import (
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

// NewManager creates a new git manager for a repository
func NewManager(repoPath string) *Manager {
	return &Manager{
		repoPath: repoPath,
	}
}

// GetCurrentBranch returns the currently checked out branch
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

// GetRemoteURL returns the remote URL of the repository
func (m *Manager) GetRemoteURL() (string, error) {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	cmd.Dir = m.repoPath

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get remote URL: %w", err)
	}

	url := strings.TrimSpace(string(output))
	return normalizeGitURL(url), nil
}

// Pull performs a git pull operation
func (m *Manager) Pull() error {
	// First, fetch to get latest changes
	fetchCmd := exec.Command("git", "fetch", "origin")
	fetchCmd.Dir = m.repoPath

	if output, err := fetchCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git fetch failed: %w\nOutput: %s", err, string(output))
	}

	// Then pull
	pullCmd := exec.Command("git", "pull", "origin")
	pullCmd.Dir = m.repoPath

	if output, err := pullCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git pull failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// IsGitRepository checks if the path is a git repository
func IsGitRepository(path string) bool {
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// normalizeGitURL converts various git URL formats to a consistent format
// for comparison (removes .git suffix, converts SSH to HTTPS format)
func normalizeGitURL(url string) string {
	// Remove .git suffix
	url = strings.TrimSuffix(url, ".git")

	// Convert SSH format (git@github.com:user/repo) to HTTPS-like format
	if strings.HasPrefix(url, "git@") {
		url = strings.Replace(url, ":", "/", 1)
		url = strings.Replace(url, "git@", "https://", 1)
	}

	// Ensure https:// prefix
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	return strings.ToLower(url)
}

// CompareURLs checks if two git URLs refer to the same repository
func CompareURLs(url1, url2 string) bool {
	return normalizeGitURL(url1) == normalizeGitURL(url2)
}
