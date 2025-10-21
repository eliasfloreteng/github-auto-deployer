# Quick Start Guide

This guide will help you get GitHub Auto-Deployer up and running quickly.

## Step 1: Build the Application

```bash
make build
```

This creates the `github-auto-deployer` binary in the current directory.

## Step 2: Create a GitHub App

1. Go to https://github.com/settings/apps/new
2. Fill in:
   - **Name**: `My Auto-Deployer` (or any unique name)
   - **Homepage URL**: Your repository or server URL
   - **Webhook URL**: `http://your-server-ip:8080/webhook`
   - **Webhook secret**: Generate a random string (save it!)
3. Set permissions:
   - Repository → Contents: Read-only
   - Repository → Metadata: Read-only
4. Subscribe to events: **Push**
5. Click "Create GitHub App"
6. Note the **App ID**
7. Generate and download a **private key** (save the .pem file)
8. Click "Install App" and install it on your repositories
9. Note the **Installation ID** from the URL (e.g., `/settings/installations/12345678`)

## Step 3: Configure the Application

```bash
./github-auto-deployer init
```

Enter the following when prompted:

**GitHub App:**

- App ID: (from step 2)
- Installation ID: (from step 2)
- Private Key Path: `/path/to/downloaded-key.pem`
- Webhook Secret: (from step 2)

**SMTP (for Gmail):**

- Host: `smtp.gmail.com`
- Port: `587`
- Username: `your-email@gmail.com`
- Password: (use App Password - see below)
- From: `your-email@gmail.com`
- To: `your-email@gmail.com`

**Server:**

- Port: `8080` (or press Enter for default)

### Gmail App Password

1. Enable 2FA on your Google account
2. Go to: https://myaccount.google.com/apppasswords
3. Generate a new app password
4. Use this password (not your regular Gmail password)

## Step 4: Add a Repository to Watch

```bash
cd /path/to/your/local/repo
./github-auto-deployer add . --command "docker compose up -d --build"
```

Replace the command with whatever you want to run after pulling changes.

## Step 5: Start the Server

### For Testing (foreground):

```bash
./github-auto-deployer start
```

### For Production (systemd service):

```bash
# Install the binary system-wide
sudo make install

# Install the systemd service
make install-service
# Enter your username when prompted

# Enable and start the service
sudo systemctl enable github-auto-deployer@yourusername
sudo systemctl start github-auto-deployer@yourusername

# Check status
sudo systemctl status github-auto-deployer@yourusername

# View logs
sudo journalctl -u github-auto-deployer@yourusername -f
```

## Step 6: Test It!

1. Make a change to your repository
2. Commit and push to the monitored branch
3. Watch the logs to see the automatic deployment happen!

```bash
# If running in foreground, you'll see output directly
# If using systemd:
sudo journalctl -u github-auto-deployer@yourusername -f
```

## Troubleshooting

### Webhook not working?

1. Check if the server is running: `./github-auto-deployer status`
2. Verify firewall allows port 8080: `sudo ufw allow 8080`
3. Check GitHub webhook deliveries in your App settings
4. View logs for errors

### Email not sending?

1. Verify SMTP credentials
2. For Gmail, ensure you're using an App Password
3. Check if port 587 is allowed outbound

### Git pull failing?

1. Ensure no uncommitted changes in the repo
2. Verify git is configured for the user running the service
3. Check SSH keys or credentials for private repos

## Common Commands

```bash
# List watched repositories
./github-auto-deployer list

# Check configuration
./github-auto-deployer status

# Remove a repository
./github-auto-deployer remove /path/to/repo

# View service logs
sudo journalctl -u github-auto-deployer@yourusername -f

# Restart service
sudo systemctl restart github-auto-deployer@yourusername
```

## Next Steps

- Read the full [README.md](README.md) for detailed documentation
- Set up multiple repositories to watch
- Configure your firewall and reverse proxy (nginx/caddy) for production
- Set up HTTPS for the webhook endpoint
