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

var (
	addPath      string
	addCommand   string
	addBranch    string
	addRepoURL   string
	addYes       bool
	addNoRestart bool
)

var addCmd = &cobra.Command{
	Use:   "add [path]",
	Short: "Add a folder to watch",
	Long: `Add a git repository folder to watch for changes.

All values can be supplied via flags for fully non-interactive use. When a
flag is omitted, the command falls back to an interactive prompt (unless
--yes is set, in which case detected/default values are used).

If the systemd service is running, it will be restarted automatically after
the folder is added. Pass --no-restart to skip the restart.

Examples:
  # Fully non-interactive
  deployer add --path /var/www/app --command "docker compose up -d" --yes

  # Use current directory and auto-detected defaults, no prompts
  deployer add --yes

  # Add without restarting the service
  deployer add --path /var/www/app --yes --no-restart

  # Mix flags and prompts
  deployer add --path /var/www/app`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := addPath
		if path == "" && len(args) > 0 {
			path = args[0]
		}

		opts := addOptions{
			path:       path,
			command:    addCommand,
			branch:     addBranch,
			repoURL:    addRepoURL,
			yes:        addYes,
			noRestart:  addNoRestart,
			commandSet: cmd.Flags().Changed("command"),
			branchSet:  cmd.Flags().Changed("branch"),
			repoURLSet: cmd.Flags().Changed("repo-url"),
		}

		if err := runAddFolder(opts); err != nil {
			log.Fatalf("Failed to add folder: %v", err)
		}
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List watched folders",
	Long:  `Display all folders currently being watched.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runListFolders(); err != nil {
			log.Fatalf("Failed to list folders: %v", err)
		}
	},
}

var removeCmd = &cobra.Command{
	Use:   "remove",
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
	// Flags for the `add` command. These allow fully non-interactive use.
	addCmd.Flags().StringVarP(&addPath, "path", "p", "", "Path to the git repository (defaults to the positional argument or current directory)")
	addCmd.Flags().StringVarP(&addCommand, "command", "c", "", "Command to execute after pulling (e.g. 'docker compose up -d --pull=auto --build')")
	addCmd.Flags().StringVarP(&addBranch, "branch", "b", "", "Branch to watch (defaults to the repository's current branch)")
	addCmd.Flags().StringVarP(&addRepoURL, "repo-url", "u", "", "Repository URL used to match webhooks (defaults to the 'origin' remote)")
	addCmd.Flags().BoolVarP(&addYes, "yes", "y", false, "Non-interactive mode: skip all prompts and use flag values or detected defaults")
	addCmd.Flags().BoolVar(&addNoRestart, "no-restart", false, "Do not restart the systemd service after adding the folder (restart is the default)")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(removeCmd)
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
	fmt.Println("You can now add folders to watch using 'deployer add'")

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

// addOptions holds parsed flag values for the `add` command.
type addOptions struct {
	path       string
	command    string
	branch     string
	repoURL    string
	yes        bool
	noRestart  bool
	commandSet bool
	branchSet  bool
	repoURLSet bool
}

func runAddFolder(opts addOptions) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	reader := bufio.NewReader(os.Stdin)
	var repoPath string

	// Resolve repository path
	if opts.path != "" {
		repoPath = opts.path
	} else if opts.yes {
		// Default to current directory in non-interactive mode
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		repoPath = cwd
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		fmt.Println("Add Folder to Watch")
		fmt.Println("===================")
		fmt.Println()
		fmt.Printf("Repository Path (default: current directory): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			repoPath = cwd
		} else {
			repoPath = input
		}
	}

	// Expand ~ to home directory
	if strings.HasPrefix(repoPath, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		repoPath = filepath.Join(home, repoPath[1:])
	}

	// Convert to absolute path
	if !filepath.IsAbs(repoPath) {
		absPath, err := filepath.Abs(repoPath)
		if err != nil {
			return fmt.Errorf("failed to convert to absolute path: %w", err)
		}
		repoPath = absPath
	}

	// Verify it's a git repository
	if !git.IsGitRepository(repoPath) {
		return fmt.Errorf("not a git repository: %s", repoPath)
	}

	// Get current branch and remote URL (used as defaults when flags aren't set)
	gitMgr := git.NewManager(repoPath)

	detectedBranch, err := gitMgr.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	detectedRepoURL, err := gitMgr.GetRemoteURL()
	if err != nil {
		return fmt.Errorf("failed to get remote URL: %w", err)
	}

	branch := detectedBranch
	if opts.branchSet && opts.branch != "" {
		branch = opts.branch
	}

	repoURL := detectedRepoURL
	if opts.repoURLSet && opts.repoURL != "" {
		repoURL = opts.repoURL
	}

	fmt.Printf("Detected branch: %s\n", detectedBranch)
	fmt.Printf("Detected repository: %s\n", detectedRepoURL)
	if branch != detectedBranch {
		fmt.Printf("Using branch override: %s\n", branch)
	}
	if repoURL != detectedRepoURL {
		fmt.Printf("Using repo URL override: %s\n", repoURL)
	}
	fmt.Println()

	// Suggest default command based on what's in the repository
	defaultCmd := suggestDefaultCommand(repoPath)

	var command string
	switch {
	case opts.commandSet:
		command = opts.command
	case opts.yes:
		command = defaultCmd
		if command != "" {
			fmt.Printf("Using default command: %s\n", command)
		}
	default:
		if defaultCmd != "" {
			fmt.Printf("Command to execute after pull (default: %s): ", defaultCmd)
		} else {
			fmt.Print("Command to execute after pull (e.g., 'docker compose up -d --pull=auto --build'): ")
		}

		input, _ := reader.ReadString('\n')
		command = strings.TrimSpace(input)

		if command == "" && defaultCmd != "" {
			command = defaultCmd
			fmt.Printf("Using default command: %s\n", command)
		}
	}

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

	// Restart the service by default if it is running. Pass --no-restart to skip.
	if isServiceRunning() {
		fmt.Println()
		if opts.noRestart {
			fmt.Println("Service is running. Skipping restart (--no-restart).")
			fmt.Println("Remember to restart: systemctl --user restart github-deployer")
		} else {
			fmt.Println("Restarting service...")
			if err := systemd.Stop(); err != nil {
				fmt.Printf("Warning: Failed to stop service: %v\n", err)
			}
			if err := systemd.Start(); err != nil {
				return fmt.Errorf("failed to start service: %w", err)
			}
			fmt.Println("Service restarted successfully!")
		}
	}

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
		fmt.Println("Add a folder using 'deployer add'")
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

	// Check if service is running and offer to restart
	if isServiceRunning() {
		fmt.Println()
		fmt.Print("Service is running. Restart to apply changes? (y/n): ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response == "y" || response == "yes" {
			fmt.Println("Restarting service...")
			if err := systemd.Stop(); err != nil {
				fmt.Printf("Warning: Failed to stop service: %v\n", err)
			}
			if err := systemd.Start(); err != nil {
				return fmt.Errorf("failed to start service: %w", err)
			}
			fmt.Println("Service restarted successfully!")
		} else {
			fmt.Println("Remember to restart the service: systemctl --user restart github-deployer")
		}
	}

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

// isServiceRunning checks if the systemd service is currently running
func isServiceRunning() bool {
	status, err := systemd.Status()
	if err != nil {
		return false
	}
	// Check if status contains "active (running)"
	return strings.Contains(status, "active (running)")
}

// suggestDefaultCommand suggests a default command based on files in the repository
func suggestDefaultCommand(repoPath string) string {
	// Check for docker-compose.yml or compose.yml
	if fileExists(filepath.Join(repoPath, "docker-compose.yml")) ||
		fileExists(filepath.Join(repoPath, "compose.yml")) ||
		fileExists(filepath.Join(repoPath, "docker-compose.yaml")) ||
		fileExists(filepath.Join(repoPath, "compose.yaml")) {
		return "docker compose up -d --pull=auto --build"
	}

	// Check for Dockerfile
	if fileExists(filepath.Join(repoPath, "Dockerfile")) {
		return "docker build -t app . && docker run -d app"
	}

	// Check for package.json (Node.js)
	if fileExists(filepath.Join(repoPath, "package.json")) {
		return "npm install && npm run build"
	}

	// Check for Makefile
	if fileExists(filepath.Join(repoPath, "Makefile")) {
		return "make deploy"
	}

	// Check for requirements.txt (Python)
	if fileExists(filepath.Join(repoPath, "requirements.txt")) {
		return "pip install -r requirements.txt"
	}

	// No suggestion
	return ""
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
