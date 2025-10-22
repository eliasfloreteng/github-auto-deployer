package systemd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const userServiceTemplate = `[Unit]
Description=GitHub Auto Deployer
After=network.target

[Service]
Type=simple
WorkingDirectory=%s
ExecStart=%s start
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=default.target
`

// Install installs the systemd user service
func Install(execPath string) error {
	// Get absolute path of executable
	absExecPath, err := filepath.Abs(execPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Get working directory (directory containing the executable)
	workDir := filepath.Dir(absExecPath)

	// Generate service file content
	serviceContent := fmt.Sprintf(userServiceTemplate, workDir, absExecPath)

	// Get user systemd directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	systemdUserDir := filepath.Join(home, ".config", "systemd", "user")
	if err := os.MkdirAll(systemdUserDir, 0755); err != nil {
		return fmt.Errorf("failed to create systemd user directory: %w", err)
	}

	// Write service file
	servicePath := filepath.Join(systemdUserDir, "github-deployer.service")
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	// Reload systemd user daemon
	if err := exec.Command("systemctl", "--user", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	// Enable service
	if err := exec.Command("systemctl", "--user", "enable", "github-deployer.service").Run(); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}

	// Enable lingering so service runs even when user is not logged in
	if err := exec.Command("loginctl", "enable-linger").Run(); err != nil {
		fmt.Println("Warning: Failed to enable lingering. Service may not start on boot.")
		fmt.Println("You can manually enable it with: loginctl enable-linger $USER")
	}

	return nil
}

// Uninstall removes the systemd user service
func Uninstall() error {
	// Stop service if running
	exec.Command("systemctl", "--user", "stop", "github-deployer.service").Run()

	// Disable service
	exec.Command("systemctl", "--user", "disable", "github-deployer.service").Run()

	// Get service file path
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	servicePath := filepath.Join(home, ".config", "systemd", "user", "github-deployer.service")
	if err := os.Remove(servicePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove service file: %w", err)
	}

	// Reload systemd user daemon
	if err := exec.Command("systemctl", "--user", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	return nil
}

// Start starts the service
func Start() error {
	return exec.Command("systemctl", "--user", "start", "github-deployer.service").Run()
}

// Stop stops the service
func Stop() error {
	return exec.Command("systemctl", "--user", "stop", "github-deployer.service").Run()
}

// Status returns the service status
func Status() (string, error) {
	output, err := exec.Command("systemctl", "--user", "status", "github-deployer.service").CombinedOutput()
	return string(output), err
}
