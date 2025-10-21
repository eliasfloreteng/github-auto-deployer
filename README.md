# GitHub Auto-Deployer

A command-line application that automatically deploys GitHub repositories when push events are received via webhooks. Built with Go for easy distribution as a single binary.

## Features

- 🔄 **Automatic Deployment**: Automatically pulls changes when a push event is received
- 🔐 **GitHub App Integration**: Secure authentication using GitHub Apps
- 📧 **Email Notifications**: Get notified via email when deployments fail or conflicts occur
- ⚙️ **Post-Deploy Commands**: Execute custom commands after successful deployments (e.g., `docker compose up -d --build`)
- 🔒 **Webhook Security**: Validates webhook signatures to ensure authenticity
- 💾 **Persistent Configuration**: Survives server restarts with systemd integration
- 📦 **Single Binary**: Easy to distribute and deploy

## Prerequisites

- Go 1.21 or higher (for building)
- Git installed on the server
- A GitHub account with permissions to create GitHub Apps
- SMTP server credentials (Gmail, custom mail server, etc.)
- Linux server with systemd (for production deployment)

## Installation

### Option 1: Build from Source

```bash
# Clone the repository
git clone https://github.com/eliasfloreteng/github-auto-deployer.git
cd github-auto-deployer

# Build the binary
make build

# Install to system (requires sudo)
make install
```

### Option 2: Cross-compile for Linux Server

If you're building on macOS for a Linux server:

```bash
make build-linux
# This creates: github-auto-deployer-linux-amd64

# Copy to your server
scp github-auto-deployer-linux-amd64 user@server:/usr/local/bin/github-auto-deployer
```

## GitHub App Setup

Before using the auto-deployer, you need to create a GitHub App:

1. Go to GitHub Settings → Developer settings → GitHub Apps → New GitHub App

2. Configure the app:

   - **GitHub App name**: Choose a unique name (e.g., "My Auto-Deployer")
   - **Homepage URL**: Your server URL or repository URL
   - **Webhook URL**: `http://your-server:8080/webhook` (or your configured port)
   - **Webhook secret**: Generate a secure random string (save this!)

3. Set permissions:

   - Repository permissions:
     - **Contents**: Read-only
     - **Metadata**: Read-only
   - Subscribe to events:
     - **Push**

4. Create the app and note down:

   - **App ID**
   - **Installation ID** (after installing the app on your repositories)
   - Download the **private key** (PEM file)

5. Install the app on the repositories you want to watch

## Configuration

### Initial Setup

Run the initialization command to configure the application:

```bash
github-auto-deployer init
```

You'll be prompted for:

**GitHub App Configuration:**

- App ID
- Installation ID
- Path to private key file
- Webhook secret

**SMTP Configuration:**

- SMTP host (e.g., `smtp.gmail.com`)
- SMTP port (e.g., `587` for TLS)
- SMTP username
- SMTP password
- From email address
- To email address (for notifications)

**Server Configuration:**

- Webhook port (default: 8080)

Configuration is saved to `~/.github-auto-deployer/config.json`

### Gmail SMTP Setup

If using Gmail:

1. Enable 2-factor authentication on your Google account
2. Generate an App Password: Google Account → Security → 2-Step Verification → App passwords
3. Use these settings:
   - Host: `smtp.gmail.com`
   - Port: `587`
   - Username: Your Gmail address
   - Password: The generated App Password

## Usage

### Add a Repository to Watch

Navigate to your local repository and add it to the watch list:

```bash
cd /path/to/your/repo
github-auto-deployer add . --command "docker compose up -d --build"
```

The command will:

- Detect the current branch automatically
- Extract repository owner and name from the remote URL
- Save the configuration

### List Watched Repositories

```bash
github-auto-deployer list
```

### Remove a Repository

```bash
github-auto-deployer remove /path/to/repo
```

### Check Status

```bash
github-auto-deployer status
```

### Start the Webhook Server

```bash
github-auto-deployer start
```

The server will:

- Listen for webhook events on the configured port
- Automatically pull changes when push events are received
- Execute post-deploy commands
- Send email notifications on failures

## Systemd Service Setup

