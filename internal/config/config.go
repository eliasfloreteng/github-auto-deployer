package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	GitHub  GitHubConfig    `json:"github"`
	SMTP    SMTPConfig      `json:"smtp"`
	Server  ServerConfig    `json:"server"`
	Folders []WatchedFolder `json:"folders"`
}

// GitHubConfig holds GitHub App credentials
type GitHubConfig struct {
	AppID          int64  `json:"app_id"`
	PrivateKeyPath string `json:"private_key_path"`
	WebhookSecret  string `json:"webhook_secret"`
}

// SMTPConfig holds email notification settings
type SMTPConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
	To       string `json:"to"`
}

// ServerConfig holds webhook server settings
type ServerConfig struct {
	Port int `json:"port"`
}

// WatchedFolder represents a folder being monitored
type WatchedFolder struct {
	Path    string `json:"path"`
	Command string `json:"command"`
	Branch  string `json:"branch"`   // Current branch (detected automatically)
	RepoURL string `json:"repo_url"` // Repository URL for matching webhooks
}

var (
	configPath string
)

// GetConfigPath returns the configuration file path
func GetConfigPath() string {
	if configPath != "" {
		return configPath
	}

	// Try /etc first (for system-wide installation)
	systemPath := "/etc/github-deployer/config.json"
	if _, err := os.Stat(systemPath); err == nil {
		configPath = systemPath
		return configPath
	}

	// Fall back to user home directory
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("cannot determine home directory: %v", err))
	}

	configPath = filepath.Join(home, ".github-deployer", "config.json")
	return configPath
}

// SetConfigPath sets a custom configuration path
func SetConfigPath(path string) {
	configPath = path
}

// Load reads the configuration from disk
func Load() (*Config, error) {
	path := GetConfigPath()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("configuration file not found at %s. Run 'deployer init' first", path)
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

// Save writes the configuration to disk
func Save(cfg *Config) error {
	path := GetConfigPath()

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Exists checks if the configuration file exists
func Exists() bool {
	_, err := os.Stat(GetConfigPath())
	return err == nil
}
