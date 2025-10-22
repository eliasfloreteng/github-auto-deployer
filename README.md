# GitHub Auto Deployer

A lightweight, self-contained Go application that automatically deploys your code when you push to GitHub. It watches git repositories and automatically pulls changes and runs commands when pushes are detected via GitHub webhooks.

## Features

- 🚀 **Automatic Deployment**: Automatically pulls and deploys when you push to GitHub
- 🔒 **Secure**: Webhook signature verification, secure credential storage
- 📧 **Email Notifications**: Get notified when deployments fail
- 🔄 **Multi-Repository**: Watch multiple repositories and branches
- 🛠️ **Custom Commands**: Run any command after pulling (e.g., Docker Compose, build scripts)
- 💾 **Persistent**: Runs as a systemd service, survives reboots
- 📦 **Single Binary**: Easy to distribute and install
- 🎯 **Branch-Specific**: Only deploys when the watched branch is pushed to

## Quick Start

### 1. Build the Application

```bash
go build -o deployer cmd/deployer/main.go
```

### 2. Set Up GitHub App

Follow the detailed instructions in [docs/GITHUB_APP_SETUP.md](docs/GITHUB_APP_SETUP.md) to:

- Create a GitHub App
- Generate a private key
- Configure webhook settings
- Install the app on your repositories

### 3. Initialize Configuration

```bash
./deployer init
```

You'll be prompted to enter:

- GitHub App ID
- Private Key Path
- Webhook Secret
- SMTP settings (for failure notifications)
- Webhook server port (default: 8080)

### 4. Add Folders to Watch

```bash
./deployer add-folder
```

Enter:

- Repository path (must be a git repository)
- Command to execute after pulling (e.g., `docker compose up -d --pull=auto --build`)

The tool will automatically detect:

- Current branch
- Remote repository URL

### 5. Install as Service

```bash
./deployer install
```

This creates a systemd user service that:

- Starts automatically on boot (with lingering enabled)
- Restarts on failure
- Logs to journalctl
- Runs as your user (no sudo required)

### 6. Start the Service

```bash
systemctl --user start github-deployer
```

## Usage

### Available Commands

```bash
deployer init              # Initialize configuration
deployer install           # Install as systemd service
deployer uninstall         # Remove systemd service
deployer start             # Start webhook server (manual mode)
deployer add-folder        # Add a folder to watch
deployer list-folders      # List all watched folders
deployer remove-folder     # Remove a watched folder
deployer status            # Check service status
```

### Managing the Service

```bash
# Start the service
systemctl --user start github-deployer

# Stop the service
systemctl --user stop github-deployer

# Restart the service
systemctl --user restart github-deployer

# Check status
systemctl --user status github-deployer
# or
./deployer status

# View logs
journalctl --user -u github-deployer -f

# Enable lingering (service runs even when not logged in)
loginctl enable-linger $USER
```

### Configuration File

Configuration is stored in:

- `/etc/github-deployer/config.json` (system-wide)
- `~/.github-deployer/config.json` (user-specific)

Example configuration:

```json
{
  "github": {
    "app_id": 123456,
    "private_key_path": "/etc/github-deployer/private-key.pem",
    "webhook_secret": "your-webhook-secret"
  },
  "smtp": {
    "host": "smtp.gmail.com",
    "port": 587,
    "username": "your-email@gmail.com",
    "password": "your-app-password",
    "from": "deployer@example.com",
    "to": "admin@example.com"
  },
  "server": {
    "port": 8080
  },
  "folders": [
    {
      "path": "/var/www/myapp",
      "command": "docker compose up -d --pull=auto --build",
      "branch": "main",
      "repo_url": "https://github.com/username/myapp"
    }
  ]
}
```

## How It Works

1. **Webhook Reception**: GitHub sends a webhook to your server when you push
2. **Signature Verification**: The webhook signature is verified using HMAC-SHA256
3. **Repository Matching**: The pushed repository and branch are matched against watched folders
4. **Git Pull**: If matched, `git fetch && git pull` is executed
5. **Command Execution**: The configured command is run (e.g., Docker Compose)
6. **Notification**: If anything fails, an email notification is sent

## Architecture

```
github-auto-deployer/
├── cmd/
│   └── deployer/
│       └── main.go              # Entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── webhook/
│   │   └── handler.go           # Webhook handling
│   ├── git/
│   │   └── manager.go           # Git operations
│   ├── executor/
│   │   └── executor.go          # Command execution
│   ├── notifier/
│   │   └── email.go             # Email notifications
│   └── cli/
│       └── commands.go          # CLI commands
├── pkg/
│   └── systemd/
│       └── service.go           # Systemd service management
├── docs/
│   └── GITHUB_APP_SETUP.md      # GitHub App setup guide
├── go.mod
└── README.md
```

## Requirements

- Go 1.21 or later (for building)
- Git installed on the server
- systemd (for service installation)
- A domain with HTTPS (for webhooks)
- Reverse proxy (nginx, caddy, etc.)

## Security Considerations

1. **Webhook Secret**: Always use a strong random webhook secret
2. **Private Key**: Store with `chmod 600` permissions
3. **HTTPS**: Always use HTTPS for the webhook endpoint
4. **Firewall**: Only expose necessary ports
5. **User Permissions**: Run as a non-root user when possible
6. **Repository Access**: Only give the GitHub App access to necessary repositories

## Troubleshooting

### Webhook not received

- Verify your domain is accessible from the internet
- Check reverse proxy configuration
- Verify webhook URL in GitHub App settings
- Check firewall rules

### Git pull fails

- Ensure SSH keys or credentials are configured
- Check repository permissions
- Verify the user running the service has access

### Command execution fails

- Check command syntax
- Verify required tools are installed (e.g., docker)
- Check user permissions

### Email notifications not working

- Verify SMTP settings
- Check if your email provider requires app-specific passwords
- Test SMTP connection manually

## Example Use Cases

### Docker Compose Deployment

```bash
# Command to run after pull
docker compose up -d --pull=auto --build
```

### Node.js Application

```bash
# Command to run after pull
npm install && npm run build && pm2 restart app
```

### Static Website

```bash
# Command to run after pull
hugo --minify && rsync -av public/ /var/www/html/
```

### Python Application

```bash
# Command to run after pull
pip install -r requirements.txt && systemctl restart myapp
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - feel free to use this in your own projects.

## Support

For issues and questions:

- Check the [GitHub App Setup Guide](docs/GITHUB_APP_SETUP.md)
- Review the troubleshooting section above
- Open an issue on GitHub

## Acknowledgments

Built with:

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management
- [gomail](https://github.com/go-gomail/gomail) - Email sending
- [go-github](https://github.com/google/go-github) - GitHub API client
