package cli

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/eliasfloreteng/github-auto-deployer/internal/config"
	"github.com/eliasfloreteng/github-auto-deployer/internal/git"
	"github.com/eliasfloreteng/github-auto-deployer/internal/webhook"
	"github.com/eliasfloreteng/github-auto-deployer/pkg/systemd"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "deployer",
	Short: "GitHub Auto Deployer - Automatically deploy on push",
	Long:  `A tool that watches git repositories and automatically pulls changes and runs commands when pushes are detected via GitHub webhooks.`,
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration",
	Long:  `Interactive setup for GitHub App credentials and SMTP settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runInit(); err != nil {
			log.Fatalf("Initialization failed: %v", err)
		}
	},
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install as systemd service",
	Long:  `Install the deployer as a systemd service that starts automatically on boot.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runInstall(); err != nil {
			log.Fatalf("Installation failed: %v", err)
		}
	},
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall systemd service",
	Long:  `Remove the deployer systemd service.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runUninstall(); err != nil {
			log.Fatalf("Uninstallation failed: %v", err)
		}
	},
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the webhook server",
	Long:  `Start the webhook server to listen for GitHub push events.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runStart(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	},
}

var addFolderCmd = &cobra.Command{
	Use:   "add-folder",
	Short: "Add a folder to watch",
	Long:  `Interactively add a git repository folder to watch for changes.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runAddFolder(); err != nil {
			log.Fatalf("Failed to add folder: %v", err)
		}
	},
}

var listFoldersCmd = &cobra.Command{
	Use:   "list-folders",
	Short: "List watched folders",
	Long:  `Display all folders currently being watched.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runListFolders(); err != nil {
			log.Fatalf("Failed to list folders: %v", err)
		}
	},
}

var removeFolderCmd = &cobra.Command{
	Use:   "remove-folder",
	Short: "Remove a watched folder",
	Long:  `Remove a folder from the watch list.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runRemoveFolder(); err != nil {
			log.Fatalf("Failed to remove folder: %v", err)
		}
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check service status",
	Long:  `Check the status of the systemd service.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runStatus(); err != nil {
			log.Fatalf("Failed to get status: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(addFolderCmd)
	rootCmd.AddCommand(listFoldersCmd)
	rootCmd.AddCommand(removeFolderCmd)
	rootCmd.AddCommand(statusCmd)
}

// Execute runs the CLI
func Execute() error {
	return rootCmd.Execute()
}

func runInit() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("GitHub Auto Deployer - Configuration Setup")
	fmt.Println("==========================================")
	fmt.Println()

	// GitHub App Configuration
	fmt.Println("GitHub App Configuration:")
	fmt.Print("App ID: ")
	appIDStr, _ := reader.ReadString('\n')
	appID, err := strconv.ParseInt(strings.TrimSpace(appIDStr), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid App ID: %w", err)
	}

	fmt.Print("Private Key Path (absolute path): ")
	privateKeyPath, _ := reader.ReadString('\n')
	privateKeyPath = strings.TrimSpace(privateKeyPath)

	// Expand ~ to home directory
	if strings.HasPrefix(privateKeyPath, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		privateKeyPath = filepath.Join(home, privateKeyPath[1:])
	}

	// Verify private key exists
	if _, err := os.Stat(privateKeyPath); err != nil {
		return fmt.Errorf("private key file not found: %w", err)
	}

	fmt.Print("Webhook Secret: ")
	webhookSecret, _ := reader.ReadString('\n')
	webhookSecret = strings.TrimSpace(webhookSecret)

	fmt.Println()

	// SMTP Configuration
	fmt.Println("SMTP Configuration (for failure notifications):")
	fmt.Print("SMTP Host: ")
	smtpHost, _ := reader.ReadString('\n')
	smtpHost = strings.TrimSpace(smtpHost)

	fmt.Print("SMTP Port: ")
	smtpPortStr, _ := reader.ReadString('\n')
	smtpPort, err := strconv.Atoi(strings.TrimSpace(smtpPortStr))
	if err != nil {
		return fmt.Errorf("invalid SMTP port: %w", err)
	}

	fmt.Print("SMTP Username: ")
	smtpUsername, _ := reader.ReadString('\n')
	smtpUsername = strings.TrimSpace(smtpUsername)

	fmt.Print("SMTP Password: ")
	smtpPassword, _ := reader.ReadString('\n')
	smtpPassword = strings.TrimSpace(smtpPassword)

	fmt.Print("From Email: ")
	fromEmail, _ := reader.ReadString('\n')
	fromEmail = strings.TrimSpace(fromEmail)

	fmt.Print("To Email (for notifications): ")
	toEmail, _ := reader.ReadString('\n')
	toEmail = strings.TrimSpace(toEmail)

	fmt.Println()

	// Server Configuration
	fmt.Println("Server Configuration:")
	fmt.Print("Webhook Server Port (default 8080): ")
	portStr, _ := reader.ReadString('\n')
	portStr = strings.TrimSpace(portStr)
	port := 8080
	if portStr != "" {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			return fmt.Errorf("invalid port: %w", err)
		}
	}

	// Create configuration
	cfg := &config.Config{
		GitHub: config.GitHubConfig{
			AppID:          appID,
			PrivateKeyPath: privateKeyPath,
			WebhookSecret:  webhookSecret,
		},
		SMTP: config.SMTPConfig{
			Host:     smtpHost,
			Port:     smtpPort,
			Username: smtpUsername,
			Password: smtpPassword,
			From:     fromEmail,
			To:       toEmail,
		},
		Server: config.ServerConfig{
			Port: port,
		},
		Folders: []config.WatchedFolder{},
	}

	// Save configuration
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println()
	fmt.Printf("Configuration saved to: %s\n", config.GetConfigPath())
	fmt.Println("You can now add folders to watch using 'deployer add-folder'")

	return nil
}

