package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	GitHub   GitHubConfig     `json:"github"`
	SMTP     SMTPConfig       `json:"smtp"`
	Server   ServerConfig     `json:"server"`
	Watchers []WatcherConfig  `json:"watchers"`
}

// GitHubConfig holds GitHub App credentials
type GitHubConfig struct {
	AppID          int64  `json:"app_id"`
	InstallationID int64  `json:"installation_id"`
	PrivateKeyPath string `json:"private_key_path"`
	WebhookSecret  string `json:"webhook_secret"`
}

// SMTPConfig holds email configuration
type SMTPConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
	To       string `json:"to"`
}

// ServerConfig holds webhook server configuration
type ServerConfig struct {
	Port int    `json:"port"`
	Host string `json:"host"`
}

// WatcherConfig represents a watched repository
type WatcherConfig struct {
	Path          string `json:"path"`
	Branch        string `json:"branch"`
	Command       string `json:"command"`
	RepoOwner     string `json:"repo_owner"`
	RepoName      string `json:"repo_name"`
}

var configPath string

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("failed to get home directory: %v", err))
	}
	configPath = filepath.Join(home, ".github-auto-deployer", "config.json")
}

// GetConfigPath returns the configuration file path
func GetConfigPath() string {
	return configPath
}

// Load reads the configuration from disk
func Load() (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				Server: ServerConfig{
					Port: 8080,
					Host: "0.0.0.0",
				},
				Watchers: []WatcherConfig{},
			}, nil
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
func (c *Config) Save() error {
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// AddWatcher adds a new watcher to the configuration
func (c *Config) AddWatcher(watcher WatcherConfig) error {
	// Check if watcher already exists
	for _, w := range c.Watchers {
		if w.Path == watcher.Path {
			return fmt.Errorf("watcher for path %s already exists", watcher.Path)
		}
	}

	c.Watchers = append(c.Watchers, watcher)
	return c.Save()
}

// RemoveWatcher removes a watcher from the configuration
func (c *Config) RemoveWatcher(path string) error {
	found := false
	newWatchers := []WatcherConfig{}
	
	for _, w := range c.Watchers {
		if w.Path != path {
			newWatchers = append(newWatchers, w)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("watcher for path %s not found", path)
	}

	c.Watchers = newWatchers
	return c.Save()
}

// GetWatcher returns a watcher by path
func (c *Config) GetWatcher(path string) (*WatcherConfig, error) {
	for _, w := range c.Watchers {
		if w.Path == path {
			return &w, nil
		}
	}
	return nil, fmt.Errorf("watcher for path %s not found", path)
}

// GetWatcherByRepo returns a watcher by repository owner and name
func (c *Config) GetWatcherByRepo(owner, name string) (*WatcherConfig, error) {
	for _, w := range c.Watchers {
		if w.RepoOwner == owner && w.RepoName == name {
			return &w, nil
		}
	}
	return nil, fmt.Errorf("watcher for repository %s/%s not found", owner, name)
}
