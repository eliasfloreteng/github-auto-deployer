package systemd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const serviceTemplate = `[Unit]
Description=GitHub Auto Deployer
After=network.target

[Service]
Type=simple
User=%s
WorkingDirectory=%s
ExecStart=%s start
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
`

// Install installs the systemd service
func Install(execPath string) error {
	// Get current user
	user := os.Getenv("USER")
	if user == "" {
		return fmt.Errorf("cannot determine current user")
	}

	// Get absolute path of executable
	absExecPath, err := filepath.Abs(execPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Get working directory (directory containing the executable)
	workDir := filepath.Dir(absExecPath)

	// Generate service file content
	serviceContent := fmt.Sprintf(serviceTemplate, user, workDir, absExecPath)

	// Write service file
	servicePath := "/etc/systemd/system/github-deployer.service"
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to write service file (try running with sudo): %w", err)
	}

	// Reload systemd
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	// Enable service
	if err := exec.Command("systemctl", "enable", "github-deployer.service").Run(); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}

	return nil
}

// Uninstall removes the systemd service
func Uninstall() error {
	// Stop service if running
	exec.Command("systemctl", "stop", "github-deployer.service").Run()

	// Disable service
	exec.Command("systemctl", "disable", "github-deployer.service").Run()

	// Remove service file
	servicePath := "/etc/systemd/system/github-deployer.service"
	if err := os.Remove(servicePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove service file: %w", err)
	}

	// Reload systemd
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	return nil
}

// Start starts the service
func Start() error {
	return exec.Command("systemctl", "start", "github-deployer.service").Run()
}

// Stop stops the service
func Stop() error {
	return exec.Command("systemctl", "stop", "github-deployer.service").Run()
}

// Status returns the service status
func Status() (string, error) {
	output, err := exec.Command("systemctl", "status", "github-deployer.service").CombinedOutput()
	return string(output), err
}