func runInstall() error {
	// Check if config exists
	if !config.Exists() {
		return fmt.Errorf("configuration not found. Run 'deployer init' first")
	}

	// Get executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	fmt.Println("Installing systemd user service...")

	if err := systemd.Install(execPath); err != nil {
		return err
	}

	fmt.Println("Service installed successfully!")
	fmt.Println("To start the service: systemctl --user start github-deployer")
	fmt.Println("To view logs: journalctl --user -u github-deployer -f")
	fmt.Println("To enable on boot: loginctl enable-linger $USER")

	return nil
}

func runUninstall() error {
	fmt.Println("Uninstalling systemd service...")

	if err := systemd.Uninstall(); err != nil {
		return err
	}

	fmt.Println("Service uninstalled successfully!")
	return nil
}

func runStart() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Create webhook handler
	handler := webhook.NewHandler(cfg)

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Starting webhook server on %s", addr)
	log.Printf("Watching %d folder(s)", len(cfg.Folders))

	http.Handle("/webhook", handler)

	if err := http.ListenAndServe(addr, nil); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

func runAddFolder() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Add Folder to Watch")
	fmt.Println("===================")
	fmt.Println()

	fmt.Print("Repository Path (absolute path): ")
	repoPath, _ := reader.ReadString('\n')
	repoPath = strings.TrimSpace(repoPath)

	// Expand ~ to home directory
	if strings.HasPrefix(repoPath, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		repoPath = filepath.Join(home, repoPath[1:])
	}

	// Verify it's a git repository
	if !git.IsGitRepository(repoPath) {
		return fmt.Errorf("not a git repository: %s", repoPath)
	}

	// Get current branch and remote URL
	gitMgr := git.NewManager(repoPath)

	branch, err := gitMgr.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	repoURL, err := gitMgr.GetRemoteURL()
	if err != nil {
		return fmt.Errorf("failed to get remote URL: %w", err)
	}

	fmt.Printf("Detected branch: %s\n", branch)
	fmt.Printf("Detected repository: %s\n", repoURL)
	fmt.Println()

	fmt.Print("Command to execute after pull (e.g., 'docker compose up -d --pull=auto --build'): ")
	command, _ := reader.ReadString('\n')
	command = strings.TrimSpace(command)

	// Add folder to configuration
	folder := config.WatchedFolder{
		Path:    repoPath,
		Command: command,
		Branch:  branch,
		RepoURL: repoURL,
	}

	cfg.Folders = append(cfg.Folders, folder)

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println()
	fmt.Println("Folder added successfully!")
	fmt.Printf("Watching: %s (branch: %s)\n", repoPath, branch)

	return nil
}

func runListFolders() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if len(cfg.Folders) == 0 {
		fmt.Println("No folders are being watched.")
		fmt.Println("Add a folder using 'deployer add-folder'")
		return nil
	}

	fmt.Println("Watched Folders:")
	fmt.Println("================")
	fmt.Println()

	for i, folder := range cfg.Folders {
		fmt.Printf("%d. Path: %s\n", i+1, folder.Path)
		fmt.Printf("   Branch: %s\n", folder.Branch)
		fmt.Printf("   Repository: %s\n", folder.RepoURL)
		fmt.Printf("   Command: %s\n", folder.Command)
		fmt.Println()
	}

	return nil
}

func runRemoveFolder() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if len(cfg.Folders) == 0 {
		fmt.Println("No folders are being watched.")
		return nil
	}

	// List folders
	fmt.Println("Watched Folders:")
	for i, folder := range cfg.Folders {
		fmt.Printf("%d. %s (branch: %s)\n", i+1, folder.Path, folder.Branch)
	}
	fmt.Println()

	// Get selection
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter number to remove (or 0 to cancel): ")
	numStr, _ := reader.ReadString('\n')
	num, err := strconv.Atoi(strings.TrimSpace(numStr))
	if err != nil || num < 0 || num > len(cfg.Folders) {
		return fmt.Errorf("invalid selection")
	}

	if num == 0 {
		fmt.Println("Cancelled.")
		return nil
	}

	// Remove folder
	removedFolder := cfg.Folders[num-1]
	cfg.Folders = append(cfg.Folders[:num-1], cfg.Folders[num:]...)

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("Removed: %s\n", removedFolder.Path)
	return nil
}

func runStatus() error {
	status, err := systemd.Status()
	if err != nil {
		// Service might not be installed or not running
		fmt.Println("Service status: Not running or not installed")
		return nil
	}

	fmt.Println(status)
	return nil
}
