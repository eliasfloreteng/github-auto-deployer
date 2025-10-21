# GitHub Auto-Deployer

A Docker-based application that automatically deploys your repositories when changes are pushed to GitHub. It watches specified folders, pulls the latest changes when webhooks are received, and executes custom deployment commands.

## Features

- 🔄 **Automatic Deployment**: Automatically pulls and deploys when you push to GitHub
- 📁 **Multi-Repository Support**: Watch multiple repositories simultaneously
- 🔍 **Smart Detection**: Auto-detects repositories and branches
- 📧 **Email Notifications**: Get notified on failures (and optionally on success)
- 🐳 **Docker-Based**: Easy to deploy and manage with Docker Compose
- 🔒 **Secure**: GitHub webhook signature validation
- ⚡ **Efficient**: Uses filesystem events for real-time monitoring
- 🛡️ **Conflict Detection**: Prevents deployments when there are uncommitted changes

## Quick Start

### Prerequisites

- Docker and Docker Compose installed
- A GitHub account with repositories to deploy
- SMTP server credentials for email notifications

### 1. Clone the Repository

```bash
git clone <your-repo-url>
cd github-auto-deployer
```

### 2. Configure Environment Variables

Copy the example environment file and fill in your details:

```bash
cp .env.example .env
```

Edit `.env` with your configuration:

```env
# GitHub App Configuration
GITHUB_APP_ID=your_app_id
GITHUB_APP_PRIVATE_KEY_PATH=/app/github-app-key.pem
GITHUB_WEBHOOK_SECRET=your_webhook_secret

# SMTP Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your_email@gmail.com
SMTP_PASSWORD=your_app_password
SMTP_FROM_EMAIL=your_email@gmail.com

# Notification Configuration
DEFAULT_NOTIFICATION_EMAIL=admin@example.com
```

### 3. Configure Docker Compose

Edit `docker-compose.yml` and update the volumes section:

```yaml
volumes:
  # Mount your repositories directory
  - /path/to/your/repos:/repos:rw

  # Mount logs directory
  - ./logs:/app/logs

  # Mount your GitHub App private key file
  - /path/to/your/github-app-key.pem:/app/github-app-key.pem:ro
```

If your deployment commands use Docker, uncomment this line:

```yaml
- /var/run/docker.sock:/var/run/docker.sock
```

### 4. Configure Your Repositories

In each repository you want to deploy, create a `.deployer.yml` file:

```yaml
# Required: Command to execute after successful pull
command: "docker compose up -d --build"

# Optional: Branch to watch (auto-detected if not specified)
branch: "main"

# Optional: Email for notifications (uses default if not specified)
notification_email: "dev@example.com"

# Optional: Command timeout in seconds (default: 300)
command_timeout: 300

# Optional: Send email on successful deployment (default: false)
send_success_email: false
```

### 5. Start the Application

```bash
docker compose up -d
```

Check the logs:

```bash
docker compose logs -f
```

### 6. Set Up GitHub App