For production deployment, set up the systemd service:

```bash
# Install the service
make install-service

# Enable and start the service (replace USERNAME with your user)
sudo systemctl enable github-auto-deployer@USERNAME
sudo systemctl start github-auto-deployer@USERNAME

# Check status
sudo systemctl status github-auto-deployer@USERNAME

# View logs
sudo journalctl -u github-auto-deployer@USERNAME -f
```

The service will:

- Start automatically on boot
- Restart automatically if it crashes
- Run as the specified user
- Log to systemd journal

## How It Works

1. **Webhook Reception**: GitHub sends a push event to your server's webhook endpoint
2. **Signature Validation**: The webhook signature is validated using the webhook secret
3. **Repository Matching**: The event is matched against configured watchers
4. **Branch Verification**: Ensures the push is for the monitored branch
5. **Git Pull**: Executes `git fetch` and `git pull` in the repository directory
6. **Conflict Detection**: Detects merge conflicts and sends email notification
7. **Command Execution**: Runs the configured post-deploy command (if any)
8. **Error Handling**: Sends email notifications for any failures

## Email Notifications

You'll receive email notifications for:

- **Merge Conflicts**: When automatic pull fails due to conflicts
- **Pull Failures**: When git operations fail for other reasons
- **Command Failures**: When post-deploy commands fail

Each notification includes:

- Repository path and branch
- Timestamp
- Detailed error information

## Security Considerations

- **Webhook Signatures**: All webhooks are validated using HMAC-SHA256
- **Private Key Security**: Store your GitHub App private key securely (permissions 600)
- **SMTP Credentials**: Configuration file has restricted permissions (600)
- **Systemd Isolation**: Service runs with security restrictions (NoNewPrivileges, PrivateTmp)
- **User Isolation**: Service runs as a specific non-root user

## Troubleshooting

### Webhook Not Receiving Events

1. Check firewall rules allow incoming connections on the webhook port
2. Verify the webhook URL in GitHub App settings
3. Check systemd service status: `sudo systemctl status github-auto-deployer@USERNAME`
4. View logs: `sudo journalctl -u github-auto-deployer@USERNAME -f`

### Authentication Errors

1. Verify App ID and Installation ID are correct
2. Ensure private key file exists and is readable
3. Check that the GitHub App is installed on the repository

### Git Pull Failures

1. Ensure the repository has no uncommitted changes
2. Verify the user running the service has git configured
3. Check SSH keys or credentials for private repositories

### Email Not Sending

1. Verify SMTP credentials are correct
2. For Gmail, ensure you're using an App Password
3. Check firewall allows outbound connections on SMTP port

## Development

### Project Structure

```
github-auto-deployer/
├── cmd/
│   └── github-auto-deployer/
│       └── main.go              # CLI entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── github/
│   │   └── auth.go              # GitHub App authentication
│   ├── git/
│   │   └── manager.go           # Git operations
│   ├── email/
│   │   └── smtp.go              # Email notifications
│   ├── executor/
│   │   └── command.go           # Command execution
│   └── server/
│       └── server.go            # Webhook server
├── systemd/
│   └── github-auto-deployer.service
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### Building

```bash
# Build for current platform
make build

# Build for Linux
make build-linux

# Clean build artifacts
make clean

# Update dependencies
make deps
```

### Testing

```bash
make test
```

## Example Workflow

1. **Initial Setup**:

   ```bash
   github-auto-deployer init
   ```

2. **Add Repository**:

   ```bash
   cd /var/www/my-app
   github-auto-deployer add . --command "docker compose up -d --build"
   ```

3. **Install Service**:

   ```bash
   make install-service
   sudo systemctl enable github-auto-deployer@myuser
   sudo systemctl start github-auto-deployer@myuser
   ```

4. **Push to GitHub**:

   - Make changes to your code
   - Commit and push to the monitored branch
   - The server automatically pulls changes and runs your command

5. **Monitor**:
   ```bash
   sudo journalctl -u github-auto-deployer@myuser -f
   ```

## License

MIT License - feel free to use this in your projects!

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

For issues and questions, please open an issue on GitHub.
