# GitHub App Setup Guide

This guide will walk you through creating a GitHub App for the GitHub Auto Deployer.

## Prerequisites

- A GitHub account
- Admin access to the repositories you want to deploy
- A server with a public domain/IP address where the deployer will run

## Step 1: Create a GitHub App

1. Go to your GitHub account settings:

   - Click your profile picture → **Settings**
   - Scroll down to **Developer settings** (in the left sidebar)
   - Click **GitHub Apps** → **New GitHub App**

2. Fill in the basic information:

   - **GitHub App name**: Choose a unique name (e.g., "My Auto Deployer")
   - **Homepage URL**: Your organization's website or the repository URL
   - **Webhook URL**: `https://your-domain.com/webhook`
     - Replace `your-domain.com` with your actual domain
     - The deployer will listen on the port you configure (default: 8080)
     - Make sure your reverse proxy (nginx, caddy, etc.) forwards requests to this port
   - **Webhook secret**: Generate a strong random secret (save this for later)
     - You can generate one with: `openssl rand -hex 32`

3. Set permissions:

   - Under **Repository permissions**:
     - **Contents**: Read-only (required to receive push events)
     - **Metadata**: Read-only (automatically selected)

4. Subscribe to events:

   - Check **Push** (this is the only event we need)

5. Set **Where can this GitHub App be installed?**:

   - Select **Only on this account** (recommended for personal use)
   - Or **Any account** if you want to share it

6. Click **Create GitHub App**

## Step 2: Generate a Private Key

1. After creating the app, scroll down to **Private keys**
2. Click **Generate a private key**
3. A `.pem` file will be downloaded to your computer
4. Move this file to your server in a secure location (e.g., `/etc/github-deployer/private-key.pem`)
5. Set appropriate permissions:
   ```bash
   chmod 600 /etc/github-deployer/private-key.pem
   ```

## Step 3: Note Your App ID

1. On the GitHub App settings page, note the **App ID** (you'll need this during initialization)

## Step 4: Install the App

1. On the GitHub App settings page, click **Install App** in the left sidebar
2. Select the account where you want to install it
3. Choose which repositories to give access to:
   - **All repositories** (if you want to deploy all repos)
   - **Only select repositories** (recommended - choose specific repos)
4. Click **Install**

## Step 5: Configure the Deployer

Now you have all the information needed to configure the deployer:

- **App ID**: From Step 3
- **Private Key Path**: Where you saved the `.pem` file (Step 2)
- **Webhook Secret**: The secret you generated in Step 1

Run the initialization command:

```bash
./deployer init
```

You'll be prompted to enter:

- GitHub App ID
- Private Key Path
- Webhook Secret
- SMTP settings (for failure notifications)
- Webhook server port

## Step 6: Set Up Your Server

### Configure Reverse Proxy

If you're using nginx, add this to your configuration:

```nginx
server {
    listen 443 ssl;
    server_name your-domain.com;

    # SSL configuration
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location /webhook {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

For Caddy, it's even simpler:

```
your-domain.com {
    reverse_proxy /webhook localhost:8080
}
```

### Install as a Service

```bash
sudo ./deployer install
```

This will create a systemd service that starts automatically on boot.

### Start the Service

```bash
sudo systemctl start github-deployer
```

### Check Status

```bash
sudo systemctl status github-deployer
# or
./deployer status
```

### View Logs

```bash
sudo journalctl -u github-deployer -f
```

## Step 7: Add Folders to Watch

1. Clone your repository to the server:

   ```bash
   git clone https://github.com/username/repo.git /path/to/repo
   cd /path/to/repo
   git checkout main  # or your desired branch
   ```

2. Add the folder to the deployer:

   ```bash
   ./deployer add-folder
   ```

3. Follow the prompts:
   - Enter the repository path
   - Enter the command to run after pulling (e.g., `docker compose up -d --pull=auto --build`)

## Step 8: Test the Setup

1. Make a commit and push to your repository:

   ```bash
   echo "test" >> README.md
   git add README.md
   git commit -m "Test auto deploy"
   git push
   ```

2. Check the logs to see if the webhook was received:

   ```bash
   sudo journalctl -u github-deployer -f
   ```

3. You should see:
   - Webhook received
   - Git pull executed
   - Command executed (if configured)

## Troubleshooting

### Webhook not received

- Check that your domain is accessible from the internet
- Verify the webhook URL in GitHub App settings
- Check firewall rules on your server
- Verify the reverse proxy configuration

### Authentication errors

- Verify the App ID is correct
- Check that the private key file exists and has correct permissions
- Ensure the webhook secret matches

### Git pull fails

- Check that the repository has the correct remote URL
- Verify SSH keys or credentials are set up for the user running the service
- Check file permissions on the repository directory

### Command execution fails

- Verify the command syntax is correct
- Check that required tools (e.g., docker) are installed
- Ensure the user running the service has necessary permissions

## Security Considerations

1. **Private Key**: Keep your private key secure with `chmod 600`
2. **Webhook Secret**: Use a strong random secret
3. **HTTPS**: Always use HTTPS for the webhook endpoint
4. **Firewall**: Only expose necessary ports
5. **User Permissions**: Run the service as a non-root user when possible
6. **Repository Access**: Only give the app access to repositories it needs

## Multiple Servers

If you need to deploy to multiple servers:

1. Create a separate GitHub App for each server
2. Each app should have its own webhook URL pointing to the respective server
3. Install each app on the repositories you want to deploy to that server

This ensures webhooks are sent to the correct server for each repository.