See [GitHub App Setup Guide](#github-app-setup) below for detailed instructions.

## GitHub App Setup

### Creating the GitHub App

1. Go to GitHub Settings → Developer settings → GitHub Apps → New GitHub App

2. Fill in the basic information:

   - **GitHub App name**: Choose a unique name (e.g., "My Auto Deployer")
   - **Homepage URL**: Your server URL or repository URL
   - **Webhook URL**: `https://your-domain.com/webhook`
   - **Webhook secret**: Generate a secure random string and save it

3. Set permissions:

   - **Repository permissions**:
     - Contents: Read-only (to read repository content)

4. Subscribe to events:

   - ✅ Push

5. Create the GitHub App

6. After creation:
   - Note down the **App ID**
   - Generate a **private key** and download the `.pem` file
   - Save the `.pem` file securely on your server

### Configuring the Private Key

1. Save the downloaded `.pem` file to a secure location on your server
2. Update `docker-compose.yml` to mount the private key file:
   ```yaml
   volumes:
     - /path/to/your/github-app-key.pem:/app/github-app-key.pem:ro
   ```
3. Set the path in your `.env` file:
   ```env
   GITHUB_APP_PRIVATE_KEY_PATH=/app/github-app-key.pem
   ```

### Installing the GitHub App

1. Go to your GitHub App settings
2. Click "Install App" in the left sidebar
3. Choose the account/organization where you want to install it
4. Select the repositories you want to give access to
5. Click "Install"

### Webhook Configuration

Your webhook endpoint will be: `https://your-domain.com/webhook`

Make sure:

- The endpoint is publicly accessible
- You're using HTTPS (required by GitHub)
- The webhook secret in your `.env` matches what you set in GitHub

## Repository Configuration

### .deployer.yml Reference

```yaml
# Required Fields
command: "docker compose up -d --build"

# Optional Fields
branch: "main" # Auto-detected if not specified
notification_email: "dev@example.com" # Uses DEFAULT_NOTIFICATION_EMAIL if not specified
command_timeout: 300 # Timeout in seconds (default: 300)
send_success_email: false # Send email on success (default: false)
```

### Example Configurations

**Simple Node.js Application:**

```yaml
command: "npm install && npm run build && pm2 restart app"
branch: "production"
```

**Docker Compose Application:**

```yaml
command: "docker compose pull && docker compose up -d --build"
command_timeout: 600
send_success_email: true
```

**Static Website:**

```yaml
command: "npm run build && rsync -av dist/ /var/www/html/"
```

## API Endpoints

The application exposes several endpoints:

- `GET /` - Service status and repository count
- `GET /health` - Health check endpoint
- `GET /repositories` - List all watched repositories
- `POST /webhook` - GitHub webhook endpoint (used by GitHub)

Example:

```bash
curl http://localhost:8080/repositories
```

## Email Notifications

The application sends email notifications in the following scenarios:

### Failure Notifications (Always Sent)

1. **Git Conflicts**: When there are uncommitted changes preventing a pull
2. **Pull Failures**: When git pull fails for any reason
3. **Command Failures**: When the deployment command fails

### Success Notifications (Optional)

Set `send_success_email: true` in `.deployer.yml` to receive notifications on successful deployments.

## Logs

Logs are stored in the `./logs` directory:

- `app.log` - General application logs
- Container logs: `docker compose logs -f`

## Troubleshooting

### Repositories Not Being Detected

1. Ensure the repository has a `.git` directory
2. Ensure the repository has a `.deployer.yml` file
3. Check the logs: `docker compose logs -f`
4. Verify the volume mapping in `docker-compose.yml`

### Webhook Not Triggering Deployments

1. Check GitHub webhook delivery status in your GitHub App settings
2. Verify the webhook secret matches in both GitHub and `.env`
3. Ensure the webhook URL is publicly accessible
4. Check the logs for webhook validation errors

### Git Pull Failures

1. Ensure the repository has no uncommitted changes
2. Check that the remote URL is accessible
3. Verify git credentials if using private repositories
4. Check the logs for detailed error messages

### Command Execution Failures

1. Verify the command syntax in `.deployer.yml`
2. Check if the command requires additional permissions
3. Increase `command_timeout` if the command takes longer
4. Check the logs for command output

### Docker Socket Permission Issues

If your deployment commands use Docker and you get permission errors:

1. Ensure the docker socket is mounted in `docker-compose.yml`
2. The container user needs access to the docker group

## Security Considerations

- **Webhook Secret**: Always use a strong, random webhook secret
- **SMTP Credentials**: Store securely and use app-specific passwords
- **Private Keys**: Never commit the `.env` file to version control
- **Command Validation**: The application validates commands for dangerous patterns
- **File Permissions**: Ensure proper permissions on mounted volumes

## Advanced Configuration

### Custom Port

Change the webhook port in `docker-compose.yml`:

```yaml
ports:
  - "9000:8080" # External:Internal
environment:
  - WEBHOOK_PORT=8080 # Keep internal port as 8080
```

### Multiple Repository Paths

You can mount multiple directories:

```yaml
volumes:
  - /path/to/repos1:/repos/group1:rw
  - /path/to/repos2:/repos/group2:rw
```

### Using with Reverse Proxy

Example Nginx configuration:

```nginx
server {
    listen 443 ssl;
    server_name deploy.example.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Development

### Running Locally

```bash
# Create virtual environment
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install dependencies
pip install -r requirements.txt

# Set environment variables
export GITHUB_APP_ID=your_app_id
export GITHUB_WEBHOOK_SECRET=your_secret
# ... other variables

# Run the application
python -m src.main
```

### Project Structure

```
github-auto-deployer/
├── src/
│   ├── __init__.py
│   ├── main.py              # Application entry point
│   ├── models.py            # Data models
│   ├── config_parser.py     # Configuration parser
│   ├── folder_monitor.py    # Filesystem monitoring
│   ├── git_manager.py       # Git operations
│   ├── command_executor.py  # Command execution
│   ├── email_notifier.py    # Email notifications
│   └── webhook_server.py    # FastAPI webhook server
├── Dockerfile
├── docker-compose.yml
├── requirements.txt
├── .env.example
├── .deployer.yml.example
└── README.md
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - feel free to use this project for any purpose.

## Support

For issues, questions, or contributions, please open an issue on GitHub.
